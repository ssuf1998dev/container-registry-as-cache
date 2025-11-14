package api

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/tarhelper"
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
	assert.Equal(t, "bd142ccf", ref.Identifier())

	img, err := tarball.Image(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}, nil)
	require.NoError(t, err)

	cf, _ := img.ConfigFile()
	metaIndex := slices.IndexFunc(cf.History, func(h v1.History) bool {
		return h.CreatedBy == utils.CreatedByCracMeta
	})
	require.GreaterOrEqual(t, metaIndex, 0)
	layers, _ := img.Layers()
	metaLayer := layers[metaIndex]
	metaReader, _ := metaLayer.Uncompressed()
	var meta utils.CracMeta
	tarhelper.WalkTar(metaReader, func(header *tar.Header, fi os.FileInfo, data []byte) (bool, error) {
		if header.Name == fmt.Sprintf("/%s/meta.yaml", utils.Crac) {
			return true, yaml.Unmarshal(data, &meta)
		}
		return false, nil
	})
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
		workdir:     basepath,
	})
	require.NoError(t, err)
}

func TestPush_Remote(t *testing.T) {
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("http_proxy")
	reg := os.Getenv("CRAC_TEST_REGISTRY")
	if reg == "" {
		t.Skipf("registry is empty, skip")
	}

	_, err := push(&options{
		context:   t.Context(),
		repo:      fmt.Sprintf("%s/%s", reg, utils.Crac),
		username:  "testuser",
		password:  "testpassword",
		forceHttp: true,
		insecure:  true,
		depFiles:  map[string]string{"../testdata/foo": "../testdata/foo"},
		files:     map[string]string{"../testdata/foo": "../testdata/foo"},
		forcePush: true,
	})
	require.NoError(t, err)

	repo, _ := name.NewRepository(fmt.Sprintf("%s/%s", reg, utils.Crac), name.Insecure)
	tags, err := remote.List(
		repo,
		remote.WithAuth(&authn.Basic{Username: "testuser", Password: "testpassword"}),
	)
	require.NoError(t, err)
	assert.Equal(t, "bd142ccf", tags[0])
}
