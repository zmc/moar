package m

import (
	"fmt"
	"unicode"

	"github.com/walles/moar/twin"
)

// From: https://www.compart.com/en/unicode/U+00A0
const NO_BREAK_SPACE = '\xa0'

func getWrapWidth(line []twin.Cell, maxWrapWidth int) int {
	if len(line) <= maxWrapWidth {
		panic(fmt.Errorf("cannot compute wrap width when input isn't longer than max (%d<=%d)",
			len(line), maxWrapWidth))
	}

	// Find the last whitespace in the input. Since we want to break *before*
	// whitespace, we loop through characters to the right of the current one.
	for nextIndex := maxWrapWidth; nextIndex > 0; nextIndex-- {
		char := line[nextIndex].Rune
		if !unicode.IsSpace(char) {
			// Want to break before whitespace, this is not it, keep looking
			continue
		}

		if char == NO_BREAK_SPACE {
			// Don't break at non-break whitespace
			continue
		}

		return nextIndex
	}

	// No breakpoint found, give up
	return maxWrapWidth
}

func wrapLine(width int, line []twin.Cell) [][]twin.Cell {
	if len(line) == 0 {
		return [][]twin.Cell{{}}
	}

	// Trailing space risks showing up by itself on a line, which would just
	// look weird.
	line = twin.TrimSpaceRight(line)

	wrapped := make([][]twin.Cell, 0, len(line)/width)
	for len(line) > width {
		wrapWidth := getWrapWidth(line, width)
		firstPart := line[:wrapWidth]
		if len(wrapped) > 0 {
			// Leading whitespace on wrapped lines would just look like
			// indentation, which would be weird for wrapped text.
			firstPart = twin.TrimSpaceLeft(firstPart)
		}

		wrapped = append(wrapped, twin.TrimSpaceRight(firstPart))

		line = twin.TrimSpaceLeft(line[wrapWidth:])
	}

	if len(wrapped) > 0 {
		// Leading whitespace on wrapped lines would just look like
		// indentation, which would be weird for wrapped text.
		line = twin.TrimSpaceLeft(line)
	}

	if len(line) > 0 {
		wrapped = append(wrapped, twin.TrimSpaceRight(line))
	}

	return wrapped
}
