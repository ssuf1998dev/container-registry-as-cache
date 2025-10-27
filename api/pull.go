package api

import (
	"archive/tar"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type PullOptions struct {
	BaseOptions
}

func Pull(opts *PullOptions) error {
	ref, err := name.ParseReference(opts.Repo)
	if err != nil {
		return err
	}

	transport := remote.DefaultTransport.(*http.Transport)
	if opts.Insecure {
		transport = transport.Clone()
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	img, err := remote.Image(
		ref,
		remote.WithAuth(&authn.Basic{Username: opts.Username, Password: opts.Password}),
		remote.WithTransport(transport),
	)
	if err != nil {
		return err
	}
	cf, _ := img.ConfigFile()
	meta := -1
	for i, history := range cf.History {
		if history.CreatedBy == "CRACMETA" {
			meta = i
			break
		}
	}
	if meta < 0 {
		return fmt.Errorf("invalid")
	}

	layers, _ := img.Layers()
	// metaLayer := layers[meta]
	cacheLayer := layers[meta+1]
	cacheReader, _ := cacheLayer.Uncompressed()
	cacheTar := tar.NewReader(cacheReader)
	for h, err := cacheTar.Next(); err != io.EOF; h, err = cacheTar.Next() {
		if err != nil {
			return err
		}
		fi := h.FileInfo()
		f, err := os.Create(fi.Name())
		if err != nil {
			return err
		}
		_, err = io.Copy(f, cacheTar)
		if err != nil {
			return err
		}
		os.Chmod(fi.Name(), fi.Mode().Perm())
		f.Close()
	}
	return nil
}
