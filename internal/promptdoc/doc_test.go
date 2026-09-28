package promptdoc

import "testing"

func TestDocumentRendersParagraphsAndLists(t *testing.T) {
	doc := New()
	doc.Paragraph("Intro text")
	doc.BulletList("first", "second\ncontinued")

	got := doc.String()
	want := "Intro text\n\n- first\n- second\n  continued"
	if got != want {
		t.Fatalf("Document.String() = %q, want %q", got, want)
	}
}
