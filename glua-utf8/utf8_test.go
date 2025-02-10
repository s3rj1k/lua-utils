package gluautf8_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	gluautf8 "github.com/projectsveltos/lua-utils/glua-utf8"
)

func TestChar(t *testing.T) {
	tests := []string{
		"A",
		"ABC",
		"你好",
		"A你B",
		"",
		"😀",
		"Hi 😀 你好",
		string(utf8.RuneError),
		" \t\n\r\u00A0\u2000\u3000",
		"©®™",
	}

	for i, input := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, input), func(t *testing.T) {
			r := require.New(t)

			L := lua.NewState()
			defer L.Close()

			gluautf8.Preload(L)

			err := L.DoString("utf8 = require('utf8')")
			r.NoError(err)

			luaCode := "return utf8.char("
			luaCode += strings.Join(gluautf8.GetRuneValuesAsStrings(input), ", ")
			luaCode += ")"

			err = L.DoString(luaCode)
			r.NoError(err)

			result := L.ToString(-1)
			r.Equal(input, result)
			L.Pop(1)
		})
	}
}

func TestCodes(t *testing.T) {
	tests := []struct {
		input    string
		expected []struct {
			pos  int
			code rune
		}
		wantErr bool
		lax     bool
	}{
		{
			input: "abc",
			expected: []struct {
				pos  int
				code rune
			}{
				{strings.Index("abc", "a") + 1, 'a'},
				{strings.Index("abc", "b") + 1, 'b'},
				{strings.Index("abc", "c") + 1, 'c'},
			},
		},
		{
			input: "世界",
			expected: []struct {
				pos  int
				code rune
			}{
				{strings.Index("世界", "世") + utf8.RuneLen('世'), '世'}, // ToDo: why no +1
				{strings.Index("世界", "界") + utf8.RuneLen('界'), '界'}, // ToDo: why no +1
			},
		},
		{
			input: "hello世界",
			expected: []struct {
				pos  int
				code rune
			}{
				{strings.Index("hello世界", "h") + 1, 'h'},
				{strings.Index("hello世界", "e") + 1, 'e'},
				{strings.Index("hello世界", "l") + 1, 'l'},
				{strings.LastIndex("hello世界", "l") + 1, 'l'},
				{strings.Index("hello世界", "o") + 1, 'o'},
				{strings.Index("hello世界", "世") + utf8.RuneLen('世'), '世'}, // ToDo: why no +1
				{strings.Index("hello世界", "界") + utf8.RuneLen('界'), '界'}, // ToDo: why no +1
			},
		},
		{
			input: "",
			expected: []struct {
				pos  int
				code rune
			}{},
		},
		{
			input:   string([]byte{0xFF, 0xFE}),
			wantErr: true,
		},
		{
			input: string([]byte{0xFF, 0xFE}),
			lax:   true,
			expected: []struct {
				pos  int
				code rune
			}{
				{1, 0xFFFD},
				{2, 0xFFFD},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, tt.input), func(t *testing.T) {
			r := require.New(t)

			L := lua.NewState()
			defer L.Close()

			gluautf8.Preload(L)
			err := L.DoString("utf8 = require('utf8')")
			r.NoError(err)

			err = L.DoString(`
                function test_codes(input, lax)
                    local result = {}
                    local idx = 1
                    for pos, code in utf8.codes(input, lax) do
                        result[idx] = {pos = pos, code = code}
                        idx = idx + 1
                    end
                    return result
                end
            `)
			r.NoError(err)

			L.Push(L.GetGlobal("test_codes"))
			L.Push(lua.LString(tt.input))
			L.Push(lua.LBool(tt.lax))

			err = L.PCall(2, 1, nil)
			if tt.wantErr {
				r.Error(err)
				return
			}
			r.NoError(err)

			result := L.Get(-1)
			r.Equal(lua.LTTable, result.Type())

			resultTable := result.(*lua.LTable)

			var got []struct {
				pos  int
				code rune
			}

			for i := 1; i <= resultTable.Len(); i++ {
				entry, ok := resultTable.RawGetInt(i).(*lua.LTable)
				r.True(ok)

				pos, ok := entry.RawGet(lua.LString("pos")).(lua.LNumber)
				r.True(ok)

				code, ok := entry.RawGet(lua.LString("code")).(lua.LNumber)
				r.True(ok)

				got = append(got, struct {
					pos  int
					code rune
				}{
					pos:  int(pos),
					code: rune(code),
				})
			}

			if len(tt.expected) == 0 {
				got = []struct {
					pos  int
					code rune
				}{}
			}

			r.Equal(tt.expected, got)
		})
	}
}

func TestCharPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "A",
			expected: "A",
		},
		{
			input:    "¢",
			expected: "¢",
		},
		{
			input:    "世",
			expected: "世",
		},
		{
			input:    "😀",
			expected: "😀",
		},
		{
			input:    "A世😀",
			expected: "A",
		},
		{
			input:    "世界",
			expected: "世",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, tt.input), func(t *testing.T) {
			r := require.New(t)

			L := lua.NewState()
			defer L.Close()

			gluautf8.Preload(L)

			L.SetGlobal("test_input", lua.LString(tt.input))

			err := L.DoString(`
				local utf8 = require('utf8')
				return string.match(test_input, utf8.charpattern)
			`)
			r.NoError(err)

			result := L.ToString(-1)
			r.Equal(tt.expected, result)
			L.Pop(1)
		})
	}
}

func TestCodepoint(t *testing.T) {
	tests := []struct {
		input    string
		i        int
		j        int
		lax      bool
		expected []int
		wantErr  bool
	}{
		{
			input:    "",
			expected: []int{},
		},
		{
			input:    "hello",
			expected: gluautf8.GetRuneValues("h"),
		},
		{
			input:    "hello",
			i:        strings.Index("hello", "e") + 1,
			expected: gluautf8.GetRuneValues("e"),
		},
		{
			input:    "hello",
			i:        strings.Index("hello", "l") + 1,
			j:        strings.LastIndex("hello", "l") + 1,
			expected: gluautf8.GetRuneValues("ll"),
		},
		{
			input:    "你好",
			i:        strings.Index("你好", "你") + 1,
			j:        strings.LastIndex("你好", "好") + utf8.RuneLen('你') + 1,
			expected: gluautf8.GetRuneValues("你好"),
		},
		{
			input:   "你好",
			i:       strings.Index("你好", "好"), // ToDo: why no +1
			j:       strings.Index("你好", "好"), // ToDo: why no +1
			wantErr: true,
		},
		{
			input:    "hi👋",
			i:        strings.Index("hi👋", "👋") + 1,
			j:        strings.Index("hi👋", "👋") + utf8.RuneLen('👋') + 1,
			expected: gluautf8.GetRuneValues("👋"),
		},
		{
			input:   "hello\xFF\xFEworld",
			i:       strings.Index("hello", "e") + 1,
			j:       strings.Index("hello", "o") + 2,
			wantErr: true,
		},
		{
			input:    "hello\xFF\xFEworld",
			i:        1,
			j:        strings.Index("hello", "o") + 3,
			lax:      true,
			expected: []int{104, 101, 108, 108, 111, 0xFFFD, 0xFFFD},
		},
		{
			input:    "hello",
			i:        10,
			expected: []int{},
		},
		{
			input:    "hello",
			i:        strings.Index("hello", "l") + 2,
			j:        strings.Index("hello", "e") + 1,
			expected: []int{},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, tt.input), func(t *testing.T) {
			r := require.New(t)

			L := lua.NewState()
			defer L.Close()

			gluautf8.Preload(L)

			err := L.DoString("utf8 = require('utf8')")
			r.NoError(err)

			L.SetGlobal("test_input", lua.LString(tt.input))

			var luaCode string
			if tt.lax {
				if tt.i == 0 {
					tt.i = 1
				}

				if tt.j == 0 {
					tt.j = tt.i
				}

				luaCode = fmt.Sprintf("return utf8.codepoint(test_input, %d, %d, true)", tt.i, tt.j)
			} else {
				if tt.i == 0 {
					luaCode = "return utf8.codepoint(test_input)"
				} else if tt.j == 0 {
					luaCode = fmt.Sprintf("return utf8.codepoint(test_input, %d)", tt.i)
				} else {
					luaCode = fmt.Sprintf("return utf8.codepoint(test_input, %d, %d)", tt.i, tt.j)
				}
			}

			err = L.DoString(luaCode)
			if tt.wantErr {
				r.Error(err)
				return
			}
			r.NoError(err)

			result, ok := L.Get(-1).(*lua.LTable)
			r.True(ok)

			codepoints := make([]int, 0)
			result.ForEach(func(_ lua.LValue, value lua.LValue) {
				codepoints = append(codepoints, int(value.(lua.LNumber)))
			})

			r.Equal(tt.expected, codepoints)
			L.Pop(1)
		})
	}
}

