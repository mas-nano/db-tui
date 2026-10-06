package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"db-tui/internal/config"
	"db-tui/internal/ui/styles"
)

func (m Model) View() tea.View {
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	header := styles.Title.Width(w).Render("db-tui • SQLite / MariaDB / PostgreSQL")
	footer := lipgloss.JoinVertical(lipgloss.Left,
		styles.StatusBar.Width(w).Render(m.status),
		styles.Help.Render("1 connections • tab editor • enter load table • ctrl+r run • j/k select • h/l scroll • q quit"),
	)
	toastH := 0
	toast := ""
	if m.toastText != "" {
		if m.toastErr {
			toast = styles.ToastError.Width(w).Render("! " + m.toastText)
		} else {
			toast = styles.ToastInfo.Width(w).Render(m.toastText)
		}
		toastH = lipgloss.Height(toast)
	}
	bodyHeight := h - lipgloss.Height(header) - toastH - lipgloss.Height(footer)
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	body := m.renderBrowserSplit(w, bodyHeight)
	if lipgloss.Height(body) < bodyHeight {
		body += strings.Repeat("\n", bodyHeight-lipgloss.Height(body))
	}

	base := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if toast != "" {
		base = lipgloss.JoinVertical(lipgloss.Left, header, toast, body, footer)
	}

	// Daftar koneksi hanya tampil sebagai modal di tengah.
	if m.showModal {
		var modal string
		if m.showForm {
			modal = m.renderFormModal()
		} else {
			modal = m.renderConnectionsModal()
		}
		overlay := lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, modal)
		v := tea.NewView(overlay)
		v.AltScreen = true
		return v
	}

	v := tea.NewView(base)
	v.AltScreen = true
	return v
}

