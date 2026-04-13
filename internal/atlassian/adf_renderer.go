package atlassian

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extAst "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
)

var _ renderer.Renderer = &adfRenderer{}

type adfRenderer struct {
	doc   *adfNode
	stack []*adfNode
}

type adfNode struct {
	Type    string            `json:"type"`
	Version int               `json:"version,omitempty"`
	Attrs   map[string]any    `json:"attrs,omitempty"`
	Content []*adfNode        `json:"content,omitempty"`
	Marks   []adfMark         `json:"marks,omitempty"`
	Text    string            `json:"text,omitempty"`
}

type adfMark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

func newADFRenderer() *adfRenderer {
	doc := &adfNode{Type: "doc", Version: 1}
	return &adfRenderer{
		doc:   doc,
		stack: []*adfNode{doc},
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
			text := string(n.Text(source))
			if len(text) > 0 {
				r.addInline(&adfNode{Type: "text", Text: text})
			}
			if n.SoftLineBreak() {
				r.addInline(&adfNode{Type: "text", Text: " "})
			}
			if n.HardLineBreak() {
				r.addInline(&adfNode{Type: "hardBreak"})
			}
		}

	case *ast.String:
		if entering {
			text := string(n.Text(source))
			if len(text) > 0 {
				r.addInline(&adfNode{Type: "text", Text: text})
			}
		}

	case *ast.TextBlock:
		if entering {
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
			text := string(n.Text(source))
			r.addInline(&adfNode{
				Type:  "text",
				Text:  text,
				Marks: []adfMark{{Type: mark}},
			})
			return ast.WalkSkipChildren, nil
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
			text := string(n.Text(source))
			r.addInline(&adfNode{
				Type:  "text",
				Text:  text,
				Marks: []adfMark{{Type: "strike"}},
			})
			return ast.WalkSkipChildren, nil
		}

	case *ast.Link:
		if entering {
			text := string(n.Text(source))
			attrs := map[string]any{"href": string(n.Destination)}
			if len(n.Title) > 0 {
				attrs["title"] = string(n.Title)
			}
			r.addInline(&adfNode{
				Type:  "text",
				Text:  text,
				Marks: []adfMark{{Type: "link", Attrs: attrs}},
			})
			return ast.WalkSkipChildren, nil
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
			text := string(n.Text(source))
			url := string(n.Destination)
			attrs := map[string]any{"href": url}
			if text != "" {
				attrs["alt"] = text
			}
			r.addInline(&adfNode{
				Type:  "text",
				Text:  text,
				Marks: []adfMark{{Type: "link", Attrs: attrs}},
			})
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
			attrs := map[string]any{}
			if lang != "" {
				attrs["language"] = lang
			}
			codeBlock := &adfNode{Type: "codeBlock"}
			if len(attrs) > 0 {
				codeBlock.Attrs = attrs
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
			listType := "bulletList"
			if n.IsOrdered() {
				listType = "orderedList"
			}
			r.push(&adfNode{Type: listType})
		} else {
			r.pop()
		}

	case *ast.ListItem:
		if entering {
			r.push(&adfNode{Type: "listItem"})
		} else {
			r.pop()
		}

	case *ast.Blockquote:
		if entering {
			r.push(&adfNode{Type: "blockquote"})
		} else {
			r.pop()
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
		if entering {
			state := "TODO"
			if n.IsChecked {
				state = "DONE"
			}
			listItem := r.current()
			if listItem.Type == "paragraph" {
				r.pop()
				parent := r.current()
				if parent.Type == "listItem" {
					if parent.Attrs == nil {
						parent.Attrs = map[string]any{}
					}
					parent.Attrs["state"] = state
					r.push(listItem)
				} else {
					r.push(listItem)
				}
			}
		}

	default:
		return ast.WalkContinue, fmt.Errorf("unsupported markdown node: %s", n.Kind().String())
	}

	return ast.WalkContinue, nil
}

func renderMarkdownToADF(source []byte) ([]byte, error) {
	r := newADFRenderer()
	gm := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAttribute()),
		goldmark.WithRenderer(r),
	)

	var buf bytes.Buffer
	if err := gm.Convert(source, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
