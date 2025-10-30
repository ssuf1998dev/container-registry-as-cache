package api

import (
	"context"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/configfile"
)

func Login(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	cf, err := configfile.NewConfigFile()
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(o.repo)
	if err != nil {
		return err
	}
	key := ref.Context().RegistryStr()
	if cf.Config.Auths == nil {
		cf.Config.Auths = map[string]configfile.ConfigAuth{}
	}
	cf.Config.Auths[key] = configfile.ConfigAuth{
		Username: o.username,
		Password: o.password,
	}
	return cf.WriteConfig()
}

func Logout(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	cf, err := configfile.NewConfigFile()
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(o.repo)
	if err != nil {
		return err
	}
	key := ref.Context().RegistryStr()
	delete(cf.Config.Auths, key)

	return cf.WriteConfig()
}
