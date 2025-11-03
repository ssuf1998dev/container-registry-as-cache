package profile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	p, err := Render(Pnpm)
	require.NoError(t, err)
	pnpmStoreOutput, err := exec.Command("pnpm", "store", "path").Output()
	require.NoError(t, err)
	cwd, _ := os.Getwd()
	pnpmStore, _ := filepath.Rel(cwd, string(pnpmStoreOutput))
	pnpmStore = strings.TrimSpace(pnpmStore)
	require.True(t, strings.HasPrefix(p.Files[0], pnpmStore))
}
