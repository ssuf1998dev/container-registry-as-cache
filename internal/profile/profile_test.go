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
	basepath := "../../testdata/pnpm"
	os.Chdir(basepath)
	cwd, _ := os.Getwd()

	cmd := exec.Command("pnpm", "install")
	cmd.Dir = basepath
	output, err := cmd.Output()
	if err != nil {
		t.Skipf("pnpm is not ready, skip, %s", err)
	}
	t.Logf("%s\n", output)

	p, err := Render(Pnpm, "")
	require.NoError(t, err)
	pnpmStoreOutput, err := exec.Command("pnpm", "store", "path").Output()
	require.NoError(t, err)
	pnpmStore, _ := filepath.Rel(cwd, string(pnpmStoreOutput))
	pnpmStore = strings.TrimSpace(pnpmStore)

	for _, v := range p.Files.Value {
		f, _ := filepath.Rel(cwd, v)
		require.True(t, strings.HasPrefix(f, pnpmStore))
	}
}
