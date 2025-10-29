package api

import "context"

type Option func(*options)

type options struct {
	context  context.Context
	repo     string
	username string
	password string
	insecure bool

	depFiles []string
	files    []string
	platform string

	tag     string
	workdir string
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

func WithInsecure(insecure bool) Option {
	return func(o *options) {
		o.insecure = insecure
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
