package gzip

import (
	"testing"
)

func TestGzipPacker(t *testing.T) {
	w := NewGzipPacker("/tmp/dd/1", "/tmp/dd/2")
	t.Log(w.Pack())
}

func TestGzipUnPacker(t *testing.T) {
	r := NewGzipUnPacker("/tmp/dd/1.gz", "/tmp/dd/2.gz")
	t.Log(r.Unpack())
}
