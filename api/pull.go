package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func Pull(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	_, err := pull(o, true)
	return err
}

func pull(opts *options, output bool) ([]byte, error) {
	tag := opts.tag
	var err error
	if len(tag) == 0 {
		tag, err = computeTag(opts.depFiles, opts.keys)
		if err != nil {
			return nil, err
		}
	}
	repo := opts.repo
	if len(repo) == 0 {
		repo = fmt.Sprintf("%s/crac", name.DefaultRegistry)
	}
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", repo, tag))
	if err != nil {
		return nil, err
	}

	transport := remote.DefaultTransport.(*http.Transport)
	if opts.insecure {
		transport = transport.Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	img, err := remote.Image(
		ref,
		remote.WithAuth(&authn.Basic{Username: opts.username, Password: opts.password}),
		remote.WithTransport(transport),
		remote.WithContext(opts.context),
	)
	if err != nil {
		return nil, err
	}
	cf, _ := img.ConfigFile()
	metaIndex := -1
	for i, history := range cf.History {
		if history.CreatedBy == "CRACMETA" {
			metaIndex = i
			break
		}
	}
	if metaIndex < 0 {
		return nil, fmt.Errorf("invalid, \"CRACMETA\" not found")
	}

	layers, _ := img.Layers()
	metaLayer := layers[metaIndex]
	metaReader, _ := metaLayer.Uncompressed()
	metaData, _ := extraFileTar(metaReader, "/crac/meta.json")
	var meta cracMeta
	_ = json.Unmarshal(metaData, &meta)
	if len(meta.Version) == 0 || !cracVersionConstraint.Check(semver.MustParse(meta.Version)) {
		return nil, fmt.Errorf("invalid, version does't meet the constraint, (%s)", cracVersionConstraint.String())
	}

	cacheIndex := metaIndex + 1
	cacheLayer := layers[cacheIndex]
	cacheReader, _ := cacheLayer.Uncompressed()
	if output {
		err := untar(cacheReader, opts.workdir)
		return nil, err
	} else {
		return io.ReadAll(cacheReader)
	}
}
