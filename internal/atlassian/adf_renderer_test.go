package atlassian

import (
	"encoding/json"
	"testing"
)

func mustRenderADF(t *testing.T, markdown string) map[string]any {
	t.Helper()
	adf, err := MarkdownToADF(markdown)
	if err != nil {
		t.Fatalf("MarkdownToADF(%q) returned error: %v", markdown, err)
	}
	if adf == nil {
		t.Fatalf("MarkdownToADF(%q) returned nil", markdown)
	}
	if adf["type"] != "doc" {
		t.Fatalf("expected doc type, got %v", adf["type"])
	}
	return adf
}

func contentNodes(t *testing.T, adf map[string]any) []map[string]any {
	t.Helper()
	raw, ok := adf["content"].([]any)
	if !ok {
		t.Fatal("expected content array")
	}
	var nodes []map[string]any
	for _, item := range raw {
		n, ok := item.(map[string]any)
		if !ok {
			t.Fatal("expected map in content array")
		}
		nodes = append(nodes, n)
	}
	return nodes
}

func nodeContent(t *testing.T, node map[string]any) []map[string]any {
	t.Helper()
	raw, ok := node["content"].([]any)
	if !ok {
		return nil
	}
	var nodes []map[string]any
	for _, item := range raw {
		n, ok := item.(map[string]any)
		if !ok {
			t.Fatal("expected map in node content")
		}
		nodes = append(nodes, n)
	}
	return nodes
}

func TestRenderer_Paragraph(t *testing.T) {
	adf := mustRenderADF(t, "Hello world")
	nodes := contentNodes(t, adf)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(nodes))
	}
	if nodes[0]["type"] != "paragraph" {
		t.Errorf("expected paragraph, got %v", nodes[0]["type"])
	}

	children := nodeContent(t, nodes[0])
	if len(children) == 0 {
		t.Fatal("expected at least 1 text child")
	}

	var combined string
	for _, child := range children {
		if child["type"] != "text" {
			t.Errorf("expected text node, got %v", child["type"])
		}
		combined += child["text"].(string)
	}
	if combined != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", combined)
	}
}

func TestRenderer_MultipleParagraphs(t *testing.T) {
	adf := mustRenderADF(t, "First paragraph\n\nSecond paragraph")
	nodes := contentNodes(t, adf)

	if len(nodes) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(nodes))
	}
	for _, n := range nodes {
		if n["type"] != "paragraph" {
			t.Errorf("expected paragraph, got %v", n["type"])
		}
	}
}

func TestRenderer_Headings(t *testing.T) {
	tests := []struct {
		markdown string
		level    float64
	}{
		{"# H1", 1},
		{"## H2", 2},
		{"### H3", 3},
		{"#### H4", 4},
		{"##### H5", 5},
		{"###### H6", 6},
	}

	for _, tt := range tests {
		t.Run(tt.markdown, func(t *testing.T) {
			adf := mustRenderADF(t, tt.markdown)
			nodes := contentNodes(t, adf)
			if len(nodes) != 1 {
				t.Fatalf("expected 1 node, got %d", len(nodes))
			}
			if nodes[0]["type"] != "heading" {
				t.Fatalf("expected heading, got %v", nodes[0]["type"])
			}
			attrs, ok := nodes[0]["attrs"].(map[string]any)
			if !ok {
				t.Fatal("expected attrs on heading")
			}
			if attrs["level"] != tt.level {
				t.Errorf("expected level %v, got %v", tt.level, attrs["level"])
			}
		})
	}
}

