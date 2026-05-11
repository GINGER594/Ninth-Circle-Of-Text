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

// takes in line no. and total lines, returns padded line no. UI element
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
	//checking for comment at line-start
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

// applies string hues to line (handles edge cases such as: (" \" "), ("\\") and: (" 'a' "))
func applyStringHues(line string, stringHues map[string]string) string {
	newLine := ""
	substrings := []string{}
	substring := ""
	line = "  " + line //adding a buffer to the line so that previous 2 bytes can be parsed for \" and "\\" edge cases
	for x := 2; x < len(line); x++ {
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
			//terminating the string unless the termination is preceded by a \ (unless that \ is also preceded by a \ e.g. "\\")
			if strSymbol := string(substring[0]); chr == strSymbol && (line[x-1] != byte('\\') || (line[x-1] == byte('\\') && line[x-2] == byte('\\'))) {
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
	line = " " + line //adding a buffer to the line so that each previous byte can be processed
	for x := 1; x < len(line); x++ {
		chr := string(line[x])
		if hue, ok := bracketHues[chr]; ok && line[x-1] != 27 {
			chr := hue + chr + fgResetHue
			line = line[:x] + chr + line[x+1:]
			x += len(chr) - 1
		}
	}
	line = strings.TrimPrefix(line, " ") //removing buffer
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

// returns the longest keyword (and hue) (with no neighbouring letters or underscores) immediately at the start of a given string
func getKeywordPrefix(str string, keywordHues map[string]string) (string, string) {
	keywordPrefix := ""
	for keyword := range keywordHues {
		if strings.HasPrefix(str, keyword) {
			nextByte := byte((str + " ")[len(keyword)])
			if isNonUnderscoreSymbol(nextByte) && len(keyword) > len(keywordPrefix) {
				keywordPrefix = keyword
			}
		}
	}
	hue, _ := keywordHues[keywordPrefix]
	return keywordPrefix, hue
}

// applies syntax hues any keywords that are not preceded or succeeded by a letter or underscore
func applyKeywordHues(line string, keywordHues map[string]string) string {
	var newLine strings.Builder
	line = " " + line                 //adding a buffer to the line so that each previous byte can be processed
	for x := 1; x <= len(line); x++ { //iterating over each byte in line and checking for a keyword if b is a non-underscore symbol
		b := line[x-1]
		if x == 0 || isNonUnderscoreSymbol(b) {
			if keyword, hue := getKeywordPrefix(line[x:], keywordHues); keyword != "" {
				newLine.WriteString(string(b) + hue + keyword + fgResetHue)
				x += len(keyword)
				continue
			}
		}
		newLine.WriteString(string(b))
	}
	hueLine := newLine.String()
	hueLine = strings.TrimPrefix(hueLine, " ") //removing buffer
	return hueLine
}

func applySyntaxHues(line string, hueMap huesettings.HueMap) string {
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

// formats the stored file for output and returns it
func formatTextForTerminal(editor *texteditor.TextEditor, hueMap huesettings.HueMap) string {
	upperUI, lowerUI, text, curX, curY, termY, termHeight := editor.GetOutputFields()

	ftextBuilder := strings.Builder{}
	ftextBuilder.WriteString("\x1b[0m" + upperUI)
	for y := termY; y < termY+termHeight; y++ {
		line := ""
		if y <= len(text)-1 {
			line = text[y]
			line = applySyntaxHues(line, hueMap)
			if y == curY {
				line = insertCursor(line, curX)
			}
		}
		lineNumUI := ""
		if editor.LineNumToggle {
			lineNumUI = generateLineNumUI(y, len(text))
		}
		ftextBuilder.WriteString("\x1b[0m" + lineNumUI + line + "\n\r")
	}

	ftextBuilder.WriteString("\x1b[0m" + lowerUI)
	return ftextBuilder.String()
}

// clears terminal and outputs text
func WriteEditorTextToTerminal(editor *texteditor.TextEditor, hueMap huesettings.HueMap) {
	ftext := formatTextForTerminal(editor, hueMap)
	fmt.Printf("\x1bc%s", ftext)
}
