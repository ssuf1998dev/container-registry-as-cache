package api

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
)

func Push(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	_, err := push(o)
	return err
}

func push(opts *options) ([]byte, error) {
	base := empty.Image

	meta, _ := yaml.Marshal(utils.CracMeta{
		Version: utils.CracVersion.String(),
	})
	metaLayer, _ := crane.Layer(map[string][]byte{fmt.Sprintf("/%s/meta.yaml", utils.Crac): meta})
	img, _ := mutate.AppendLayers(base, metaLayer)

	files := make(map[string][]byte, len(opts.files))
	for fn, f := range opts.files {
		b, err := os.ReadFile(filepath.Join(opts.workdir, f))
		if err != nil {
			return nil, err
		}
		files[fn] = b
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("empty image is not allowed")
	}
	cacheLayer, _ := crane.Layer(files)
	img, _ = mutate.AppendLayers(img, cacheLayer)

	cf, _ := img.ConfigFile()
	cf = cf.DeepCopy()
	cf.Created = v1.Time{Time: time.Now()}
	cf.History = []v1.History{
		{Created: cf.Created, CreatedBy: utils.CreatedByCracMeta},
		{Created: cf.Created, CreatedBy: utils.CreatedByCracCopy},
	}
	img, _ = mutate.ConfigFile(img, cf)

	var keys []string
	keys = append(keys, opts.keys...)
	if len(opts.platform) > 0 {
		keys = append(keys, opts.platform)
	}
	tag, err := utils.ComputeTag(opts.depFiles, keys, opts.workdir)
	if err != nil {
		return nil, err
	}
	repo := opts.repo
	if len(repo) == 0 {
		repo = fmt.Sprintf("%s/%s", name.DefaultRegistry, utils.Crac)
	}
	nameOpts := []name.Option{}
	if opts.insecure {
		nameOpts = append(nameOpts, name.Insecure)
	}
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", repo, tag), nameOpts...)
	if err != nil {
		return nil, err
	}

	if opts.outputStdout {
		err := tarball.Write(ref, img, os.Stdout)
		return nil, err
	}

	if opts.outputBytes {
		var buf bytes.Buffer
		if err := tarball.Write(ref, img, &buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	if len(opts.outputFile) != 0 {
		f, err := os.OpenFile(opts.outputFile, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		err = tarball.Write(ref, img, f)
		return nil, err
	}

	err = remote.Write(
		ref, img,
		remote.WithAuth(&authn.Basic{Username: opts.username, Password: opts.password}),
		remote.WithContext(opts.context),
	)
	return nil, err
}
