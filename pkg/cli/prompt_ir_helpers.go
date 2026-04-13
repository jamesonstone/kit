package cli

import (
	"strings"

	"github.com/jamesonstone/kit/internal/promptdoc"
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

func renderNonEmptySectionRules(documentName string) string {
	return renderBuilderText(func(sb *strings.Builder) {
		appendNonEmptySectionRules(sb, documentName)
	})
}

func renderNumberedSteps(startStep int, steps []string) string {
	return renderBuilderText(func(sb *strings.Builder) {
		appendNumberedSteps(sb, startStep, steps)
	})
}
