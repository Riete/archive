package zip

import (
	"testing"
)

func TestZipPacker(t *testing.T) {
	w := NewZipPacker("/tmp/dd/1.zip", "/tmp", "b", "a", "33")
	t.Log(w.Pack())
}

func TestZipUnPacker(t *testing.T) {
	r := NewZipUnPacker("/tmp/dd/1.zip", "/tmp/dd")
	t.Log(r.Unpack())
}
