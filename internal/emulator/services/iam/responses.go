package iam

import (
	"fmt"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *IAMService) successResponse(action string, data interface{}) (*emulator.AWSResponse, error) {
	return emulator.BuildQueryResponse(action, data, emulator.ResponseBuilderConfig{
		ServiceName: "iam",
		Version:     "2010-05-08",
	})
}

func (s *IAMService) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	return emulator.BuildErrorResponse("iam", statusCode, code, message)
}

func (s *IAMService) parseTags(params map[string]interface{}) []XMLTag {
	var tags []XMLTag
	tagIndex := 1

	for {
		var keyParam, valueParam string
		var key, value string
		var hasKey, hasValue bool

		// Try Tags.member.N format (AWS SDK)
		keyParam = fmt.Sprintf("Tags.member.%d.Key", tagIndex)
		valueParam = fmt.Sprintf("Tags.member.%d.Value", tagIndex)
		key, hasKey = params[keyParam].(string)
		value, hasValue = params[valueParam].(string)

		// Try Tags.Tag.N format (Terraform)
		if !hasKey || !hasValue {
			keyParam = fmt.Sprintf("Tags.Tag.%d.Key", tagIndex)
			valueParam = fmt.Sprintf("Tags.Tag.%d.Value", tagIndex)
			key, hasKey = params[keyParam].(string)
			value, hasValue = params[valueParam].(string)
		}

		if !hasKey || !hasValue {
			break
		}

		tags = append(tags, XMLTag{
			Key:   key,
			Value: value,
		})

		tagIndex++
	}

	return tags
}
