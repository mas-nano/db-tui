package app

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"db-tui/internal/config"
	"db-tui/internal/db"
)

func seedTempDB(t *testing.T) (config.Connection, func()) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn := config.Connection{Name: "tes", Driver: "sqlite", FilePath: path}
	handle, err := db.Open(conn)
	if err != nil {
		t.Fatalf("seed open: %v", err)
	}
	defer handle.Close()
	ctx := context.Background()
	for _, q := range []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`,
		`INSERT INTO users (name) VALUES ('Andi'), ('Budi')`,
	} {
		if _, err := handle.Query(ctx, q); err != nil {
			t.Fatalf("seed query %q: %v", q, err)
		}
	}
	return conn, func() {}
}

func TestEditorInsertBackspaceMove(t *testing.T) {
	var m Model
	m.editorInsert("ab")
	m.editorInsert("c")
	if m.editorText != "abc" || m.editorCursor != 3 {
		t.Fatalf("insert salah: %q cursor=%d", m.editorText, m.editorCursor)
	}
	m.editorMoveLeft()
	m.editorMoveLeft()
	m.editorInsert("X")
	if m.editorText != "aXbc" {
		t.Fatalf("insert tengah salah: %q", m.editorText)
	}
	m.editorBackspace()
	if m.editorText != "abc" {
		t.Fatalf("backspace salah: %q", m.editorText)
	}
	m.editorMoveUp() // satu baris: tetap
	m.editorInsert("\nde")
	m.editorMoveUp()
	line, _ := editorLineCol(m.editorText, m.editorCursor)
	if line != 0 {
		t.Fatalf("move up harus ke baris 0, dapat %d (%q)", line, m.editorText)
	}
	m.editorMoveDown()
	line, _ = editorLineCol(m.editorText, m.editorCursor)
	if line != 1 {
		t.Fatalf("move down harus ke baris 1, dapat %d", line)
	}
}

func TestLoadPreviewCmd(t *testing.T) {
	conn, _ := seedTempDB(t)
	handle, err := db.Open(conn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer handle.Close()

	msg := loadPreviewCmd(handle, "sqlite", "users")()
	pm, ok := msg.(previewMsg)
	if !ok {
		t.Fatalf("harus previewMsg, dapat %T (%v)", msg, msg)
	}
	if len(pm.res.Rows) != 2 || len(pm.res.Columns) != 2 {
		t.Fatalf("preview salah: %+v", pm.res)
	}
}

func TestRunQueryCmd(t *testing.T) {
	conn, _ := seedTempDB(t)
	handle, err := db.Open(conn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer handle.Close()

	msg := runQueryCmd(handle, `SELECT name FROM users ORDER BY name`)()
	qm, ok := msg.(queryDoneMsg)
	if !ok {
		t.Fatalf("harus queryDoneMsg, dapat %T (%v)", msg, msg)
	}
	if len(qm.res.Rows) != 2 || qm.res.Rows[0][0] != "Andi" {
		t.Fatalf("hasil salah: %+v", qm.res)
	}

	if msg := runQueryCmd(handle, `SELECT * FROM tidak_ada`)(); isQueryErr(msg) == false {
		t.Fatalf("SQL salah harus queryErrMsg")
	}
	if msg := runQueryCmd(handle, `   `)(); isQueryErr(msg) == false {
		t.Fatalf("query kosong harus queryErrMsg")
	}
}

func isQueryErr(msg interface{}) bool {
	_, ok := msg.(queryErrMsg)
	return ok
}

func TestRenderGrid(t *testing.T) {
	res := db.Result{
		Columns: []string{"id", "name"},
		Rows:    [][]string{{"1", "Andi"}, {"2", "Budi"}},
	}
	out := renderGrid(res, 60, 10, 0)
	for _, want := range []string{"id", "name", "Andi", "Budi"} {
		if !strings.Contains(out, want) {
			t.Fatalf("grid harus memuat %q:\n%s", want, out)
		}
	}
	// Sel panjang dipotong.
	res.Rows[0][1] = strings.Repeat("x", 100)
	out = renderGrid(res, 60, 10, 0)
	if strings.Contains(out, strings.Repeat("x", 100)) {
		t.Fatalf("sel panjang harus dipotong")
	}
	if !strings.Contains(out, "…") {
		t.Fatalf("pemotongan harus ditandai …")
	}
}

func TestRenderGridScroll(t *testing.T) {
	cols := []string{"c1", "c2", "c3", "c4", "c5", "c6"}
	res := db.Result{Columns: cols, Rows: [][]string{{"1", "2", "3", "4", "5", "6"}}}
	// Layar sempit: tidak semua kolom muat.
	off0 := renderGrid(res, 25, 10, 0)
	if !strings.Contains(off0, "c1") || strings.Contains(off0, "c6") {
		t.Fatalf("offset 0 harus mulai dari c1 dan potong kanan:\n%s", off0)
	}
	if !strings.Contains(off0, "cols right") {
		t.Fatalf("harus ada indikator kolom kanan:\n%s", off0)
	}
	// Geser 3 kolom.
	off3 := renderGrid(res, 25, 10, 3)
	if strings.Contains(off3, "c1") || !strings.Contains(off3, "c4") {
		t.Fatalf("offset 3 harus mulai dari c4:\n%s", off3)
	}
	if !strings.Contains(off3, "cols left") {
		t.Fatalf("harus ada indikator kolom kiri:\n%s", off3)
	}
	// Sel baris ikut offset (kolom c4 = "4").
	found := false
	for _, line := range strings.Split(off3, "\n") {
		if strings.Contains(line, "4") && !strings.Contains(line, "c1") {
			found = true
		}
	}
	if !found {
		t.Fatalf("sel harus ikut offset:\n%s", off3)
	}
}

func TestRightPaneShowsQueryResult(t *testing.T) {
	m := testModel()
	m.showModal = false
	conn := config.Connection{Name: "COba", Driver: "sqlite"}
	m.activeConn = &conn
	m.tables = []string{"users"}
	m.queryResult = &db.Result{Columns: []string{"name"}, Rows: [][]string{{"Andi"}}}
	v := m.View().Content
	if !strings.Contains(v, "Query result") || !strings.Contains(v, "Andi") {
		t.Fatalf("panel kanan harus tampilkan hasil query")
	}
}

func TestQueryErrorToast(t *testing.T) {
	m := testModel()
	m.showModal = false
	nm, cmd := m.Update(queryErrMsg{err: "syntax error near FRO"})
	um := nm.(Model)
	if um.toastText == "" || !um.toastErr {
		t.Fatalf("error query harus jadi toast: %+v", um)
	}
	if cmd == nil {
		t.Fatalf("toast harus bawa cmd auto-clear")
	}
	if !strings.Contains(um.View().Content, "syntax error near FRO") {
		t.Fatalf("toast harus tampil di view")
	}
	// Seq cocok: toast hilang.
	nm2, _ := um.Update(toastMsg{seq: um.toastSeq})
	if nm2.(Model).toastText != "" {
		t.Fatalf("toast harus hilang setelah timeout")
	}
	// Seq basi: diabaikan.
	m2 := testModel()
	m2.toastText = "baru"
	m2.toastSeq = 5
	nm3, _ := m2.Update(toastMsg{seq: 4})
	if nm3.(Model).toastText != "baru" {
		t.Fatalf("clear basi tidak boleh hapus toast baru")
	}
}

func TestPreviewErrorKeepsContent(t *testing.T) {
	m := testModel()
	m.previewTable = "users"
	m.previewResult = &db.Result{Columns: []string{"id"}, Rows: [][]string{{"1"}}}
	nm, _ := m.Update(previewErrMsg{table: "users", err: "no such table"})
	um := nm.(Model)
	if um.toastText != "no such table" {
		t.Fatalf("preview error harus jadi toast")
	}
	if um.previewResult == nil {
		t.Fatalf("isi lama panel kanan harus dipertahankan")
	}
}

func manyTables(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "tabel_" + itoa2(i)
	}
	return out
}

func itoa2(i int) string {
	if i < 10 {
		return "0" + string(rune('0'+i))
	}
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}

func TestTableListScroll(t *testing.T) {
	m := testModel()
	conn := config.Connection{Name: "COba", Driver: "sqlite"}
	m.activeConn = &conn
	m.tables = manyTables(50)

	// Kursor atas: tampilkan awal, potong akhir + indikator bawah.
	m.tableCursor = 0
	top := m.renderTableList(30, 12)
	if !strings.Contains(top, "tabel_00") || strings.Contains(top, "tabel_49") {
		t.Fatalf("jendela atas salah:\n%s", top)
	}
	if !strings.Contains(top, "↓") {
		t.Fatalf("harus ada indikator bawah:\n%s", top)
	}

	// Kursor bawah: tampilkan akhir, potong awal + indikator atas.
	m.tableCursor = 49
	bottom := m.renderTableList(30, 12)
	if strings.Contains(bottom, "tabel_00") || !strings.Contains(bottom, "tabel_49") {
		t.Fatalf("jendela bawah salah:\n%s", bottom)
	}
	if !strings.Contains(bottom, "↑") {
		t.Fatalf("harus ada indikator atas:\n%s", bottom)
	}

	// Kursor tengah: kursor terlihat di jendela.
	m.tableCursor = 25
	mid := m.renderTableList(30, 12)
	if !strings.Contains(mid, "tabel_25") {
		t.Fatalf("kursor harus terlihat:\n%s", mid)
	}
}