func TestRenderer_Bold(t *testing.T) {
	adf := mustRenderADF(t, "This is **bold** text")
	nodes := contentNodes(t, adf)
	children := nodeContent(t, nodes[0])

	found := false
	for _, child := range children {
		marks, ok := child["marks"].([]any)
		if ok && len(marks) > 0 {
			mark := marks[0].(map[string]any)
			if mark["type"] == "strong" && child["text"] == "bold" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected strong mark on 'bold' text")
	}
}

func TestRenderer_Italic(t *testing.T) {
	adf := mustRenderADF(t, "This is *italic* text")
	nodes := contentNodes(t, adf)
	children := nodeContent(t, nodes[0])

	found := false
	for _, child := range children {
		marks, ok := child["marks"].([]any)
		if ok && len(marks) > 0 {
			mark := marks[0].(map[string]any)
			if mark["type"] == "em" && child["text"] == "italic" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected em mark on 'italic' text")
	}
}

func TestRenderer_Strikethrough(t *testing.T) {
	adf := mustRenderADF(t, "This is ~~struck~~ text")
	nodes := contentNodes(t, adf)
	children := nodeContent(t, nodes[0])

	found := false
	for _, child := range children {
		marks, ok := child["marks"].([]any)
		if ok && len(marks) > 0 {
			mark := marks[0].(map[string]any)
			if mark["type"] == "strike" && child["text"] == "struck" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected strike mark on 'struck' text")
	}
}

func TestRenderer_InlineCode(t *testing.T) {
	adf := mustRenderADF(t, "Use `fmt.Println` here")
	nodes := contentNodes(t, adf)
	children := nodeContent(t, nodes[0])

	found := false
	for _, child := range children {
		marks, ok := child["marks"].([]any)
		if ok && len(marks) > 0 {
			mark := marks[0].(map[string]any)
			if mark["type"] == "code" && child["text"] == "fmt.Println" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected code mark on 'fmt.Println' text")
	}
}

func TestRenderer_FencedCodeBlock(t *testing.T) {
	adf := mustRenderADF(t, "```go\nfunc main() {}\n```")
	nodes := contentNodes(t, adf)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0]["type"] != "codeBlock" {
		t.Fatalf("expected codeBlock, got %v", nodes[0]["type"])
	}

	attrs, ok := nodes[0]["attrs"].(map[string]any)
	if !ok {
		t.Fatal("expected attrs on codeBlock")
	}
	if attrs["language"] != "go" {
		t.Errorf("expected language 'go', got %v", attrs["language"])
	}

	children := nodeContent(t, nodes[0])
	if len(children) != 1 {
		t.Fatalf("expected 1 text child, got %d", len(children))
	}
	if children[0]["text"] != "func main() {}\n" {
		t.Errorf("expected code content, got %q", children[0]["text"])
	}
}

func TestRenderer_FencedCodeBlockNoLanguage(t *testing.T) {
	adf := mustRenderADF(t, "```\nsome code\n```")
	nodes := contentNodes(t, adf)

	if nodes[0]["type"] != "codeBlock" {
		t.Fatalf("expected codeBlock, got %v", nodes[0]["type"])
	}
	if nodes[0]["attrs"] != nil {
		t.Errorf("expected no attrs for code block without language, got %v", nodes[0]["attrs"])
	}
}

func TestRenderer_BulletList(t *testing.T) {
	adf := mustRenderADF(t, "- Item 1\n- Item 2\n- Item 3")
	nodes := contentNodes(t, adf)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0]["type"] != "bulletList" {
		t.Fatalf("expected bulletList, got %v", nodes[0]["type"])
	}

	items := nodeContent(t, nodes[0])
	if len(items) != 3 {
		t.Fatalf("expected 3 list items, got %d", len(items))
	}
	for _, item := range items {
		if item["type"] != "listItem" {
			t.Errorf("expected listItem, got %v", item["type"])
		}
	}
}

func TestRenderer_OrderedList(t *testing.T) {
	adf := mustRenderADF(t, "1. First\n2. Second\n3. Third")
	nodes := contentNodes(t, adf)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0]["type"] != "orderedList" {
		t.Fatalf("expected orderedList, got %v", nodes[0]["type"])
	}

	items := nodeContent(t, nodes[0])
	if len(items) != 3 {
		t.Fatalf("expected 3 list items, got %d", len(items))
	}
}

func TestRenderer_NestedList(t *testing.T) {
	md := "- Parent 1\n  - Child 1\n  - Child 2\n- Parent 2"
	adf := mustRenderADF(t, md)
	nodes := contentNodes(t, adf)

	if nodes[0]["type"] != "bulletList" {
		t.Fatalf("expected bulletList, got %v", nodes[0]["type"])
	}

	items := nodeContent(t, nodes[0])
	if len(items) != 2 {
		t.Fatalf("expected 2 top-level items, got %d", len(items))
	}

	firstItemContent := nodeContent(t, items[0])
	hasNestedList := false
	for _, child := range firstItemContent {
		if child["type"] == "bulletList" {
			hasNestedList = true
			nestedItems := nodeContent(t, child)
			if len(nestedItems) != 2 {
				t.Errorf("expected 2 nested items, got %d", len(nestedItems))
			}
		}
	}
	if !hasNestedList {
		t.Error("expected nested bullet list in first item")
	}
}

func TestRenderer_Blockquote(t *testing.T) {
	adf := mustRenderADF(t, "> This is a quote")
	nodes := contentNodes(t, adf)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0]["type"] != "blockquote" {
		t.Fatalf("expected blockquote, got %v", nodes[0]["type"])
	}

	children := nodeContent(t, nodes[0])
	if len(children) != 1 {
		t.Fatalf("expected 1 paragraph in blockquote, got %d", len(children))
	}
	if children[0]["type"] != "paragraph" {
		t.Errorf("expected paragraph in blockquote, got %v", children[0]["type"])
	}
}

