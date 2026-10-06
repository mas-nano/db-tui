package app

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"db-tui/internal/config"
	"db-tui/internal/db"
)

func testModel() Model {
	m := New()
	m.width = 100
	m.height = 30
	m.connections = []config.Connection{
		{Name: "local-sqlite (./dev.db)", Driver: "sqlite", FilePath: "./dev.db"},
		{Name: "local-postgres", Driver: "postgres", Host: "localhost", Port: 5432, User: "postgres", Database: "postgres"},
	}
	return m
}

func TestModalOpenByDefault(t *testing.T) {
	m := New()
	if !m.showModal {
		t.Fatalf("modal koneksi harus terbuka saat start")
	}
}

func TestConnectionsModalLists(t *testing.T) {
	m := testModel() // showModal=true dari New()
	v := m.View()
	if !v.AltScreen {
		t.Fatalf("view harus pakai AltScreen supaya baris shell tidak tersisa")
	}
	c := v.Content
	if !strings.Contains(c, "Select connection") {
		t.Fatalf("modal must show 'Select connection' title")
	}
	if !strings.Contains(c, "local-sqlite") {
		t.Fatalf("modal must list connections")
	}
	if !strings.Contains(c, "[enter] connect") {
		t.Fatalf("modal must show connect action")
	}
}

func TestBrowserBackgroundWhenModalClosed(t *testing.T) {
	m := testModel()
	m.showModal = false
	v := m.View().Content
	if strings.Contains(v, "[enter] connect") {
		t.Fatalf("modal tidak boleh tampil saat showModal=false")
	}
	if !strings.Contains(v, "Tables") {
		t.Fatalf("background harus browser saat modal tutup")
	}
}

func TestListHintMentionsNew(t *testing.T) {
	m := testModel()
	if !strings.Contains(m.View().Content, "[n] new") {
		t.Fatalf("modal list harus sebut tombol n")
	}
}

func TestFormRenders(t *testing.T) {
	m := testModel()
	m.showForm = true
	m.form = defaultFormFields()
	v := m.View()
	if !v.AltScreen {
		t.Fatalf("form harus tetap pakai AltScreen")
	}
	if !strings.Contains(v.Content, "Add connection") {
		t.Fatalf("form harus memuat judul 'Add connection'")
	}
	if !strings.Contains(v.Content, "[ctrl+s] save") {
		t.Fatalf("form harus memuat hint simpan")
	}
}

func TestFormValidation(t *testing.T) {
	m := testModel()
	m.form = defaultFormFields()

	// Nama kosong -> error.
	if _, err := m.formConnection(); err == nil {
		t.Fatalf("nama kosong harus error")
	}

	// SQLite valid.
	m.form[0].Value = "tes"
	m.form[7].Value = "./tes.db"
	conn, err := m.formConnection()
	if err != nil {
		t.Fatalf("sqlite valid harus lolos: %v", err)
	}
	if conn.Driver != "sqlite" || conn.FilePath != "./tes.db" {
		t.Fatalf("hasil sqlite salah: %+v", conn)
	}

	// Driver asing -> error.
	m.form[1].Value = "oracle"
	if _, err := m.formConnection(); err == nil {
		t.Fatalf("driver asing harus error")
	}

	// Port bukan angka -> error.
	m.form[1].Value = "postgres"
	m.form[3].Value = "bukan-angka"
	if _, err := m.formConnection(); err == nil {
		t.Fatalf("port bukan angka harus error")
	}
}

func TestConnectCmdSQLite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	seed, err := db.Open(config.Connection{Driver: "sqlite", FilePath: path})
	if err != nil {
		t.Fatalf("seed open: %v", err)
	}
	ctx := context.Background()
	for _, q := range []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`,
		`CREATE TABLE orders (id INTEGER PRIMARY KEY)`,
	} {
		if _, err := seed.Query(ctx, q); err != nil {
			seed.Close()
			t.Fatalf("seed query: %v", err)
		}
	}
	seed.Close()

	msg := connectCmd(config.Connection{Name: "tes", Driver: "sqlite", FilePath: path})()
	ok, isOK := msg.(connectedMsg)
	if !isOK {
		t.Fatalf("harus connectedMsg, dapat %T (%v)", msg, msg)
	}
	defer ok.handle.Close()
	if len(ok.tables) != 2 || ok.tables[0] != "orders" || ok.tables[1] != "users" {
		t.Fatalf("tabel salah: %+v", ok.tables)
	}
}

func TestConnectCmdError(t *testing.T) {
	msg := connectCmd(config.Connection{Name: "x", Driver: "ngawur"})()
	if _, isOK := msg.(connectedMsg); isOK {
		t.Fatalf("driver ngawur harus gagal")
	}
	if _, isErr := msg.(connectErrMsg); !isErr {
		t.Fatalf("harus connectErrMsg, dapat %T", msg)
	}
}

func TestBrowserShowsTables(t *testing.T) {
	m := testModel()
	m.showModal = false
	conn := config.Connection{Name: "COba", Driver: "sqlite", FilePath: "./dev.db"}
	m.activeConn = &conn
	m.tables = []string{"orders", "products", "users"}
	v := m.View().Content
	if strings.Contains(v, "connected (dummy)") {
		t.Fatalf("tidak boleh ada tulisan dummy lagi")
	}
	// Layout split: panel Tables + Content + SQL.
	for _, want := range []string{"Tables", "Content", "SQL", "orders", "products", "users"} {
		if !strings.Contains(v, want) {
			t.Fatalf("browser split harus memuat %q", want)
		}
	}
}

func TestModalShowsConnectError(t *testing.T) {
	m := testModel()
	m.modalErr = "ping: connection refused"
	if !strings.Contains(m.View().Content, "connection refused") {
		t.Fatalf("modal harus tampilkan error koneksi")
	}
}
