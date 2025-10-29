package api

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPull(t *testing.T) {
	opts := BaseOptions{
		Repo:     "localhost:5000/crac",
		Username: "testuser",
		Password: "testpassword",
		Insecure: true,
	}
	deps := []string{"../testdata/foo"}

	_, err := push(&PushOptions{
		BaseOptions: opts,
		DepFiles:    deps,
		Files:       []string{"../testdata/foo"},
	}, true)
	require.NoError(t, err)

	cache, err := pull(&PullOptions{BaseOptions: opts, DepFiles: deps}, false)
	require.NoError(t, err)
	b, err := extraFileInTar(bytes.NewReader(cache), "../testdata/foo")
	require.NoError(t, err)
	assert.Equal(t, string(b), "bar")
}
