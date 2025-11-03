package api

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPush_Local(t *testing.T) {
	data, err := push(&options{
		depFiles:    map[string]string{"../testdata/foo": "../testdata/foo"},
		files:       map[string]string{"../testdata/foo": "../testdata/foo"},
		outputBytes: true,
	})
	require.NoError(t, err)

	mft, err := tarball.LoadManifest(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	})
	require.NoError(t, err)

	ref, err := name.ParseReference(mft[0].RepoTags[0])
	require.NoError(t, err)
	assert.Equal(t, "5cffc09f", ref.Identifier())

	img, err := tarball.Image(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}, nil)
	require.NoError(t, err)

	layers, _ := img.Layers()
	metaLayer := layers[0]
	metaReader, _ := metaLayer.Uncompressed()
	metaTar := tar.NewReader(metaReader)
	var meta utils.CracMeta
	for h, err := metaTar.Next(); err != io.EOF; h, err = metaTar.Next() {
		require.NoError(t, err)
		if h.Name == fmt.Sprintf("/%s/meta.yaml", utils.Crac) {
			b, _ := io.ReadAll(metaTar)
			_ = yaml.Unmarshal(b, &meta)
			break
		}
	}
	assert.Equal(t, utils.CracVersion.String(), meta.Version)
}

func TestPush_Local_Pnpm(t *testing.T) {
	basepath := "../testdata/pnpm"

	depFiles, err := utils.ScanFiles([]string{filepath.Join(basepath, "pnpm-lock.yaml")})
	require.NoError(t, err)
	files, err := utils.ScanFiles([]string{filepath.Join(basepath, ".pnpm/store/**")})
	require.NoError(t, err)

	_, err = push(&options{
		depFiles:    depFiles,
		files:       files,
		outputBytes: true,
	})
	require.NoError(t, err)
}

func TestPush_Remote(t *testing.T) {
	_, err := push(&options{
		context:  t.Context(),
		repo:     fmt.Sprintf("host.docker.internal:5000/%s", utils.Crac),
		username: "testuser",
		password: "testpassword",
		insecure: true,
		depFiles: map[string]string{"../testdata/foo": "../testdata/foo"},
		files:    map[string]string{"../testdata/foo": "../testdata/foo"},
	})
	require.NoError(t, err)

	repo, _ := name.NewRepository(fmt.Sprintf("host.docker.internal:5000/%s", utils.Crac), name.Insecure)
	tags, err := remote.List(
		repo,
		remote.WithAuth(&authn.Basic{Username: "testuser", Password: "testpassword"}),
	)
	require.NoError(t, err)
	assert.Equal(t, "5cffc09f", tags[0])
}
