package twin

import (
	"fmt"
	"math"

	"github.com/alecthomas/chroma/v2"
)

// Create using NewColor16(), NewColor256 or NewColor24Bit(), or use
// ColorDefault.
type Color uint32
type ColorType uint8

const (
	// Default foreground / background color
	ColorTypeDefault ColorType = iota

	// https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit
	//
	// Note that this type is only used for output, on input we store 3 bit
	// colors as 4 bit colors since they map to the same values.
	ColorType8

	// https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit
	ColorType16

	// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
	ColorType256

	// RGB: https://en.wikipedia.org/wiki/ANSI_escape_code#24-bit
	ColorType24bit
)

// Reset to default foreground / background color
var ColorDefault = newColor(ColorTypeDefault, 0)

// From: https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit
var colorNames16 = map[int]string{
	0:  "0 black",
	1:  "1 red",
	2:  "2 green",
	3:  "3 yellow (orange)",
	4:  "4 blue",
	5:  "5 magenta",
	6:  "6 cyan",
	7:  "7 white (light gray)",
	8:  "8 bright black (dark gray)",
	9:  "9 bright red",
	10: "10 bright green",
	11: "11 bright yellow",
	12: "12 bright blue",
	13: "13 bright magenta",
	14: "14 bright cyan",
	15: "15 bright white",
}

func newColor(colorType ColorType, value uint32) Color {
	return Color(value | (uint32(colorType) << 24))
}

// Four bit colors as defined here:
// https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit
func NewColor16(colorNumber0to15 int) Color {
	return newColor(ColorType16, uint32(colorNumber0to15))
}

func NewColor256(colorNumber uint8) Color {
	return newColor(ColorType256, uint32(colorNumber))
}

func NewColor24Bit(red uint8, green uint8, blue uint8) Color {
	return newColor(ColorType24bit, (uint32(red)<<16)+(uint32(green)<<8)+(uint32(blue)<<0))
}

func NewColorHex(rgb uint32) Color {
	return newColor(ColorType24bit, rgb)
}

func (color Color) ColorType() ColorType {
	return ColorType(color >> 24)
}

func (color Color) colorValue() uint32 {
	return uint32(color & 0xff_ff_ff)
}

// Render color into an ANSI string.
//
// Ref: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
func (color Color) ansiString(foreground bool, terminalColorCount ColorType) string {
	fgBgMarker := "3"
	if !foreground {
		fgBgMarker = "4"
	}

	if color.ColorType() == ColorTypeDefault {
		return fmt.Sprint("\x1b[", fgBgMarker, "9m")
	}

	color = color.downsampleTo(terminalColorCount)

	if color.ColorType() == ColorType16 {
		value := color.colorValue()
		if value < 8 {
			return fmt.Sprint("\x1b[", fgBgMarker, value, "m")
		} else if value <= 15 {
			fgBgMarker := "9"
			if !foreground {
				fgBgMarker = "10"
			}
			return fmt.Sprint("\x1b[", fgBgMarker, value-8, "m")
		}
	}

	if color.ColorType() == ColorType256 {
		value := color.colorValue()
		if value <= 255 {
			return fmt.Sprint("\x1b[", fgBgMarker, "8;5;", value, "m")
		}
	}

	if color.ColorType() == ColorType24bit {
		value := color.colorValue()
		red := (value & 0xff0000) >> 16
		green := (value & 0xff00) >> 8
		blue := value & 0xff

		return fmt.Sprint("\x1b[", fgBgMarker, "8;2;", red, ";", green, ";", blue, "m")
	}

	panic(fmt.Errorf("unhandled color type=%d %s", color.ColorType(), color.String()))
}

func (color Color) ForegroundAnsiString(terminalColorCount ColorType) string {
	// FIXME: Test this function with all different color types.
	return color.ansiString(true, terminalColorCount)
}

func (color Color) BackgroundAnsiString(terminalColorCount ColorType) string {
	// FIXME: Test this function with all different color types.
	return color.ansiString(false, terminalColorCount)
}

func (color Color) String() string {
	switch color.ColorType() {
	case ColorTypeDefault:
		return "Default color"

	case ColorType16:
		return colorNames16[int(color.colorValue())]

	case ColorType256:
		if color.colorValue() < 16 {
			return colorNames16[int(color.colorValue())]
		}
		return fmt.Sprintf("#%02x", color.colorValue())

	case ColorType24bit:
		return fmt.Sprintf("#%06x", color.colorValue())
	}

	panic(fmt.Errorf("unhandled color type %d", color.ColorType()))
}

func (color Color) to24Bit() Color {
	if color.ColorType() == ColorType24bit {
		return color
	}

	if color.ColorType() == ColorType8 || color.ColorType() == ColorType16 || color.ColorType() == ColorType256 {
		r0, g0, b0 := color256ToRGB(uint8(color.colorValue()))
		return NewColor24Bit(r0, g0, b0)
	}

	panic(fmt.Errorf("unhandled color type %d", color.ColorType()))
}

func (color Color) downsampleTo(terminalColorCount ColorType) Color {
	if color.ColorType() == ColorTypeDefault || terminalColorCount == ColorTypeDefault {
		panic(fmt.Errorf("downsampling to or from default color not supported, %s -> %#v", color.String(), terminalColorCount))
	}

	if color.ColorType() <= terminalColorCount {
		// Already low enough
		return color
	}

	target := color.to24Bit()

	// Find the closest match in the terminal color palette
	scanRange := 255
	switch terminalColorCount {
	case ColorType8:
		scanRange = 7
	case ColorType16:
		scanRange = 15
	case ColorType256:
		scanRange = 255
	default:
		panic(fmt.Errorf("unhandled terminal color count %#v", terminalColorCount))
	}

	// Iterate over the scan range and find the best matching index
	bestMatch := 0
	bestDistance := math.MaxFloat64
	for i := 0; i <= scanRange; i++ {
		r, g, b := color256ToRGB(uint8(i))
		candidate := NewColor24Bit(r, g, b)

		distance := target.Distance(candidate)
		if distance < bestDistance {
			bestDistance = distance
			bestMatch = i
		}
	}

	if bestMatch <= 15 {
		return NewColor16(bestMatch)
	}
	return NewColor256(uint8(bestMatch))
}

// Wrapper for Chroma's color distance function.
//
// That one says it uses this formula: https://www.compuphase.com/cmetric.htm
//
// The result from this function has been scaled to 0.0-1.0, where 1.0 is the
// distance between black and white.
func (color Color) Distance(other Color) float64 {
	if color.ColorType() != ColorType24bit {
		panic(fmt.Errorf("contrast only supported for 24 bit colors, got %s vs %s", color.String(), other.String()))
	}

	baseColor := chroma.NewColour(
		uint8(color.colorValue()>>16&0xff),
		uint8(color.colorValue()>>8&0xff),
		uint8(color.colorValue()&0xff),
	)

	otherColor := chroma.NewColour(
		uint8(other.colorValue()>>16&0xff),
		uint8(other.colorValue()>>8&0xff),
		uint8(other.colorValue()&0xff),
	)

	// Magic constant comes from testing
	maxDistance := 764.8333151739665
	return baseColor.Distance(otherColor) / maxDistance
}
