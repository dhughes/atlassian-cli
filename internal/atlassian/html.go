package atlassian

import (
	"regexp"
	"strings"
)

// HTMLToText converts Confluence HTML storage format to readable plain text
// This is a simple converter - not perfect but makes content readable
func HTMLToText(html string) string {
	if html == "" {
		return ""
	}

	text := html

	// Replace headings with markdown-style headers
	text = regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`).ReplaceAllString(text, "\n# $1\n")
	text = regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`).ReplaceAllString(text, "\n## $1\n")
	text = regexp.MustCompile(`<h3[^>]*>(.*?)</h3>`).ReplaceAllString(text, "\n### $1\n")
	text = regexp.MustCompile(`<h4[^>]*>(.*?)</h4>`).ReplaceAllString(text, "\n#### $1\n")
	text = regexp.MustCompile(`<h5[^>]*>(.*?)</h5>`).ReplaceAllString(text, "\n##### $1\n")
	text = regexp.MustCompile(`<h6[^>]*>(.*?)</h6>`).ReplaceAllString(text, "\n###### $1\n")

	// Replace paragraphs with newlines
	text = regexp.MustCompile(`<p[^>]*>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`</p>`).ReplaceAllString(text, "\n")

	// Replace breaks
	text = regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(text, "\n")

	// Replace list items
	text = regexp.MustCompile(`<li[^>]*>`).ReplaceAllString(text, "\n• ")
	text = regexp.MustCompile(`</li>`).ReplaceAllString(text, "")

	// Replace unordered lists
	text = regexp.MustCompile(`<ul[^>]*>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`</ul>`).ReplaceAllString(text, "\n")

	// Replace ordered lists
	text = regexp.MustCompile(`<ol[^>]*>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`</ol>`).ReplaceAllString(text, "\n")

	// Replace code blocks
	text = regexp.MustCompile(`<pre[^>]*><code[^>]*>(.*?)</code></pre>`).ReplaceAllString(text, "\n```\n$1\n```\n")
	text = regexp.MustCompile(`<pre[^>]*>(.*?)</pre>`).ReplaceAllString(text, "\n```\n$1\n```\n")

	// Replace inline code
	text = regexp.MustCompile(`<code[^>]*>(.*?)</code>`).ReplaceAllString(text, "`$1`")

	// Replace strong/bold
	text = regexp.MustCompile(`<strong[^>]*>(.*?)</strong>`).ReplaceAllString(text, "**$1**")
	text = regexp.MustCompile(`<b[^>]*>(.*?)</b>`).ReplaceAllString(text, "**$1**")

	// Replace emphasis/italic
	text = regexp.MustCompile(`<em[^>]*>(.*?)</em>`).ReplaceAllString(text, "*$1*")
	text = regexp.MustCompile(`<i[^>]*>(.*?)</i>`).ReplaceAllString(text, "*$1*")

	// Replace links - show URL
	text = regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>`).ReplaceAllString(text, "$2 ($1)")

	// Replace images - just show alt text or URL
	text = regexp.MustCompile(`<img[^>]*alt="([^"]*)"[^>]*>`).ReplaceAllString(text, "[Image: $1]")
	text = regexp.MustCompile(`<img[^>]*src="([^"]*)"[^>]*>`).ReplaceAllString(text, "[Image: $1]")

	// Handle table elements - basic rendering
	text = regexp.MustCompile(`<table[^>]*>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`</table>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`<tr[^>]*>`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`</tr>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`<th[^>]*>`).ReplaceAllString(text, "| ")
	text = regexp.MustCompile(`</th>`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`<td[^>]*>`).ReplaceAllString(text, "| ")
	text = regexp.MustCompile(`</td>`).ReplaceAllString(text, " ")

	// Handle divs and spans - just remove tags
	text = regexp.MustCompile(`<div[^>]*>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`</div>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`<span[^>]*>`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`</span>`).ReplaceAllString(text, "")

	// Handle blockquotes
	text = regexp.MustCompile(`<blockquote[^>]*>`).ReplaceAllString(text, "\n> ")
	text = regexp.MustCompile(`</blockquote>`).ReplaceAllString(text, "\n")

	// Remove remaining HTML tags
	text = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	// Clean up excessive whitespace
	text = regexp.MustCompile(`\n\s*\n\s*\n+`).ReplaceAllString(text, "\n\n")
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")

	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)

	return text
}
