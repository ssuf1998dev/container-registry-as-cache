package api

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/configfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var f = "../testdata/config.json"

func login(t *testing.T) *configfile.ConfigFile {
	err := Login(
		withConfigfile(f),
		WithRepository("docker.io/library/golang"),
		WithUsername("testuser"),
		WithPassword("testpassword"),
	)
	require.NoError(t, err)

	cf := configfile.NewConfigFile(&configfile.NewOptions{File: f})
	err = cf.Read()
	require.NoError(t, err)
	assert.Equal(t, cf.Config.Auths, map[string]configfile.ConfigAuth{
		"index.docker.io": {Username: "testuser", Password: "testpassword"},
	})
	return cf
}

func TestLogin(t *testing.T) {
	cf := login(t)
	err := cf.Reset()
	require.NoError(t, err)
}

func TestLogout(t *testing.T) {
	cf := login(t)

	err := Logout(
		withConfigfile(f),
		WithRepository("docker.io/library/golang"),
	)
	require.NoError(t, err)

	err = cf.Read()
	require.NoError(t, err)

	if cf.Config.Auths != nil {
		require.Fail(t, "auth exists")
		return
	}

	if _, ok := cf.Config.Auths[name.DefaultRegistry]; ok {
		require.Fail(t, "auth exists")
		return
	}

	err = cf.Reset()
	require.NoError(t, err)
}
