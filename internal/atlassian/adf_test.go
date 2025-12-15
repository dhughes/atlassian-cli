package atlassian

import (
	"strings"
	"testing"
)

func TestADFToText_Nil(t *testing.T) {
	result := ADFToText(nil)
	if result != "" {
		t.Errorf("Expected empty string for nil input, got %q", result)
	}
}

func TestADFToText_InvalidType(t *testing.T) {
	result := ADFToText("not a map")
	if result != "" {
		t.Errorf("Expected empty string for invalid type, got %q", result)
	}
}

func TestADFToText_SimpleParagraph(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Hello world",
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	expected := "Hello world"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestADFToText_Heading(t *testing.T) {
	tests := []struct {
		name     string
		level    float64
		text     string
		expected string
	}{
		{"H1", 1, "Title", "# Title"},
		{"H2", 2, "Subtitle", "## Subtitle"},
		{"H3", 3, "Section", "### Section"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adf := map[string]interface{}{
				"type": "doc",
				"content": []interface{}{
					map[string]interface{}{
						"type": "heading",
						"attrs": map[string]interface{}{
							"level": tt.level,
						},
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": tt.text,
							},
						},
					},
				},
			}

			result := ADFToText(adf)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestADFToText_TextFormatting(t *testing.T) {
	tests := []struct {
		name     string
		markType string
		text     string
		expected string
	}{
		{"Bold", "strong", "bold text", "**bold text**"},
		{"Italic", "em", "italic text", "*italic text*"},
		{"Code", "code", "code text", "`code text`"},
		{"Strike", "strike", "strike text", "~~strike text~~"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adf := map[string]interface{}{
				"type": "doc",
				"content": []interface{}{
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": tt.text,
								"marks": []interface{}{
									map[string]interface{}{
										"type": tt.markType,
									},
								},
			},
						},
					},
				},
			}

			result := ADFToText(adf)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestADFToText_BulletList(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "bulletList",
				"content": []interface{}{
					map[string]interface{}{
						"type": "listItem",
						"content": []interface{}{
							map[string]interface{}{
								"type": "paragraph",
								"content": []interface{}{
									map[string]interface{}{
										"type": "text",
										"text": "First item",
									},
								},
							},
						},
					},
					map[string]interface{}{
						"type": "listItem",
						"content": []interface{}{
							map[string]interface{}{
								"type": "paragraph",
								"content": []interface{}{
									map[string]interface{}{
										"type": "text",
										"text": "Second item",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, "• First item") {
		t.Errorf("Expected bullet list to contain '• First item', got %q", result)
	}
	if !strings.Contains(result, "• Second item") {
		t.Errorf("Expected bullet list to contain '• Second item', got %q", result)
	}
}

func TestADFToText_OrderedList(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "orderedList",
				"content": []interface{}{
					map[string]interface{}{
						"type": "listItem",
						"content": []interface{}{
							map[string]interface{}{
								"type": "paragraph",
								"content": []interface{}{
									map[string]interface{}{
										"type": "text",
										"text": "First item",
									},
								},
							},
						},
					},
					map[string]interface{}{
						"type": "listItem",
						"content": []interface{}{
							map[string]interface{}{
								"type": "paragraph",
								"content": []interface{}{
									map[string]interface{}{
										"type": "text",
										"text": "Second item",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, "1. First item") {
		t.Errorf("Expected ordered list to contain '1. First item', got %q", result)
	}
	if !strings.Contains(result, "2. Second item") {
		t.Errorf("Expected ordered list to contain '2. Second item', got %q", result)
	}
}

func TestADFToText_CodeBlock(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "codeBlock",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "func main() {\n  fmt.Println(\"Hello\")\n}",
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, "```") {
		t.Errorf("Expected code block to contain '```', got %q", result)
	}
	if !strings.Contains(result, "func main()") {
		t.Errorf("Expected code block to contain code, got %q", result)
	}
}

func TestADFToText_Blockquote(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "blockquote",
				"content": []interface{}{
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Quoted text",
							},
						},
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, "> ") {
		t.Errorf("Expected blockquote to contain '> ', got %q", result)
	}
	if !strings.Contains(result, "Quoted text") {
		t.Errorf("Expected blockquote to contain text, got %q", result)
	}
}

func TestADFToText_HardBreak(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Line 1",
					},
					map[string]interface{}{
						"type": "hardBreak",
					},
					map[string]interface{}{
						"type": "text",
						"text": "Line 2",
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, "Line 1") || !strings.Contains(result, "Line 2") {
		t.Errorf("Expected both lines in output, got %q", result)
	}
}

func TestADFToText_Rule(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "rule",
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, "---") {
		t.Errorf("Expected rule to contain '---', got %q", result)
	}
}

func TestADFToText_Mention(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "mention",
						"attrs": map[string]interface{}{
							"text": "John Doe",
						},
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, "@John Doe") {
		t.Errorf("Expected mention to contain '@John Doe', got %q", result)
	}
}

func TestADFToText_Emoji(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "emoji",
						"attrs": map[string]interface{}{
							"shortName": ":smile:",
						},
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, ":smile:") {
		t.Errorf("Expected emoji to contain ':smile:', got %q", result)
	}
}

func TestADFToText_Panel(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "panel",
				"attrs": map[string]interface{}{
					"panelType": "info",
				},
				"content": []interface{}{
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Info message",
							},
						},
					},
				},
			},
		},
	}

	result := ADFToText(adf)
	if !strings.Contains(result, "[INFO]") {
		t.Errorf("Expected panel to contain '[INFO]', got %q", result)
	}
	if !strings.Contains(result, "Info message") {
		t.Errorf("Expected panel to contain message, got %q", result)
	}
}

