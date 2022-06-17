package dockerless

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"io/ioutil"
	"os"
)

func (c Client) ForwardImage(src, target string) (string, error) {
	srcRef, srcCraneOpts, srcRemoteOpts, err := c.GetCraneReference(src)
	if err != nil {
		return "", fmt.Errorf("source %q is an invalid docker reference: %s", src, err)
	}
	targetRef, _, targetRemoteOpts, err := c.GetCraneReference(target)
	if err != nil {
		return "", fmt.Errorf("target %q is an invalid docker reference: %s", target, err)
	}

	// Retrieve metadata about source image and build image map for pulling image
	rmt, err := remote.Get(srcRef, srcRemoteOpts...)
	if err != nil {
		return "", fmt.Errorf("error retrieving metadata for source image: %s", err)
	}
	img, err := rmt.Image()
	if err != nil {
		return "", fmt.Errorf("error preparing source image for pull: %s", err)
	}
	imageMap := map[string]v1.Image{src: img}

	// Pull docker image using crane and save it as a tarball to 'path'

	file, err := ioutil.TempFile(".", "tmp_remote_image_*.tgz")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file for source image: %s", err)
	}
	file.Close() // close immediately to allow pull to work

	path := file.Name()
	if err := crane.MultiSave(imageMap, path, srcCraneOpts...); err != nil {
		return "", fmt.Errorf("error pulling source image: %s", err)
	}

	// Load image from tarball and push it
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("error finding source tarball: %s", err)
	}
	pushImg, err := crane.Load(path)
	if err != nil {
		return "", fmt.Errorf("loading %s as tarball: %w", path, err)
	}
	if err := remote.Write(targetRef, pushImg, targetRemoteOpts...); err != nil {
		return "", fmt.Errorf("error pushing image: %w", err)
	}

	h, err := pushImg.Digest()
	if err != nil {
		return "", err
	}
	return targetRef.Context().Digest(h.String()).String(), nil
}
