package api

import (
	"bytes"
	"context"
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

func Push(opts ...Option) error {
	o := &options{
		context: context.Background(),
	}
	for _, option := range opts {
		option(o)
	}

	_, err := push(o, true)
	return err
}

func push(opts *options, isRemote bool) ([]byte, error) {
	base := empty.Image

	meta, _ := json.Marshal(cracMeta{
		Version: CracVersion.String(),
	})
	metaLayer, _ := crane.Layer(map[string][]byte{"/crac/meta.json": meta})
	img, _ := mutate.AppendLayers(base, metaLayer)

	files := make(map[string][]byte, len(opts.files))
	for _, f := range opts.files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		files[f] = b
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("empty image is not allowed")
	}
	cacheLayer, _ := crane.Layer(files)
	img, _ = mutate.AppendLayers(img, cacheLayer)

	cf, _ := img.ConfigFile()
	cf = cf.DeepCopy()
	if len(opts.platform) > 0 {
		archOs := strings.SplitN(opts.platform, "/", 2)
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

	tag, err := computeTag(opts.depFiles)
	if err != nil {
		return nil, err
	}
	repo := opts.repo
	if len(repo) == 0 {
		repo = fmt.Sprintf("%s/crac", name.DefaultRegistry)
	}
	ref, err := name.ParseReference(fmt.Sprintf("%s:%s", repo, tag))
	if err != nil {
		return nil, err
	}
	if isRemote {
		transport := remote.DefaultTransport.(*http.Transport)
		if opts.insecure {
			transport = transport.Clone()
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		err = remote.Write(
			ref, img,
			remote.WithAuth(&authn.Basic{Username: opts.username, Password: opts.password}),
			remote.WithTransport(transport),
			remote.WithContext(opts.context),
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
