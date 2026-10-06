package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"db-tui/internal/db"
)

// Pesan async panel kanan.

type previewMsg struct {
	table string
	res   db.Result
}

type previewErrMsg struct {
	table string
	err   string
}

type queryDoneMsg struct{ res db.Result }

type queryErrMsg struct{ err string }

// toastMsg hapus toast; seq cegah clear basi menimpa toast baru.
type toastMsg struct{ seq int }

func clearToastCmd(seq int) tea.Cmd {
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return toastMsg{seq: seq}
	})
}

// notify set toast + cmd auto-hilang.
func (m *Model) notify(text string, isErr bool) tea.Cmd {
	m.toastSeq++
	m.toastText = text
	m.toastErr = isErr
	return clearToastCmd(m.toastSeq)
}

func quoteIdent(driver, name string) string {
	if driver == "mysql" || driver == "mariadb" {
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// gridCols jumlah kolom hasil yang tampil di panel kanan.
func (m Model) gridCols() int {
	if m.queryResult != nil {
		return len(m.queryResult.Columns)
	}
	if m.previewResult != nil {
		return len(m.previewResult.Columns)
	}
	return 0
}

func loadPreviewCmd(handle db.DB, driver, table string) tea.Cmd {
	return func() tea.Msg {
		if handle == nil {
			return previewErrMsg{table, "not connected"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		res, err := handle.Query(ctx, fmt.Sprintf(`SELECT * FROM %s LIMIT 100`, quoteIdent(driver, table)))
		if err != nil {
			return previewErrMsg{table, err.Error()}
		}
		return previewMsg{table, res}
	}
}

func runQueryCmd(handle db.DB, q string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(q) == "" {
			return queryErrMsg{"empty query"}
		}
		if handle == nil {
			return queryErrMsg{"not connected"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		res, err := handle.Query(ctx, q)
		if err != nil {
			return queryErrMsg{err.Error()}
		}
		return queryDoneMsg{res}
	}
}

// updateEditor tangani semua key saat fokus di SQL editor.
func (m Model) updateEditor(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if m.activeDB != nil {
			m.activeDB.Close()
			m.activeDB = nil
		}
		return m, tea.Quit
	case "esc", "tab":
		m.editorFocused = false
		return m, nil
	case "ctrl+r", "ctrl+enter":
		if m.activeDB == nil {
			return m, m.notify("not connected", true)
		}
		m.queryLoading = true
		m.queryResult = nil
		return m, runQueryCmd(m.activeDB, m.editorText)
	case "enter":
		m.editorInsert("\n")
		return m, nil
	case "backspace":
		m.editorBackspace()
		return m, nil
	case "left":
		m.editorMoveLeft()
		return m, nil
	case "right":
		m.editorMoveRight()
		return m, nil
	case "up":
		m.editorMoveUp()
		return m, nil
	case "down":
		m.editorMoveDown()
		return m, nil
	case "space":
		m.editorInsert(" ")
		return m, nil
	default:
		if t := msg.Key().Text; t != "" {
			m.editorInsert(t)
		}
		return m, nil
	}
}

// renderGrid gambar db.Result sebagai teks kolom.
// off = offset kolom untuk scroll horizontal.
func renderGrid(res db.Result, maxW, maxH, off int) string {
	if len(res.Columns) == 0 {
		return "(query OK, no results)"
	}
	const maxColW = 24
	widths := make([]int, len(res.Columns))
	for i, c := range res.Columns {
		widths[i] = len([]rune(c))
	}
	for _, row := range res.Rows {
		for i := range res.Columns {
			if i < len(row) && len([]rune(row[i])) > widths[i] {
				widths[i] = len([]rune(row[i]))
			}
		}
	}
	for i := range widths {
		if widths[i] < 3 {
			widths[i] = 3
		}
		if widths[i] > maxColW {
			widths[i] = maxColW
		}
	}
	// Kolom tampil mulai dari off selama muat.
	if off < 0 {
		off = 0
	}
	if off >= len(widths) {
		off = len(widths) - 1
	}
	used, end := 0, off
	for i := off; i < len(widths); i++ {
		if used+widths[i]+3 > maxW && i > off {
			break
		}
		used += widths[i] + 3
		end = i + 1
	}
	visWidths := widths[off:end]
	visColumns := res.Columns[off:end]

	fit := func(s string, w int) string {
		r := []rune(s)
		if len(r) > w {
			if w <= 1 {
				return string(r[:w])
			}
			return string(r[:w-1]) + "…"
		}
		return s + strings.Repeat(" ", w-len(r))
	}

	var b strings.Builder
	for i, c := range visColumns {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(fit(c, visWidths[i]))
	}
	b.WriteString("\n")
	total := 0
	for _, w := range visWidths {
		total += w + 3
	}
	b.WriteString(strings.Repeat("-", total-3) + "\n")

	lines := 2
	for _, row := range res.Rows {
		if lines >= maxH {
			b.WriteString(fmt.Sprintf("... (%d rows total)", len(res.Rows)))
			break
		}
		for i := range visColumns {
			if i > 0 {
				b.WriteString(" | ")
			}
			cell := ""
			if off+i < len(row) {
				cell = row[off+i]
			}
			b.WriteString(fit(cell, visWidths[i]))
		}
		b.WriteString("\n")
		lines++
	}
	if off > 0 || end < len(res.Columns) {
		b.WriteString(fmt.Sprintf("← %d cols left • %d cols right → (h/l scroll)", off, len(res.Columns)-end))
	}
	return b.String()
}
