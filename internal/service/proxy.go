package service

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/zkep/mygeektime/internal/global"
	"golang.org/x/net/html"
)

func PorxyMatch(uri string) bool {
	for _, p := range global.CONF.Site.Proxy.Urls {
		if strings.HasPrefix(uri, p) {
			return true
		}
	}
	return false
}

func URLProxyReplace(uri string) string {
	if PorxyMatch(uri) {
		if strings.Contains(global.CONF.Site.Proxy.ProxyUrl, "{url}") {
			uri = strings.Replace(global.CONF.Site.Proxy.ProxyUrl, "{url}", uri, 1)
		} else {
			uri = fmt.Sprintf("%s?url=%s", global.CONF.Site.Proxy.ProxyUrl, uri)
		}
	}
	return uri
}

var voidElements = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"keygen": true, // "keygen" has been removed from the spec, but are kept here for backwards compatibility.
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

const escapedChars = "&'<>\"\r"

func escape(w *strings.Builder, s string) error {
	i := strings.IndexAny(s, escapedChars)
	for i != -1 {
		if _, err := w.WriteString(s[:i]); err != nil {
			return err
		}
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
			esc = "&#39;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			// "&#34;" is shorter than "&quot;".
			esc = "&#34;"
		case '\r':
			esc = "&#13;"
		default:
			panic("unrecognized escape character")
		}
		s = s[i+1:]
		if _, err := w.WriteString(esc); err != nil {
			return err
		}
		i = strings.IndexAny(s, escapedChars)
	}
	_, err := w.WriteString(s)
	return err
}

func childTextNodesAreLiteral(n *html.Node) bool {
	if n.Namespace != "" {
		return false
	}
	switch n.Data {
	case "iframe", "noembed", "noframes", "noscript", "plaintext", "script", "style", "xmp":
		return true
	default:
		return false
	}
}

func OutputHTML(n *html.Node) string {
	var output func(*strings.Builder, *html.Node)
	output = func(b *strings.Builder, n *html.Node) {
		switch n.Type {
		case html.DocumentNode:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				output(b, c)
			}
			return
		case html.ElementNode:
			if err := b.WriteByte('<'); err != nil {
				return
			}
			if _, err := b.WriteString(n.Data); err != nil {
				return
			}
			for _, a := range n.Attr {
				if err := b.WriteByte(' '); err != nil {
					return
				}
				if a.Namespace != "" {
					if _, err := b.WriteString(a.Namespace); err != nil {
						return
					}
					if err := b.WriteByte(':'); err != nil {
						return
					}
				}
				if _, err := b.WriteString(a.Key); err != nil {
					return
				}
				if _, err := b.WriteString(`="`); err != nil {
					return
				}
				if a.Key == "href" || a.Key == "src" {
					a.Val = URLProxyReplace(a.Val)
				}
				if err := escape(b, a.Val); err != nil {
					return
				}
				if err := b.WriteByte('"'); err != nil {
					return
				}
			}
			if voidElements[n.Data] {
				if n.FirstChild != nil {
					return
				}
				_, err := b.WriteString("/>")
				if err != nil {
					return
				}
			}

			if err := b.WriteByte('>'); err != nil {
				return
			}
			// Add initial newline where there is danger of a newline beging ignored.
			if c := n.FirstChild; c != nil && c.Type == html.TextNode && strings.HasPrefix(c.Data, "\n") {
				switch n.Data {
				case "pre", "listing", "textarea":
					if err := b.WriteByte('\n'); err != nil {
						return
					}
				}
			}
			// Render any child nodes
			if childTextNodesAreLiteral(n) {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						if _, err := b.WriteString(c.Data); err != nil {
							return
						}
					} else {
						output(b, c)
					}
				}
				if n.Data == "plaintext" {
					return
				}
			} else {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					output(b, c)
				}
			}

			// Render the </xxx> closing tag.
			if _, err := b.WriteString("</"); err != nil {
				return
			}
			if _, err := b.WriteString(n.Data); err != nil {
				return
			}
			b.WriteByte('>')
			return
		default:
			_ = html.Render(b, n)
			return
		}
	}

	var b strings.Builder
	output(&b, n)
	return b.String()
}

func HtmlURLProxyReplace(rawHtml string) (string, error) {
	node, err := html.Parse(bytes.NewBufferString(rawHtml))
	if err != nil {
		return "", err
	}
	output := OutputHTML(node)
	return output, nil
}
