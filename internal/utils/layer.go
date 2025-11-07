package utils

import (
	"archive/tar"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type TarLayer struct {
	v1.Layer
	File string
}

func NewTarLayer(files map[string]string) (*TarLayer, error) {
	id, err := gonanoid.New()
	if err != nil {
		return nil, err
	}
	dst := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.tar", Crac, id))
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	tw := tar.NewWriter(file)
	fn := []string{}
	for f := range files {
		fn = append(fn, f)
	}
	sort.Strings(fn)
	for _, f := range fn {
		b, err := os.ReadFile(files[f])
		if err != nil {
			return nil, err
		}
		if err := tw.WriteHeader(&tar.Header{
			Name: f,
			Size: int64(len(b)),
		}); err != nil {
			return nil, err
		}
		if _, err := tw.Write(b); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}

	layer, err := tarball.LayerFromFile(dst)
	if err != nil {
		return nil, err
	}
	return &TarLayer{Layer: layer, File: dst}, nil
}