func TestLen(t *testing.T) {
	tests := []struct {
		input    string
		i        int
		j        int
		lax      bool
		expected any
		pos      int
	}{
		{
			input:    "",
			expected: 0,
		},
		{
			input: "hello",
			expected: gluautf8.CountChars(
				"hello", 0, -1,
			),
		},
		// {
		// 	input: "hello",
		// 	i:     strings.Index("hello", "e") + 1,
		// 	expected: gluautf8.CountChars(
		// 		"hello", strings.Index("hello", "e")+1, -1,
		// 	),
		// },
		{
			input: "hello",
			i:     strings.Index("hello", "e") + 1,
			j:     strings.LastIndex("hello", "l") + 1,
			expected: gluautf8.CountChars(
				"hello", strings.Index("hello", "e")+1, strings.LastIndex("hello", "l")+1,
			),
		},
		// {
		// 	input: "你好",
		// 	expected: gluautf8.CountChars(
		// 		"你好", 0, 0,
		// 	),
		// },
		// {
		// 	input: "你好",
		// 	i:     utf8.RuneLen('你') + 1,
		// 	j:     strings.LastIndex("你好", "好") + utf8.RuneLen('好') + 1,
		// 	expected: gluautf8.CountChars(
		// 		"你好", utf8.RuneLen('你')+1, strings.LastIndex("你好", "好")+utf8.RuneLen('好')+1,
		// 	),
		// },
		{
			input:    "hi👋",
			expected: 3,
		},
		{
			input:    "hi👋",
			i:        strings.Index("hi👋", "👋") + 1,
			j:        strings.Index("hi👋", "👋") + utf8.RuneLen('👋') + 1,
			expected: 1,
		},
		{
			input:    "hello\xFF\xFEworld",
			i:        1,
			j:        strings.Index("hello", "o") + 3,
			expected: nil,
			pos:      6,
		},
		{
			input:    "hello\xFF\xFEworld",
			i:        1,
			j:        strings.Index("hello", "o") + 3,
			lax:      true,
			expected: 7,
		},
		{
			input:    "café",
			expected: 4,
		},
		{
			input:    "hello",
			i:        10,
			expected: 0,
		},
		{
			input:    "hello",
			i:        strings.Index("hello", "l") + 2,
			j:        strings.Index("hello", "e") + 1,
			expected: 0,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, tt.input), func(t *testing.T) {
			r := require.New(t)

			L := lua.NewState()
			defer L.Close()

			gluautf8.Preload(L)

			err := L.DoString("utf8 = require('utf8')")
			r.NoError(err)

			L.SetGlobal("test_input", lua.LString(tt.input))

			var luaCode string
			if tt.lax {
				if tt.i == 0 {
					tt.i = 1
				}

				if tt.j == 0 {
					tt.j = -1
				}

				luaCode = fmt.Sprintf("return utf8.len(test_input, %d, %d, true)", tt.i, tt.j)
			} else {
				if tt.i == 0 {
					luaCode = "return utf8.len(test_input)"
				} else if tt.j == 0 {
					luaCode = fmt.Sprintf("return utf8.len(test_input, %d)", tt.i)
				} else {
					luaCode = fmt.Sprintf("return utf8.len(test_input, %d, %d)", tt.i, tt.j)
				}
			}

			err = L.DoString(luaCode)
			r.NoError(err)

			if tt.expected == nil {
				r.Equal(lua.LNil, L.Get(-2))
				r.Equal(lua.LNumber(tt.pos), L.Get(-1))
				L.Pop(2)
			} else {
				r.Equal(lua.LNumber(tt.expected.(int)), L.Get(-1))
				L.Pop(1)
			}
		})
	}
}

