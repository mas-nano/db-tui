package config

import (
	"path/filepath"
	"testing"
)

func TestSaveRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "connections.yaml")

	if err := Save(path, Connection{Name: "a", Driver: "sqlite", FilePath: "./a.db"}); err != nil {
		t.Fatalf("save pertama: %v", err)
	}
	if err := Save(path, Connection{Name: "b", Driver: "mysql", Host: "h", Port: 3306, User: "u", Database: "d"}); err != nil {
		t.Fatalf("save kedua: %v", err)
	}

	conns, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(conns) != 2 || conns[0].Name != "a" || conns[1].Name != "b" {
		t.Fatalf("isi salah: %+v", conns)
	}

	// Nama sama -> ganti, bukan append.
	if err := Save(path, Connection{Name: "a", Driver: "sqlite", FilePath: "./a2.db"}); err != nil {
		t.Fatalf("save replace: %v", err)
	}
	conns, _ = Load(path)
	if len(conns) != 2 || conns[0].FilePath != "./a2.db" {
		t.Fatalf("replace gagal: %+v", conns)
	}
}
