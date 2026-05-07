package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHtmlToText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strips tags",
			input: "<p>Hello <b>world</b></p>",
			want:  "Hello world",
		},
		{
			name:  "removes script block",
			input: `<html><script>alert("x")</script><p>content</p></html>`,
			want:  "content",
		},
		{
			name:  "removes style block",
			input: `<html><style>.foo { color: red }</style><p>text</p></html>`,
			want:  "text",
		},
		{
			name:  "decodes entities",
			input: `<p>AT&amp;T &lt;3</p>`,
			want:  "AT&T <3",
		},
		{
			name:  "collapses whitespace",
			input: "<p>foo   bar</p>",
			want:  "foo bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, htmlToText(tt.input))
		})
	}
}
