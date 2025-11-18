package api

import (
	"context"
	"io/fs"
	"maps"
	"os"
	"path/filepath"

	cracprofile "github.com/ssuf1998dev/container-registry-as-cache/internal/profile"
)

type Option func(*options)

type options struct {
	context   context.Context
	repo      string
	username  string
	password  string
	forceHttp bool
	insecure  bool

	keys     []string
	depFiles map[string]string
	files    map[string]string
	platform string
	filePerm fs.FileMode

	tag     string
	workdir string

	outputStdout bool
	outputBytes  bool
	outputFile   string

	forcePush bool
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

func WithPassword(password string) Option {
	return func(o *options) {
		o.password = password
	}
}

func WithForceHttp(forceHttp bool) Option {
	return func(o *options) {
		o.forceHttp = forceHttp
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

func WithDepFiles(depFiles map[string]string) Option {
	return func(o *options) {
		o.depFiles = depFiles
	}
}

func WithFiles(files map[string]string) Option {
	return func(o *options) {
		o.files = files
	}
}

func WithPlatform(platform string) Option {
	return func(o *options) {
		o.platform = platform
	}
}

func WithFilePerm(perm fs.FileMode) Option {
	return func(o *options) {
		o.filePerm = perm
	}
}

func WithTag(tag string) Option {
	return func(o *options) {
		o.tag = tag
	}
}

func WithWorkdir(workdir string) Option {
	return func(o *options) {
		if len(workdir) != 0 {
			o.workdir, _ = filepath.Abs(workdir)
			os.MkdirAll(o.workdir, 0766)
		}
	}
}

func WithProfile(profile string, profileType string) Option {
	return func(o *options) {
		var text string

		switch profileType {
		case "content":
			text = profile
		case "file":
			if b, err := os.ReadFile(profile); err == nil {
				text = string(b)
			}
		default:
			switch profile {
			case "pnpm":
				text = cracprofile.Pnpm
			}
		}

		if len(text) == 0 {
			return
		}
		p, err := cracprofile.Render(text, o.workdir)
		if err != nil {
			return
		}

		if o.keys == nil {
			o.keys = []string{}
		}
		o.keys = append(o.keys, p.Keys...)
		if o.depFiles == nil {
			o.depFiles = map[string]string{}
		}
		maps.Copy(o.depFiles, p.DepFiles.Value)
		if o.files == nil {
			o.files = map[string]string{}
		}
		maps.Copy(o.files, p.Files.Value)
	}
}

func WithOutputStdout(enable bool) Option {
	return func(o *options) {
		o.outputStdout = enable
	}
}

// func withOutputBytes(enable bool) Option {
// 	return func(o *options) {
// 		o.outputBytes = enable
// 	}
// }

func WithOutputFile(file string) Option {
	return func(o *options) {
		o.outputFile = file
	}
}

func WithForcePush(forcePush bool) Option {
	return func(o *options) {
		o.forcePush = forcePush
	}
}
