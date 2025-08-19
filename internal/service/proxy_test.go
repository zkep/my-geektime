package service

import (
	"strings"
	"testing"

	"github.com/zkep/my-geektime/internal/config"
	"github.com/zkep/my-geektime/internal/global"
)

// TestHtmlURLProxyReplace_VoidElements tests that void elements are correctly handled
func TestHtmlURLProxyReplace_VoidElements(t *testing.T) {
	// Initialize global config to avoid panic
	if global.CONF == nil {
		global.CONF = &config.Config{
			Site: config.Site{
				Proxy: config.Proxy{
					Urls:     []string{},
					ProxyUrl: "",
				},
			},
		}
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "br_tag_not_self_closed",
			input:    `<p>Line 1<br>Line 2</p>`,
			expected: `<p>Line 1<br/>Line 2</p>`,
		},
		{
			name:     "br_tag_self_closed",
			input:    `<p>Line 1<br/>Line 2</p>`,
			expected: `<p>Line 1<br/>Line 2</p>`,
		},
		{
			name:     "img_tag_not_self_closed",
			input:    `<p>Text<img src="test.jpg" alt="Test">More text</p>`,
			expected: `<p>Text<img src="test.jpg" alt="Test"/>More text</p>`,
		},
		{
			name:     "img_tag_self_closed",
			input:    `<p>Text<img src="test.jpg" alt="Test"/>More text</p>`,
			expected: `<p>Text<img src="test.jpg" alt="Test"/>More text</p>`,
		},
		{
			name:     "multiple_br_tags",
			input:    `<p>Line 1<br>Line 2<br>Line 3<br/></p>`,
			expected: `<p>Line 1<br/>Line 2<br/>Line 3<br/></p>`,
		},
		{
			name:  "mixed_void_elements",
			input: `<p>Text<br>More<img src="test.jpg"><hr>End</p>`,
			// Note: HTML parser auto-closes <p> before <hr> as per HTML spec
			// <hr> cannot be nested inside <p>, so parser restructures the DOM
			expected: `<p>Text<br/>More<img src="test.jpg"/></p><hr/>End<p></p>`,
		},
		{
			name:     "blockquote_with_br",
			input:    `<blockquote>Quote line 1<br>Quote line 2<br>Quote line 3</blockquote>`,
			expected: `<blockquote>Quote line 1<br/>Quote line 2<br/>Quote line 3</blockquote>`,
		},
		{
			name:     "real_world_html_from_test_file",
			input:    `<p>第一段。</p><blockquote><p>引用内容<br>第二行<br><img src="image.jpg" alt="图片"></p></blockquote>`,
			expected: `<p>第一段。</p><blockquote><p>引用内容<br/>第二行<br/><img src="image.jpg" alt="图片"/></p></blockquote>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Input:    %s", tt.input)

			result, err := HtmlURLProxyReplace(tt.input)
			if err != nil {
				t.Fatalf("HtmlURLProxyReplace() error = %v", err)
			}

			// Remove the HTML boilerplate that html.Parse adds
			// The parser wraps content in <html><head></head><body>...</body></html>
			result = extractBodyContent(result)

			t.Logf("Expected: %s", tt.expected)
			t.Logf("Got:      %s", result)

			// Special check for the problematic pattern
			if strings.Contains(result, ">>") {
				t.Errorf("Result contains problematic '>>'")
			}

			if result != tt.expected {
				t.Errorf("Result mismatch")
			}
		})
	}
}

// Helper function to extract body content from the full HTML document
func extractBodyContent(html string) string {
	start := strings.Index(html, "<body>")
	end := strings.Index(html, "</body>")
	if start != -1 && end != -1 {
		return html[start+6 : end]
	}
	return html
}