func TestRenderer_HorizontalRule(t *testing.T) {
	adf := mustRenderADF(t, "Before\n\n---\n\nAfter")
	nodes := contentNodes(t, adf)

	foundRule := false
	for _, n := range nodes {
		if n["type"] == "rule" {
			foundRule = true
		}
	}
	if !foundRule {
		t.Error("expected rule node in output")
	}
}

func TestRenderer_Link(t *testing.T) {
	adf := mustRenderADF(t, "[Example](https://example.com)")
	nodes := contentNodes(t, adf)
	children := nodeContent(t, nodes[0])

	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	child := children[0]
	if child["text"] != "Example" {
		t.Errorf("expected text 'Example', got %v", child["text"])
	}

	marks, ok := child["marks"].([]any)
	if !ok || len(marks) == 0 {
		t.Fatal("expected link mark")
	}
	mark := marks[0].(map[string]any)
	if mark["type"] != "link" {
		t.Errorf("expected link mark type, got %v", mark["type"])
	}
	attrs, ok := mark["attrs"].(map[string]any)
	if !ok {
		t.Fatal("expected mark attrs")
	}
	if attrs["href"] != "https://example.com" {
		t.Errorf("expected href, got %v", attrs["href"])
	}
}

func TestRenderer_HardBreak(t *testing.T) {
	adf := mustRenderADF(t, "Line 1  \nLine 2")
	nodes := contentNodes(t, adf)
	children := nodeContent(t, nodes[0])

	foundBreak := false
	for _, child := range children {
		if child["type"] == "hardBreak" {
			foundBreak = true
		}
	}
	if !foundBreak {
		t.Error("expected hardBreak node")
	}
}

func TestRenderer_Table(t *testing.T) {
	md := "| Header 1 | Header 2 |\n| --- | --- |\n| Cell 1 | Cell 2 |\n| Cell 3 | Cell 4 |"
	adf := mustRenderADF(t, md)
	nodes := contentNodes(t, adf)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0]["type"] != "table" {
		t.Fatalf("expected table, got %v", nodes[0]["type"])
	}

	rows := nodeContent(t, nodes[0])
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (1 header + 2 body), got %d", len(rows))
	}

	for _, row := range rows {
		if row["type"] != "tableRow" {
			t.Errorf("expected tableRow, got %v", row["type"])
		}
	}

	headerCells := nodeContent(t, rows[0])
	if len(headerCells) != 2 {
		t.Fatalf("expected 2 header cells, got %d", len(headerCells))
	}
	for _, cell := range headerCells {
		if cell["type"] != "tableHeader" {
			t.Errorf("expected tableHeader, got %v", cell["type"])
		}
		paras := nodeContent(t, cell)
		if len(paras) == 0 || paras[0]["type"] != "paragraph" {
			t.Error("expected paragraph inside table header cell")
		}
	}

	bodyCells := nodeContent(t, rows[1])
	if len(bodyCells) != 2 {
		t.Fatalf("expected 2 body cells, got %d", len(bodyCells))
	}
	for _, cell := range bodyCells {
		if cell["type"] != "tableCell" {
			t.Errorf("expected tableCell, got %v", cell["type"])
		}
	}
}

func TestRenderer_TableCellContent(t *testing.T) {
	md := "| Name | Value |\n| --- | --- |\n| foo | bar |"
	adf := mustRenderADF(t, md)
	nodes := contentNodes(t, adf)
	rows := nodeContent(t, nodes[0])

	bodyRow := rows[1]
	cells := nodeContent(t, bodyRow)
	firstCell := cells[0]
	paras := nodeContent(t, firstCell)
	textNodes := nodeContent(t, paras[0])

	if len(textNodes) == 0 {
		t.Fatal("expected text in table cell")
	}
	if textNodes[0]["text"] != "foo" {
		t.Errorf("expected 'foo', got %v", textNodes[0]["text"])
	}
}

func TestRenderer_Checkbox_Unchecked(t *testing.T) {
	md := "- [ ] Todo item"
	adf := mustRenderADF(t, md)
	nodes := contentNodes(t, adf)

	if nodes[0]["type"] != "bulletList" {
		t.Fatalf("expected bulletList, got %v", nodes[0]["type"])
	}

	items := nodeContent(t, nodes[0])
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	attrs, ok := item["attrs"].(map[string]any)
	if !ok {
		t.Fatal("expected attrs on task list item")
	}
	if attrs["state"] != "TODO" {
		t.Errorf("expected state 'TODO', got %v", attrs["state"])
	}
}

func TestRenderer_Checkbox_Checked(t *testing.T) {
	md := "- [x] Done item"
	adf := mustRenderADF(t, md)
	nodes := contentNodes(t, adf)

	items := nodeContent(t, nodes[0])
	item := items[0]
	attrs, ok := item["attrs"].(map[string]any)
	if !ok {
		t.Fatal("expected attrs on task list item")
	}
	if attrs["state"] != "DONE" {
		t.Errorf("expected state 'DONE', got %v", attrs["state"])
	}
}

