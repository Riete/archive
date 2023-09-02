package common

import (
	"io/fs"
	"os"
	"path/filepath"
)

// WalkSource recurse find all files
func WalkSource(srcs []string) ([]string, error) {
	var sources []string
	for _, src := range srcs {
		if f, err := os.Stat(src); err != nil {
			return sources, err
		} else if !f.IsDir() {
			sources = append(sources, src)
			continue
		}
		if err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			if f, err := os.Stat(path); err != nil {
				return err
			} else if !f.IsDir() {
				sources = append(sources, path)
			}
			return nil
		}); err != nil {
			return sources, err
		}
	}
	return sources, nil
}
