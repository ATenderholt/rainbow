package service

import (
	"github.com/go-rel/rel"
	"regexp"
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
	return nil
}

func (s SqsService) DecorateAttributes(payload string, response []byte) ([]byte, error) {
	return []byte(""), nil
}
