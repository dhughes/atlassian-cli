package atlassian

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownToADF_SimpleParagraph(t *testing.T) {
	markdown := "Hello world"
	adf, _, err := MarkdownToADF(markdown)
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
	adf, _, err := MarkdownToADF(markdown)
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
	adf, _, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_ItalicText(t *testing.T) {
	markdown := "This is *italic* text"
	adf, _, err := MarkdownToADF(markdown)
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
	adf, _, err := MarkdownToADF(markdown)
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
	adf, _, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_CodeBlock(t *testing.T) {
	markdown := "```go\nfunc main() {}\n```"
	adf, _, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_InlineCode(t *testing.T) {
	markdown := "This is `inline code` text"
	adf, _, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_Link(t *testing.T) {
	markdown := "[Example](https://example.com)"
	adf, _, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_Blockquote(t *testing.T) {
	markdown := "> This is a quote"
	adf, _, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if adf == nil {
		t.Fatal("Expected ADF output, got nil")
	}
}

func TestMarkdownToADF_EmptyString(t *testing.T) {
	markdown := ""
	adf, _, err := MarkdownToADF(markdown)
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
	adf, _, err := MarkdownToADF(markdown)
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

func TestExtractLocalImages_NoImages(t *testing.T) {
	markdown := "Hello world, no images here"
	images, cleaned := ExtractLocalImages(markdown)

	if len(images) != 0 {
		t.Errorf("Expected 0 images, got %d", len(images))
	}
	if cleaned != markdown {
		t.Errorf("Expected cleaned to equal original, got %q", cleaned)
	}
}

func TestExtractLocalImages_URLsIgnored(t *testing.T) {
	markdown := "![logo](https://example.com/logo.png)"
	images, cleaned := ExtractLocalImages(markdown)

	if len(images) != 0 {
		t.Errorf("Expected 0 images (URLs should be ignored), got %d", len(images))
	}
	if cleaned != markdown {
		t.Errorf("Expected cleaned to equal original for URL images, got %q", cleaned)
	}
}

func TestExtractLocalImages_HttpIgnored(t *testing.T) {
	markdown := "![logo](http://example.com/logo.png)"
	images, cleaned := ExtractLocalImages(markdown)

	if len(images) != 0 {
		t.Errorf("Expected 0 images (http URLs should be ignored), got %d", len(images))
	}
	if cleaned != markdown {
		t.Errorf("Expected cleaned to equal original for http URL images, got %q", cleaned)
	}
}

func TestExtractLocalImages_LocalFile(t *testing.T) {
	// Create a temp file to represent a local image
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(imgPath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	markdown := "Bug report: ![screenshot](" + imgPath + ")"
	images, cleaned := ExtractLocalImages(markdown)

	if len(images) != 1 {
		t.Fatalf("Expected 1 image, got %d", len(images))
	}

	if images[0].AltText != "screenshot" {
		t.Errorf("Expected alt text 'screenshot', got %q", images[0].AltText)
	}
	if images[0].FilePath != imgPath {
		t.Errorf("Expected file path %q, got %q", imgPath, images[0].FilePath)
	}

	if cleaned != "Bug report: ATLIMG_PLACEHOLDER_0" {
		t.Errorf("Expected cleaned markdown with placeholder, got %q", cleaned)
	}
}

func TestExtractLocalImages_NonexistentFile(t *testing.T) {
	markdown := "![screenshot](/nonexistent/path/image.png)"
	images, cleaned := ExtractLocalImages(markdown)

	if len(images) != 0 {
		t.Errorf("Expected 0 images (file doesn't exist), got %d", len(images))
	}
	if cleaned != markdown {
		t.Errorf("Expected cleaned to equal original for nonexistent file, got %q", cleaned)
	}
}

func TestExtractLocalImages_MixedContent(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "local.png")
	if err := os.WriteFile(imgPath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	markdown := "Text ![web](https://example.com/img.png) and ![local](" + imgPath + ") end"
	images, cleaned := ExtractLocalImages(markdown)

	if len(images) != 1 {
		t.Fatalf("Expected 1 local image, got %d", len(images))
	}
	if images[0].FilePath != imgPath {
		t.Errorf("Expected local image path, got %q", images[0].FilePath)
	}

	// URL image should be untouched
	if cleaned != "Text ![web](https://example.com/img.png) and ATLIMG_PLACEHOLDER_0 end" {
		t.Errorf("Expected mixed cleaned markdown, got %q", cleaned)
	}
}

func TestExtractLocalImages_EmptyAltText(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "noalt.png")
	if err := os.WriteFile(imgPath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	markdown := "Image: ![](" + imgPath + ")"
	images, cleaned := ExtractLocalImages(markdown)

	if len(images) != 1 {
		t.Fatalf("Expected 1 image, got %d", len(images))
	}
	if images[0].AltText != "" {
		t.Errorf("Expected empty alt text, got %q", images[0].AltText)
	}

	// Placeholder uses index regardless of alt text
	if cleaned != "Image: ATLIMG_PLACEHOLDER_0" {
		t.Errorf("Expected placeholder, got %q", cleaned)
	}
}

func TestBuildMediaSingleNode_WithAlt(t *testing.T) {
	node := BuildMediaSingleNode("abc-123-def", "screenshot")

	if node["type"] != "mediaSingle" {
		t.Errorf("Expected type 'mediaSingle', got %v", node["type"])
	}

	attrs, ok := node["attrs"].(map[string]any)
	if !ok {
		t.Fatal("Expected attrs map")
	}
	if attrs["layout"] != "center" {
		t.Errorf("Expected layout 'center', got %v", attrs["layout"])
	}

	content, ok := node["content"].([]any)
	if !ok || len(content) != 1 {
		t.Fatal("Expected 1 content child")
	}

	media, ok := content[0].(map[string]any)
	if !ok {
		t.Fatal("Expected media node")
	}
	if media["type"] != "media" {
		t.Errorf("Expected type 'media', got %v", media["type"])
	}

	mediaAttrs, ok := media["attrs"].(map[string]any)
	if !ok {
		t.Fatal("Expected media attrs")
	}
	if mediaAttrs["id"] != "abc-123-def" {
		t.Errorf("Expected id 'abc-123-def', got %v", mediaAttrs["id"])
	}
	if mediaAttrs["type"] != "file" {
		t.Errorf("Expected type 'file', got %v", mediaAttrs["type"])
	}
	if mediaAttrs["alt"] != "screenshot" {
		t.Errorf("Expected alt 'screenshot', got %v", mediaAttrs["alt"])
	}
}

func TestBuildMediaSingleNode_NoAlt(t *testing.T) {
	node := BuildMediaSingleNode("abc-123-def", "")

	content := node["content"].([]any)
	media := content[0].(map[string]any)
	mediaAttrs := media["attrs"].(map[string]any)

	if _, hasAlt := mediaAttrs["alt"]; hasAlt {
		t.Error("Expected no 'alt' key when alt text is empty")
	}
}

func TestMarkdownToADFWithImages_NoMedia(t *testing.T) {
	adf, _, err := MarkdownToADFWithImages("Hello world", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	docType, _ := adf["type"].(string)
	if docType != "doc" {
		t.Errorf("Expected doc type, got %v", docType)
	}
}

func TestMarkdownToADFWithImages_WithMedia(t *testing.T) {
	mediaNodes := []map[string]any{
		BuildMediaSingleNode("uuid-1", "first"),
		BuildMediaSingleNode("uuid-2", "second"),
	}

	// Input markdown with placeholders already in place (as ExtractLocalImages would produce).
	// Each placeholder on its own line becomes its own paragraph in ADF.
	markdown := "Some text\n\nATLIMG_PLACEHOLDER_0\n\nMore text\n\nATLIMG_PLACEHOLDER_1"

	adf, _, err := MarkdownToADFWithImages(markdown, mediaNodes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, ok := adf["content"].([]any)
	if !ok {
		t.Fatal("Expected content array")
	}

	// Should have: paragraph("Some text"), mediaSingle(0), paragraph("More text"), mediaSingle(1)
	if len(content) < 4 {
		t.Fatalf("Expected at least 4 content nodes, got %d", len(content))
	}

	// Check that mediaSingle nodes are at positions 1 and 3 (in-place, not appended)
	node1, ok := content[1].(map[string]any)
	if !ok {
		t.Fatal("Expected map at index 1")
	}
	if node1["type"] != "mediaSingle" {
		t.Errorf("Expected mediaSingle at index 1, got %v", node1["type"])
	}

	node2, ok := content[2].(map[string]any)
	if !ok {
		t.Fatal("Expected map at index 2")
	}
	if node2["type"] != "paragraph" {
		t.Errorf("Expected paragraph at index 2, got %v", node2["type"])
	}

	node3, ok := content[3].(map[string]any)
	if !ok {
		t.Fatal("Expected map at index 3")
	}
	if node3["type"] != "mediaSingle" {
		t.Errorf("Expected mediaSingle at index 3, got %v", node3["type"])
	}
}
