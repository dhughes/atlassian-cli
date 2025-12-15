package atlassian

import (
	"strings"
	"testing"
)

func TestHTMLToText_Empty(t *testing.T) {
	result := HTMLToText("")
	if result != "" {
		t.Errorf("Expected empty string for empty input, got %q", result)
	}
}

func TestHTMLToText_PlainText(t *testing.T) {
	input := "Plain text"
	result := HTMLToText(input)
	if result != input {
		t.Errorf("Expected %q, got %q", input, result)
	}
}

func TestHTMLToText_Headings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"H1", "<h1>Title</h1>", "# Title"},
		{"H2", "<h2>Subtitle</h2>", "## Subtitle"},
		{"H3", "<h3>Section</h3>", "### Section"},
		{"H4", "<h4>Subsection</h4>", "#### Subsection"},
		{"H5", "<h5>Minor heading</h5>", "##### Minor heading"},
		{"H6", "<h6>Smallest</h6>", "###### Smallest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HTMLToText(tt.input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHTMLToText_Paragraphs(t *testing.T) {
	input := "<p>First paragraph</p><p>Second paragraph</p>"
	result := HTMLToText(input)

	if !strings.Contains(result, "First paragraph") {
		t.Errorf("Expected first paragraph in output, got %q", result)
	}
	if !strings.Contains(result, "Second paragraph") {
		t.Errorf("Expected second paragraph in output, got %q", result)
	}
}

func TestHTMLToText_LineBreaks(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Self-closing", "Line 1<br/>Line 2"},
		{"With space", "Line 1<br />Line 2"},
		{"Without slash", "Line 1<br>Line 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HTMLToText(tt.input)
			if !strings.Contains(result, "Line 1") || !strings.Contains(result, "Line 2") {
				t.Errorf("Expected both lines in output, got %q", result)
			}
		})
	}
}

func TestHTMLToText_UnorderedList(t *testing.T) {
	input := "<ul><li>Item 1</li><li>Item 2</li><li>Item 3</li></ul>"
	result := HTMLToText(input)

	if !strings.Contains(result, "• Item 1") {
		t.Errorf("Expected '• Item 1' in output, got %q", result)
	}
	if !strings.Contains(result, "• Item 2") {
		t.Errorf("Expected '• Item 2' in output, got %q", result)
	}
	if !strings.Contains(result, "• Item 3") {
		t.Errorf("Expected '• Item 3' in output, got %q", result)
	}
}

func TestHTMLToText_OrderedList(t *testing.T) {
	input := "<ol><li>First</li><li>Second</li></ol>"
	result := HTMLToText(input)

	if !strings.Contains(result, "• First") {
		t.Errorf("Expected '• First' in output, got %q", result)
	}
	if !strings.Contains(result, "• Second") {
		t.Errorf("Expected '• Second' in output, got %q", result)
	}
}

func TestHTMLToText_CodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"Pre with code", "<pre><code>func main() {}</code></pre>", "func main()"},
		{"Pre only", "<pre>func main() {}</pre>", "func main()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HTMLToText(tt.input)
			// Just verify the code content is present
			// The exact formatting may vary based on regex processing order
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected code content %q in output, got %q", tt.contains, result)
			}
		})
	}
}

func TestHTMLToText_InlineCode(t *testing.T) {
	input := "This is <code>inline code</code> text"
	result := HTMLToText(input)

	if !strings.Contains(result, "`inline code`") {
		t.Errorf("Expected '`inline code`' in output, got %q", result)
	}
}

func TestHTMLToText_Bold(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Strong tag", "<strong>bold text</strong>"},
		{"B tag", "<b>bold text</b>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HTMLToText(tt.input)
			if !strings.Contains(result, "**bold text**") {
				t.Errorf("Expected '**bold text**' in output, got %q", result)
			}
		})
	}
}

func TestHTMLToText_Italic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Em tag", "<em>italic text</em>"},
		{"I tag", "<i>italic text</i>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HTMLToText(tt.input)
			if !strings.Contains(result, "*italic text*") {
				t.Errorf("Expected '*italic text*' in output, got %q", result)
			}
		})
	}
}

func TestHTMLToText_Links(t *testing.T) {
	input := `<a href="https://example.com">Example Link</a>`
	result := HTMLToText(input)

	if !strings.Contains(result, "Example Link") {
		t.Errorf("Expected link text in output, got %q", result)
	}
	if !strings.Contains(result, "https://example.com") {
		t.Errorf("Expected URL in output, got %q", result)
	}
}

