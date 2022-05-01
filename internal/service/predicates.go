package service

import (
	"github.com/ATenderholt/rainbow/internal/domain"
	"strings"
)

type persistPredicate func(request domain.MotoRequest) bool

func persistIamRequest(request domain.MotoRequest) bool {
	excludes := []string{
		"GetRole",
		"ListAttachedRolePolicies",
		"ListRolePolicies",
		"ListRoles",
	}

	for _, exclude := range excludes {
		if strings.Contains(request.Payload, exclude) {
			return false
		}
	}

	return true
}

func persistSsmRequest(request domain.MotoRequest) bool {
	excludes := []string{
		"AmazonSSM.GetParameter",
		"AmazonSSM.DescribeParameters",
		"AmazonSSM.ListTagsForResource",
	}

	for _, exclude := range excludes {
		if request.Target == exclude {
			return false
		}
	}

	return true
}
