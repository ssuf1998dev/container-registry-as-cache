package api

import (
	"archive/tar"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type PullOptions struct {
	BaseOptions
	DepFiles []string
}

func Pull(opts *PullOptions) error {
	_, err := pull(opts, true)
	return err
}

func pull(opts *PullOptions, output bool) ([]byte, error) {
	tag, err := computeTag(opts.DepFiles)
	if err != nil {
		return nil, err
	}
	repo := opts.Repo
	if len(repo) == 0 {
		repo = fmt.Sprintf("%s/crac", name.DefaultRegistry)
	}
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", repo, tag))
	if err != nil {
		return nil, err
	}

	transport := remote.DefaultTransport.(*http.Transport)
	if opts.Insecure {
		transport = transport.Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	img, err := remote.Image(
		ref,
		remote.WithAuth(&authn.Basic{Username: opts.Username, Password: opts.Password}),
		remote.WithTransport(transport),
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
	metaData, _ := extraFileInTar(metaReader, "/crac/meta.json")
	var meta Meta
	_ = json.Unmarshal(metaData, &meta)
	if len(meta.Version) == 0 || !CRAC_VERSION_CONSTRAINT.Check(semver.MustParse(meta.Version)) {
		return nil, fmt.Errorf("invalid, version does't meet the constraint, (%s)", CRAC_VERSION_CONSTRAINT.String())
	}

	cacheIndex := metaIndex + 1
	cacheLayer := layers[cacheIndex]
	cacheReader, _ := cacheLayer.Uncompressed()
	if output {
		err := walkInTar(cacheReader, func(head *tar.Header, fi os.FileInfo, data []byte) (bool, error) {
			f, err := os.Create(fi.Name())
			if err != nil {
				return false, err
			}
			defer f.Close()
			_, err = f.Write(data)
			if err != nil {
				return false, err
			}
			os.Chmod(fi.Name(), fi.Mode().Perm())
			return false, nil
		})
		return nil, err
	} else {
		return io.ReadAll(cacheReader)
	}
}
