package provider

import "github.com/google/go-containerregistry/pkg/authn"

var _ authn.Authenticator = registryAuth{}

type registryAuth struct {
	Username string `tfsdk:"username"`
	Password string `tfsdk:"password"`
}

func (r registryAuth) Authorization() (*authn.AuthConfig, error) {
	return &authn.AuthConfig{
		Username:      r.Username,
		Password:      r.Password,
	}, nil
}
