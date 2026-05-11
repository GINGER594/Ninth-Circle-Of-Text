package main

import (
	"bufio"
	"fmt"
	huesettings "ninthcircleoftext/hueSettings"
	"ninthcircleoftext/output"
	texteditor "ninthcircleoftext/textEditor"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"
)

const settingsFlag string = "-settings"
const helpFlag string = "-help"
const helpMessage string = "\n\r" +
	"--- Ninth Circle of text (NCT) ---\n\r" +
	"\n\r" +
	"flags:\n\r" +
	"> NCT can be run with the flags '-help' & '-settings'\n\r" +
	"\n\r" +
	"settings:\n\r" +
	"> running NCT with the '-settings' flag will open settings\n\r" +
	"> note - opening the settings file with '-settings' will open the settings file without applying any settings\n\r" +
	"\n\r" +
	"> entering a new setting involves a file extension, a type, an identifier, & a hue\n\r" +
	"> if an entry is repeated in the settings file, the latest entry takes priority\n\r" +
	"> only one comment can be configured per file extension - the latest entry takes priority\n\r" +
	"> valid types are: 'keyword', 'bracket', 'strings' & 'comment'\n\r" +
	"> the different types dictate how the syntax highlighting is applied\n\r" +
	"> note - there are only 6 valid bracket identifiers: '(', ')', '[', ']', '{', '}'\n\r" +
	"\n\r" +
	"> an example for setting python comments to green would be: 'py|comment|#|32'\n\r" +
	"\n\r" +
	"hues:\n\r" +
	"  hue              code\n\r" +
	"  black            30\n\r" +
	"  red              31\n\r" +
	"  green            32\n\r" +
	"  yellow           33\n\r" +
	"  blue             34\n\r" +
	"  magenta          35\n\r" +
	"  cyan             36\n\r" +
	"  white            37\n\r" +
	"  default          39\n\r" +
	"\n\r" +
	"editing files:\n\r" +
	"> to move the cursor, use the arrow, pg up/pg dn & home/end keys\n\r" +
	"> type any char to insert it\n\r" +
	"> tabs are 4 spaces\n\r" +
	"> the REPLACE ALL function takes an input of 'a'/'b' (quote marks required) & replaces a with b\n\r" +
	"> other CTRL sequences are detailed in the file-editing UI\n\r" +
	"\n\r"

var keyByteMap map[string][]byte = map[string][]byte{
	"up":        {27, 91, 65, 0},
	"down":      {27, 91, 66, 0},
	"right":     {27, 91, 67, 0},
	"left":      {27, 91, 68, 0},
	"home":      {27, 91, 72, 0},
	"end":       {27, 91, 70, 0},
	"pg up":     {27, 91, 53, 126},
	"pg dn":     {27, 91, 54, 126},
	"backspace": {127, 0, 0, 0},
	"delete":    {27, 91, 51, 126},
	"^?":        {31, 0, 0, 0},
	"^X":        {24, 0, 0, 0},
	"^C":        {3, 0, 0, 0},
	"^V":        {22, 0, 0, 0},
	"^R":        {18, 0, 0, 0},
	"^T":        {20, 0, 0, 0},
	"^S":        {19, 0, 0, 0},
	"^Q":        {17, 0, 0, 0},
	"enter":     {13, 0, 0, 0},
	"tab":       {9, 0, 0, 0},
}

// exits the program, if the file has not been saved the user will be given the option to save it
func saveAndExit(editor *texteditor.TextEditor) error {
	fmt.Printf("\x1bc")
	if editor.Saved() {
		return nil
	}

	fmt.Printf("changes may have been made to this file since last saving - save? (y/N): ")
	inp := ""
	fmt.Scanln(&inp) //using fmt scan to accept only the first arg in an input
	fmt.Printf("\x1bc")
	if inp != "y" && inp != "Y" {
		return nil
	}
	err := editor.SaveFile()
	if err != nil {
		return err
	}
	return nil
}

// (to be used from a function where the terminal has been made raw) switches terminal state & gets user input
// parses user input into a & b for TextEditor.ReplaceAllOccurrences, returns true if input is valid
func getReplaceAllInp(oldState *term.State, inpMsg string) (string, string, bool, error) {
	err := term.Restore(int(os.Stdin.Fd()), oldState)
	if err != nil {
		return "", "", false, err
	}
	defer term.MakeRaw(int(os.Stdin.Fd()))
	fmt.Printf("\x1b[0m%s", inpMsg)
	inpScanner := bufio.NewReader(os.Stdin)
	inp, _, err := inpScanner.ReadLine()

	//parsing input
	splitInp := strings.Split(string(inp), "'/'")
	if len(splitInp) != 2 {
		return "", "", false, nil
	}
	a, b := splitInp[0], splitInp[1]
	trimA, trimB := strings.TrimPrefix(a, "'"), strings.Trim(b, "'")
	if a == trimA || b == trimB || len(trimA) == 0 { //checking for surrounding ' and if a is empty
		return "", "", false, nil
	}
	return trimA, trimB, true, nil
}

