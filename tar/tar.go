package tar

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/riete/archive"

	"github.com/riete/archive/common"

	set "github.com/riete/go-set"
)

type tarPacker struct {
	w        *tar.Writer
	compress bool
	filename string
	cwd      string
	sources  []string
}

func (t *tarPacker) writeHeader(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	fileInfo, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return err
	}
	fileInfo.Name = strings.TrimLeft(path, "/")
	return t.w.WriteHeader(fileInfo)
}

func (t *tarPacker) pack(src string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err = t.writeHeader(src); err != nil {
		return err
	}
	_, err = io.Copy(t.w, r)
	return err
}

func (t *tarPacker) Pack() error {
	if t.cwd != "" {
		if err := os.Chdir(t.cwd); err != nil {
			return err
		}
	}
	_ = os.Remove(t.filename)
	if _, err := os.Stat(path.Dir(t.filename)); errors.Is(err, fs.ErrNotExist) {
		if err = os.MkdirAll(path.Dir(t.filename), 0777); err != nil {
			return err
		}
	}

	sources, err := common.WalkSource(t.sources)
	if err != nil {
		return err
	}

	f, err := os.Create(t.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if t.compress {
		gw := gzip.NewWriter(f)
		defer gw.Close()
		t.w = tar.NewWriter(gw)
	} else {
		t.w = tar.NewWriter(f)
	}
	defer t.w.Close()

	for _, src := range sources {
		if err = t.pack(src); err != nil {
			return err
		}
	}
	return nil
}

// NewTarPacker cwd is working directory when packing, use this argument to exclude the leading directories
// cwd == "" (do not chdir), source = "/tmp/aa/1" -->  "tmp/aa/1" in tar file
// to remove leading directories
// cwd == "/tmp/aa", source = "1" --> "1" in tar file
// cwd == "/tmp/aa", source = "."  pack all items in /tmp/aa to tar file with no "tmp/aa" prefixed
func NewTarPacker(filename, cwd string, compress bool, sources ...string) archive.Pack {
	return &tarPacker{compress: compress, filename: filename, cwd: cwd, sources: sources}
}

type tarUnPacker struct {
	r          *tar.Reader
	compress   bool
	filename   string
	todir      string
	createdDir set.Set
	src        io.ReadCloser
}

// unpack read original filename from header
func (t *tarUnPacker) unpack(h *tar.Header) error {
	if h.FileInfo().IsDir() { // ignore directory
		return nil
	}
	filePath := path.Join(t.todir, h.Name)
	fileDir := path.Dir(filePath)
	if !t.createdDir.Has(fileDir) {
		if err := os.MkdirAll(fileDir, 0777); err != nil {
			return err
		}
		t.createdDir.Add(fileDir)
	}
	w, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(h.Mode))
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, t.r)
	return err
}

func (t *tarUnPacker) Unpack() error {
	if t.src == nil {
		var err error
		t.src, err = os.Open(t.filename)
		if err != nil {
			return err
		}
	}
	defer t.src.Close()

	if t.compress {
		gr, err := gzip.NewReader(t.src)
		if err != nil {
			return err
		}
		defer gr.Close()
		t.r = tar.NewReader(gr)
	} else {
		t.r = tar.NewReader(t.src)
	}

	t.createdDir = set.NewSet()
	for {
		h, err := t.r.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err = t.unpack(h); err != nil {
			return err
		}
	}
}

func NewTarUnPacker(filename, todir string, compressed bool) archive.Unpack {
	return &tarUnPacker{compress: compressed, filename: filename, todir: todir}
}

func NewTarUnPackerFromReader(src io.ReadCloser, todir string, compressed bool) archive.Unpack {
	return &tarUnPacker{compress: compressed, src: src, todir: todir}
}
