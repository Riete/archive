package tar

import (
	"os"
	"testing"
)

func TestTarPacker(t *testing.T) {
	w := NewTarPacker("/tmp/cc/a.tar", "/tmp/cc", false, ".", "/tmp/aa")
	t.Log(w.Pack())
}

func TestTarUnPacker(t *testing.T) {
	r := NewTarUnPacker("/tmp/dd/a.tar", "/tmp/dd", false)
	t.Log(r.Unpack())
}

func TestNewTarUnPackerFromReader(t *testing.T) {
	f, _ := os.Open("/tmp/dd/a.tar")
	r := NewTarUnPackerFromReader(f, "/tmp/dd", false)
	t.Log(r.Unpack())
}
