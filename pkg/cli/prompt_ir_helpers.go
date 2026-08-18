package cli

import (
	"strings"

	"github.com/jamesonstone/kit/v3/internal/promptdoc"
)

func renderPromptDocument(build func(*promptdoc.Document)) string {
	doc := promptdoc.New()
	build(doc)
	return doc.String()
}

func renderBuilderText(build func(*strings.Builder)) string {
	var sb strings.Builder
	build(&sb)
	return strings.Trim(sb.String(), "\n")
}
