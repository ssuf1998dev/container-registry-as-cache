package api

import (
	"archive/tar"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushLocal(t *testing.T) {
	data, err := push(&PushOptions{
		DepFiles: []string{"../testdata/foo"},
		Files:    []string{"../testdata/foo"},
	}, false)
	require.NoError(t, err)

	mft, err := tarball.LoadManifest(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	})
	require.NoError(t, err)

	ref, err := name.ParseReference(mft[0].RepoTags[0])
	require.NoError(t, err)
	assert.Equal(t, "4d6d865b", ref.Identifier())

	img, err := tarball.Image(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}, nil)
	require.NoError(t, err)

	layers, _ := img.Layers()
	metaLayer := layers[0]
	metaReader, _ := metaLayer.Uncompressed()
	metaTar := tar.NewReader(metaReader)
	var meta Meta
	for h, err := metaTar.Next(); err != io.EOF; h, err = metaTar.Next() {
		require.NoError(t, err)
		if h.Name == "/crac/meta.json" {
			b, _ := io.ReadAll(metaTar)
			_ = json.Unmarshal(b, &meta)
			break
		}
	}
	assert.Equal(t, CRAC_VERSION.String(), meta.Version)
}

func TestPushRemote(t *testing.T) {
	_, err := push(&PushOptions{
		BaseOptions: BaseOptions{
			Repo:     "localhost:5000/crac",
			Username: "testuser",
			Password: "testpassword",
			Insecure: true,
		},
		DepFiles: []string{"../testdata/foo"},
		Files:    []string{"../testdata/foo"},
	}, true)
	require.NoError(t, err)

	transport := remote.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	repo, _ := name.NewRepository("localhost:5000/crac")
	tags, err := remote.List(
		repo,
		remote.WithAuth(&authn.Basic{Username: "testuser", Password: "testpassword"}),
		remote.WithTransport(transport),
	)
	require.NoError(t, err)
	assert.Equal(t, "4d6d865b", tags[0])
}
