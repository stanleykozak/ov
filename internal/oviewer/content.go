package oviewer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// Content represents one character on the terminal.
type Content struct {
	width int
	mainc rune
	combc []rune
	style tcell.Style
}

// lineContents represents one line of contents.
type lineContents struct {
	// contents contains one line of contents.
	contents []Content
	// byteMap is the number of contents corresponding to the number of bytes.
	byteMap map[int]int
}

// The states of the ANSI escape code parser.
const (
	ansiText = iota
	ansiEscape
	ansiSubstring
	ansiControlSequence
)

// DefaultContent is a blank Content.
var DefaultContent = Content{
	mainc: 0,
	combc: nil,
	width: 0,
	style: tcell.StyleDefault,
}

// parseString converts a string to lineContents.
// parseString includes escape sequences and tabs.
func parseString(line string, tabWidth int) lineContents {
	lc := lineContents{
		contents: nil,
		byteMap:  make(map[int]int),
	}

	state := ansiText
	csiParameter := new(bytes.Buffer)
	style := tcell.StyleDefault
	x := 0
	n := 0
	bsFlag := false
	var bsContent Content
	for _, runeValue := range line {
		c := DefaultContent
		switch state {
		case ansiEscape:
			switch runeValue {
			case '[': // Control Sequence Introducer.
				csiParameter.Reset()
				state = ansiControlSequence
				continue
			case 'c': // Reset.
				style = tcell.StyleDefault
				state = ansiText
				continue
			case 'P', ']', 'X', '^', '_': // Substrings and commands.
				state = ansiSubstring
				continue
			default: // Ignore.
				state = ansiText
			}
		case ansiSubstring:
			if runeValue == 0x1b {
				state = ansiEscape
				continue
			}
		case ansiControlSequence:
			if runeValue == 'm' {
				style = csToStyle(style, csiParameter)
			} else if runeValue >= 'A' && runeValue <= 'T' {
				// Ignore.
			} else {
				if runeValue >= 0x30 && runeValue <= 0x3f {
					csiParameter.WriteRune(runeValue)
					continue
				}
			}
			state = ansiText
			continue
		}

		switch runeValue {
		case 0x1b:
			state = ansiEscape
			continue
		case '\n':
			continue
		}

		lc.byteMap[n] = len(lc.contents)
		n += len(string(runeValue))

		switch runewidth.RuneWidth(runeValue) {
		case 0:
			switch runeValue {
			case '\t': // TAB
				switch {
				case tabWidth > 0:
					tabStop := tabWidth - (x % tabWidth)
					c.mainc = rune(' ')
					c.width = 1
					c.style = style
					for i := 0; i < tabStop; i++ {
						lc.contents = append(lc.contents, c)
						x++
					}
				case tabWidth < 0:
					lc.byteMap[n] = len(lc.contents)
					n += len(string(runeValue))
					c.width = 1
					c.style = style.Reverse(true)
					c.mainc = rune('\\')
					lc.contents = append(lc.contents, c)
					c.mainc = rune('t')
					lc.contents = append(lc.contents, c)
					x += 2
				default:
				}
				continue
			case '\b': // BackSpace
				if len(lc.contents) == 0 {
					continue
				}
				bsFlag = true
				bsContent = lastContent(lc.contents)
				if bsContent.width > 1 {
					lc.contents = lc.contents[:len(lc.contents)-2]
				} else {
					lc.contents = lc.contents[:len(lc.contents)-1]
				}
				continue
			}
			content := lastContent(lc.contents)
			content.combc = append(content.combc, runeValue)
			n := len(lc.contents) - content.width
			if n >= 0 && len(lc.contents) > 0 {
				lc.contents[n] = content
			}
		case 1:
			c.mainc = runeValue
			c.width = 1
			c.style = style
			if bsFlag {
				c.style = overstrike(bsContent.mainc, runeValue, style)
				bsFlag = false
				bsContent = DefaultContent
			}
			lc.contents = append(lc.contents, c)
			x++
		case 2:
			c.mainc = runeValue
			c.width = 2
			c.style = style
			if bsFlag {
				c.style = overstrike(bsContent.mainc, runeValue, style)
				bsFlag = false
				bsContent = DefaultContent
			}
			lc.contents = append(lc.contents, c, DefaultContent)
			x += 2
		}
	}
	lc.byteMap[n] = len(lc.contents)
	return lc
}

