package app

import (
	tea "charm.land/bubbletea/v2"
	"db-tui/internal/config"
	"db-tui/internal/db"
)

// Screen aktif.
type Screen int

const (
	ScreenConnections Screen = iota
	ScreenBrowser
)

type Model struct {
	screen      Screen
	width       int
	height      int
	connections []config.Connection
	cursor      int
	status      string
	showModal   bool
	// Form tambah koneksi (dibuka dengan "n" dari modal list).
	showForm bool
	form     []formField
	formIdx  int
	formErr  string
	// Koneksi aktif.
	activeDB    db.DB
	activeConn  *config.Connection
	tables      []string
	tableCursor int
	connecting  bool
	modalErr    string
	// Panel kanan: preview tabel / hasil query.
	previewTable   string
	previewResult  *db.Result
	previewLoading bool
	queryResult    *db.Result
	queryLoading   bool
	// Toast notifikasi (auto-hilang).
	toastText string
	toastErr  bool
	toastSeq  int
	// SQL editor bawah.
	editorText    string
	editorCursor  int // offset rune
	editorFocused bool
	// Scroll horizontal grid (offset kolom) dan vertikal daftar tabel.
	gridOff  int
	tableOff int
}

func New() Model {
	return Model{
		screen: ScreenBrowser,
		width:  80,
		height: 24,
		// Modal koneksi langsung terbuka saat start.
		showModal: true,
		status:    "enter connect • n new • esc close • 1 connections • 2 browser • q quit",
	}
}

func (m Model) Init() tea.Cmd { return LoadConnections() }