// renderConnectionsModal: daftar koneksi di kotak tengah + detail yang disorot.
func (m Model) renderConnectionsModal() string {
	var b strings.Builder
	b.WriteString(styles.ModalTitle.Render("Select connection") + "\n\n")
	if len(m.connections) == 0 {
		b.WriteString("  (loading...)\n")
	}
	for i, c := range m.connections {
		line := fmt.Sprintf("%s [%s] %s", c.Name, c.Driver, describe(c))
		if i == m.cursor {
			b.WriteString(styles.Selected.Render("→ "+line) + "\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}
	b.WriteString("\n")
	if m.connecting {
		b.WriteString(styles.Dim.Render("  connecting...") + "\n")
	}
	if m.modalErr != "" {
		b.WriteString(styles.FormError.Render("  ! "+m.modalErr) + "\n")
	}
	b.WriteString(styles.Dim.Render("[enter] connect   [n] new   [↑/↓] select   [esc] close") + "\n")
	return styles.Modal.Render(b.String())
}

// renderFormModal: form tambah koneksi di kotak tengah.
func (m Model) renderFormModal() string {
	var b strings.Builder
	b.WriteString(styles.ModalTitle.Render("Add connection") + "\n\n")
	for i, f := range m.form {
		val := f.Value
		if f.Mask && val != "" {
			val = strings.Repeat("•", len([]rune(val)))
		}
		if val == "" {
			val = styles.Dim.Render("—")
		}
		if i == m.formIdx {
			b.WriteString(styles.Selected.Render(fmt.Sprintf("→ %-28s: %s█", f.Label, val)) + "\n")
		} else {
			b.WriteString(fmt.Sprintf("  %-28s: %s\n", f.Label, val))
		}
	}
	if m.formErr != "" {
		b.WriteString("\n" + styles.FormError.Render("! "+m.formErr) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(styles.Dim.Render("[tab/enter] next   [ctrl+s] save   [esc] cancel") + "\n")
	return styles.Modal.Render(b.String())
}

// renderBrowserSplit: kiri daftar tabel, kanan isi, bawah SQL editor.
func (m Model) renderBrowserSplit(w, bodyHeight int) string {
	editorH := 7
	if bodyHeight < 12 {
		editorH = 5
	}
	topH := bodyHeight - editorH
	if topH < 3 {
		topH = 3
		editorH = bodyHeight - topH
	}

	leftW := w / 3
	if leftW < 20 {
		leftW = 20
	}
	if leftW > 42 {
		leftW = 42
	}
	rightW := w - leftW
	if rightW < 10 {
		rightW = 10
	}

	leftStyle := styles.Pane
	if !m.editorFocused {
		leftStyle = styles.PaneFocused
	}
	rightStyle := styles.Pane

	left := leftStyle.Width(leftW).Height(topH).Render(m.renderTableList(leftW-2, topH-2))
	right := rightStyle.Width(rightW).Height(topH).Render(m.renderContent(rightW-2, topH-2))
	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	edStyle := styles.Pane
	if m.editorFocused {
		edStyle = styles.PaneFocused
	}
	editor := edStyle.Width(w).Height(editorH).Render(m.renderEditor(w-2, editorH-2))

	return lipgloss.JoinVertical(lipgloss.Left, top, editor)
}

// renderTableList isi panel kiri (innerW lebar dalam, innerH tinggi konten).
// Hanya tampilkan jendela tabel yang muat; kursor selalu terlihat (scroll vertikal).
func (m Model) renderTableList(innerW, innerH int) string {
	var b strings.Builder
	if m.activeConn == nil {
		b.WriteString(styles.PaneTitle.Render("Tables") + "\n")
		b.WriteString("Not connected.\nPress 1.\n")
		return b.String()
	}
	total := len(m.tables)
	b.WriteString(styles.PaneTitle.Render(fmt.Sprintf("Tables (%d)", total)) + "\n")
	if total == 0 {
		b.WriteString("(no tables)\n")
		return b.String()
	}
	visible := innerH - 2 // judul + baris indikator
	if visible < 1 {
		visible = 1
	}
	off := m.tableOff
	if off < 0 {
		off = 0
	}
	if off > total-1 {
		off = total - 1
	}
	if m.tableCursor < off {
		off = m.tableCursor
	}
	if m.tableCursor >= off+visible {
		off = m.tableCursor - visible + 1
	}
	if off < 0 {
		off = 0
	}
	end := off + visible
	if end > total {
		end = total
	}
	for i := off; i < end; i++ {
		name := m.tables[i]
		if r := []rune(name); len(r) > innerW-2 {
			name = string(r[:innerW-3]) + "…"
		}
		if i == m.tableCursor && !m.editorFocused {
			b.WriteString(styles.Selected.Render("→ "+name) + "\n")
		} else {
			b.WriteString("  " + name + "\n")
		}
	}
	if off > 0 || end < total {
		b.WriteString(styles.Dim.Render(fmt.Sprintf("↑ %d more • ↓ %d more (j/k)", off, total-end)) + "\n")
	}
	return b.String()
}

// renderContent isi panel kanan: hasil query / preview tabel.
func (m Model) renderContent(innerW, innerH int) string {
	var b strings.Builder
	switch {
	case m.activeConn == nil:
		return styles.PaneTitle.Render("Content") + "\nSelect a connection first (1).\n"
	case m.queryLoading:
		return styles.PaneTitle.Render("Query result") + "\nRunning...\n"
	case m.queryResult != nil:
		b.WriteString(styles.PaneTitle.Render(fmt.Sprintf("Query result (%d rows)", len(m.queryResult.Rows))) + "\n")
		b.WriteString(renderGrid(*m.queryResult, innerW, innerH-1, m.gridOff))
		return b.String()
	case m.previewLoading:
		return styles.PaneTitle.Render("Content: "+m.previewTable) + "\nLoading...\n"
	case m.previewResult != nil:
		b.WriteString(styles.PaneTitle.Render(fmt.Sprintf("Content: %s (%d rows)", m.previewTable, len(m.previewResult.Rows))) + "\n")
		b.WriteString(renderGrid(*m.previewResult, innerW, innerH-1, m.gridOff))
		return b.String()
	default:
		return styles.PaneTitle.Render("Content") + "\nSelect a table, then enter.\n"
	}
}

// renderEditor isi panel bawah (innerH baris untuk teks).
func (m Model) renderEditor(innerW, innerH int) string {
	var b strings.Builder
	focus := ""
	if m.editorFocused {
		focus = " ●focused"
	}
	b.WriteString(styles.PaneTitle.Render("SQL"+focus) + "\n")
	text := m.editorText
	if text == "" && !m.editorFocused {
		b.WriteString(styles.Dim.Render("tab to focus, type SQL, ctrl+r to run") + "\n")
		return b.String()
	}
	lines := strings.Split(text, "\n")
	curLine, _ := editorLineCol(text, m.editorCursor)
	// Scroll stik ke baris kursor.
	start := 0
	visible := innerH - 1
	if visible < 1 {
		visible = 1
	}
	if curLine >= start+visible {
		start = curLine - visible + 1
	}
	end := start + visible
	if end > len(lines) {
		end = len(lines)
	}
	for i := start; i < end; i++ {
		line := lines[i]
		if r := []rune(line); len(r) > innerW-1 {
			line = string(r[:innerW-1])
		}
		if m.editorFocused && i == curLine {
			_, col := editorLineCol(text, m.editorCursor)
			lr := []rune(line)
			if col > len(lr) {
				col = len(lr)
			}
			line = string(lr[:col]) + "█" + string(lr[col:])
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func describe(c config.Connection) string {
	if c.Driver == "sqlite" {
		return c.FilePath
	}
	return fmt.Sprintf("%s:%d/%s", c.Host, c.Port, c.Database)
}
