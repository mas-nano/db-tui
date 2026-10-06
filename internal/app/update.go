package app

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"db-tui/internal/config"
	"db-tui/internal/db"
)

type connectionsLoadedMsg []config.Connection

// connectedMsg hasil koneksi beneran (Open + Ping + Tables sukses).
type connectedMsg struct {
	conn   config.Connection
	handle db.DB
	tables []string
}

type connectErrMsg struct{ err string }

func connectCmd(conn config.Connection) tea.Cmd {
	return func() tea.Msg {
		handle, err := db.Open(conn)
		if err != nil {
			return connectErrMsg{"open: " + err.Error()}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := handle.Ping(ctx); err != nil {
			handle.Close()
			return connectErrMsg{"ping: " + err.Error()}
		}
		tables, err := handle.Tables(ctx)
		if err != nil {
			handle.Close()
			return connectErrMsg{"tables: " + err.Error()}
		}
		return connectedMsg{conn: conn, handle: handle, tables: tables}
	}
}

func LoadConnections() tea.Cmd {
	return func() tea.Msg {
		path, err := config.DefaultPath()
		if err != nil {
			return connectionsLoadedMsg(nil)
		}
		conns, _ := config.Load(path)
		if len(conns) == 0 {
			// fallback contoh supaya TUI tidak kosong saat pertama jalan
			conns = []config.Connection{
				{Name: "local-sqlite (./dev.db)", Driver: "sqlite", FilePath: "./dev.db"},
				{Name: "local-postgres", Driver: "postgres", Host: "localhost", Port: 5432, User: "postgres", Database: "postgres"},
				{Name: "local-mariadb", Driver: "mysql", Host: "127.0.0.1", Port: 3306, User: "root", Database: "test"},
			}
		}
		return connectionsLoadedMsg(conns)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case connectionsLoadedMsg:
		m.connections = []config.Connection(msg)
		if m.cursor >= len(m.connections) {
			m.cursor = max(0, len(m.connections)-1)
		}
		return m, nil

	case connectedMsg:
		if m.activeDB != nil {
			m.activeDB.Close()
		}
		m.activeDB = msg.handle
		cp := msg.conn
		m.activeConn = &cp
		m.tables = msg.tables
		m.tableCursor = 0
		m.tableOff = 0
		m.gridOff = 0
		m.connecting = false
		m.modalErr = ""
		m.showModal = false
		m.screen = ScreenBrowser
		m.status = "connected: " + cp.Name + " • 1 connections • 2 browser • q quit"
		return m, nil

	case connectErrMsg:
		m.connecting = false
		m.modalErr = msg.err
		return m, nil

	case previewMsg:
		if msg.table != m.previewTable {
			return m, nil // basi: user sudah pindah tabel
		}
		m.previewLoading = false
		r := msg.res
		m.previewResult = &r
		m.gridOff = 0
		m.status = "table: " + msg.table + fmt.Sprintf(" (%d rows) • 1 connections • 2 browser • q quit", len(r.Rows))
		return m, nil

	case previewErrMsg:
		if msg.table != m.previewTable {
			return m, nil
		}
		m.previewLoading = false
		return m, m.notify(msg.err, true)

	case queryDoneMsg:
		m.queryLoading = false
		r := msg.res
		m.queryResult = &r
		m.gridOff = 0
		m.status = fmt.Sprintf("query: %d rows • 1 connections • 2 browser • q quit", len(r.Rows))
		return m, nil

	case queryErrMsg:
		m.queryLoading = false
		return m, m.notify(msg.err, true)

	case toastMsg:
		if msg.seq == m.toastSeq {
			m.toastText = ""
		}
		return m, nil

	case tea.KeyPressMsg:
		// Form tambah koneksi menangkap semua ketikan.
		if m.showForm {
			return m.updateForm(msg)
		}
		// SQL editor fokus menangkap semua ketikan.
		if m.editorFocused && !m.showModal && m.screen == ScreenBrowser {
			return m.updateEditor(msg)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			if m.activeDB != nil {
				m.activeDB.Close()
				m.activeDB = nil
			}
			return m, tea.Quit
		case "esc":
			m.showModal = false
			return m, nil
		case "enter":
			// Di modal: konek beneran (async).
			if m.showModal {
				if len(m.connections) == 0 || m.connecting {
					return m, nil
				}
				m.connecting = true
				m.modalErr = ""
				return m, connectCmd(m.connections[m.cursor])
			}
			// Di browser yang sudah connect: muat isi tabel.
			if m.screen == ScreenBrowser && m.activeDB != nil && len(m.tables) > 0 {
				t := m.tables[m.tableCursor]
				m.previewTable = t
				m.previewLoading = true
				m.previewResult = nil
				m.queryResult = nil
				driver := ""
				if m.activeConn != nil {
					driver = m.activeConn.Driver
				}
				return m, loadPreviewCmd(m.activeDB, driver, t)
			}
			m.showModal = true
			return m, nil
		case "c":
			// Ganti koneksi: selalu buka modal.
			m.showModal = true
			return m, nil
		case "tab":
			// Fokus ke SQL editor (browser + sudah connect).
			if !m.showModal && m.screen == ScreenBrowser && m.activeDB != nil {
				m.editorFocused = true
				return m, nil
			}
			return m, nil
		case "h", "left":
			// Geser grid ke kiri (tampilkan kolom awal).
			if !m.showModal && m.screen == ScreenBrowser && m.gridOff > 0 {
				m.gridOff--
			}
			return m, nil
		case "l", "right":
			// Geser grid ke kanan.
			if !m.showModal && m.screen == ScreenBrowser && m.gridOff < m.gridCols()-1 {
				m.gridOff++
			}
			return m, nil
		case "1":
			// Daftar koneksi hanya ada sebagai modal.
			m.showModal = true
			return m, nil
		case "n", "N":
			// Buka form tambah koneksi dari modal list.
			if m.showModal && !m.connecting {
				m.showForm = true
				m.formIdx = 0
				m.formErr = ""
				if len(m.form) == 0 {
					m.form = defaultFormFields()
				}
				return m, nil
			}
			return m, nil
		case "2":
			m.showModal = false
			m.screen = ScreenBrowser
			return m, nil
		case "up", "k":
			if m.showModal {
				if m.connecting {
					return m, nil
				}
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			}
			// Di browser: gerakkan kursor daftar tabel.
			if m.screen == ScreenBrowser && m.tableCursor > 0 {
				m.tableCursor--
			}
			return m, nil
		case "down", "j":
			if m.showModal {
				if m.connecting {
					return m, nil
				}
				if m.cursor < len(m.connections)-1 {
					m.cursor++
				}
				return m, nil
			}
			if m.screen == ScreenBrowser && m.tableCursor < len(m.tables)-1 {
				m.tableCursor++
			}
			return m, nil
		}
	}
	return m, nil
}

// updateForm tangani semua key saat form tambah koneksi terbuka.
func (m Model) updateForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if len(m.form) == 0 {
		m.form = defaultFormFields()
	}
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.showForm = false
		m.formErr = ""
		return m, nil
	case "tab", "enter", "down":
		m.formIdx = (m.formIdx + 1) % len(m.form)
		return m, nil
	case "shift+tab", "up":
		m.formIdx = (m.formIdx - 1 + len(m.form)) % len(m.form)
		return m, nil
	case "backspace":
		m.formBackspace()
		return m, nil
	case "space":
		m.formType(" ")
		return m, nil
	case "ctrl+s":
		conn, err := m.formConnection()
		if err != nil {
			m.formErr = err.Error()
			return m, nil
		}
		if err := config.AppendConnection(conn); err != nil {
			m.formErr = err.Error()
			return m, nil
		}
		m.connections = append(m.connections, conn)
		m.cursor = len(m.connections) - 1
		m.showForm = false
		m.form = nil
		m.formIdx = 0
		m.status = "saved: " + conn.Name + " • 1 connections • 2 browser • q quit"
		return m, nil
	default:
		// Ketikan biasa (termasuk q/n) masuk ke field, bukan shortcut.
		if t := msg.Key().Text; t != "" {
			m.formType(t)
		}
		return m, nil
	}
}
