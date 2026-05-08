package render

import (
	"strings"
	"testing"
)

func TestMarkdownToSafeHTML(t *testing.T) {
	html, err := MarkdownToSafeHTML("# 标题\n\n<script>alert(1)</script>\n\n**粗体**")
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}
	if strings.Contains(html, "<script>") {
		t.Fatalf("unsafe script was not removed: %s", html)
	}
	if !strings.Contains(html, "<strong>") {
		t.Fatalf("expected markdown formatting: %s", html)
	}
}