// handles raw inputs, calling edit/display functions/methods
func editMode(editor *texteditor.TextEditor, hueMap huesettings.HueMap) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	inpScanner := bufio.NewReader(os.Stdin)
	for editLoop := true; editLoop; {
		err := editor.UpdateTerminalAndUIFields()
		if err != nil {
			return err
		}
		output.WriteEditorTextToTerminal(editor, hueMap)

		inp := make([]byte, 4)
		inpScanner.Read(inp)
		switch {
		case slices.Equal(inp, keyByteMap["up"]):
			editor.AdjustCursorY(-1)
		case slices.Equal(inp, keyByteMap["down"]):
			editor.AdjustCursorY(1)
		case slices.Equal(inp, keyByteMap["right"]):
			editor.AdjustCursorX(1)
		case slices.Equal(inp, keyByteMap["left"]):
			editor.AdjustCursorX(-1)
		case slices.Equal(inp, keyByteMap["home"]):
			editor.HomeCursor()
		case slices.Equal(inp, keyByteMap["end"]):
			editor.EndCursor()
		case slices.Equal(inp, keyByteMap["pg up"]):
			editor.PageUpCursor()
		case slices.Equal(inp, keyByteMap["pg dn"]):
			editor.PageDownCursor()
		case slices.Equal(inp, keyByteMap["backspace"]):
			editor.Backspace()
		case slices.Equal(inp, keyByteMap["delete"]):
			editor.Delete()
		case slices.Equal(inp, keyByteMap["^?"]):
			editor.CommentLine()
		case slices.Equal(inp, keyByteMap["^X"]):
			editor.XLine()
		case slices.Equal(inp, keyByteMap["^C"]):
			editor.StoreLine()
		case slices.Equal(inp, keyByteMap["^V"]):
			editor.InsertStoredLine()
		case slices.Equal(inp, keyByteMap["^R"]):
			a, b, valid, err := getReplaceAllInp(oldState, texteditor.ReplaceUI)
			if err != nil {
				return err
			}
			if valid {
				editor.ReplaceAllOccurrences(a, b)
			}
		case slices.Equal(inp, keyByteMap["^T"]):
			editor.LineNumToggle = !editor.LineNumToggle
		case slices.Equal(inp, keyByteMap["^S"]):
			err := editor.SaveFile()
			if err != nil {
				return err
			}
		case slices.Equal(inp, keyByteMap["^Q"]):
			editLoop = false
		case slices.Equal(inp, keyByteMap["enter"]):
			editor.InsertLine()
		case slices.Equal(inp, keyByteMap["tab"]):
			editor.InsertTab()
		default:
			//checking if the key is a symbol or letter and not an unwanted ctrl/esc sequence
			if inp[0] >= 32 && inp[0] <= 126 && inp[1] == 0 && inp[2] == 0 && inp[3] == 0 {
				editor.InsertString(string(inp[0]))
			}
		}
	}
	return nil
}

// returns path and path extension
func validateOsArgs() (string, string, error) {
	if len(os.Args) != 2 {
		return "", "", fmt.Errorf("invalid no. of arguments recieved - epected: 1, got: %d", len(os.Args)-1)
	}

	//handling flags
	path := os.Args[1]
	if path == helpFlag || path == settingsFlag {
		return path, "", nil
	}

	//getting file extension
	if len(strings.Split(path, ".")) != 2 {
		return "", "", fmt.Errorf("invalid path argument: no file extension present")
	}
	ext := strings.Split(path, ".")[1]
	return path, ext, nil

}

// runs program
func runTextEditor() error {
	//arg validation
	path, ext, err := validateOsArgs()
	if err != nil {
		return fmt.Errorf("error during argument validation: %s", err.Error())
	}
	if path == helpFlag {
		fmt.Printf(helpMessage)
		return nil
	}

	//ensuring the existence of a settings file if the args are valid and path != helpFlag
	err = huesettings.CreateInternalSettingsFile()
	if err != nil {
		return fmt.Errorf("error ensuring existence of settings file: %s", err.Error())
	}

	//processing syntax hue settings (unless editing NCT settings)
	hueMap := huesettings.HueMap{
		Keywords: map[string]string{},
		Brackets: map[string]string{},
		Strings:  map[string]string{},
	}
	if path == settingsFlag {
		path, err = huesettings.GetInternalSettingsPath()
		if err != nil {
			return fmt.Errorf("error getting path to settings file: %s", err.Error())
		}
	} else {
		hueMap, err = huesettings.GetSyntaxHueSettings(ext)
		if err != nil {
			return fmt.Errorf("error processing settings: %s", err.Error())
		}
	}

	//setting up editor
	editor := texteditor.TextEditor{}
	err = editor.SetDefaultValues(path, hueMap.Comment)
	if err != nil {
		return fmt.Errorf("error during text-editor set-up: %s", err.Error())
	}

	//edit mode
	err = editMode(&editor, hueMap)
	if err != nil {
		return fmt.Errorf("error during file-edit session: %s", err.Error())
	}

	//exiting program
	err = saveAndExit(&editor)
	if err != nil {
		return fmt.Errorf("error during save/exit: %s", err.Error())
	}
	return nil
}

func main() {
	err := runTextEditor()
	if err != nil {
		fmt.Printf("%s\n\rrun with the '-help' flag for more info\n\r", err.Error())
	}
}
