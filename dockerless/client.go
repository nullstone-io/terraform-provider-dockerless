package dockerless

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Client struct {
	Registries map[string]RegistryAuth
}

func (c Client) FindRegistryAuth(address string) *RegistryAuth {
	for key, registryAuth := range c.Registries {
		if key == address {
			return &registryAuth
		}
	}
	return nil
}

func (c Client) GetCraneReference(dockerName string) (name.Reference, []crane.Option, []remote.Option, error) {
	craneOptions := make([]crane.Option, 0)
	remoteOptions := make([]remote.Option, 0)

	ref, err := name.ParseReference(dockerName)
	if err != nil {
		return ref, craneOptions, remoteOptions, fmt.Errorf("%q is an invalid docker reference: %s", dockerName, err)
	}
	address := ref.Context().RegistryStr()
	registryAuth := c.FindRegistryAuth(address)
	if registryAuth != nil {
		craneOptions = append(craneOptions, crane.WithAuth(registryAuth))
		remoteOptions = append(remoteOptions, remote.WithAuth(registryAuth))
	}
	return ref, craneOptions, remoteOptions, nil
}
