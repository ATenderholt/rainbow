package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/ATenderholt/rainbow/settings"
	"github.com/docker/docker/api/types/mount"
	"github.com/go-rel/rel"
	"github.com/go-rel/rel/where"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const configFileTemplate = `include classpath("application.conf")

messages-storage {
  enabled = true
}

aws {
  region = %s
  accountId = %s
}
`

var actionRegex *regexp.Regexp
var queueNameRegex *regexp.Regexp
var queueUrlRegex *regexp.Regexp
var createQueueAttributeRegex *regexp.Regexp
var createQueueTagRegex *regexp.Regexp

func init() {
	var err error

	actionRegex, err = regexp.Compile(`Action=([^&]*)`)
	if err != nil {
		panic("unable to compile action regex")
	}

	queueNameRegex, err = regexp.Compile(`QueueName=([^&]*)`)
	if err != nil {
		panic("unable to compile queue name regex")
	}

	// name of queue is everything after trailing slash
	queueUrlRegex, err = regexp.Compile(`QueueUrl=.*%2F([^&]+)`)
	if err != nil {
		panic("unable to compile queue url regex")
	}

	createQueueAttributeRegex, err = regexp.Compile(`Attribute.\d+.Name=([^&]*)&Attribute.\d+.Value=([^&]*)`)
	if err != nil {
		panic("unable to compile create queue attribute regex")
	}

	createQueueTagRegex, err = regexp.Compile(`Tag.\d+.Key=([^&]*)&Tag.\d+.Value=([^&]*)`)
	if err != nil {
		panic("unable to compile create queue tag regex")
	}
}

type SqsService struct {
	cfg    *settings.Config
	docker *dockerlib.DockerController
	repo   rel.Repository
}

func NewSqsService(cfg *settings.Config, docker *dockerlib.DockerController, repo rel.Repository) SqsService {
	return SqsService{
		cfg:    cfg,
		docker: docker,
		repo:   repo,
	}
}

func (sqs SqsService) Start(ctx context.Context) error {
	dataPath := sqs.cfg.DataPath()
	basePath := filepath.Join(dataPath, "sqs")
	err := os.MkdirAll(basePath, 0775)
	if err != nil {
		e := fmt.Errorf("unable to create path for sqs: %v", err)
		logger.Error(e)
		return e
	}

	configPath := filepath.Join(dataPath, "sqs.conf")

	err = writeConfigFile(sqs.cfg, configPath)
	if err != nil {
		return err
	}

	image := sqs.cfg.Sqs.Image
	err = sqs.docker.EnsureImage(ctx, image)
	if err != nil {
		e := fmt.Errorf("unable to ensure image for sqs: %v", err)
		logger.Error(e)
		return e
	}

	container := dockerlib.Container{
		Name:  "sqs",
		Image: image,
		Mounts: []mount.Mount{
			{
				Source: basePath,
				Target: "/data",
				Type:   mount.TypeBind,
			},
			{
				Source: configPath,
				Target: "/opt/elasticmq.conf",
				Type:   mount.TypeBind,
			},
		},
		Ports: map[int]int{
			9324: sqs.cfg.Sqs.Port,
			9325: sqs.cfg.Sqs.Port + 1,
		},
		Network: []string{sqs.cfg.Network},
	}

	ready, err := sqs.docker.Start(ctx, &container, "started in")
	if err != nil {
		e := fmt.Errorf("unable to start sqs container: %v", err)
		logger.Error(e)
		return e
	}

	<-ready

	logger.Info("SQS is ready")

	return nil
}

func (sqs SqsService) ParseAction(payload string) string {
	logger.Infof("payload: %s", payload)
	action := actionRegex.FindStringSubmatch(payload)
	return action[1]
}

func (sqs SqsService) SaveAttributes(payload string) error {
	logger.Infof("Saving attributes for payload: %s", payload)

	name := queueNameRegex.FindStringSubmatch(payload)
	if name == nil {
		err := fmt.Errorf("unable to find queue name in %s", payload)
		logger.Error(err)
		return err
	}

	attributes := createQueueAttributeRegex.FindAllStringSubmatch(payload, -1)
	if attributes == nil {
		logger.Warnf("unable to find attributes in %s", payload)
	}

	tags := createQueueTagRegex.FindAllStringSubmatch(payload, -1)
	if tags == nil {
		logger.Warnf("unable to find tags in %s", payload)
	}

	queue := domain.NewSqsQueue(name[1])
	for _, groups := range attributes {
		queue.AddAttribute(groups[1], groups[2])
	}
	for _, groups := range tags {
		queue.AddTag(groups[1], groups[2])
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := sqs.repo.Insert(ctx, &queue)

	if err != nil {
		e := fmt.Errorf("error inserting attributes/tags for queue %s: %v", name, err)
		logger.Error(e)
		return e
	}

	return nil
}

func (sqs SqsService) DecorateAttributes(payload string, response []byte) ([]byte, error) {
	logger.Infof("Decorating attributes for payload: %s", payload)
	logger.Infof("Current response: %s", string(response))

	name := queueUrlRegex.FindStringSubmatch(payload)
	if name == nil {
		err := fmt.Errorf("unable to find queue name in %s", payload)
		logger.Error(err)
		return response, err
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var queue domain.SqsQueue
	err := sqs.repo.Find(ctx, &queue, where.Eq("name", name[1]))
	switch {
	case err == rel.NotFoundError{}:
		logger.Warnf("unable to find queue named %s", name[1])
		return response, nil
	case err != nil:
		e := fmt.Errorf("unable to find queue named %s: %v", name[1], err)
		logger.Error(e)
		return response, e
	}

	var output domain.GetQueueAttributesResponse
	err = xml.Unmarshal(response, &output)
	if err != nil {
		e := fmt.Errorf("unable to unmarshal %s: %v", string(response), err)
		logger.Error(e)
		return response, e
	}

	for _, attr := range queue.Attributes {
		output.AddAttributeIfNotExists(attr.Name, attr.Value)
	}

	bytes, err := xml.Marshal(output)
	if err != nil {
		e := fmt.Errorf("unable to marshal %+v: %v", output, err)
		logger.Error(e)
		return response, e
	}

	return bytes, nil
}

func writeConfigFile(cfg *settings.Config, configPath string) error {
	stat, err := os.Stat(configPath)
	if err == nil && stat.IsDir() {
		e := fmt.Errorf("expecting %s to be a file, but is a directory", configPath)
		logger.Error(e)
		return e
	}

	f, err := os.Create(configPath)
	if err != nil {
		e := fmt.Errorf("unable to create %s: %v", configPath, err)
		logger.Error(e)
		return e
	}

	contents := fmt.Sprintf(configFileTemplate, cfg.Region, cfg.AccountNumber)
	_, err = f.WriteString(contents)
	if err != nil {
		e := fmt.Errorf("unable to write sqs config file to %s: %v", configPath, err)
		logger.Error(e)
		return e
	}

	return nil
}
