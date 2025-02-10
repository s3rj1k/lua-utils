package gluautf8

import (
	"unicode/utf8"

	lua "github.com/yuin/gopher-lua"
)

func Preload(L *lua.LState) {
	L.PreloadModule("utf8", Loader)
}

func Loader(L *lua.LState) int {
	mod := L.NewTable()

	L.SetFuncs(mod, utf8Funcs)
	L.SetField(mod, "charpattern", lua.LString("[\x00-\x7F\xC2-\xFD][\x80-\xBF]*"))

	L.Push(mod)
	return 1
}

var utf8Funcs = map[string]lua.LGFunction{
	"char": func(L *lua.LState) int {
		top := L.GetTop()
		runes := make([]rune, top)

		for i := 1; i <= top; i++ {
			runes[i-1] = rune(L.CheckInt(i))
		}

		L.Push(lua.LString(string(runes)))
		return 1
	},
	"codes": func(L *lua.LState) int {
		s := L.CheckString(1)
		lax := L.OptBool(2, false)

		iter := func(L *lua.LState) int {
			pos := L.ToInt(2)
			if pos >= len(s) {
				L.Push(lua.LNil)
				return 1
			}

			r, size := utf8.DecodeRuneInString(s[pos:])
			if r == utf8.RuneError && !lax {
				L.RaiseError("invalid UTF-8 sequence at position %d", pos)
			}

			L.Push(lua.LNumber(pos + size)) // note, values here do not match vanilla Lua runtime for UTF8
			L.Push(lua.LNumber(r))
			return 2
		}

		L.Push(L.NewFunction(iter))
		L.Push(lua.LNumber(0))
		L.Push(lua.LNil)
		return 3
	},
	"codepoint": func(L *lua.LState) int {
		s := L.CheckString(1)
		i := L.OptInt(2, 1) - 1
		j := L.OptInt(3, i+1) - 1
		lax := L.OptBool(4, false)

		if i < 0 {
			i = 0
		}

		if j >= len(s) {
			j = len(s) - 1
		}

		results := L.NewTable()
		if len(s) == 0 || i > j || i >= len(s) {
			L.Push(results)
			return 1
		}

		pos := i
		for pos <= j {
			r, size := utf8.DecodeRuneInString(s[pos:])
			if r == utf8.RuneError && !lax {
				L.RaiseError("invalid UTF-8 sequence at position %d", pos)
			}

			results.Append(lua.LNumber(r))
			pos += size
		}

		L.Push(results)
		return 1
	},
	"len": func(L *lua.LState) int {
		s := L.CheckString(1)
		i := L.OptInt(2, 1) - 1
		j := L.OptInt(3, -1) - 1
		lax := L.OptBool(4, false)

		if i < 0 {
			i = 0
		}

		if j < 0 {
			j = len(s) - 1
		}

		count := 0
		pos := 0

		for pos < len(s) && pos < i {
			_, size := utf8.DecodeRuneInString(s[pos:])
			pos += size
		}

		for pos <= j && pos < len(s) {
			r, size := utf8.DecodeRuneInString(s[pos:])
			if r == utf8.RuneError && !lax {
				L.Push(lua.LNil)
				L.Push(lua.LNumber(pos + 1))
				return 2
			}

			count++
			pos += size
		}

		L.Push(lua.LNumber(count))
		return 1
	},
	"offset": func(L *lua.LState) int { // ToDo: FixMe
		s := L.CheckString(1)
		n := L.CheckInt(2)

		// Handle default i position based on n
		var i int
		if L.Get(3) == lua.LNil {
			if n >= 0 {
				i = 1
			} else {
				i = len(s)
			}
		} else {
			i = L.CheckInt(3)
		}

		// Convert from 1-based to 0-based index
		i--

		if i < 0 || i > len(s) {
			L.Push(lua.LNil)
			return 1
		}

		// Special case: n = 0 (find character boundary)
		if n == 0 {
			pos := 0
			for pos < len(s) && pos < i {
				_, size := utf8.DecodeRuneInString(s[pos:])
				if pos+size > i {
					L.Push(lua.LNumber(pos + 1))
					return 1
				}
				pos += size
			}
			L.Push(lua.LNumber(i + 1))
			return 1
		}

		// Find the character index at position i
		startCharIdx := 0
		pos := 0
		for pos < i && pos < len(s) {
			_, size := utf8.DecodeRuneInString(s[pos:])
			pos += size
			startCharIdx++
		}

		// Calculate target character position
		targetCharIdx := startCharIdx + n
		if targetCharIdx <= 0 {
			if n < 0 {
				L.Push(lua.LNumber(1))
				return 1
			}
			L.Push(lua.LNil)
			return 1
		}

		// Convert character index back to byte position
		pos = 0
		currentCharIdx := 0
		for pos < len(s) && currentCharIdx < targetCharIdx-1 {
			_, size := utf8.DecodeRuneInString(s[pos:])
			pos += size
			currentCharIdx++
		}

		if pos >= len(s) {
			L.Push(lua.LNil)
			return 1
		}

		// Return byte position (converting to 1-based index)
		L.Push(lua.LNumber(pos + 1))
		return 1
	},
}
