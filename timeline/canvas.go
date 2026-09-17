package timeline

import "strings"

// cell is a single character on the canvas plus its style spec.
type cell struct {
	r    rune
	spec string
}

// canvas is a simple grid of styled cells that grows on demand.
type canvas struct {
	rows [][]cell
}

func newCanvas() *canvas { return &canvas{} }

// grow ensures a cell at (x, y) exists, padding with blanks as needed.
func (c *canvas) grow(x, y int) {
	for len(c.rows) <= y {
		c.rows = append(c.rows, nil)
	}
	row := c.rows[y]
	for len(row) <= x {
		row = append(row, cell{r: ' '})
	}
	c.rows[y] = row
}

// set writes a single styled rune at (x, y). Out-of-range coordinates are
// ignored so callers don't have to bounds-check every write.
func (c *canvas) set(x, y int, r rune, spec string) {
	if x < 0 || y < 0 {
		return
	}
	c.grow(x, y)
	c.rows[y][x] = cell{r: r, spec: spec}
}

// puts writes a string starting at (x, y) and returns the column just past
// the last rune written. Each rune occupies a single column.
func (c *canvas) puts(x, y int, s, spec string) int {
	col := x
	for _, r := range s {
		c.set(col, y, r, spec)
		col++
	}
	return col
}

// render serializes the canvas to a string. Trailing blanks on each row are
// trimmed. When noColor is true all style information is dropped.
func (c *canvas) render(noColor bool) string {
	var b strings.Builder
	for _, row := range c.rows {
		last := -1
		for i, cl := range row {
			if cl.r != ' ' {
				last = i
			}
		}
		cur := ""
		for i := 0; i <= last; i++ {
			cl := row[i]
			spec := cl.spec
			if cl.r == ' ' {
				// Never style blanks: avoids stray background runs and keeps
				// the escape sequences minimal.
				spec = ""
			}
			if !noColor && spec != cur {
				if cur != "" {
					b.WriteString(reset)
					cur = ""
				}
				if open := ansiOpen(spec); open != "" {
					b.WriteString(open)
					cur = spec
				}
			}
			b.WriteRune(cl.r)
		}
		if !noColor && cur != "" {
			b.WriteString(reset)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
