package service_test

import (
	"github.com/ATenderholt/rainbow/internal/domain"
	"github.com/ATenderholt/rainbow/internal/service"
	"github.com/go-rel/reltest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqsParseAction(t *testing.T) {
	s := service.NewSqsService(nil)
	action := s.ParseAction("Action=CreateQueue&Version=2012-11-05&QueueName=test5")

	assert.Equal(t, "CreateQueue", action)
}

func TestSqsSaveWithoutAttributesOrTags(t *testing.T) {
	payload := "Action=CreateQueue&Version=2012-11-05&QueueName=test6"

	repo := reltest.New()
	repo.ExpectInsert().For(&domain.SqsQueue{
		Name: "test6",
	})

	s := service.NewSqsService(repo)
	err := s.SaveAttributes(payload)
	if err != nil {
		t.Fatalf("Unable to save: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestSqsSaveWithoutTags(t *testing.T) {
	payload := "Action=CreateQueue&Version=2012-11-05&QueueName=test6&Attribute.1.Name=Key1&Attribute.1.Value=Value1&Attribute.2.Name=Key2&Attribute.2.Value=Value2"

	repo := reltest.New()
	repo.ExpectInsert().For(&domain.SqsQueue{
		Name: "test6",
		Attributes: []domain.SqsQueueAttribute{
			{
				Name:  "Key1",
				Value: "Value1",
			},
			{
				Name:  "Key2",
				Value: "Value2",
			},
		},
	})

	s := service.NewSqsService(repo)
	err := s.SaveAttributes(payload)
	if err != nil {
		t.Fatalf("Unable to save: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestSqsSaveWithoutAttributes(t *testing.T) {
	payload := "Action=CreateQueue&Version=2012-11-05&QueueName=test6&Tag.1.Key=Key1&Tag.1.Value=Value1&Tag.2.Key=Key2&Tag.2.Value=Value2"

	repo := reltest.New()
	repo.ExpectInsert().For(&domain.SqsQueue{
		Name: "test6",
		Tags: []domain.SqsQueueTag{
			{
				Name:  "Key1",
				Value: "Value1",
			},
			{
				Name:  "Key2",
				Value: "Value2",
			},
		},
	})

	s := service.NewSqsService(repo)
	err := s.SaveAttributes(payload)
	if err != nil {
		t.Fatalf("Unable to save: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestSqsSave(t *testing.T) {
	payload := "Action=CreateQueue&Version=2012-11-05&QueueName=test6&Attribute.1.Name=Key1&Attribute.1.Value=Value1&Attribute.2.Name=Key2&Attribute.2.Value=Value2&Tag.1.Key=Key1&Tag.1.Value=Value1&Tag.2.Key=Key2&Tag.2.Value=Value2"

	repo := reltest.New()
	repo.ExpectInsert().For(&domain.SqsQueue{
		Name: "test6",
		Tags: []domain.SqsQueueTag{
			{
				Name:  "Key1",
				Value: "Value1",
			},
			{
				Name:  "Key2",
				Value: "Value2",
			},
		},
		Attributes: []domain.SqsQueueAttribute{
			{
				Name:  "Key1",
				Value: "Value1",
			},
			{
				Name:  "Key2",
				Value: "Value2",
			},
		},
	})

	s := service.NewSqsService(repo)
	err := s.SaveAttributes(payload)
	if err != nil {
		t.Fatalf("Unable to save: %v", err)
	}

	repo.AssertExpectations(t)
}
