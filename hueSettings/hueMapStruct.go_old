package huesettings

import "strings"

// a way of keeping all of the syntax hue data for a single file ext together
type HueMap struct {
	Keywords   map[string]string
	Brackets   map[string]string
	Strings    map[string]string
	Comment    string
	CommentHue string
	MLinComPre string
	MLinComSuf string
	MLinComHue string
}

// takes in parsed items from an entry in the settings file & inserts them into the hue map
func (h *HueMap) insertIntoHueMap(valueType, identifier, hue string) {
	switch valueType {
	case "keyword":
		h.Keywords[identifier] = hue
	case "bracket":
		h.Brackets[identifier] = hue
	case "strings":
		h.Strings[identifier] = hue
	case "comment":
		h.Comment = identifier
		h.CommentHue = hue
	case "mlincom":
		splitId := strings.Split(identifier, " ")
		h.MLinComPre, h.MLinComSuf = splitId[0], splitId[1]
		h.MLinComHue = hue
	}
}