func TestADFToText_Table(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "table",
				"content": []interface{}{
					map[string]interface{}{
						"type": "tableRow",
						"content": []interface{}{
							map[string]interface{}{
								"type": "tableHeader",
								"content": []interface{}{
									map[string]interface{}{
										"type": "paragraph",
										"content": []interface{}{
											map[string]interface{}{
												"type": "text",
												"text": "Header 1",
											},
										},
									},
								},
							},
							map[string]interface{}{
								"type": "tableHeader",
								"content": []interface{}{
									map[string]interface{}{
										"type": "paragraph",
										"content": []interface{}{
											map[string]interface{}{
												"type": "text",
												"text": "Header 2",
											},
										},
									},
								},
							},
						},
					},
					map[string]interface{}{
						"type": "tableRow",
						"content": []interface{}{
							map[string]interface{}{
								"type": "tableCell",
								"content": []interface{}{
									map[string]interface{}{
										"type": "paragraph",
										"content": []interface{}{
											map[string]interface{}{
												"type": "text",
												"text": "Cell 1",
											},
										},
									},
								},
							},
							map[string]interface{}{
								"type": "tableCell",
								"content": []interface{}{
									map[string]interface{}{
										"type": "paragraph",
										"content": []interface{}{
											map[string]interface{}{
												"type": "text",
												"text": "Cell 2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := ADFToText(adf)
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

func TestADFToText_ComplexDocument(t *testing.T) {
	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "heading",
				"attrs": map[string]interface{}{
					"level": float64(1),
				},
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Document Title",
					},
				},
			},
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "This is ",
					},
					map[string]interface{}{
						"type": "text",
						"text": "bold",
						"marks": []interface{}{
							map[string]interface{}{
								"type": "strong",
							},
						},
					},
					map[string]interface{}{
						"type": "text",
						"text": " text.",
					},
				},
			},
			map[string]interface{}{
				"type": "bulletList",
				"content": []interface{}{
					map[string]interface{}{
						"type": "listItem",
						"content": []interface{}{
							map[string]interface{}{
								"type": "paragraph",
								"content": []interface{}{
									map[string]interface{}{
										"type": "text",
										"text": "Item 1",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := ADFToText(adf)

	// Check all parts are present
	if !strings.Contains(result, "# Document Title") {
		t.Errorf("Expected heading in output, got %q", result)
	}
	if !strings.Contains(result, "**bold**") {
		t.Errorf("Expected bold text in output, got %q", result)
	}
	if !strings.Contains(result, "• Item 1") {
		t.Errorf("Expected list item in output, got %q", result)
	}
}
