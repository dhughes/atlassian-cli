package atlassian

import (
	"fmt"
	"strings"
)

// ADFToText converts Atlassian Document Format (ADF) to plain text with basic formatting
// ADF is used by both Jira and Confluence for rich text content
func ADFToText(adf interface{}) string {
	if adf == nil {
		return ""
	}

	doc, ok := adf.(map[string]interface{})
	if !ok {
		return ""
	}

	var sb strings.Builder
	processNode(doc, &sb, 0)
	return strings.TrimSpace(sb.String())
}

func processNode(node map[string]interface{}, sb *strings.Builder, indent int) {
	nodeType, _ := node["type"].(string)
	content, _ := node["content"].([]interface{})

	switch nodeType {
	case "doc":
		// Root document node
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
			}
		}

	case "paragraph":
		writeIndent(sb, indent)
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
			}
		}
		sb.WriteString("\n")

	case "heading":
		level, _ := node["attrs"].(map[string]interface{})["level"].(float64)
		sb.WriteString("\n")
		writeIndent(sb, indent)
		// Add heading markers
		sb.WriteString(strings.Repeat("#", int(level)))
		sb.WriteString(" ")
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
			}
		}
		sb.WriteString("\n")

	case "text":
		text, _ := node["text"].(string)
		marks, _ := node["marks"].([]interface{})

		// Apply text formatting based on marks
		formatted := text
		for _, mark := range marks {
			if markMap, ok := mark.(map[string]interface{}); ok {
				markType, _ := markMap["type"].(string)
				switch markType {
				case "strong":
					formatted = "**" + formatted + "**"
				case "em":
					formatted = "*" + formatted + "*"
				case "code":
					formatted = "`" + formatted + "`"
				case "strike":
					formatted = "~~" + formatted + "~~"
				}
			}
		}
		sb.WriteString(formatted)

	case "bulletList":
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processListItem(childMap, sb, indent, "•")
			}
		}
		sb.WriteString("\n")

	case "orderedList":
		for i, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				marker := fmt.Sprintf("%d.", i+1)
				processListItem(childMap, sb, indent, marker)
			}
		}
		sb.WriteString("\n")

	case "listItem":
		// Handled by parent list nodes
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
			}
		}

	case "codeBlock":
		sb.WriteString("\n")
		writeIndent(sb, indent)
		sb.WriteString("```\n")
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
			}
		}
		writeIndent(sb, indent)
		sb.WriteString("```\n")

	case "blockquote":
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				writeIndent(sb, indent)
				sb.WriteString("> ")
				processNode(childMap, sb, indent)
			}
		}

	case "hardBreak":
		sb.WriteString("\n")

	case "rule":
		sb.WriteString("\n")
		writeIndent(sb, indent)
		sb.WriteString("---\n")

	case "mention":
		attrs, _ := node["attrs"].(map[string]interface{})
		text, _ := attrs["text"].(string)
		sb.WriteString("@")
		sb.WriteString(text)

	case "emoji":
		attrs, _ := node["attrs"].(map[string]interface{})
		shortName, _ := attrs["shortName"].(string)
		sb.WriteString(shortName)

	case "inlineCard", "blockCard":
		attrs, _ := node["attrs"].(map[string]interface{})
		url, _ := attrs["url"].(string)
		sb.WriteString(url)

	case "panel":
		attrs, _ := node["attrs"].(map[string]interface{})
		panelType, _ := attrs["panelType"].(string)
		sb.WriteString("\n")
		writeIndent(sb, indent)
		sb.WriteString(fmt.Sprintf("[%s]\n", strings.ToUpper(panelType)))
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent+2)
			}
		}
		sb.WriteString("\n")

	case "table":
		// Simple table rendering - just show content
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
			}
		}

	case "tableRow":
		writeIndent(sb, indent)
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
				sb.WriteString(" | ")
			}
		}
		sb.WriteString("\n")

	case "tableHeader", "tableCell":
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
			}
		}

	default:
		// For unknown nodes, process children if they exist
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				processNode(childMap, sb, indent)
			}
		}
	}
}

func processListItem(node map[string]interface{}, sb *strings.Builder, indent int, marker string) {
	content, _ := node["content"].([]interface{})
	writeIndent(sb, indent)
	sb.WriteString(marker)
	sb.WriteString(" ")

	for _, child := range content {
		if childMap, ok := child.(map[string]interface{}); ok {
			childType, _ := childMap["type"].(string)
			if childType == "paragraph" {
				// For paragraphs in list items, don't add extra newline
				childContent, _ := childMap["content"].([]interface{})
				for _, grandChild := range childContent {
					if grandChildMap, ok := grandChild.(map[string]interface{}); ok {
						processNode(grandChildMap, sb, indent)
					}
				}
			} else {
				processNode(childMap, sb, indent+2)
			}
		}
	}
	sb.WriteString("\n")
}

func writeIndent(sb *strings.Builder, indent int) {
	sb.WriteString(strings.Repeat(" ", indent))
}
