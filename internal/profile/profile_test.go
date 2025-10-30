package profile

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	rendered, err := Render(pnpm)
	require.NoError(t, err)

	var p Profile
	err = yaml.Unmarshal(rendered, &p)
	require.NoError(t, err)

	require.Contains(t, p.Files[0], "pnpm/store")
}
