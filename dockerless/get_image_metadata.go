package dockerless

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func (c Client) GetImageMetadata(dockerName string) (*v1.Descriptor, error) {
	_, opts, _, err := c.GetCraneReference(dockerName)
	if err != nil {
		return nil, err
	}

	meta, err := crane.Head(dockerName, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve image metadata: %w", err)
	}
	return meta, nil
}