func TestOffset(t *testing.T) { // ToDo: FixMe
	tests := []struct {
		input    string
		n        int
		i        int
		expected any
	}{
		{
			input:    "hello",
			n:        1,
			expected: 1,
		},
		// {
		// 	input:    "hello",
		// 	n:        3,
		// 	expected: 3,
		// },
		// {
		// 	input:    "hello",
		// 	n:        -1,
		// 	expected: 5,
		// },
		// {
		// 	input:    "你好",
		// 	n:        1,
		// 	expected: 1,
		// },
		// {
		// 	input:    "你好",
		// 	n:        2,
		// 	expected: 4,
		// },
		// {
		// 	input:    "你好",
		// 	n:        -1,
		// 	expected: 4,
		// },
		// {
		// 	input:    "你好",
		// 	n:        -2,
		// 	expected: 1,
		// },
		// {
		// 	input:    "hi你好",
		// 	n:        1,
		// 	expected: 1,
		// },
		// {
		// 	input:    "hi你好",
		// 	n:        3,
		// 	expected: 3,
		// },
		// {
		// 	input:    "hi你好",
		// 	n:        4,
		// 	expected: 6,
		// },
		// {
		// 	input:    "你好",
		// 	n:        0,
		// 	i:        2, // middle of first character
		// 	expected: 1,
		// },
		// {
		// 	input:    "你好",
		// 	n:        0,
		// 	i:        4, // start of second character
		// 	expected: 4,
		// },
		// {
		// 	input:    "hello世界",
		// 	n:        1,
		// 	i:        3,
		// 	expected: 3,
		// },
		// {
		// 	input:    "hello世界",
		// 	n:        2,
		// 	i:        3,
		// 	expected: 6,
		// },
		// {
		// 	input:    "hello世界",
		// 	n:        -1,
		// 	i:        6, // starting from '世'
		// 	expected: 3,
		// },
		// {
		// 	input:    "",
		// 	n:        1,
		// 	expected: nil,
		// },
		// {
		// 	input:    "hello",
		// 	n:        10,
		// 	expected: nil,
		// },
		// {
		// 	input:    "hello",
		// 	n:        -10,
		// 	expected: nil,
		// },
		// {
		// 	input:    "你好",
		// 	n:        0,
		// 	i:        10,
		// 	expected: nil,
		// },
		// {
		// 	input:    "hi👋bye",
		// 	n:        3,
		// 	expected: 3,
		// },
		// {
		// 	input:    "hi👋bye",
		// 	n:        4,
		// 	expected: 7,
		// },
		// {
		// 	input:    "hi👋bye",
		// 	n:        -2,
		// 	expected: 7,
		// },
		// {
		// 	input:    "hi👋bye",
		// 	n:        0,
		// 	i:        4, // middle of emoji
		// 	expected: 3,
		// },
		// {
		// 	input:    "hello",
		// 	n:        1,
		// 	i:        1,
		// 	expected: 1,
		// },
		// {
		// 	input:    "你好",
		// 	n:        1,
		// 	i:        1,
		// 	expected: 1,
		// },
		// {
		// 	input:    "hello",
		// 	n:        -1,
		// 	i:        6,
		// 	expected: 5,
		// },
		// {
		// 	input:    "你好",
		// 	n:        -1,
		// 	i:        7,
		// 	expected: 4,
		// },
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q/n=%d/i=%d", i, tt.input, tt.n, tt.i), func(t *testing.T) {
			r := require.New(t)

			L := lua.NewState()
			defer L.Close()

			gluautf8.Preload(L)

			err := L.DoString("utf8 = require('utf8')")
			r.NoError(err)

			L.SetGlobal("test_input", lua.LString(tt.input))

			var luaCode string
			if tt.i != 0 {
				luaCode = fmt.Sprintf("return utf8.offset(test_input, %d, %d)", tt.n, tt.i)
			} else {
				luaCode = fmt.Sprintf("return utf8.offset(test_input, %d)", tt.n)
			}

			err = L.DoString(luaCode)
			r.NoError(err)

			if tt.expected == nil {
				r.Equal(lua.LNil, L.Get(-1))
			} else {
				r.Equal(lua.LNumber(tt.expected.(int)), L.Get(-1))
			}
			L.Pop(1)
		})
	}
}
