package api

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/name"
)

var CracVersion = semver.MustParse("1.0.0")
var cracVersionConstraint, _ = semver.NewConstraint(fmt.Sprintf(">= %d < %d", CracVersion.Major(), CracVersion.Major()+1))

type cracMeta struct {
	Version string `json:"created,omitempty"`
}

func computeTag(files []string, keys []string) (string, error) {
	tag := name.DefaultTag
	hashes := []string{}

	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}
		f, _ = filepath.Abs(f)
		b = append(fmt.Appendf(nil, "%s:", f), b...)
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

func walkTar(r io.Reader, callback func(header *tar.Header, fi os.FileInfo, data []byte) (bool, error)) error {
	tr := tar.NewReader(r)
	for header, err := tr.Next(); err != io.EOF; header, err = tr.Next() {
		if err != nil {
			return err
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			return err
		}
		stop, err := callback(header, header.FileInfo(), b)
		if stop {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func extraFileTar(r io.Reader, path string) ([]byte, error) {
	var b []byte
	err := walkTar(r, func(header *tar.Header, fi os.FileInfo, data []byte) (bool, error) {
		if header.Name == path {
			b = data
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("file \"%s\" not found", path)
	}
	return b, nil
}

func untar(r io.Reader, dst string) error {
	return walkTar(r, func(header *tar.Header, fi os.FileInfo, data []byte) (bool, error) {
		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {

		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return false, err
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return false, err
			}

			if _, err := io.Copy(f, r); err != nil {
				return false, err
			}

			f.Close()
		}

		return false, nil
	})
}
