package atlassian

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extAst "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
)

var _ renderer.Renderer = &adfRenderer{}

type adfRenderer struct {
	doc       *adfNode
	stack     []*adfNode
	markStack []adfMark
	warnings  []string
	seenWarns map[string]bool
}

type adfNode struct {
	Type    string         `json:"type"`
	Version int            `json:"version,omitempty"`
	Attrs   map[string]any `json:"attrs,omitempty"`
	Content []*adfNode     `json:"content,omitempty"`
	Marks   []adfMark      `json:"marks,omitempty"`
	Text    string         `json:"text,omitempty"`
}

type adfMark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

func newADFRenderer() *adfRenderer {
	doc := &adfNode{Type: "doc", Version: 1}
	return &adfRenderer{
		doc:       doc,
		stack:     []*adfNode{doc},
		seenWarns: map[string]bool{},
	}
}

func (r *adfRenderer) warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if !r.seenWarns[msg] {
		r.seenWarns[msg] = true
		r.warnings = append(r.warnings, msg)
	}
}

func (r *adfRenderer) current() *adfNode {
	return r.stack[len(r.stack)-1]
}

func (r *adfRenderer) push(n *adfNode) {
	r.current().Content = append(r.current().Content, n)
	r.stack = append(r.stack, n)
}

func (r *adfRenderer) pop() {
	r.stack = r.stack[:len(r.stack)-1]
}

func (r *adfRenderer) addInline(n *adfNode) {
	r.current().Content = append(r.current().Content, n)
}

func (r *adfRenderer) pushMark(m adfMark) {
	r.markStack = append(r.markStack, m)
}

func (r *adfRenderer) popMark() {
	r.markStack = r.markStack[:len(r.markStack)-1]
}

func (r *adfRenderer) currentMarks() []adfMark {
	if len(r.markStack) == 0 {
		return nil
	}
	marks := make([]adfMark, len(r.markStack))
	copy(marks, r.markStack)
	return marks
}

func (r *adfRenderer) ancestorHasType(nodeType string) bool {
	for _, n := range r.stack {
		if n.Type == nodeType {
			return true
		}
	}
	return false
}

func (r *adfRenderer) parentType() string {
	if len(r.stack) < 2 {
		return ""
	}
	return r.stack[len(r.stack)-1].Type
}

var commonMarkEscapable = func() map[byte]bool {
	m := map[byte]bool{}
	for _, c := range []byte(`!\#$%&'()*+,-./:;<=>?@[\]^_{|}~` + "`") {
		m[c] = true
	}
	return m
}()

