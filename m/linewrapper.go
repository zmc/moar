package m

import (
	"fmt"

	"github.com/walles/moar/twin"
)

func getWrapWidth(line []twin.Cell, maxWrapWidth int) int {
	if len(line) <= maxWrapWidth {
		panic(fmt.Errorf("cannot compute wrap width when input isn't longer than max (%d<=%d)",
			len(line), maxWrapWidth))
	}

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
