package atlassian

import (
	"testing"
)

func TestMarkdownToADF_SimpleParagraph(t *testing.T) {
	markdown := "Hello world"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}

	// Check document type
	docType, ok := adf["type"].(string)
	if !ok || docType != "doc" {
		t.Errorf("Expected document type 'doc', got %v", adf["type"])
	}

	// Check that content exists
	_, hasContent := adf["content"]
	if !hasContent {
		t.Error("Expected document to have content")
	}
}

func TestMarkdownToADF_Heading(t *testing.T) {
	markdown := "# Main Title"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}

	// Verify it's a document
	docType, _ := adf["type"].(string)
	if docType != "doc" {
		t.Errorf("Expected document type 'doc', got %v", docType)
	}
}

func TestMarkdownToADF_BoldText(t *testing.T) {
	markdown := "This is **bold** text"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_ItalicText(t *testing.T) {
	markdown := "This is *italic* text"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_BulletList(t *testing.T) {
	markdown := `- Item 1
- Item 2
- Item 3`
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_OrderedList(t *testing.T) {
	markdown := `1. First
2. Second
3. Third`
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_CodeBlock(t *testing.T) {
	markdown := "```go\nfunc main() {}\n```"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_InlineCode(t *testing.T) {
	markdown := "This is `inline code` text"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_Link(t *testing.T) {
	markdown := "[Example](https://example.com)"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_Blockquote(t *testing.T) {
	markdown := "> This is a quote"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_EmptyString(t *testing.T) {
	markdown := ""
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error for empty string, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}

	// Empty markdown should still produce a valid document
	docType, _ := adf["type"].(string)
	if docType != "doc" {
		t.Errorf("Expected document type 'doc', got %v", docType)
	}
}

func TestMarkdownToADF_ComplexDocument(t *testing.T) {
	markdown := "# Main Title\n\n" +
		"This is a **bold** paragraph with *italic* text and `code`.\n\n" +
		"## Section 1\n\n" +
		"- Item 1\n" +
		"- Item 2\n\n" +
		"## Section 2\n\n" +
		"1. First\n" +
		"2. Second\n\n" +
		"> A quote\n\n" +
		"```go\n" +
		"func main() {\n" +
		"	println(\"Hello\")\n" +
		"}\n" +
		"```\n"
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}

	// Verify it's a document
	docType, _ := adf["type"].(string)
	if docType != "doc" {
		t.Errorf("Expected document type 'doc', got %v", docType)
	}

	// Verify it has content
	content, hasContent := adf["content"].([]any)
	if !hasContent {
		t.Fatal("Expected document to have content")
	}

	// Should have multiple blocks for all the markdown sections
	if len(content) == 0 {
		t.Error("Expected non-empty content array")
	}
}
