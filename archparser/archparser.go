package archparser

import (
	"strings"

	"github.com/cbbond/newsie/termstyle"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

// Simple struct for modeling text as code, or not.
type textChunk struct {
	text   string // Anything goes here, could just be a newline.
	isCode bool   // Are we in a <code> </code> segment?
}

// All text in a post is stored here in (text, isCode) chunks.
var text []textChunk

// True when the tokenizer is currently in a code block.
var isCode bool

// handleStartTag is called when the tokenizer encounters a start tag "<...".
func handleStartTag(tokenizer *html.Tokenizer) {
	token := tokenizer.Token()

	if token.Data == "code" {
		isCode = true
	}
}

// handleEndTag is called when the tokenizer encounters an end tag "...>".
func handleEndTag(tokenizer *html.Tokenizer) {
	token := tokenizer.Token()

	if token.Data == "code" {
		isCode = false
	} else if token.Data == "p" {
		text = append(text, textChunk{"\n", isCode})
	}
}

// handleText is called when the tokenizer encounters text/data.
func handleText(tokenizer *html.Tokenizer) {
	token := tokenizer.Token()
	text = append(text, textChunk{token.Data, isCode})
}

// parsePost takes the given post description and parses text + code segments from it.
func parsePost(desc string) {
	reader := strings.NewReader(desc)
	tokenizer := html.NewTokenizer(reader)

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken { // EOF or actual error. Bail out either way.
			break
		}

		switch tokenType {
		case html.StartTagToken:
			handleStartTag(tokenizer)
		case html.EndTagToken:
			handleEndTag(tokenizer)
		case html.TextToken:
			handleText(tokenizer)
		default:
		}
	}
}

// MakePretty takes a gofeed Item and makes a pretty, formatted string from it.
func MakePretty(post *gofeed.Item) string {
	var prettyPost string

	isCode = false
	text = nil
	parsePost(post.Description)

	prettyPost += termstyle.StyleText(post.Title+"\n", []string{"bold", "red"})
	prettyPost += termstyle.StyleText(post.Published+"\n\n", []string{"blue"})

	for _, chunk := range text {
		if chunk.isCode {
			prettyPost += termstyle.StyleText(chunk.text, []string{"green"})
		} else {
			prettyPost += chunk.text
		}
	}

	prettyPost += termstyle.StyleText("\n"+post.Link, []string{"underline", "blue"})
	return prettyPost
}