func overstrike(p, m rune, style tcell.Style) tcell.Style {
	if p == m {
		style = OverStrikeStyle
	} else if p == '_' {
		style = OverLineStyle
	}
	return style
}

func lastContent(contents []Content) Content {
	n := len(contents)
	if n == 0 {
		return Content{}
	}
	if (n > 1) && (contents[n-2].width > 1) {
		return contents[n-2]
	}
	return contents[n-1]
}

func csToStyle(style tcell.Style, csiParameter *bytes.Buffer) tcell.Style {
	fields := strings.Split(csiParameter.String(), ";")
	if len(fields) == 0 || len(fields) == 1 && fields[0] == "0" {
		style = tcell.StyleDefault.Normal()
	}
FieldLoop:
	for index, field := range fields {
		switch field {
		case "1", "01":
			style = style.Bold(true)
		case "2", "02":
			style = style.Dim(true)
		case "4", "04":
			style = style.Underline(true)
		case "5", "05":
			style = style.Blink(true)
		case "7", "07":
			style = style.Reverse(true)
		case "22", "24", "25", "27":
			style = style.Normal()
		case "30", "31", "32", "33", "34", "35", "36", "37":
			colorNumber, _ := strconv.Atoi(field)
			style = style.Foreground(tcell.Color(colorNumber - 30))
		case "39":
			style = style.Foreground(tcell.ColorDefault)
		case "40", "41", "42", "43", "44", "45", "46", "47":
			colorNumber, _ := strconv.Atoi(field)
			style = style.Background(tcell.Color(colorNumber - 40))
		case "49":
			style = style.Background(tcell.ColorDefault)
		case "90", "91", "92", "93", "94", "95", "96", "97":
			colorNumber, _ := strconv.Atoi(field)
			style = style.Foreground(tcell.Color(colorNumber - 82))
		case "100", "101", "102", "103", "104", "105", "106", "107":
			colorNumber, _ := strconv.Atoi(field)
			style = style.Background(tcell.Color(colorNumber - 92))
		case "38", "48":
			var color string
			if len(fields) > index+1 {
				if fields[index+1] == "5" && len(fields) > index+2 { // 8-bit colors.
					colorNumber, _ := strconv.Atoi(fields[index+2])
					if colorNumber <= 7 {
						color = lookupColor(colorNumber)
					} else if colorNumber <= 15 {
						color = lookupColor(colorNumber)
					} else if colorNumber <= 231 {
						red := (colorNumber - 16) / 36
						green := ((colorNumber - 16) / 6) % 6
						blue := (colorNumber - 16) % 6
						color = fmt.Sprintf("#%02x%02x%02x", 255*red/5, 255*green/5, 255*blue/5)
					} else if colorNumber <= 255 {
						grey := 255 * (colorNumber - 232) / 23
						color = fmt.Sprintf("#%02x%02x%02x", grey, grey, grey)
					}
				} else if fields[index+1] == "2" && len(fields) > index+4 { // 24-bit colors.
					red, _ := strconv.Atoi(fields[index+2])
					green, _ := strconv.Atoi(fields[index+3])
					blue, _ := strconv.Atoi(fields[index+4])
					color = fmt.Sprintf("#%02x%02x%02x", red, green, blue)
				}
			}
			if len(color) > 0 {
				if field == "38" {
					style = style.Foreground(tcell.GetColor(color))
				} else {
					style = style.Background(tcell.GetColor(color))
				}
				break FieldLoop
			}
		}
	}
	return style
}

func lookupColor(colorNumber int) string {
	if colorNumber < 0 || colorNumber > 15 {
		return "black"
	}
	return [...]string{
		"black",
		"maroon",
		"green",
		"olive",
		"navy",
		"purple",
		"teal",
		"silver",
		"gray",
		"red",
		"lime",
		"yellow",
		"blue",
		"fuchsia",
		"aqua",
		"white",
	}[colorNumber]
}

// strToContents converts a single-line string into a content array.
func strToContents(str string, tabWidth int) []Content {
	lc := parseString(str, tabWidth)
	return lc.contents
}
