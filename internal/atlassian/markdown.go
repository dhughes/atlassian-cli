package atlassian

import (
	"bytes"
	"encoding/json"

	"github.com/summonio/markdown-to-adf/renderer"
)

// MarkdownToADF converts markdown text to Atlassian Document Format (ADF)
func MarkdownToADF(markdown string) (map[string]interface{}, error) {
	var buf bytes.Buffer

	// Convert markdown to ADF using the renderer
	if err := renderer.Render(&buf, []byte(markdown)); err != nil {
		return nil, err
	}

	// Parse the ADF JSON
	var adf map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &adf); err != nil {
		return nil, err
	}

	return adf, nil
}
