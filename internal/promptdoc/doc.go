// Package promptdoc renders the short Markdown prompts Kit prints: paragraphs
// and bullet lists.
package promptdoc

import "strings"

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

func (d *Document) Paragraph(text string) {
	d.Add(Paragraph{Text: text})
}

func (d *Document) BulletList(items ...string) {
	d.Add(BulletList{Items: items})
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

type Paragraph struct {
	Text string
}

func (p Paragraph) render(sb *strings.Builder) {
	sb.WriteString(strings.Trim(p.Text, "\n"))
}

type BulletList struct {
	Items []string
}

func (l BulletList) render(sb *strings.Builder) {
	for i, item := range l.Items {
		if i > 0 {
			sb.WriteString("\n")
		}
		renderListItem(sb, "- ", item)
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
