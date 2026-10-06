package app

import (
	"fmt"
	"strconv"
	"strings"

	"db-tui/internal/config"
)

// Satu baris input di form tambah koneksi.
type formField struct {
	Label string
	Value string
	Mask  bool // tampilkan sebagai ••• (password)
}

func defaultFormFields() []formField {
	return []formField{
		{Label: "Name"},
		{Label: "Driver (sqlite/postgres/mysql)", Value: "sqlite"},
		{Label: "Host", Value: "localhost"},
		{Label: "Port"},
		{Label: "User"},
		{Label: "Password", Mask: true},
		{Label: "Database"},
		{Label: "File path (sqlite)", Value: "./dev.db"},
	}
}

// formConnection validasi + rakit config.Connection dari form.
func (m Model) formConnection() (config.Connection, error) {
	val := func(i int) string { return strings.TrimSpace(m.form[i].Value) }

	name := val(0)
	if name == "" {
		return config.Connection{}, fmt.Errorf("name is required")
	}
	driver := strings.ToLower(val(1))
	switch driver {
	case "sqlite":
		if val(7) == "" {
			return config.Connection{}, fmt.Errorf("file path is required for sqlite")
		}
		return config.Connection{Name: name, Driver: "sqlite", FilePath: val(7)}, nil
	case "postgres", "postgresql", "pg":
		port := 5432
		if val(3) != "" {
			p, err := strconv.Atoi(val(3))
			if err != nil {
				return config.Connection{}, fmt.Errorf("port must be a number")
			}
			port = p
		}
		return config.Connection{
			Name: name, Driver: "postgres",
			Host: val(2), Port: port, User: val(4),
			Password: val(5), Database: val(6),
		}, nil
	case "mysql", "mariadb":
		port := 3306
		if val(3) != "" {
			p, err := strconv.Atoi(val(3))
			if err != nil {
				return config.Connection{}, fmt.Errorf("port must be a number")
			}
			port = p
		}
		return config.Connection{
			Name: name, Driver: "mysql",
			Host: val(2), Port: port, User: val(4),
			Password: val(5), Database: val(6),
		}, nil
	default:
		return config.Connection{}, fmt.Errorf("driver must be sqlite / postgres / mysql")
	}
}

// appendRune tambah karakter ke field aktif.
func (m *Model) formType(t string) {
	if t == "" || len(m.form) == 0 {
		return
	}
	m.form[m.formIdx].Value += t
}

// formBackspace hapus satu karakter terakhir (aman untuk UTF-8).
func (m *Model) formBackspace() {
	if len(m.form) == 0 {
		return
	}
	v := m.form[m.formIdx].Value
	r := []rune(v)
	if len(r) > 0 {
		m.form[m.formIdx].Value = string(r[:len(r)-1])
	}
}
