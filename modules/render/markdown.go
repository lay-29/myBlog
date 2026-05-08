package render

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

var sanitizer = bluemonday.UGCPolicy()

func MarkdownToSafeHTML(input string) (string, error) {
	var buf bytes.Buffer
	if err := markdown.Convert([]byte(input), &buf); err != nil {
		return "", err
	}
	return sanitizer.Sanitize(buf.String()), nil
}
