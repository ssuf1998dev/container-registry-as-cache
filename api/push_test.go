package api

import (
	"archive/tar"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
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

func TestPushLocal(t *testing.T) {
	data, err := push(&options{
		depFiles:    []string{"../testdata/foo"},
		files:       []string{"../testdata/foo"},
		outputBytes: true,
	})
	require.NoError(t, err)

	mft, err := tarball.LoadManifest(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	})
	require.NoError(t, err)

	ref, err := name.ParseReference(mft[0].RepoTags[0])
	require.NoError(t, err)
	assert.Equal(t, "716639f2", ref.Identifier())

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

func TestPushRemote(t *testing.T) {
	_, err := push(&options{
		context:  t.Context(),
		repo:     fmt.Sprintf("localhost:5000/%s", utils.Crac),
		username: "testuser",
		password: "testpassword",
		insecure: true,
		depFiles: []string{"../testdata/foo"},
		files:    []string{"../testdata/foo"},
	})
	require.NoError(t, err)

	transport := remote.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	repo, _ := name.NewRepository(fmt.Sprintf("localhost:5000/%s", utils.Crac))
	tags, err := remote.List(
		repo,
		remote.WithAuth(&authn.Basic{Username: "testuser", Password: "testpassword"}),
		remote.WithTransport(transport),
	)
	require.NoError(t, err)
	assert.Equal(t, "716639f2", tags[0])
}
