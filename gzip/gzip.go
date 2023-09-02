package gzip

import (
	"compress/gzip"
	"io"
	"os"
	"path"

	"github.com/riete/archive"
)

type gzipPacker struct {
	sources []string
}

// pack user source.gz as filename, remove source file when gzip file created success
func (g gzipPacker) pack(src string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	dst := src + ".gz"
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	w := gzip.NewWriter(f)
	w.Header.Name = src
	defer w.Close()

	_, err = io.Copy(w, r)
	if err == nil {
		if err := os.Remove(src); err != nil {
			return err
		}
	}
	return err
}

func (g gzipPacker) Pack() error {
	for _, src := range g.sources {
		if err := g.pack(src); err != nil {
			return err
		}
	}
	return nil
}

// unpack read original filename from header
func (g gzipPacker) unpack(src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(path.Join(path.Dir(src), path.Base(r.Header.Name)))
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	if err == nil {
		if err := os.Remove(src); err != nil {
			return err
		}
	}
	return err
}

func (g gzipPacker) Unpack() error {
	for _, source := range g.sources {
		if err := g.unpack(source); err != nil {
			return err
		}
	}
	return nil
}

func NewGzipPacker(sources ...string) archive.Pack {
	return &gzipPacker{sources: sources}
}

func NewGzipUnPacker(sources ...string) archive.Unpack {
	return &gzipPacker{sources: sources}
}
