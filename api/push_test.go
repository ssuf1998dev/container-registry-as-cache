package api

import (
	"io"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	buf, err := push(&PushOptions{
		DepFiles: []string{"../testdata/foo"},
		Files:    []string{"../testdata/foo"},
	}, false)
	require.NoError(t, err)

	_, err = tarball.Image(func() (io.ReadCloser, error) {
		return io.NopCloser(buf), nil
	}, nil)
	require.NoError(t, err)
}
