package dockerless

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/crane"
)

func (c Client) DeleteImageTag(dockerName string) error {
	_, opts, _, err := c.GetCraneReference(dockerName)
	if err != nil {
		return err
	}

	if err := crane.Delete(dockerName, opts...); err != nil {
		return fmt.Errorf("unable to delete remote image: %w", err)
	}
	return nil
}
