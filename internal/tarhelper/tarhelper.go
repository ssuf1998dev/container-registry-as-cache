package tarhelper

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func WalkTar(r io.Reader, callback func(header *tar.Header, fi os.FileInfo, data []byte) (bool, error)) error {
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

func UntarFile(r io.Reader, path string) ([]byte, error) {
	var b []byte
	err := WalkTar(r, func(header *tar.Header, fi os.FileInfo, data []byte) (bool, error) {
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

func Untar(r io.Reader, dst string) error {
	return WalkTar(r, func(header *tar.Header, fi os.FileInfo, data []byte) (bool, error) {
		// force to make header.Name relative to dst
		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {

		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return false, err
				}
			}

		case tar.TypeReg:
			dir := filepath.Dir(target)
			if _, err := os.Stat(dir); err != nil {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return false, err
				}
			}

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
