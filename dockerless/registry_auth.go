package dockerless

import "github.com/google/go-containerregistry/pkg/authn"

var _ authn.Authenticator = RegistryAuth{}

type RegistryAuth struct {
	Username string `tfsdk:"username"`
	Password string `tfsdk:"password"`
}

func (r RegistryAuth) Authorization() (*authn.AuthConfig, error) {
	return &authn.AuthConfig{
		Username: r.Username,
		Password: r.Password,
	}, nil
}
