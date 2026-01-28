package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

func (s *EC2Service) createLaunchTemplate(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	templateName, ok := params["LaunchTemplateName"].(string)
	if !ok || templateName == "" {
		return s.errorResponse(400, "MissingParameter", "LaunchTemplateName is required"), nil
	}

	templateId := fmt.Sprintf("lt-%s", uuid.New().String()[:17])
	versionNumber := int64(1)

	template := LaunchTemplate{
		LaunchTemplateId:     &templateId,
		LaunchTemplateName:   &templateName,
		CreateTime:           helpers.TimePtr(time.Now()),
		CreatedBy:            helpers.StringPtr("arn:aws:iam::123456789012:root"),
		DefaultVersionNumber: &versionNumber,
		LatestVersionNumber:  &versionNumber,
	}

	if err := s.state.Set(fmt.Sprintf("ec2:launch-templates:%s", templateId), &template); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store launch template"), nil
	}

	return s.createLaunchTemplateResponse(template)
}

func (s *EC2Service) describeLaunchTemplates(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	templateIds := s.parseLaunchTemplateIds(params)

	var templates []LaunchTemplate

	if len(templateIds) > 0 {
		for _, templateId := range templateIds {
			var template LaunchTemplate
			if err := s.state.Get(fmt.Sprintf("ec2:launch-templates:%s", templateId), &template); err != nil {
				return s.errorResponse(400, "InvalidLaunchTemplateId.NotFound", fmt.Sprintf("The launch template ID '%s' does not exist", templateId)), nil
			}
			templates = append(templates, template)
		}
	} else {
		keys, err := s.state.List("ec2:launch-templates:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list launch templates"), nil
		}

		for _, key := range keys {
			var template LaunchTemplate
			if err := s.state.Get(key, &template); err == nil {
				templates = append(templates, template)
			}
		}
	}

	return s.describeLaunchTemplatesResponse(templates)
}

func (s *EC2Service) deleteLaunchTemplate(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	templateId, ok := params["LaunchTemplateId"].(string)
	if !ok || templateId == "" {
		return s.errorResponse(400, "MissingParameter", "LaunchTemplateId is required"), nil
	}

	var template LaunchTemplate
	if err := s.state.Get(fmt.Sprintf("ec2:launch-templates:%s", templateId), &template); err != nil {
		return s.errorResponse(400, "InvalidLaunchTemplateId.NotFound", fmt.Sprintf("The launch template ID '%s' does not exist", templateId)), nil
	}

	s.state.Delete(fmt.Sprintf("ec2:launch-templates:%s", templateId))

	return s.deleteLaunchTemplateResponse(template)
}
