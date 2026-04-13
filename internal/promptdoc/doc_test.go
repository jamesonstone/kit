package promptdoc

import "testing"

func TestDocumentRendersStructuredBlocks(t *testing.T) {
	doc := New()
	doc.Heading(2, "Section")
	doc.Paragraph("Intro text")
	doc.BulletList("first", "second")
	doc.OrderedList(3, "step one", "step two")
	doc.Table([]string{"A", "B"}, [][]string{{"1", "2"}})
	doc.CodeBlock("text", "hello")

	got := doc.String()
	want := "## Section\n\n" +
		"Intro text\n\n" +
		"- first\n" +
		"- second\n\n" +
		"3. step one\n" +
		"4. step two\n\n" +
		"| A | B |\n" +
		"| --- | --- |\n" +
		"| 1 | 2 |\n\n" +
		"```text\n" +
		"hello\n" +
		"```"

	if got != want {
		t.Fatalf("Document.String() mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestOrderedListIndentsContinuationLines(t *testing.T) {
	doc := New()
	doc.OrderedList(1, "first line\n- nested detail\nsecond line")

	got := doc.String()
	want := "1. first line\n   - nested detail\n   second line"
	if got != want {
		t.Fatalf("Document.String() = %q, want %q", got, want)
	}
}
