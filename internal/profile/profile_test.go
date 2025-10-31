package profile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	p, err := Render(Pnpm)
	require.NoError(t, err)

	require.Contains(t, p.Files[0], "pnpm/store")
}
