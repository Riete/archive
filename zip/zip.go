package zip

import (
	"archive/zip"
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

type zipPacker struct {
	w        *zip.Writer
	filename string
	cwd      string
	sources  []string
}

func (z *zipPacker) pack(src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	fInfo, err := f.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(fInfo)
	if err != nil {
		return err
	}
	header.Name = strings.TrimLeft(src, "/")
	w, err := z.w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func (z *zipPacker) Pack() error {
	if z.cwd != "" {
		if err := os.Chdir(z.cwd); err != nil {
			return err
		}
	}
	_ = os.Remove(z.filename)
	if _, err := os.Stat(path.Dir(z.filename)); errors.Is(err, fs.ErrNotExist) {
		if err = os.MkdirAll(path.Dir(z.filename), 0777); err != nil {
			return err
		}
	}

	sources, err := common.WalkSource(z.sources)
	if err != nil {
		return err
	}

	f, err := os.Create(z.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	z.w = zip.NewWriter(f)
	defer z.w.Close()

	for _, src := range sources {
		if err = z.pack(src); err != nil {
			return err
		}
	}
	return nil
}

// NewZipPacker cwd is working directory when packing, use this argument to exclude the leading directories
// cwd == "" (do not chdir), source = "/tmp/aa/1" -->  "tmp/aa/1" in zip file
// to remove leading directories
// cwd == "/tmp/aa", source = "1" --> "1" in zip file
// cwd == "/tmp/aa", source = "."  pack all items in /tmp/aa with no "tmp/aa" prefixed
func NewZipPacker(filename, cwd string, sources ...string) archive.Pack {
	return &zipPacker{filename: filename, cwd: cwd, sources: sources}
}

type zipUnPacker struct {
	r          *zip.ReadCloser
	filename   string
	todir      string
	createdDir set.Set
}

// unpack read original filename from header
func (z *zipUnPacker) unpack(f *zip.File) error {
	var fileDir string
	filePath := path.Join(z.todir, f.FileHeader.Name)
	if f.FileInfo().IsDir() {
		fileDir = filePath
	} else {
		fileDir = path.Dir(filePath)
	}
	if !z.createdDir.Has(fileDir) {
		if err := os.MkdirAll(fileDir, 0777); err != nil {
			return err
		}
		z.createdDir.Add(fileDir)
	}
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.FileHeader.Mode())
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	return err
}

func (z *zipUnPacker) Unpack() error {
	var err error
	z.r, err = zip.OpenReader(z.filename)
	if err != nil {
		return err
	}
	defer z.r.Close()

	z.createdDir = set.NewSet()
	for _, f := range z.r.File {
		if err = z.unpack(f); err != nil {
			return err
		}
	}
	return nil
}

func NewZipUnPacker(filename, todir string) archive.Unpack {
	return &zipUnPacker{filename: filename, todir: todir}
}
