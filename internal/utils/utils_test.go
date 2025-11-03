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
	t.Logf("%s\n", output)
	require.NoError(t, err)

	files, err := ScanFiles([]string{filepath.Join(basepath, ".pnpm/store/**")})
	require.NoError(t, err)
	require.Greater(t, len(files), 0)
}
