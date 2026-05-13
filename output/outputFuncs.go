package output

import (
	"fmt"
	huesettings "ninthcircleoftext/hueSettings"
	texteditor "ninthcircleoftext/textEditor"
	"strings"
)

const curHue string = "\x1b[47m"
const bgResetHue string = "\x1b[49m"
const fgResetHue string = "\x1b[39m"

// takes in line no. & total lines, returns padded line no. UI element
func generateLineNumUI(lineNum, lineCount int) string {
	strLineNum := fmt.Sprintf("%d", lineNum+1)
	strLineCount := fmt.Sprintf("%d", lineCount)
	whiteSpace := strings.Repeat(" ", len(strLineCount))
	if lineNum >= lineCount {
		return whiteSpace + "| "
	}
	whiteSpace = strings.Repeat(" ", len(strLineCount)-len(strLineNum))
	return whiteSpace + strLineNum + "| "
}

// inserts cursor into a line, regardless of whether or not the line includes any ansi escape codes
func insertCursor(line string, curX int) string {
	xAdj := 0
	realX := -1
	for x := 0; x < len(line); x++ {
		if line[x] == byte(27) {
			escCodeLen := len(fgResetHue) //all ANSI codes accepted by huesettings are the basic hues 30->39 (all the same length)
			xAdj += escCodeLen
			x += escCodeLen - 1
		} else {
			if realX >= curX {
				break
			}
			realX += 1
		}
	}
	curX += xAdj
	if curX >= len(line)-1 {
		return line + curHue + " " + bgResetHue
	}
	return line[:curX+1] + curHue + string(line[curX+1]) + bgResetHue + line[curX+2:]
}

// removes all valid foreground hues (specified in the huesettings pkg as all 10 of the basic foreground hues)
func removeFgHues(str string) string {
	for _, hue := range huesettings.ValidHues {
		str = strings.ReplaceAll(str, "\x1b["+hue+"m", "")
	}
	return str
}

// applies comment hues to line (including in-line comments)
func applyCommentHue(line, commentSymbol, commentHue string) string {
	if commentSymbol == "" {
		return line
	}
	//checking for comment at line-start (skipping whitespace)
	trimmedLine := strings.TrimSpace(line)
	if strings.HasPrefix(trimmedLine, commentSymbol) {
		return commentHue + removeFgHues(line) + fgResetHue
	}
	//in-line comment highlighting
	splitLine := strings.Split(line, commentSymbol)
	if len(splitLine) < 2 {
		return line
	}
	prefix := splitLine[0]
	suffix := strings.Join(splitLine[1:], commentSymbol)
	suffix = commentSymbol + suffix
	return prefix + commentHue + removeFgHues(suffix) + fgResetHue
}

func applyStringHues(line string, stringHues map[string]string) string {
	newLine := ""
	substrings := []string{}
	substring := ""
	for x := 0; x < len(line); x++ {
		chr := string(line[x])
		if len(substring) == 0 {
			if _, ok := stringHues[chr]; ok {
				newLine += string(byte(0))
				substring += chr
			} else {
				newLine += chr
			}
		} else {
			substring += chr
			if chr == "\\" { //skipping over esc chars in strings
				if x < len(line)-1 {
					substring += string(line[x+1])
				}
				x += 1
			} else if strSymbol := string(substring[0]); chr == strSymbol {
				hue, _ := stringHues[strSymbol]
				substrings = append(substrings, hue+removeFgHues(substring)+fgResetHue)
				substring = ""
			}
		}
	}
	if len(substring) > 0 {
		hue, _ := stringHues[string(substring[0])]
		substrings = append(substrings, hue+removeFgHues(substring)+fgResetHue)
		substring = ""
	}
	for _, substring := range substrings {
		newLine = strings.Replace(newLine, string(byte(0)), substring, 1)
	}
	return newLine
}

// applies bracket hues to any bracket not preceded by an escape sequence
func applyBracketHues(line string, bracketHues map[string]string) string {
	for x := 0; x < len(line); x++ {
		b := line[x]
		if b == 27 {
			x += 1
			continue
		}
		if hue, ok := bracketHues[string(b)]; ok {
			bracket := hue + string(b) + fgResetHue
			line = line[:x] + bracket + line[x+1:]
			x += len(bracket) - 1
		}
	}
	return line
}

func isNonUnderscoreSymbol(a byte) bool {
	if (a != 95) && ((a >= 32 && a <= 47) ||
		(a >= 58 && a <= 64) ||
		(a >= 91 && a <= 96) ||
		(a >= 123 && a <= 126)) {
		return true
	}
	return false
}

// returns the longest keyword (& hue) (with no neighbouring letters or underscores) immediately at the start of a given string
func getKeywordPrefix(str string, keywordHues map[string]string) (string, string) {
	foundKeyword := ""
	for keyword := range keywordHues {
		if strings.HasPrefix(str, keyword) {
			nextByte := byte((str + " ")[len(keyword)])
			if isNonUnderscoreSymbol(nextByte) && len(keyword) > len(foundKeyword) {
				foundKeyword = keyword
			}
		}
	}
	hue, _ := keywordHues[foundKeyword]
	return foundKeyword, hue
}

