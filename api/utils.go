package api

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

func computeTag(files []string) (string, error) {
	tag := name.DefaultTag
	if len(files) != 0 {
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
		hash := sha256.Sum256([]byte(strings.Join(hashes, "\n")))
		tag = hex.EncodeToString(hash[:])[0:8]
	}
	return tag, nil
}

func walkInTar(r io.Reader, callback func(head *tar.Header, fi os.FileInfo, data []byte) (bool, error)) error {
	tr := tar.NewReader(r)
	for head, err := tr.Next(); err != io.EOF; head, err = tr.Next() {
		if err != nil {
			return err
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			return err
		}
		stop, err := callback(head, head.FileInfo(), b)
		if stop {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func extraFileInTar(r io.Reader, path string) ([]byte, error) {
	var b []byte
	err := walkInTar(r, func(head *tar.Header, fi os.FileInfo, data []byte) (bool, error) {
		if head.Name == path {
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
