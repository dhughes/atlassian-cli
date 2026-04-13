package atlassian

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// MarkdownToADF converts markdown text to Atlassian Document Format (ADF).
// It returns the ADF document, any conversion warnings (unsupported features
// that were rendered as plain text fallbacks), and an error if conversion fails.
func MarkdownToADF(markdown string) (map[string]any, []string, error) {
	b, warnings, err := renderMarkdownToADF([]byte(markdown))
	if err != nil {
		return nil, warnings, err
	}

	var adf map[string]any
	if err := json.Unmarshal(b, &adf); err != nil {
		return nil, warnings, err
	}

	return adf, warnings, nil
}

// ImageRef represents a local image reference found in markdown
type ImageRef struct {
	AltText  string // alt text from ![alt](path)
	FilePath string // local file path
	Original string // original markdown syntax e.g. ![alt](path)
}

// imageRegexp matches markdown image syntax: ![alt](path)
var imageRegexp = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

// obsidianImageRegexp matches Obsidian wiki-link image syntax: ![[filename]]
var obsidianImageRegexp = regexp.MustCompile(`!\[\[([^\]]+)\]\]`)

// imagePlaceholderPrefix is used to create unique indexed placeholders that can
// be located in the ADF tree after markdown conversion, allowing media nodes to
// be inserted at the correct position rather than appended at the end.
const imagePlaceholderPrefix = "ATLIMG_PLACEHOLDER_"

// ExtractLocalImages finds local image references in markdown and returns them
// along with the markdown text with image references replaced by indexed
// placeholder text. URLs (http/https) are left untouched.
// Each placeholder is indexed (ATLIMG_PLACEHOLDER_0, ATLIMG_PLACEHOLDER_1, etc.)
// so they can be mapped back to specific media nodes for in-place insertion.
func ExtractLocalImages(markdown string) ([]ImageRef, string) {
	var images []ImageRef
	cleaned := markdown

	type localMatch struct {
		start, end int
		altText    string
		path       string
		fullMatch  string
	}
	var locals []localMatch

	// Standard markdown images: ![alt](path)
	for _, m := range imageRegexp.FindAllStringSubmatchIndex(markdown, -1) {
		altText := markdown[m[2]:m[3]]
		path := markdown[m[4]:m[5]]

		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}

		locals = append(locals, localMatch{
			start:     m[0],
			end:       m[1],
			altText:   altText,
			path:      path,
			fullMatch: markdown[m[0]:m[1]],
		})
	}

	// Obsidian wiki-link images: ![[filename]]
	for _, m := range obsidianImageRegexp.FindAllStringSubmatchIndex(markdown, -1) {
		path := markdown[m[2]:m[3]]

		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}

		locals = append(locals, localMatch{
			start:     m[0],
			end:       m[1],
			altText:   filepath.Base(path),
			path:      path,
			fullMatch: markdown[m[0]:m[1]],
		})
	}

	// Sort by start position so indices are assigned in document order
	sort.Slice(locals, func(i, j int) bool {
		return locals[i].start < locals[j].start
	})

	images = make([]ImageRef, len(locals))
	for i, lm := range locals {
		images[i] = ImageRef{
			AltText:  lm.altText,
			FilePath: lm.path,
			Original: lm.fullMatch,
		}
	}

	// Replace in reverse order to preserve indices
	for i := len(locals) - 1; i >= 0; i-- {
		lm := locals[i]
		placeholder := fmt.Sprintf("%s%d", imagePlaceholderPrefix, i)
		cleaned = cleaned[:lm.start] + placeholder + cleaned[lm.end:]
	}

	return images, cleaned
}

// BuildMediaSingleNode builds an ADF mediaSingle node for an inline image
func BuildMediaSingleNode(mediaID, altText string) map[string]any {
	mediaAttrs := map[string]any{
		"id":         mediaID,
		"type":       "file",
		"collection": "",
	}
	if altText != "" {
		mediaAttrs["alt"] = altText
	}

	return map[string]any{
		"type": "mediaSingle",
		"attrs": map[string]any{
			"layout": "center",
		},
		"content": []any{
			map[string]any{
				"type":  "media",
				"attrs": mediaAttrs,
			},
		},
	}
}

