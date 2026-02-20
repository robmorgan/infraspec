package ec2

import (
	"context"
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

func (s *EC2Service) describeImages(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	imageIds := s.parseImageIds(params)

	var images []Image

	if len(imageIds) > 0 {
		for _, imageId := range imageIds {
			var image Image
			if err := s.state.Get(fmt.Sprintf("ec2:images:%s", imageId), &image); err != nil {
				return s.errorResponse(400, "InvalidAMIID.NotFound", fmt.Sprintf("The image id '[%s]' does not exist", imageId)), nil
			}
			images = append(images, image)
		}
	} else {
		keys, err := s.state.List("ec2:images:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list images"), nil
		}

		for _, key := range keys {
			var image Image
			if err := s.state.Get(key, &image); err == nil {
				images = append(images, image)
			}
		}
	}

	return s.describeImagesResponse(images)
}
