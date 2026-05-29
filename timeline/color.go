package timeline

import "strings"

// reset is the ANSI SGR sequence that clears all styling.
const reset = "\x1b[0m"

// sgrCodes maps human-friendly color/style names to ANSI SGR parameter codes.
// Specs are case-insensitive and may combine several tokens separated by
// spaces, commas or '+', e.g. "bold cyan" or "underline,red".
var sgrCodes = map[string]string{
	"reset":     "0",
	"bold":      "1",
	"dim":       "2",
	"faint":     "2",
	"italic":    "3",
	"underline": "4",

	// Standard foreground colors.
	"black":   "30",
	"red":     "31",
	"green":   "32",
	"yellow":  "33",
	"blue":    "34",
	"magenta": "35",
	"cyan":    "36",
	"white":   "37",
	"gray":    "90",
	"grey":    "90",

	// Bright foreground colors.
	"brightblack":   "90",
	"brightred":     "91",
	"brightgreen":   "92",
	"brightyellow":  "93",
	"brightblue":    "94",
	"brightmagenta": "95",
	"brightcyan":    "96",
	"brightwhite":   "97",
}

// ansiOpen returns the SGR opening sequence for the given style spec, or the
// empty string if the spec is empty or contains no recognized tokens.
func ansiOpen(spec string) string {
	if spec == "" {
		return ""
	}
	var codes []string
	for _, tok := range strings.FieldsFunc(spec, func(r rune) bool {
		return r == ' ' || r == ',' || r == '+' || r == '\t'
	}) {
		if c, ok := sgrCodes[strings.ToLower(tok)]; ok {
			codes = append(codes, c)
		}
	}
	if len(codes) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}