// MarkdownToADFWithImages converts markdown to ADF and inserts media nodes at
// the positions where image placeholders appear, rather than appending them at
// the end. The mediaNodes slice must be indexed to match the placeholders
// produced by ExtractLocalImages (i.e., mediaNodes[0] corresponds to
// ATLIMG_PLACEHOLDER_0, etc.).
func MarkdownToADFWithImages(markdown string, mediaNodes []map[string]any) (map[string]any, []string, error) {
	images, cleaned := ExtractLocalImages(markdown)
	_ = images // images are used by the caller to upload files

	adf, warnings, err := MarkdownToADF(cleaned)
	if err != nil {
		return nil, warnings, err
	}

	if len(mediaNodes) > 0 {
		adf["content"] = replacePlaceholdersInContent(adf["content"], mediaNodes)
	}

	return adf, warnings, nil
}

// replacePlaceholdersInContent walks the top-level ADF content array and
// replaces any block-level node that contains an image placeholder with the
// corresponding mediaSingle node. It handles placeholders that appear as
// standalone paragraphs, inside list items, or nested in other block structures.
func replacePlaceholdersInContent(rawContent any, mediaNodes []map[string]any) []any {
	content, ok := rawContent.([]any)
	if !ok {
		return nil
	}

	var result []any
	for _, item := range content {
		node, ok := item.(map[string]any)
		if !ok {
			result = append(result, item)
			continue
		}

		// Check if this node (paragraph, etc.) contains a placeholder
		idx, found := findPlaceholderIndex(node)
		if found && idx < len(mediaNodes) && mediaNodes[idx] != nil {
			result = append(result, mediaNodes[idx])
			continue
		}

		// Recurse into block nodes that have content arrays (lists, blockquotes, etc.)
		nodeType, _ := node["type"].(string)
		switch nodeType {
		case "bulletList", "orderedList":
			node["content"] = replaceInListItems(node["content"], mediaNodes)
		case "blockquote", "panel", "expand", "layoutSection", "layoutColumn":
			node["content"] = replacePlaceholdersInContent(node["content"], mediaNodes)
		case "listItem":
			node["content"] = replacePlaceholdersInContent(node["content"], mediaNodes)
		}

		result = append(result, node)
	}
	return result
}

// replaceInListItems recurses into list items to replace placeholders
func replaceInListItems(rawContent any, mediaNodes []map[string]any) []any {
	content, ok := rawContent.([]any)
	if !ok {
		return nil
	}

	var result []any
	for _, item := range content {
		node, ok := item.(map[string]any)
		if !ok {
			result = append(result, item)
			continue
		}
		// List items contain content arrays with paragraphs, nested lists, etc.
		node["content"] = replacePlaceholdersInContent(node["content"], mediaNodes)
		result = append(result, node)
	}
	return result
}

// placeholderRegexp matches ATLIMG_PLACEHOLDER_N and captures the index
var placeholderRegexp = regexp.MustCompile(`^` + imagePlaceholderPrefix + `(\d+)$`)

// findPlaceholderIndex checks if a node is a paragraph whose concatenated text
// content matches an image placeholder. The markdown-to-ADF renderer may split
// the placeholder text across multiple text nodes, so we concatenate all text
// children before matching. Returns the placeholder index and true if found.
func findPlaceholderIndex(node map[string]any) (int, bool) {
	nodeType, _ := node["type"].(string)
	if nodeType != "paragraph" {
		return 0, false
	}

	content, ok := node["content"].([]any)
	if !ok || len(content) == 0 {
		return 0, false
	}

	// Concatenate all text nodes; bail if any non-text node is present
	var sb strings.Builder
	for _, child := range content {
		childMap, ok := child.(map[string]any)
		if !ok {
			return 0, false
		}
		if childMap["type"] != "text" {
			return 0, false
		}
		text, _ := childMap["text"].(string)
		sb.WriteString(text)
	}

	combined := strings.TrimSpace(sb.String())

	matches := placeholderRegexp.FindStringSubmatch(combined)
	if len(matches) < 2 {
		return 0, false
	}

	var idx int
	fmt.Sscanf(matches[1], "%d", &idx)
	return idx, true
}
