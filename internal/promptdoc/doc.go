package promptdoc

import (
	"fmt"
	"strings"
)

type Block interface {
	render(*strings.Builder)
}

type Document struct {
	blocks []Block
}

func New() *Document {
	return &Document{}
}

func (d *Document) Add(block Block) {
	if block == nil {
		return
	}
	d.blocks = append(d.blocks, block)
}

func (d *Document) Heading(level int, text string) {
	d.Add(Heading{Level: level, Text: text})
}

func (d *Document) Paragraph(text string) {
	d.Add(Paragraph{Text: text})
}

func (d *Document) Raw(text string) {
	d.Add(Raw{Text: text})
}

func (d *Document) Rawf(format string, args ...any) {
	d.Raw(fmt.Sprintf(format, args...))
}

func (d *Document) BulletList(items ...string) {
	d.Add(BulletList{Items: items})
}

func (d *Document) OrderedList(start int, items ...string) {
	d.Add(OrderedList{Start: start, Items: items})
}

func (d *Document) Table(headers []string, rows [][]string) {
	d.Add(Table{Headers: headers, Rows: rows})
}

func (d *Document) CodeBlock(language, content string) {
	d.Add(CodeBlock{Language: language, Content: content})
}

func (d *Document) String() string {
	var rendered []string
	for _, block := range d.blocks {
		var sb strings.Builder
		block.render(&sb)
		text := strings.Trim(sb.String(), "\n")
		if text == "" {
			continue
		}
		rendered = append(rendered, text)
	}

	return strings.Join(rendered, "\n\n")
}

type Heading struct {
	Level int
	Text  string
}

func (h Heading) render(sb *strings.Builder) {
	level := h.Level
	if level < 1 {
		level = 1
	}
	sb.WriteString(strings.Repeat("#", level))
	sb.WriteString(" ")
	sb.WriteString(strings.TrimSpace(h.Text))
}

type Paragraph struct {
	Text string
}

func (p Paragraph) render(sb *strings.Builder) {
	sb.WriteString(strings.Trim(p.Text, "\n"))
}

type Raw struct {
	Text string
}

func (r Raw) render(sb *strings.Builder) {
	sb.WriteString(strings.Trim(r.Text, "\n"))
}

type BulletList struct {
	Items []string
}

func (l BulletList) render(sb *strings.Builder) {
	renderList(sb, "- ", l.Items)
}

type OrderedList struct {
	Start int
	Items []string
}

func (l OrderedList) render(sb *strings.Builder) {
	start := l.Start
	if start <= 0 {
		start = 1
	}

	for i, item := range l.Items {
		if i > 0 {
			sb.WriteString("\n")
		}
		renderListItem(sb, fmt.Sprintf("%d. ", start+i), item)
	}
}

type Table struct {
	Headers []string
	Rows    [][]string
}

func (t Table) render(sb *strings.Builder) {
	if len(t.Headers) == 0 {
		return
	}

	sb.WriteString("| ")
	sb.WriteString(strings.Join(t.Headers, " | "))
	sb.WriteString(" |\n")

	separators := make([]string, len(t.Headers))
	for i, header := range t.Headers {
		width := len(strings.TrimSpace(header))
		if width < 3 {
			width = 3
		}
		separators[i] = strings.Repeat("-", width)
	}
	sb.WriteString("| ")
	sb.WriteString(strings.Join(separators, " | "))
	sb.WriteString(" |")

	for _, row := range t.Rows {
		sb.WriteString("\n| ")
		cells := make([]string, len(t.Headers))
		for i := range t.Headers {
			if i < len(row) {
				cells[i] = row[i]
			}
		}
		sb.WriteString(strings.Join(cells, " | "))
		sb.WriteString(" |")
	}
}

type CodeBlock struct {
	Language string
	Content  string
}

func (c CodeBlock) render(sb *strings.Builder) {
	sb.WriteString("```")
	sb.WriteString(strings.TrimSpace(c.Language))
	sb.WriteString("\n")
	sb.WriteString(strings.Trim(c.Content, "\n"))
	sb.WriteString("\n```")
}

func renderList(sb *strings.Builder, prefix string, items []string) {
	for i, item := range items {
		if i > 0 {
			sb.WriteString("\n")
		}
		renderListItem(sb, prefix, item)
	}
}

func renderListItem(sb *strings.Builder, prefix, item string) {
	lines := strings.Split(strings.Trim(item, "\n"), "\n")
	if len(lines) == 0 {
		return
	}

	sb.WriteString(prefix)
	sb.WriteString(lines[0])
	for _, line := range lines[1:] {
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat(" ", len(prefix)))
		sb.WriteString(line)
	}
}
