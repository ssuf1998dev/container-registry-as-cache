package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/goccy/go-yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/tarhelper"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
)

func Pull(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	_, err := pull(o)
	return err
}

func pull(opts *options) (tars []byte, err error) {
	tag := opts.tag
	if len(tag) == 0 {
		var keys []string
		keys = append(keys, opts.keys...)
		if len(opts.platform) > 0 {
			keys = append(keys, opts.platform)
		}
		tag, err = utils.ComputeTag(opts.depFiles, keys, opts.workdir)
		if err != nil {
			return nil, err
		}
	}
	repo := opts.repo
	if len(repo) == 0 {
		repo = fmt.Sprintf("%s/%s", name.DefaultRegistry, utils.Crac)
	}
	refOpts := []name.Option{}
	if opts.insecure {
		refOpts = append(refOpts, name.Insecure)
	}
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", repo, tag), refOpts...)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(
		ref,
		remote.WithAuth(&authn.Basic{Username: opts.username, Password: opts.password}),
		remote.WithContext(opts.context),
	)
	if err != nil {
		return nil, err
	}
	cf, _ := img.ConfigFile()
	metaIndex := slices.IndexFunc(cf.History, func(h v1.History) bool {
		return h.CreatedBy == utils.CreatedByCracMeta
	})
	if metaIndex < 0 {
		return nil, fmt.Errorf("invalid, \"%s\" not found", utils.CreatedByCracMeta)
	}

	layers, _ := img.Layers()
	metaLayer := layers[metaIndex]
	metaReader, _ := metaLayer.Uncompressed()
	metaData, _ := tarhelper.ExtraFileTar(metaReader, fmt.Sprintf("/%s/meta.yaml", utils.Crac))
	var meta utils.CracMeta
	_ = yaml.Unmarshal(metaData, &meta)
	if len(meta.Version) == 0 || !utils.CracVersionConstraint.Check(semver.MustParse(meta.Version)) {
		return nil, fmt.Errorf("invalid, version does't meet the constraint, (%s)", utils.CracVersionConstraint.String())
	}

	cacheIndex := slices.IndexFunc(cf.History, func(h v1.History) bool {
		return h.CreatedBy == utils.CreatedByCracCopy
	})
	if cacheIndex < 0 {
		return nil, fmt.Errorf("invalid, \"%s\" not found", utils.CreatedByCracCopy)
	}

	cacheLayer := layers[metaIndex]
	cacheReader, _ := cacheLayer.Uncompressed()
	if opts.outputStdout {
		_, err := io.Copy(os.Stdout, cacheReader)
		return nil, err
	}

	if opts.outputBytes {
		return io.ReadAll(cacheReader)
	}

	err = tarhelper.Untar(cacheReader, opts.workdir)
	return nil, err
}
