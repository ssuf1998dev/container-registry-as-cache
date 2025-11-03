package utils

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanFiles_Pnpm(t *testing.T) {
	basepath := "../../testdata/pnpm"

	cmd := exec.Command("pnpm", "install")
	cmd.Dir = basepath
	output, err := cmd.Output()
	if err != nil {
		t.Skipf("pnpm is not ready, skip, %s", err)
	}
	t.Logf("%s\n", output)

	files, err := ScanFiles([]string{filepath.Join(basepath, ".pnpm/store/**")})
	require.NoError(t, err)
	require.Greater(t, len(files), 0)
}
