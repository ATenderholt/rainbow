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

type ElasticService struct {
	repo rel.Repository
}

func NewElasticService(repo rel.Repository) ElasticService {
	return ElasticService{
		repo: repo,
	}
}

func (s ElasticService) ParseAction(payload string) string {
	logger.Infof("payload: %s", payload)
	action := actionRegex.FindStringSubmatch(payload)
	return action[1]
}

func (s ElasticService) SaveAttributes(payload string) error {
	return nil
}

func (s ElasticService) DecorateAttributes(payload string, response []byte) ([]byte, error) {
	return []byte(""), nil
}
