package brigade_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/fujiwara/brigade"
)

func TestParseEOF(t *testing.T) {
	for _, line := range []string{"--", "----", "-- "} {
		_, err := brigade.ParseLine(line)
		if err != io.EOF {
			t.Fatal("unexpected err", err)
		}
	}
}

func TestParseLine(t *testing.T) {
	d, err := brigade.ParseLine("foo -> bar")
	if err != nil {
		t.Fatal(err)
	}
	if d.Src() != "foo" {
		t.Fatal("unexpected Src", d.Src())
	}
	if d.Dest() != "bar" {
		t.Fatal("unexpected Dest", d.Dest())
	}
	if d.String() != "foo->bar" {
		t.Fatal("unexpected String", d.String())
	}
	t.Log(d.String())

	d2, err := brigade.ParseLine(d.String())
	if err != nil {
		t.Fatal(err)
	}
	if d.Src() != d2.Src() || d.Dest() != d2.Dest() {
		t.Fatal("unexpected", d.String(), d2.String())
	}
}

func TestParseSrc(t *testing.T) {
	src := strings.NewReader("foo -> bar\nbar->baz\n--\ndata")
	ds, r, err := brigade.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if ds.String() != "foo->bar\nbar->baz\n--\n" {
		t.Error("unexpected ds", ds.String())
	}
	b := new(bytes.Buffer)
	n, err := io.Copy(b, r)
	if err != nil {
		t.Fatal("copy failed", err)
	}
	if n != 4 {
		t.Error("unexpected length", n)
	}
	if b.String() != "data" {
		t.Error("unexpecetd wrote", b.String())
	}
}
