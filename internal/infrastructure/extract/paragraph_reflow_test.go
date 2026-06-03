package extract_test

import (
	"strings"
	"testing"

	"github.com/tikhomirovv/book-translator/internal/infrastructure/extract"
)

func TestReflowPlainText_joinsWrappedProse(t *testing.T) {
	t.Parallel()

	raw := "The ability to use a selected variety of interventions to interpret our past for\n" +
		"the client's improved future is a rare dynamic, one shared, perhaps, with\n" +
		"doctors and lawyers."

	out := extract.ReflowPlainText(raw)
	if strings.Count(out, "\n\n") != 0 {
		t.Fatalf("expected one paragraph, got %q", out)
	}
	if !strings.Contains(out, "client's improved future") {
		t.Fatalf("unexpected reflow: %q", out)
	}
}

func TestReflowPlainText_keepsTOCShortLinesSeparate(t *testing.T) {
	t.Parallel()

	raw := "Cover\nTitle Page\nCopyright"
	out := extract.ReflowPlainText(raw)
	parts := strings.Split(out, "\n\n")
	if len(parts) != 3 {
		t.Fatalf("expected 3 paragraphs, got %d: %q", len(parts), out)
	}
}

func TestReflowPlainText_joinsWrappedHeading(t *testing.T) {
	t.Parallel()

	raw := "THE IMPORTANCE OF BUYER COMMITMENT, NOT\nCOMPLIANCE"
	out := extract.ReflowPlainText(raw)
	if strings.Count(out, "\n\n") != 0 {
		t.Fatalf("expected joined heading, got %q", out)
	}
	if !strings.Contains(out, "COMMITMENT, NOT COMPLIANCE") {
		t.Fatalf("unexpected join: %q", out)
	}
}

func TestReflowPlainText_splitsHeadingAndBody(t *testing.T) {
	t.Parallel()

	raw := "THE SUBTLE TRANSFORMATION:\nCONSULTANT PAST TO CLIENT FUTURE\nThe final major consideration"
	out := extract.ReflowPlainText(raw)
	parts := strings.Split(out, "\n\n")
	if len(parts) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d: %q", len(parts), out)
	}
	if !strings.HasPrefix(parts[1], "The final") {
		t.Fatalf("second paragraph: %q", parts[1])
	}
}

func TestReflowPlainText_splitsFigureCaption(t *testing.T) {
	t.Parallel()

	raw := "Figure 3.1 Value Distance\nNo one cares, really, about how good you are."
	out := extract.ReflowPlainText(raw)
	parts := strings.Split(out, "\n\n")
	if len(parts) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d: %q", len(parts), out)
	}
}

func TestReflowPlainText_dropsWatermark(t *testing.T) {
	t.Parallel()

	raw := "OceanofPDF.com\nReal paragraph line that is long enough to stay.\nOceanofPDF.com"
	out := extract.ReflowPlainText(raw)
	if strings.Contains(strings.ToLower(out), "oceanofpdf") {
		t.Fatalf("watermark not removed: %q", out)
	}
}

func TestNormalizeParagraphs_pdfWithReflow(t *testing.T) {
	t.Parallel()

	raw := "Table of Contents\nCover\nTitle Page\n" +
		"Chapter body line one that continues here\nand finishes the thought."
	paras := extract.NormalizeParagraphs(raw)
	if len(paras) < 4 {
		t.Fatalf("expected multiple paragraphs from PDF-style text, got %d", len(paras))
	}
	last := paras[len(paras)-1].Text
	if !strings.Contains(last, "finishes the thought") {
		t.Fatalf("last paragraph not reflowed: %q", last)
	}
}
