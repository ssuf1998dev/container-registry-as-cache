package api

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPull(t *testing.T) {
	deps := []string{"../testdata/foo"}

	_, err := push(&options{
		context:  t.Context(),
		repo:     fmt.Sprintf("localhost:5000/%s", Crac),
		username: "testuser",
		password: "testpassword",
		insecure: true,
		depFiles: deps,
		files:    []string{"../testdata/foo"},
	}, true)
	require.NoError(t, err)

	cache, err := pull(&options{
		context:  t.Context(),
		repo:     fmt.Sprintf("localhost:5000/%s", Crac),
		username: "testuser",
		password: "testpassword",
		insecure: true,
		depFiles: deps,
	}, false)
	require.NoError(t, err)
	b, err := extraFileTar(bytes.NewReader(cache), "../testdata/foo")
	require.NoError(t, err)
	assert.Equal(t, string(b), "bar")
}