func TestRenderer_Checkbox_Mixed(t *testing.T) {
	md := "- [x] Done\n- [ ] Todo\n- Regular item"
	adf := mustRenderADF(t, md)
	nodes := contentNodes(t, adf)

	items := nodeContent(t, nodes[0])
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	item0Attrs, ok := items[0]["attrs"].(map[string]any)
	if !ok || item0Attrs["state"] != "DONE" {
		t.Errorf("expected first item state 'DONE', got %v", items[0]["attrs"])
	}

	item1Attrs, ok := items[1]["attrs"].(map[string]any)
	if !ok || item1Attrs["state"] != "TODO" {
		t.Errorf("expected second item state 'TODO', got %v", items[1]["attrs"])
	}

	if items[2]["attrs"] != nil {
		t.Errorf("expected no attrs on regular list item, got %v", items[2]["attrs"])
	}
}

func TestRenderer_EmptyInput(t *testing.T) {
	adf := mustRenderADF(t, "")
	if adf["type"] != "doc" {
		t.Errorf("expected doc type, got %v", adf["type"])
	}
}

func TestRenderer_ComplexDocument(t *testing.T) {
	md := `# Title

This is **bold** and *italic* and ` + "`code`" + ` text.

## Section

- Item 1
- Item 2

1. First
2. Second

> A quote

---

| Col A | Col B |
| ----- | ----- |
| val 1 | val 2 |

- [x] Done task
- [ ] Todo task

` + "```go\nfunc main() {}\n```"

	adf := mustRenderADF(t, md)
	nodes := contentNodes(t, adf)

	typeCount := map[string]int{}
	for _, n := range nodes {
		nodeType, _ := n["type"].(string)
		typeCount[nodeType]++
	}

	if typeCount["heading"] < 2 {
		t.Error("expected at least 2 headings")
	}
	if typeCount["paragraph"] < 1 {
		t.Error("expected at least 1 paragraph")
	}
	if typeCount["bulletList"] < 2 {
		t.Error("expected at least 2 bullet lists (regular + task)")
	}
	if typeCount["orderedList"] < 1 {
		t.Error("expected at least 1 ordered list")
	}
	if typeCount["blockquote"] < 1 {
		t.Error("expected at least 1 blockquote")
	}
	if typeCount["rule"] < 1 {
		t.Error("expected at least 1 rule")
	}
	if typeCount["table"] < 1 {
		t.Error("expected at least 1 table")
	}
	if typeCount["codeBlock"] < 1 {
		t.Error("expected at least 1 code block")
	}
}

func TestRenderer_NoPanic_OnAnyInput(t *testing.T) {
	inputs := []string{
		"- [ ] checkbox",
		"- [x] checked",
		"| a | b |\n| - | - |\n| 1 | 2 |",
		"<div>html block</div>",
		"text with <em>inline html</em>",
		"![alt](https://example.com/img.png)",
		"![alt](local_file.png)",
		"~~strike~~",
		"***bold italic***",
		"- [ ] task\n  - [ ] subtask",
		"",
		"   ",
		"\n\n\n",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on input %q: %v", input, r)
				}
			}()
			_, err := MarkdownToADF(input)
			if err != nil {
				t.Errorf("unexpected error on input %q: %v", input, err)
			}
		})
	}
}

func TestRenderer_ValidJSON(t *testing.T) {
	md := "# Test\n\n**bold** and *italic*\n\n- item\n\n| a | b |\n| - | - |\n| 1 | 2 |"
	adf := mustRenderADF(t, md)

	b, err := json.Marshal(adf)
	if err != nil {
		t.Fatalf("failed to marshal ADF back to JSON: %v", err)
	}

	var roundTrip map[string]any
	if err := json.Unmarshal(b, &roundTrip); err != nil {
		t.Fatalf("failed to unmarshal round-tripped JSON: %v", err)
	}

	if roundTrip["type"] != "doc" {
		t.Error("round-tripped document lost its type")
	}
}

func TestRenderer_TableWithFormatting(t *testing.T) {
	md := "| **Bold** | *Italic* |\n| --- | --- |\n| `code` | ~~strike~~ |"
	adf := mustRenderADF(t, md)
	nodes := contentNodes(t, adf)

	if nodes[0]["type"] != "table" {
		t.Fatalf("expected table, got %v", nodes[0]["type"])
	}

	b, _ := json.Marshal(adf)
	adfStr := string(b)

	for _, mark := range []string{"strong", "em", "code", "strike"} {
		if !contains(adfStr, mark) {
			t.Errorf("expected %s mark in table output", mark)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
