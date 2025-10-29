package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type PushOptions struct {
	BaseOptions
	DepFiles []string
	Files    []string
	Platform string
}

func Push(opts *PushOptions) error {
	_, err := push(opts, true)
	return err
}

func push(opts *PushOptions, isRemote bool) ([]byte, error) {
	base := empty.Image

	meta, _ := json.Marshal(Meta{
		Version: CRAC_VERSION.String(),
	})
	metaLayer, _ := crane.Layer(map[string][]byte{"/crac/meta.json": meta})
	img, _ := mutate.AppendLayers(base, metaLayer)

	files := make(map[string][]byte, len(opts.Files))
	for _, f := range opts.Files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		files[f] = b
	}
	cacheLayer, _ := crane.Layer(files)
	img, _ = mutate.AppendLayers(img, cacheLayer)

	cf, _ := img.ConfigFile()
	cf = cf.DeepCopy()
	if len(opts.Platform) > 0 {
		archOs := strings.SplitN(opts.Platform, "/", 2)
		cf.OS = archOs[0]
		cf.Architecture = archOs[1]
	} else {
		cf.OS = runtime.GOOS
		cf.Architecture = runtime.GOARCH
	}
	cf.Created = v1.Time{Time: time.Now()}
	cf.History = []v1.History{
		{Created: cf.Created, CreatedBy: "CRACMETA"},
		{Created: cf.Created, CreatedBy: "CRACCOPY"},
	}
	img, _ = mutate.ConfigFile(img, cf)

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
	if isRemote {
		transport := remote.DefaultTransport.(*http.Transport)
		if opts.Insecure {
			transport = transport.Clone()
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		err = remote.Write(
			ref, img,
			remote.WithAuth(&authn.Basic{Username: opts.Username, Password: opts.Password}),
			remote.WithTransport(transport),
		)
		return nil, err
	} else {
		var buf bytes.Buffer
		if err := tarball.Write(ref, img, &buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}
