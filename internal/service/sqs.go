package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/go-rel/rel"
	"github.com/go-rel/rel/where"
	"regexp"
	"time"
)

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
	repo rel.Repository
}

func NewSqsService(repo rel.Repository) SqsService {
	return SqsService{
		repo: repo,
	}
}

func (s SqsService) ParseAction(payload string) string {
	logger.Infof("payload: %s", payload)
	action := actionRegex.FindStringSubmatch(payload)
	return action[1]
}

func (s SqsService) SaveAttributes(payload string) error {
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

	err := s.repo.Insert(ctx, &queue)

	if err != nil {
		e := fmt.Errorf("error inserting attributes/tags for queue %s: %v", name, err)
		logger.Error(e)
		return e
	}

	return nil
}

func (s SqsService) DecorateAttributes(payload string, response []byte) ([]byte, error) {
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
	err := s.repo.Find(ctx, &queue, where.Eq("name", name[1]))
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
