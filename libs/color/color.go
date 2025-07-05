package color

import (
	"fmt"
	"strings"
)

const (
	black = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

func Black(text string) string {
	return Col(text, black)
}

func Red(text string) string {
	return Col(text, red)
}

func Green(text string) string {
	return Col(text, green)
}

func Yellow(text string) string {
	return Col(text, yellow)
}

func Blue(text string) string {
	return Col(text, blue)
}

func Magenta(text string) string {
	return Col(text, magenta)
}

func Cyan(text string) string {
	return Col(text, cyan)
}

func White(text string) string {
	return Col(text, white)
}

var ColorEnabled = true

func Col(text string, col int) string {
	if !ColorEnabled {
		return text
	}
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 0, col, text, 0x1B)
}

var rainbowCols = []func(string) string{Red, Yellow, Green, Cyan, Blue, Magenta}

func Rainbow(text string) string {
	var builder strings.Builder
	for i := 0; i < len(text); i++ {
		fn := rainbowCols[i%len(rainbowCols)]
		builder.WriteString(fn(text[i : i+1]))
	}
	return builder.String()
}
