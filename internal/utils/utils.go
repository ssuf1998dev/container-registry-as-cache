package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/go-containerregistry/pkg/name"
)

type CracMeta struct {
	Version string `yaml:"version,omitempty"`
}

func ComputeTag(files map[string]string, keys []string, workdir string) (string, error) {
	tag := name.DefaultTag
	hashes := []string{}

	for _, f := range files {
		b, err := os.ReadFile(PathJoinRespectAbs(workdir, f))
		if err != nil {
			return "", err
		}
		b = append(fmt.Appendf(nil, "%s\n", filepath.Base(f)), b...)
		hashes = append(hashes, fmt.Sprintf("%x", sha256.Sum256(b)))
	}

	for _, k := range keys {
		hashes = append(hashes, fmt.Sprintf("%x", sha256.Sum256([]byte(k))))
	}

	sort.Strings(hashes)

	if len(hashes) != 0 {
		hash := sha256.Sum256([]byte(strings.Join(hashes, "\n")))
		tag = hex.EncodeToString(hash[:])[0:8]
	}

	return tag, nil
}

func ScanFiles(patterns []string) (map[string]string, error) {
	m := map[string]string{}
	for _, item := range patterns {
		basepath, pattern := doublestar.SplitPattern(filepath.ToSlash(item))
		fsys := os.DirFS(basepath)
		matches, err := doublestar.Glob(fsys, pattern, doublestar.WithFilesOnly())
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			abs, err := filepath.Abs(filepath.Join(basepath, match))
			if err != nil {
				return nil, err
			}
			m[match] = abs
		}
	}
	return m, nil
}

func PathJoinRespectAbs(elem ...string) string {
	for _, item := range elem[1:] {
		if filepath.IsAbs(item) {
			return item
		}
	}
	return filepath.Join(elem...)
}
