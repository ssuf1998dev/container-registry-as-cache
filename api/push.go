package api

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
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

func push(opts *PushOptions, isRemote bool) (*bytes.Buffer, error) {
	base := empty.Image

	meta, _ := json.Marshal(Meta{
		Version: CRAC_VERSION,
	})
	metaLayer, _ := crane.Layer(map[string][]byte{"/meta.json": meta})
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

	var tag string
	layers, _ := img.Layers()
	if len(opts.DepFiles) == 0 {
		ver, _ := layers[0].Digest()
		tag = ver.String()[0:8]
	} else {
		hashes := []string{}
		for _, f := range opts.DepFiles {
			b, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			hashes = append(hashes, fmt.Sprintf("%x", sha256.Sum256(b)))
		}
		hash := sha256.Sum256([]byte(strings.Join(hashes, "\n")))
		tag = hex.EncodeToString(hash[:])[0:8]
	}
	repo := opts.Repo
	if len(repo) == 0 {
		repo = "docker.io/library/crac"
	}
	if len(tag) == 0 {
		tag = "latest"
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
		return &buf, nil
	}
}
