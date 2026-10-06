package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Driver type: sqlite | postgres | mysql (MariaDB compatible).
type Connection struct {
	Name     string `yaml:"name"`
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	Database string `yaml:"database,omitempty"`
	FilePath string `yaml:"file_path,omitempty"` // khusus sqlite
}

type File struct {
	Connections []Connection `yaml:"connections"`
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "db-tui", "connections.yaml"), nil
}

func Load(path string) ([]Connection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return f.Connections, nil
}

// Save menulis satu koneksi ke path (append, atau ganti bila nama sama).
func Save(path string, conn Connection) error {
	var conns []Connection
	if existing, err := Load(path); err != nil {
		return err
	} else {
		conns = existing
	}
	replaced := false
	for i, c := range conns {
		if c.Name == conn.Name {
			conns[i] = conn
			replaced = true
			break
		}
	}
	if !replaced {
		conns = append(conns, conn)
	}
	data, err := yaml.Marshal(File{Connections: conns})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// AppendConnection simpan ke lokasi config default.
func AppendConnection(conn Connection) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	return Save(path, conn)
}
