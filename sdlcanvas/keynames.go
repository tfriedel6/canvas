package sdlcanvas

import "github.com/veandco/go-sdl2/sdl"

var keyNameMap [262]string
var keyRuneMap [262]rune

func init() {
	keyNameMap[sdl.SCANCODE_ESCAPE] = "Escape"
	keyNameMap[sdl.SCANCODE_0] = "Digit0"
	keyNameMap[sdl.SCANCODE_1] = "Digit1"
	keyNameMap[sdl.SCANCODE_2] = "Digit2"
	keyNameMap[sdl.SCANCODE_3] = "Digit3"
	keyNameMap[sdl.SCANCODE_4] = "Digit4"
	keyNameMap[sdl.SCANCODE_5] = "Digit5"
	keyNameMap[sdl.SCANCODE_6] = "Digit6"
	keyNameMap[sdl.SCANCODE_7] = "Digit7"
	keyNameMap[sdl.SCANCODE_8] = "Digit8"
	keyNameMap[sdl.SCANCODE_9] = "Digit9"
	keyNameMap[sdl.SCANCODE_MINUS] = "Minus"
	keyNameMap[sdl.SCANCODE_EQUALS] = "Equal"
	keyNameMap[sdl.SCANCODE_BACKSPACE] = "Backspace"
	keyNameMap[sdl.SCANCODE_TAB] = "Tab"
	keyNameMap[sdl.SCANCODE_Q] = "KeyQ"
	keyNameMap[sdl.SCANCODE_W] = "KeyW"
	keyNameMap[sdl.SCANCODE_E] = "KeyE"
	keyNameMap[sdl.SCANCODE_R] = "KeyR"
	keyNameMap[sdl.SCANCODE_T] = "KeyT"
	keyNameMap[sdl.SCANCODE_Y] = "KeyY"
	keyNameMap[sdl.SCANCODE_U] = "KeyU"
	keyNameMap[sdl.SCANCODE_I] = "KeyI"
	keyNameMap[sdl.SCANCODE_O] = "KeyO"
	keyNameMap[sdl.SCANCODE_P] = "KeyP"
	keyNameMap[sdl.SCANCODE_LEFTBRACKET] = "BracketLeft"
	keyNameMap[sdl.SCANCODE_RIGHTBRACKET] = "BracketRight"
	keyNameMap[sdl.SCANCODE_RETURN] = "Enter"
	keyNameMap[sdl.SCANCODE_LCTRL] = "ControlLeft"
	keyNameMap[sdl.SCANCODE_A] = "KeyA"
	keyNameMap[sdl.SCANCODE_S] = "KeyS"
	keyNameMap[sdl.SCANCODE_D] = "KeyD"
	keyNameMap[sdl.SCANCODE_F] = "KeyF"
	keyNameMap[sdl.SCANCODE_G] = "KeyG"
	keyNameMap[sdl.SCANCODE_H] = "KeyH"
	keyNameMap[sdl.SCANCODE_J] = "KeyJ"
	keyNameMap[sdl.SCANCODE_K] = "KeyK"
	keyNameMap[sdl.SCANCODE_L] = "KeyL"
	keyNameMap[sdl.SCANCODE_SEMICOLON] = "Semicolon"
	keyNameMap[sdl.SCANCODE_APOSTROPHE] = "Quote"
	keyNameMap[sdl.SCANCODE_GRAVE] = "Backquote"
	keyNameMap[sdl.SCANCODE_LSHIFT] = "ShiftLeft"
	keyNameMap[sdl.SCANCODE_BACKSLASH] = "Backslash"
	keyNameMap[sdl.SCANCODE_Z] = "KeyZ"
	keyNameMap[sdl.SCANCODE_X] = "KeyX"
	keyNameMap[sdl.SCANCODE_C] = "KeyC"
	keyNameMap[sdl.SCANCODE_V] = "KeyV"
	keyNameMap[sdl.SCANCODE_B] = "KeyB"
	keyNameMap[sdl.SCANCODE_N] = "KeyN"
	keyNameMap[sdl.SCANCODE_M] = "KeyM"
	keyNameMap[sdl.SCANCODE_COMMA] = "Comma"
	keyNameMap[sdl.SCANCODE_PERIOD] = "Period"
	keyNameMap[sdl.SCANCODE_SLASH] = "Slash"
	keyNameMap[sdl.SCANCODE_RSHIFT] = "ShiftRight"
	keyNameMap[sdl.SCANCODE_KP_MULTIPLY] = "NumpadMultiply"
	keyNameMap[sdl.SCANCODE_LALT] = "AltLeft"
	keyNameMap[sdl.SCANCODE_SPACE] = "Space"
	keyNameMap[sdl.SCANCODE_CAPSLOCK] = "CapsLock"
	keyNameMap[sdl.SCANCODE_F1] = "F1"
	keyNameMap[sdl.SCANCODE_F2] = "F2"
	keyNameMap[sdl.SCANCODE_F3] = "F3"
	keyNameMap[sdl.SCANCODE_F4] = "F4"
	keyNameMap[sdl.SCANCODE_F5] = "F5"
	keyNameMap[sdl.SCANCODE_F6] = "F6"
	keyNameMap[sdl.SCANCODE_F7] = "F7"
	keyNameMap[sdl.SCANCODE_F8] = "F8"
	keyNameMap[sdl.SCANCODE_F9] = "F9"
	keyNameMap[sdl.SCANCODE_F10] = "F10"
	keyNameMap[sdl.SCANCODE_PAUSE] = "Pause"
	keyNameMap[sdl.SCANCODE_SCROLLLOCK] = "ScrollLock"
	keyNameMap[sdl.SCANCODE_KP_7] = "Numpad7"
	keyNameMap[sdl.SCANCODE_KP_8] = "Numpad8"
	keyNameMap[sdl.SCANCODE_KP_9] = "Numpad9"
	keyNameMap[sdl.SCANCODE_KP_MINUS] = "NumpadSubtract"
	keyNameMap[sdl.SCANCODE_KP_4] = "Numpad4"
	keyNameMap[sdl.SCANCODE_KP_5] = "Numpad5"
	keyNameMap[sdl.SCANCODE_KP_6] = "Numpad6"
	keyNameMap[sdl.SCANCODE_KP_PLUS] = "NumpadAdd"
	keyNameMap[sdl.SCANCODE_KP_1] = "Numpad1"
	keyNameMap[sdl.SCANCODE_KP_2] = "Numpad2"
	keyNameMap[sdl.SCANCODE_KP_3] = "Numpad3"
	keyNameMap[sdl.SCANCODE_KP_0] = "Numpad0"
	keyNameMap[sdl.SCANCODE_KP_DECIMAL] = "NumpadDecimal"
	keyNameMap[sdl.SCANCODE_PRINTSCREEN] = "PrintScreen"
	keyNameMap[sdl.SCANCODE_NONUSBACKSLASH] = "IntlBackslash"
	keyNameMap[sdl.SCANCODE_F11] = "F11"
	keyNameMap[sdl.SCANCODE_F12] = "F12"
	keyNameMap[sdl.SCANCODE_KP_EQUALS] = "NumpadEqual"
	keyNameMap[sdl.SCANCODE_F13] = "F13"
	keyNameMap[sdl.SCANCODE_F14] = "F14"
	keyNameMap[sdl.SCANCODE_F15] = "F15"
	keyNameMap[sdl.SCANCODE_F16] = "F16"
	keyNameMap[sdl.SCANCODE_F17] = "F17"
	keyNameMap[sdl.SCANCODE_F18] = "F18"
	keyNameMap[sdl.SCANCODE_F19] = "F19"
	keyNameMap[sdl.SCANCODE_UNDO] = "Undo"
	keyNameMap[sdl.SCANCODE_PASTE] = "Paste"
	keyNameMap[sdl.SCANCODE_AUDIOPREV] = "MediaTrackPrevious"
	keyNameMap[sdl.SCANCODE_CUT] = "Cut"
	keyNameMap[sdl.SCANCODE_COPY] = "Copy"
	keyNameMap[sdl.SCANCODE_AUDIONEXT] = "MediaTrackNext"
	keyNameMap[sdl.SCANCODE_KP_ENTER] = "NumpadEnter"
	keyNameMap[sdl.SCANCODE_RCTRL] = "ControlRight"
	keyNameMap[sdl.SCANCODE_MUTE] = "AudioVolumeMute"
	keyNameMap[sdl.SCANCODE_AUDIOPLAY] = "MediaPlayPause"
	keyNameMap[sdl.SCANCODE_AUDIOSTOP] = "MediaStop"
	keyNameMap[sdl.SCANCODE_VOLUMEDOWN] = "AudioVolumeDown"
	keyNameMap[sdl.SCANCODE_VOLUMEUP] = "AudioVolumeUp"
	keyNameMap[sdl.SCANCODE_KP_DIVIDE] = "NumpadDivide"
	keyNameMap[sdl.SCANCODE_RALT] = "AltRight"
	keyNameMap[sdl.SCANCODE_HELP] = "Help"
	keyNameMap[sdl.SCANCODE_HOME] = "Home"
	keyNameMap[sdl.SCANCODE_UP] = "ArrowUp"
	keyNameMap[sdl.SCANCODE_PAGEUP] = "PageUp"
	keyNameMap[sdl.SCANCODE_LEFT] = "ArrowLeft"
	keyNameMap[sdl.SCANCODE_RIGHT] = "ArrowRight"
	keyNameMap[sdl.SCANCODE_END] = "End"
	keyNameMap[sdl.SCANCODE_DOWN] = "ArrowDown"
	keyNameMap[sdl.SCANCODE_INSERT] = "Insert"
	keyNameMap[sdl.SCANCODE_DELETE] = "Delete"
	keyNameMap[sdl.SCANCODE_APPLICATION] = "ContextMenu"

	keyRuneMap[sdl.SCANCODE_0] = '0'
	keyRuneMap[sdl.SCANCODE_1] = '1'
	keyRuneMap[sdl.SCANCODE_2] = '2'
	keyRuneMap[sdl.SCANCODE_3] = '3'
	keyRuneMap[sdl.SCANCODE_4] = '4'
	keyRuneMap[sdl.SCANCODE_5] = '5'
	keyRuneMap[sdl.SCANCODE_6] = '6'
	keyRuneMap[sdl.SCANCODE_7] = '7'
	keyRuneMap[sdl.SCANCODE_8] = '8'
	keyRuneMap[sdl.SCANCODE_9] = '9'
	keyRuneMap[sdl.SCANCODE_MINUS] = '-'
	keyRuneMap[sdl.SCANCODE_EQUALS] = '='
	keyRuneMap[sdl.SCANCODE_TAB] = '\t'
	keyRuneMap[sdl.SCANCODE_Q] = 'Q'
	keyRuneMap[sdl.SCANCODE_W] = 'W'
	keyRuneMap[sdl.SCANCODE_E] = 'E'
	keyRuneMap[sdl.SCANCODE_R] = 'R'
	keyRuneMap[sdl.SCANCODE_T] = 'T'
	keyRuneMap[sdl.SCANCODE_Y] = 'Y'
	keyRuneMap[sdl.SCANCODE_U] = 'U'
	keyRuneMap[sdl.SCANCODE_I] = 'I'
	keyRuneMap[sdl.SCANCODE_O] = 'O'
	keyRuneMap[sdl.SCANCODE_P] = 'P'
	keyRuneMap[sdl.SCANCODE_LEFTBRACKET] = '['
	keyRuneMap[sdl.SCANCODE_RIGHTBRACKET] = ']'
	keyRuneMap[sdl.SCANCODE_RETURN] = '\n'
	keyRuneMap[sdl.SCANCODE_A] = 'A'
	keyRuneMap[sdl.SCANCODE_S] = 'S'
	keyRuneMap[sdl.SCANCODE_D] = 'D'
	keyRuneMap[sdl.SCANCODE_F] = 'F'
	keyRuneMap[sdl.SCANCODE_G] = 'G'
	keyRuneMap[sdl.SCANCODE_H] = 'H'
	keyRuneMap[sdl.SCANCODE_J] = 'J'
	keyRuneMap[sdl.SCANCODE_K] = 'K'
	keyRuneMap[sdl.SCANCODE_L] = 'L'
	keyRuneMap[sdl.SCANCODE_SEMICOLON] = ';'
	keyRuneMap[sdl.SCANCODE_APOSTROPHE] = '\''
	keyRuneMap[sdl.SCANCODE_GRAVE] = '`'
	keyRuneMap[sdl.SCANCODE_BACKSLASH] = '\\'
	keyRuneMap[sdl.SCANCODE_Z] = 'Z'
	keyRuneMap[sdl.SCANCODE_X] = 'X'
	keyRuneMap[sdl.SCANCODE_C] = 'C'
	keyRuneMap[sdl.SCANCODE_V] = 'V'
	keyRuneMap[sdl.SCANCODE_B] = 'B'
	keyRuneMap[sdl.SCANCODE_N] = 'N'
	keyRuneMap[sdl.SCANCODE_M] = 'M'
	keyRuneMap[sdl.SCANCODE_COMMA] = ','
	keyRuneMap[sdl.SCANCODE_PERIOD] = '.'
	keyRuneMap[sdl.SCANCODE_SLASH] = '/'
	keyRuneMap[sdl.SCANCODE_KP_MULTIPLY] = '*'
	keyRuneMap[sdl.SCANCODE_SPACE] = ' '
	keyRuneMap[sdl.SCANCODE_KP_7] = '7'
	keyRuneMap[sdl.SCANCODE_KP_8] = '8'
	keyRuneMap[sdl.SCANCODE_KP_9] = '9'
	keyRuneMap[sdl.SCANCODE_KP_MINUS] = '-'
	keyRuneMap[sdl.SCANCODE_KP_4] = '4'
	keyRuneMap[sdl.SCANCODE_KP_5] = '5'
	keyRuneMap[sdl.SCANCODE_KP_6] = '6'
	keyRuneMap[sdl.SCANCODE_KP_PLUS] = '+'
	keyRuneMap[sdl.SCANCODE_KP_1] = '1'
	keyRuneMap[sdl.SCANCODE_KP_2] = '2'
	keyRuneMap[sdl.SCANCODE_KP_3] = '3'
	keyRuneMap[sdl.SCANCODE_KP_0] = '0'
	keyRuneMap[sdl.SCANCODE_KP_DECIMAL] = '.'
	keyRuneMap[sdl.SCANCODE_KP_EQUALS] = '='
	keyRuneMap[sdl.SCANCODE_KP_ENTER] = '\n'
	keyRuneMap[sdl.SCANCODE_KP_DIVIDE] = '/'
}

func keyName(s sdl.Scancode) string {
	if int(s) >= len(keyNameMap) {
		return "Unidentified"
	}
	name := keyNameMap[s]
	if name == "" {
		return "Unidentified"
	}
	return name
}

func keyRune(s sdl.Scancode) rune {
	if int(s) >= len(keyNameMap) {
		return 0
	}
	return keyRuneMap[s]
}
