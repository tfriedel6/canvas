package glfwcanvas

import "github.com/go-gl/glfw/v3.3/glfw"

var keyNameMap [347]string
var keyRuneMap [347]rune

func init() {
	keyNameMap[glfw.KeyEscape] = "Escape"
	keyNameMap[glfw.Key0] = "Digit0"
	keyNameMap[glfw.Key1] = "Digit1"
	keyNameMap[glfw.Key2] = "Digit2"
	keyNameMap[glfw.Key3] = "Digit3"
	keyNameMap[glfw.Key4] = "Digit4"
	keyNameMap[glfw.Key5] = "Digit5"
	keyNameMap[glfw.Key6] = "Digit6"
	keyNameMap[glfw.Key7] = "Digit7"
	keyNameMap[glfw.Key8] = "Digit8"
	keyNameMap[glfw.Key9] = "Digit9"
	keyNameMap[glfw.KeyMinus] = "Minus"
	keyNameMap[glfw.KeyEqual] = "Equal"
	keyNameMap[glfw.KeyBackspace] = "Backspace"
	keyNameMap[glfw.KeyTab] = "Tab"
	keyNameMap[glfw.KeyQ] = "KeyQ"
	keyNameMap[glfw.KeyW] = "KeyW"
	keyNameMap[glfw.KeyE] = "KeyE"
	keyNameMap[glfw.KeyR] = "KeyR"
	keyNameMap[glfw.KeyT] = "KeyT"
	keyNameMap[glfw.KeyY] = "KeyY"
	keyNameMap[glfw.KeyU] = "KeyU"
	keyNameMap[glfw.KeyI] = "KeyI"
	keyNameMap[glfw.KeyO] = "KeyO"
	keyNameMap[glfw.KeyP] = "KeyP"
	keyNameMap[glfw.KeyLeftBracket] = "BracketLeft"
	keyNameMap[glfw.KeyRightBracket] = "BracketRight"
	keyNameMap[glfw.KeyEnter] = "Enter"
	keyNameMap[glfw.KeyLeftControl] = "ControlLeft"
	keyNameMap[glfw.KeyA] = "KeyA"
	keyNameMap[glfw.KeyS] = "KeyS"
	keyNameMap[glfw.KeyD] = "KeyD"
	keyNameMap[glfw.KeyF] = "KeyF"
	keyNameMap[glfw.KeyG] = "KeyG"
	keyNameMap[glfw.KeyH] = "KeyH"
	keyNameMap[glfw.KeyJ] = "KeyJ"
	keyNameMap[glfw.KeyK] = "KeyK"
	keyNameMap[glfw.KeyL] = "KeyL"
	keyNameMap[glfw.KeySemicolon] = "Semicolon"
	keyNameMap[glfw.KeyApostrophe] = "Quote"
	keyNameMap[glfw.KeyGraveAccent] = "Backquote"
	keyNameMap[glfw.KeyLeftShift] = "ShiftLeft"
	keyNameMap[glfw.KeyBackslash] = "Backslash"
	keyNameMap[glfw.KeyZ] = "KeyZ"
	keyNameMap[glfw.KeyX] = "KeyX"
	keyNameMap[glfw.KeyC] = "KeyC"
	keyNameMap[glfw.KeyV] = "KeyV"
	keyNameMap[glfw.KeyB] = "KeyB"
	keyNameMap[glfw.KeyN] = "KeyN"
	keyNameMap[glfw.KeyM] = "KeyM"
	keyNameMap[glfw.KeyComma] = "Comma"
	keyNameMap[glfw.KeyPeriod] = "Period"
	keyNameMap[glfw.KeySlash] = "Slash"
	keyNameMap[glfw.KeyRightShift] = "ShiftRight"
	keyNameMap[glfw.KeyKPMultiply] = "NumpadMultiply"
	keyNameMap[glfw.KeyLeftAlt] = "AltLeft"
	keyNameMap[glfw.KeySpace] = "Space"
	keyNameMap[glfw.KeyCapsLock] = "CapsLock"
	keyNameMap[glfw.KeyF1] = "F1"
	keyNameMap[glfw.KeyF2] = "F2"
	keyNameMap[glfw.KeyF3] = "F3"
	keyNameMap[glfw.KeyF4] = "F4"
	keyNameMap[glfw.KeyF5] = "F5"
	keyNameMap[glfw.KeyF6] = "F6"
	keyNameMap[glfw.KeyF7] = "F7"
	keyNameMap[glfw.KeyF8] = "F8"
	keyNameMap[glfw.KeyF9] = "F9"
	keyNameMap[glfw.KeyF10] = "F10"
	keyNameMap[glfw.KeyPause] = "Pause"
	keyNameMap[glfw.KeyScrollLock] = "ScrollLock"
	keyNameMap[glfw.KeyKP7] = "Numpad7"
	keyNameMap[glfw.KeyKP8] = "Numpad8"
	keyNameMap[glfw.KeyKP9] = "Numpad9"
	keyNameMap[glfw.KeyKPSubtract] = "NumpadSubtract"
	keyNameMap[glfw.KeyKP4] = "Numpad4"
	keyNameMap[glfw.KeyKP5] = "Numpad5"
	keyNameMap[glfw.KeyKP6] = "Numpad6"
	keyNameMap[glfw.KeyKPAdd] = "NumpadAdd"
	keyNameMap[glfw.KeyKP1] = "Numpad1"
	keyNameMap[glfw.KeyKP2] = "Numpad2"
	keyNameMap[glfw.KeyKP3] = "Numpad3"
	keyNameMap[glfw.KeyKP0] = "Numpad0"
	keyNameMap[glfw.KeyKPDecimal] = "NumpadDecimal"
	keyNameMap[glfw.KeyPrintScreen] = "PrintScreen"
	// keyNameMap[glfw.KeyNonUSBackslash] = "IntlBackslash"
	keyNameMap[glfw.KeyF11] = "F11"
	keyNameMap[glfw.KeyF12] = "F12"
	keyNameMap[glfw.KeyKPEqual] = "NumpadEqual"
	keyNameMap[glfw.KeyF13] = "F13"
	keyNameMap[glfw.KeyF14] = "F14"
	keyNameMap[glfw.KeyF15] = "F15"
	keyNameMap[glfw.KeyF16] = "F16"
	keyNameMap[glfw.KeyF17] = "F17"
	keyNameMap[glfw.KeyF18] = "F18"
	keyNameMap[glfw.KeyF19] = "F19"
	// keyNameMap[glfw.KeyUndo] = "Undo"
	// keyNameMap[glfw.KeyPaste] = "Paste"
	// keyNameMap[glfw.KeyAudioNext] = "MediaTrackPrevious"
	// keyNameMap[glfw.KeyCut] = "Cut"
	// keyNameMap[glfw.KeyCopy] = "Copy"
	// keyNameMap[glfw.KeyAudioNext] = "MediaTrackNext"
	keyNameMap[glfw.KeyKPEnter] = "NumpadEnter"
	keyNameMap[glfw.KeyRightControl] = "ControlRight"
	// keyNameMap[glfw.KeyMute] = "AudioVolumeMute"
	// keyNameMap[glfw.KeyAudioPlay] = "MediaPlayPause"
	// keyNameMap[glfw.KeyAudioStop] = "MediaStop"
	// keyNameMap[glfw.KeyVolumeDown] = "AudioVolumeDown"
	// keyNameMap[glfw.KeyVolumeUp] = "AudioVolumeUp"
	keyNameMap[glfw.KeyKPDivide] = "NumpadDivide"
	keyNameMap[glfw.KeyRightAlt] = "AltRight"
	// keyNameMap[glfw.KeyHelp] = "Help"
	keyNameMap[glfw.KeyHome] = "Home"
	keyNameMap[glfw.KeyUp] = "ArrowUp"
	keyNameMap[glfw.KeyPageUp] = "PageUp"
	keyNameMap[glfw.KeyLeft] = "ArrowLeft"
	keyNameMap[glfw.KeyRight] = "ArrowRight"
	keyNameMap[glfw.KeyEnd] = "End"
	keyNameMap[glfw.KeyDown] = "ArrowDown"
	keyNameMap[glfw.KeyInsert] = "Insert"
	keyNameMap[glfw.KeyDelete] = "Delete"
	// keyNameMap[glfw.KeyApplication] = "ContextMenu"

	keyRuneMap[glfw.Key0] = '0'
	keyRuneMap[glfw.Key1] = '1'
	keyRuneMap[glfw.Key2] = '2'
	keyRuneMap[glfw.Key3] = '3'
	keyRuneMap[glfw.Key4] = '4'
	keyRuneMap[glfw.Key5] = '5'
	keyRuneMap[glfw.Key6] = '6'
	keyRuneMap[glfw.Key7] = '7'
	keyRuneMap[glfw.Key8] = '8'
	keyRuneMap[glfw.Key9] = '9'
	keyRuneMap[glfw.KeyMinus] = '-'
	keyRuneMap[glfw.KeyEqual] = '='
	keyRuneMap[glfw.KeyTab] = '\t'
	keyRuneMap[glfw.KeyQ] = 'Q'
	keyRuneMap[glfw.KeyW] = 'W'
	keyRuneMap[glfw.KeyE] = 'E'
	keyRuneMap[glfw.KeyR] = 'R'
	keyRuneMap[glfw.KeyT] = 'T'
	keyRuneMap[glfw.KeyY] = 'Y'
	keyRuneMap[glfw.KeyU] = 'U'
	keyRuneMap[glfw.KeyI] = 'I'
	keyRuneMap[glfw.KeyO] = 'O'
	keyRuneMap[glfw.KeyP] = 'P'
	keyRuneMap[glfw.KeyLeftBracket] = '['
	keyRuneMap[glfw.KeyRightBracket] = ']'
	keyRuneMap[glfw.KeyEnter] = '\n'
	keyRuneMap[glfw.KeyA] = 'A'
	keyRuneMap[glfw.KeyS] = 'S'
	keyRuneMap[glfw.KeyD] = 'D'
	keyRuneMap[glfw.KeyF] = 'F'
	keyRuneMap[glfw.KeyG] = 'G'
	keyRuneMap[glfw.KeyH] = 'H'
	keyRuneMap[glfw.KeyJ] = 'J'
	keyRuneMap[glfw.KeyK] = 'K'
	keyRuneMap[glfw.KeyL] = 'L'
	keyRuneMap[glfw.KeySemicolon] = ';'
	keyRuneMap[glfw.KeyApostrophe] = '\''
	keyRuneMap[glfw.KeyGraveAccent] = '`'
	keyRuneMap[glfw.KeyBackslash] = '\\'
	keyRuneMap[glfw.KeyZ] = 'Z'
	keyRuneMap[glfw.KeyX] = 'X'
	keyRuneMap[glfw.KeyC] = 'C'
	keyRuneMap[glfw.KeyV] = 'V'
	keyRuneMap[glfw.KeyB] = 'B'
	keyRuneMap[glfw.KeyN] = 'N'
	keyRuneMap[glfw.KeyM] = 'M'
	keyRuneMap[glfw.KeyComma] = ','
	keyRuneMap[glfw.KeyPeriod] = '.'
	keyRuneMap[glfw.KeySlash] = '/'
	keyRuneMap[glfw.KeyKPMultiply] = '*'
	keyRuneMap[glfw.KeySpace] = ' '
	keyRuneMap[glfw.KeyKP7] = '7'
	keyRuneMap[glfw.KeyKP8] = '8'
	keyRuneMap[glfw.KeyKP9] = '9'
	keyRuneMap[glfw.KeyKPSubtract] = '-'
	keyRuneMap[glfw.KeyKP4] = '4'
	keyRuneMap[glfw.KeyKP5] = '5'
	keyRuneMap[glfw.KeyKP6] = '6'
	keyRuneMap[glfw.KeyKPAdd] = '+'
	keyRuneMap[glfw.KeyKP1] = '1'
	keyRuneMap[glfw.KeyKP2] = '2'
	keyRuneMap[glfw.KeyKP3] = '3'
	keyRuneMap[glfw.KeyKP0] = '0'
	keyRuneMap[glfw.KeyKPDecimal] = '.'
	keyRuneMap[glfw.KeyKPEqual] = '='
	keyRuneMap[glfw.KeyKPEnter] = '\n'
	keyRuneMap[glfw.KeyKPDivide] = '/'
}

func keyName(key glfw.Key) string {
	if int(key) >= len(keyNameMap) {
		return "Unidentified"
	}
	name := keyNameMap[key]
	if name == "" {
		return "Unidentified"
	}
	return name
}

func keyRune(key glfw.Key) rune {
	if int(key) >= len(keyNameMap) {
		return 0
	}
	return keyRuneMap[key]
}
