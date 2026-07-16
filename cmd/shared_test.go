package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDigits(t *testing.T) {
	cases := map[string]bool{
		"6790641087": true,
		"":           false,
		"1.0":        false,
		"com.x.y":    false,
		"42a":        false,
	}
	for in, want := range cases {
		if got := isDigits(in); got != want {
			t.Errorf("isDigits(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestValueOrFile(t *testing.T) {
	if got, err := valueOrFile("plain"); err != nil || got != "plain" {
		t.Fatalf("valueOrFile(plain) = %q, %v", got, err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "desc.txt")
	if err := os.WriteFile(path, []byte("from file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := valueOrFile("@" + path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "from file" {
		t.Fatalf("valueOrFile(@file) = %q, want %q (trailing newline trimmed)", got, "from file")
	}
}
