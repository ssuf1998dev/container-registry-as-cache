package api

import (
	"context"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/configfile"
	cracprofile "github.com/ssuf1998dev/container-registry-as-cache/internal/profile"
)

type Option func(*options)

type options struct {
	context  context.Context
	repo     string
	username string
	password string
	insecure bool

	keys     []string
	depFiles []string
	files    []string
	platform string

	tag     string
	workdir string

	configfile string
}

func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.context = ctx
	}
}

func WithRepository(repository string) Option {
	return func(o *options) {
		o.repo = repository
	}
}

func WithUsername(username string) Option {
	return func(o *options) {
		o.username = username
	}
}

func WithLoginUsername() Option {
	return func(o *options) {
		o.username = func() string {
			cf := configfile.NewConfigFile(nil)
			err := cf.Read()
			if err != nil || cf.Config.Auths == nil {
				return o.username
			}
			ref, err := name.ParseReference(o.repo)
			if err != nil {
				return o.username
			}
			key := ref.Context().RegistryStr()
			if auth, ok := cf.Config.Auths[key]; ok {
				return auth.Username
			} else {
				return o.username
			}
		}()
	}
}

func WithPassword(password string) Option {
	return func(o *options) {
		o.password = password
	}
}

func WithLoginPassword() Option {
	return func(o *options) {
		o.password = func() string {
			cf := configfile.NewConfigFile(nil)
			err := cf.Read()
			if err != nil || cf.Config.Auths == nil {
				return o.password
			}
			ref, err := name.ParseReference(o.repo)
			if err != nil {
				return o.password
			}
			key := ref.Context().RegistryStr()
			if auth, ok := cf.Config.Auths[key]; ok {
				return auth.Password
			} else {
				return o.password
			}
		}()
	}
}

func WithInsecure(insecure bool) Option {
	return func(o *options) {
		o.insecure = insecure
	}
}

func WithKeys(keys []string) Option {
	return func(o *options) {
		o.keys = keys
	}
}

func WithDepFiles(depFiles []string) Option {
	return func(o *options) {
		o.depFiles = depFiles
	}
}

func WithFiles(files []string) Option {
	return func(o *options) {
		o.files = files
	}
}

func WithPlatform(platform string) Option {
	return func(o *options) {
		o.platform = platform
	}
}

func WithTag(tag string) Option {
	return func(o *options) {
		o.tag = tag
	}
}

func WithWorkdir(workdir string) Option {
	return func(o *options) {
		o.workdir = workdir
	}
}

func withConfigfile(configfile string) Option {
	return func(o *options) {
		o.configfile = configfile
	}
}

func WithProfile(profile string, file bool) Option {
	return func(o *options) {
		var text string

		if file {
			if b, err := os.ReadFile(profile); err == nil {
				text = string(b)
			}
		} else {
			switch profile {
			case "pnpm":
				text = cracprofile.Pnpm
			}
		}

		if len(text) == 0 {
			return
		}

		if p, err := cracprofile.Render(text); err == nil {
			o.keys = append(o.keys, p.Keys...)
			o.depFiles = append(o.depFiles, p.DepFiles...)
			o.files = append(o.keys, p.Files...)
		}
	}
}
