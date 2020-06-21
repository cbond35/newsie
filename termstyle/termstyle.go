package termstyle

// Prefix for all terminal escape sequences.
var prefix string = "\033["

// Translates readable formats to terminal escape sequences.
var formatter = map[string]string{
	"blue":      prefix + "94m",
	"bold":      prefix + "1m",
	"end":       prefix + "0m",
	"green":     prefix + "92m",
	"red":       prefix + "91m",
	"underline": prefix + "4m",
	"yellow":    prefix + "93m",
}

// StyleText returns a stylized version of text using
// styles.
func StyleText(text string, styles []string) string {
	for _, style := range styles {
		text = formatter[style] + text
	}

	return text + formatter["end"]
}
