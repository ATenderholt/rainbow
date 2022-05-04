package service

import (
	"context"
	"fmt"
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/go-rel/rel"
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
	return []byte(""), nil
}
