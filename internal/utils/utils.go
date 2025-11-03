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

func ComputeTag(files []string, keys []string) (string, error) {
	tag := name.DefaultTag
	hashes := []string{}

	for _, f := range files {
		f = strings.TrimSpace(f)
		b, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}
		f, _ = filepath.Abs(f)
		b = append(fmt.Appendf(nil, "%s\n", f), b...)
		hashes = append(hashes, fmt.Sprintf("%x", sha256.Sum256(b)))
	}

	hashes = append(hashes, keys...)
	sort.Strings(hashes)

	if len(hashes) != 0 {
		hash := sha256.Sum256([]byte(strings.Join(hashes, "\n")))
		tag = hex.EncodeToString(hash[:])[0:8]
	}

	return tag, nil
}

func ScanFiles(patterns []string) ([]string, error) {
	list := []string{}
	for _, item := range patterns {
		basepath, pattern := doublestar.SplitPattern(item)
		fsys := os.DirFS(basepath)
		matches, err := doublestar.Glob(fsys, pattern, doublestar.WithFilesOnly())
		if err != nil {
			return nil, err
		}
		list = append(list, matches...)
	}
	return list, nil
}