func stripEscapes(s string) string {
	if !strings.Contains(s, "\\") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && commonMarkEscapable[s[i+1]] {
			b.WriteByte(s[i+1])
			i++
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func (r *adfRenderer) AddOptions(...renderer.Option) {}

func (r *adfRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	err := ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		return r.walkNode(source, node, entering)
	})
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(r.doc, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func listHasTaskChildren(n *ast.List) bool {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if li, ok := child.(*ast.ListItem); ok {
			for gc := li.FirstChild(); gc != nil; gc = gc.NextSibling() {
				if fc := gc.FirstChild(); fc != nil {
					if _, ok := fc.(*extAst.TaskCheckBox); ok {
						return true
					}
				}
			}
		}
	}
	return false
}

func (r *adfRenderer) walkNode(source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	switch n := node.(type) {
	case *ast.Document:
		return ast.WalkContinue, nil

	case *ast.Paragraph:
		if entering {
			r.push(&adfNode{Type: "paragraph"})
		} else {
			r.pop()
		}

	case *ast.Heading:
		if entering {
			r.push(&adfNode{
				Type:  "heading",
				Attrs: map[string]any{"level": n.Level},
			})
		} else {
			r.pop()
		}

	case *ast.Text:
		if entering {
			text := stripEscapes(string(n.Text(source)))
			if len(text) > 0 {
				textNode := &adfNode{Type: "text", Text: text}
				if marks := r.currentMarks(); marks != nil {
					textNode.Marks = marks
				}
				r.addInline(textNode)
			}
			if n.SoftLineBreak() {
				textNode := &adfNode{Type: "text", Text: " "}
				if marks := r.currentMarks(); marks != nil {
					textNode.Marks = marks
				}
				r.addInline(textNode)
			}
			if n.HardLineBreak() {
				r.addInline(&adfNode{Type: "hardBreak"})
			}
		}

	case *ast.String:
		if entering {
			text := string(n.Text(source))
			if len(text) > 0 {
				textNode := &adfNode{Type: "text", Text: text}
				if marks := r.currentMarks(); marks != nil {
					textNode.Marks = marks
				}
				r.addInline(textNode)
			}
		}

	case *ast.TextBlock:
		if r.current().Type == "taskItem" {
			// taskItem expects inline content directly, no paragraph wrapper
		} else if entering {
			r.push(&adfNode{Type: "paragraph"})
		} else {
			r.pop()
		}

	case *ast.Emphasis:
		if entering {
			mark := "em"
			if n.Level >= 2 {
				mark = "strong"
			}
			r.pushMark(adfMark{Type: mark})
		} else {
			r.popMark()
		}

	case *ast.CodeSpan:
		if entering {
			text := string(n.Text(source))
			r.addInline(&adfNode{
				Type:  "text",
				Text:  text,
				Marks: []adfMark{{Type: "code"}},
			})
			return ast.WalkSkipChildren, nil
		}

	case *extAst.Strikethrough:
		if entering {
			r.pushMark(adfMark{Type: "strike"})
		} else {
			r.popMark()
		}

	case *ast.Link:
		if entering {
			attrs := map[string]any{"href": string(n.Destination)}
			if len(n.Title) > 0 {
				attrs["title"] = string(n.Title)
			}
			r.pushMark(adfMark{Type: "link", Attrs: attrs})
		} else {
			r.popMark()
		}

	case *ast.AutoLink:
		if entering {
			url := string(n.URL(source))
			r.addInline(&adfNode{
				Type:  "text",
				Text:  url,
				Marks: []adfMark{{Type: "link", Attrs: map[string]any{"href": url}}},
			})
			return ast.WalkSkipChildren, nil
		}

	case *ast.Image:
		if entering {
			alt := string(n.Text(source))
			url := string(n.Destination)
			display := alt
			if display == "" {
				display = url
			}
			display = "[Image: " + display + "]"
			r.addInline(&adfNode{
				Type:  "text",
				Text:  display,
				Marks: []adfMark{{Type: "link", Attrs: map[string]any{"href": url}}},
			})
			r.warn("remote image rendered as link (ADF has no external image embed): %s", url)
			return ast.WalkSkipChildren, nil
		}

	case *ast.FencedCodeBlock:
		if entering {
			lang := string(n.Language(source))
			var content string
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				segment := lines.At(i)
				content += string(segment.Value(source))
			}
			codeBlock := &adfNode{Type: "codeBlock"}
			if lang != "" {
				codeBlock.Attrs = map[string]any{"language": lang}
			}
			codeBlock.Content = []*adfNode{{Type: "text", Text: content}}
			r.current().Content = append(r.current().Content, codeBlock)
			return ast.WalkSkipChildren, nil
		}

	case *ast.CodeBlock:
		if entering {
			var content string
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				segment := lines.At(i)
				content += string(segment.Value(source))
			}
			codeBlock := &adfNode{
				Type:    "codeBlock",
				Content: []*adfNode{{Type: "text", Text: content}},
			}
			r.current().Content = append(r.current().Content, codeBlock)
			return ast.WalkSkipChildren, nil
		}

	case *ast.List:
		if entering {
			if !n.IsOrdered() && listHasTaskChildren(n) {
				// If we're inside a taskItem, pop it first so the nested
				// taskList becomes a sibling in the parent taskList (Jira
				// requires nested taskLists as siblings, not children).
				if r.current().Type == "taskItem" {
					r.pop()
				}
				r.push(&adfNode{
					Type:  "taskList",
					Attrs: map[string]any{"localId": ""},
				})
			} else if n.IsOrdered() {
				r.push(&adfNode{Type: "orderedList"})
			} else {
				r.push(&adfNode{Type: "bulletList"})
			}
		} else {
			r.pop()
		}

	case *ast.ListItem:
		if entering {
			if r.current().Type == "taskList" {
				r.push(&adfNode{
					Type:  "taskItem",
					Attrs: map[string]any{"localId": "", "state": "TODO"},
				})
			} else {
				r.push(&adfNode{Type: "listItem"})
			}
		} else {
			// Only pop if current is still a taskItem/listItem. When a
			// nested task list is encountered, the taskItem is popped early
			// (in the List handler above), so we must not double-pop.
			cur := r.current().Type
			if cur == "taskItem" || cur == "listItem" {
				r.pop()
			}
		}

	case *ast.Blockquote:
		if entering {
			if r.ancestorHasType("blockquote") {
				r.warn("nested blockquote flattened (not supported in ADF)")
			} else if r.parentType() == "listItem" {
				r.warn("blockquote inside list item flattened (not supported in ADF)")
			}
			if r.ancestorHasType("blockquote") || r.parentType() == "listItem" {
			} else {
				r.push(&adfNode{Type: "blockquote"})
			}
		} else {
			if !r.ancestorHasType("blockquote") {
				// Only pop if we actually pushed a blockquote (not flattened).
				// We need to check by looking at current node type.
			}
			if r.current().Type == "blockquote" {
				r.pop()
			}
		}

	case *ast.ThematicBreak:
		if entering {
			r.addInline(&adfNode{Type: "rule"})
		}

	case *ast.HTMLBlock:
		if entering {
			var content string
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				segment := lines.At(i)
				content += string(segment.Value(source))
			}
			if content != "" {
				r.push(&adfNode{Type: "paragraph"})
				r.addInline(&adfNode{Type: "text", Text: content})
				r.pop()
			}
			return ast.WalkSkipChildren, nil
		}

	case *ast.RawHTML:
		if entering {
			segments := n.Segments
			var content string
			for i := 0; i < segments.Len(); i++ {
				segment := segments.At(i)
				content += string(segment.Value(source))
			}
			if content != "" {
				r.addInline(&adfNode{Type: "text", Text: content})
			}
			return ast.WalkSkipChildren, nil
		}

	case *extAst.Table:
		if entering {
			for _, a := range n.Alignments {
				if a == extAst.AlignCenter || a == extAst.AlignRight {
					r.warn("table column alignment not supported in ADF (center/right alignment dropped)")
					break
				}
			}
			r.push(&adfNode{Type: "table"})
		} else {
			r.pop()
		}

	case *extAst.TableHeader:
		if entering {
			r.push(&adfNode{Type: "tableRow"})
		} else {
			r.pop()
		}

	case *extAst.TableRow:
		if entering {
			r.push(&adfNode{Type: "tableRow"})
		} else {
			r.pop()
		}

	case *extAst.TableCell:
		if entering {
			cellType := "tableCell"
			if _, ok := n.Parent().(*extAst.TableHeader); ok {
				cellType = "tableHeader"
			}
			r.push(&adfNode{Type: cellType})
			r.push(&adfNode{Type: "paragraph"})
		} else {
			r.pop() // paragraph
			r.pop() // cell
		}

	case *extAst.TaskCheckBox:
		if entering && n.IsChecked {
			for i := len(r.stack) - 1; i >= 0; i-- {
				if r.stack[i].Type == "taskItem" {
					r.stack[i].Attrs["state"] = "DONE"
					break
				}
			}
		}

	default:
		if entering {
			text := string(node.Text(source))
			if text != "" {
				textNode := &adfNode{Type: "text", Text: text}
				if marks := r.currentMarks(); marks != nil {
					textNode.Marks = marks
				}
				r.addInline(textNode)
			}
			r.warn("unsupported markdown element rendered as plain text: %s", node.Kind().String())
			return ast.WalkSkipChildren, nil
		}
	}

	return ast.WalkContinue, nil
}

func renderMarkdownToADF(source []byte) ([]byte, []string, error) {
	r := newADFRenderer()
	gm := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAttribute()),
		goldmark.WithRenderer(r),
	)

	var buf bytes.Buffer
	if err := gm.Convert(source, &buf); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), r.warnings, nil
}