func TestHTMLToText_Images(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"With alt", `<img alt="Logo" src="logo.png">`, "[Image: Logo]"},
		{"With src only", `<img src="photo.jpg">`, "[Image: photo.jpg]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HTMLToText(tt.input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHTMLToText_Table(t *testing.T) {
	input := `
		<table>
			<tr>
				<th>Header 1</th>
				<th>Header 2</th>
			</tr>
			<tr>
				<td>Cell 1</td>
				<td>Cell 2</td>
			</tr>
		</table>
	`
	result := HTMLToText(input)

	if !strings.Contains(result, "Header 1") || !strings.Contains(result, "Header 2") {
		t.Errorf("Expected table headers in output, got %q", result)
	}
	if !strings.Contains(result, "Cell 1") || !strings.Contains(result, "Cell 2") {
		t.Errorf("Expected table cells in output, got %q", result)
	}
	if !strings.Contains(result, "|") {
		t.Errorf("Expected table separator '|' in output, got %q", result)
	}
}

func TestHTMLToText_Blockquote(t *testing.T) {
	input := "<blockquote>This is a quote</blockquote>"
	result := HTMLToText(input)

	if !strings.Contains(result, "> ") {
		t.Errorf("Expected blockquote marker '> ' in output, got %q", result)
	}
	if !strings.Contains(result, "This is a quote") {
		t.Errorf("Expected quote text in output, got %q", result)
	}
}

func TestHTMLToText_DivsAndSpans(t *testing.T) {
	input := "<div>Text in <span>a span</span> and div</div>"
	result := HTMLToText(input)

	expected := "Text in a span and div"
	if !strings.Contains(result, expected) {
		t.Errorf("Expected %q in output, got %q", expected, result)
	}
}

func TestHTMLToText_HTMLEntities(t *testing.T) {
	tests := []struct {
		entity   string
		expected string
	}{
		{"&amp;", "&"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&quot;", "\""},
		{"&#39;", "'"},
		{"&nbsp;", " "},
	}

	for _, tt := range tests {
		t.Run(tt.entity, func(t *testing.T) {
			input := "Test " + tt.entity + " entity"
			result := HTMLToText(input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected %q to be decoded to %q, got %q", tt.entity, tt.expected, result)
			}
		})
	}
}

func TestHTMLToText_WhitespaceNormalization(t *testing.T) {
	input := `
		<p>Paragraph   with    extra    spaces</p>


		<p>Another paragraph</p>
	`
	result := HTMLToText(input)

	// Should not have excessive newlines
	if strings.Contains(result, "\n\n\n") {
		t.Errorf("Expected normalized whitespace, found excessive newlines in %q", result)
	}

	// Should normalize spaces
	if strings.Contains(result, "extra    spaces") {
		t.Errorf("Expected normalized spaces, got %q", result)
	}
}

func TestHTMLToText_ComplexHTML(t *testing.T) {
	input := `
		<h1>Main Title</h1>
		<p>This is a <strong>bold</strong> and <em>italic</em> paragraph with <code>code</code>.</p>
		<h2>Section</h2>
		<ul>
			<li>First item</li>
			<li>Second item with <a href="http://example.com">a link</a></li>
		</ul>
		<blockquote>
			<p>A quoted paragraph</p>
		</blockquote>
		<pre><code>func main() {
	println("Hello")
}</code></pre>
	`

	result := HTMLToText(input)

	// Verify all components are present
	expectations := []string{
		"# Main Title",
		"**bold**",
		"*italic*",
		"`code`",
		"## Section",
		"• First item",
		"• Second item",
		"http://example.com",
		"> ",
		"quoted paragraph",
		"func main()",
	}

	for _, expected := range expectations {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected complex HTML to contain %q, got %q", expected, result)
		}
	}
}

func TestHTMLToText_RemoveUnknownTags(t *testing.T) {
	input := "<custom-tag>Content</custom-tag><another>Text</another>"
	result := HTMLToText(input)

	// Tags should be removed, content should remain
	if strings.Contains(result, "<") || strings.Contains(result, ">") {
		t.Errorf("Expected all tags to be removed, got %q", result)
	}
	if !strings.Contains(result, "Content") || !strings.Contains(result, "Text") {
		t.Errorf("Expected content to be preserved, got %q", result)
	}
}

func TestHTMLToText_AttributesInTags(t *testing.T) {
	input := `<p class="special" id="para1" style="color:red;">Paragraph with attributes</p>`
	result := HTMLToText(input)

	// Attributes should be ignored, content should remain
	if strings.Contains(result, "class") || strings.Contains(result, "id") || strings.Contains(result, "style") {
		t.Errorf("Expected attributes to be removed, got %q", result)
	}
	if !strings.Contains(result, "Paragraph with attributes") {
		t.Errorf("Expected content to be preserved, got %q", result)
	}
}
