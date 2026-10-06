package app

import "strings"

// Helper SQL editor: buffer teks + kursor offset rune.

func (m *Model) editorInsert(t string) {
	if t == "" {
		return
	}
	r := []rune(m.editorText)
	if m.editorCursor < 0 {
		m.editorCursor = 0
	}
	if m.editorCursor > len(r) {
		m.editorCursor = len(r)
	}
	tr := []rune(t)
	out := make([]rune, 0, len(r)+len(tr))
	out = append(out, r[:m.editorCursor]...)
	out = append(out, tr...)
	out = append(out, r[m.editorCursor:]...)
	m.editorText = string(out)
	m.editorCursor += len(tr)
}

func (m *Model) editorBackspace() {
	r := []rune(m.editorText)
	if m.editorCursor <= 0 || m.editorCursor > len(r) {
		return
	}
	out := make([]rune, 0, len(r)-1)
	out = append(out, r[:m.editorCursor-1]...)
	out = append(out, r[m.editorCursor:]...)
	m.editorText = string(out)
	m.editorCursor--
}

func (m *Model) editorMoveLeft() {
	if m.editorCursor > 0 {
		m.editorCursor--
	}
}

func (m *Model) editorMoveRight() {
	if m.editorCursor < len([]rune(m.editorText)) {
		m.editorCursor++
	}
}

// editorLineCol posisi baris/kolom (0-based) dari kursor.
func editorLineCol(text string, cursor int) (line, col int) {
	r := []rune(text)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(r) {
		cursor = len(r)
	}
	for i := 0; i < cursor; i++ {
		if r[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return line, col
}

func editorCursorFromLineCol(text string, line, col int) int {
	lines := strings.Split(text, "\n")
	if line < 0 {
		line = 0
	}
	if line >= len(lines) {
		line = len(lines) - 1
	}
	w := len([]rune(lines[line]))
	if col < 0 {
		col = 0
	}
	if col > w {
		col = w
	}
	pos := 0
	for i := 0; i < line; i++ {
		pos += len([]rune(lines[i])) + 1 // + newline
	}
	return pos + col
}

func (m *Model) editorMoveUp() {
	line, col := editorLineCol(m.editorText, m.editorCursor)
	m.editorCursor = editorCursorFromLineCol(m.editorText, line-1, col)
}

func (m *Model) editorMoveDown() {
	line, col := editorLineCol(m.editorText, m.editorCursor)
	m.editorCursor = editorCursorFromLineCol(m.editorText, line+1, col)
}
