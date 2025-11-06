package api

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ssuf1998dev/container-registry-as-cache/internal/tarhelper"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPull(t *testing.T) {
	deps := map[string]string{"../testdata/foo": "../testdata/foo"}

	_, err := push(&options{
		context:  t.Context(),
		repo:     fmt.Sprintf("host.docker.internal:5000/%s", utils.Crac),
		username: "testuser",
		password: "testpassword",
		insecure: true,
		depFiles: deps,
		files:    map[string]string{"../testdata/foo": "../testdata/foo"},
	})
	require.NoError(t, err)

	cache, err := pull(&options{
		context:     t.Context(),
		repo:        fmt.Sprintf("host.docker.internal:5000/%s", utils.Crac),
		username:    "testuser",
		password:    "testpassword",
		insecure:    true,
		depFiles:    deps,
		outputBytes: true,
	})
	require.NoError(t, err)
	b, err := tarhelper.UntarFile(bytes.NewReader(cache), "../testdata/foo")
	require.NoError(t, err)
	assert.Equal(t, string(b), "bar")
}
