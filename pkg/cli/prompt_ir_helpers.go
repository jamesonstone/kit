package cli

import (
	"github.com/jamesonstone/kit/v3/internal/promptdoc"
)

func renderPromptDocument(build func(*promptdoc.Document)) string {
	doc := promptdoc.New()
	build(doc)
	return doc.String()
}
