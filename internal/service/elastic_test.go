package service_test

import (
	"github.com/ATenderholt/rainbow/internal/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestElasticParseAction(t *testing.T) {
	s := service.NewElasticService(nil)
	action := s.ParseAction("Action=CreateQueue&Version=2012-11-05&QueueName=test5")

	assert.Equal(t, "CreateQueue", action)
}