func applyKeywordHues(line string, keywordHues map[string]string) string {
	var newLine strings.Builder
	line = " " + line                 //adding a buffer space to the line so that each previous byte can be processed
	for x := 1; x <= len(line); x++ { //iterating over each byte in line & checking for a keyword if b is a non-underscore symbol
		b := line[x-1]
		if x == 1 || isNonUnderscoreSymbol(b) {
			if keyword, hue := getKeywordPrefix(line[x:], keywordHues); keyword != "" {
				newLine.WriteString(string(b) + hue + keyword + fgResetHue)
				x += len(keyword)
				continue
			}
		}
		newLine.WriteString(string(b))
	}
	return strings.TrimPrefix(newLine.String(), " ") //removing buffer space
}

// applies regular syntax hues to a line (keywords, brackets, strings & code-comments)
func applyRegularSyntaxHues(line string, hueMap *huesettings.HueMap) string {
	//keywords
	line = applyKeywordHues(line, hueMap.Keywords)
	//symbols
	line = applyBracketHues(line, hueMap.Brackets)
	//strings (applied 2nd-last as any strings need to have all hues removed)
	line = applyStringHues(line, hueMap.Strings)
	//comments (applied last as any comments need to have all hues removed)
	line = applyCommentHue(line, hueMap.Comment, hueMap.CommentHue)
	return line
}

// applies syntax hue settings to a line
// also handles multi-line-code-comments (MLinComs) & returns whether or not the next line is in an ongoing MLinCom
// note: only applies MLinCom hues to the first instance of an MLinCom identifier in any given line, subsequent MLinCom identifiers are treated as regular line parts
func applyHueMap(line string, hueMap *huesettings.HueMap, inMLinCom bool) (string, bool) {
	//not applying multi-line-comment hues if they have not been configured
	if hueMap.MLinComPre == "" || hueMap.MLinComSuf == "" {
		return applyRegularSyntaxHues(line, hueMap), false

	}

	if !inMLinCom { //applying multi-line-comment hues to a line that is not in an ongoing multi-line-comment
		if splitLine := strings.Split(line, hueMap.MLinComPre); len(splitLine) >= 2 {
			pre := applyRegularSyntaxHues(splitLine[0], hueMap)
			post := strings.Join(splitLine[1:], hueMap.MLinComPre) //joining the rest of the line so that any remaining MLinComPre strings are still printed, but not highlighted
			post = hueMap.MLinComHue + hueMap.MLinComPre + post + fgResetHue
			return pre + post, true
		}
		return applyRegularSyntaxHues(line, hueMap), false

	} else { //applying multi-line-comment hues to a line that is in an ongoing multi-line-comment
		if splitLine := strings.Split(line, hueMap.MLinComSuf); len(splitLine) >= 2 {
			pre := hueMap.MLinComHue + splitLine[0] + hueMap.MLinComSuf + fgResetHue
			post := strings.Join(splitLine[1:], hueMap.MLinComSuf)
			post = applyRegularSyntaxHues(post, hueMap)
			return pre + post, false
		}
		return hueMap.MLinComHue + line, true
	}
}

// formats the stored file for output & returns it
func formatTextForTerminal(editor *texteditor.TextEditor, hueMap *huesettings.HueMap) string {
	upperUI, lowerUI, text, curX, curY, termY, termHeight := editor.GetOutputFields()

	ftextBuilder := strings.Builder{}
	ftextBuilder.WriteString("\x1b[0m" + upperUI)

	//iterating over the lines in the file below the end of the terminal (only applying syntax hues to visible lines)
	inMLinCom := false
	for y := 0; y < termY+termHeight; y++ {
		//only checking if the above-off-screen line has a multi-line code comment prefix/suffix so that multi-line code comment hues can be applied even when they start off-screen
		if y < termY && y <= len(text)-1 {
			if line := text[y]; (!inMLinCom && strings.Contains(line, hueMap.MLinComPre)) || (inMLinCom && strings.Contains(line, hueMap.MLinComSuf)) {
				inMLinCom = !inMLinCom
			}
		} else if y >= termY {
			line := ""
			if y <= len(text)-1 {
				line = text[y]
				line, inMLinCom = applyHueMap(line, hueMap, inMLinCom)
				if y == curY {
					line = insertCursor(line, curX)
				}
			}
			lineNumUI := ""
			if editor.LineNumToggle() {
				lineNumUI = generateLineNumUI(y, len(text))
			}
			ftextBuilder.WriteString("\x1b[0m" + lineNumUI + line + "\n\r")
		}
	}

	ftextBuilder.WriteString("\x1b[0m" + lowerUI)
	return ftextBuilder.String()
}

// clears terminal & outputs text
func WriteEditorTextToTerminal(editor *texteditor.TextEditor, hueMap *huesettings.HueMap) {
	ftext := formatTextForTerminal(editor, hueMap)
	fmt.Printf("\x1bc%s", ftext)
}
