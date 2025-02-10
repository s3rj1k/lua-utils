package gluautf8

import (
	"fmt"
	"unicode/utf8"
)

type RunePosition struct {
	Position  int
	CodePoint int
}

func GetRuneValues(s string) []int {
	out := make([]int, 0)

	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		out = append(out, int(r))
		s = s[size:]
	}

	return out
}

func GetRuneValuesAsStrings(s string) []string {
	in := GetRuneValues(s)
	out := make([]string, len(in))

	for i, r := range in {
		out[i] = fmt.Sprintf("%d", r)
	}

	return out
}

func GetRunePositions(s string, lax bool, offset int) []RunePosition {
	out := make([]RunePosition, 0)
	pos := 0

	for pos < len(s) {
		r, size := utf8.DecodeRuneInString(s[pos:])
		if r == utf8.RuneError && !lax {
			break
		}

		out = append(out, RunePosition{
			Position:  pos + offset,
			CodePoint: int(r),
		})

		pos += size
	}

	return out
}

func CountChars(s string, start, end int) int {
	if start < 0 {
		start = 0
	}

	if end < 0 {
		end = len(s) - 1
	}

	count := 0
	pos := 0

	// skip to start position
	for pos < len(s) && pos < start {
		_, size := utf8.DecodeRuneInString(s[pos:])
		pos += size
	}

	// count runes between start and end
	for pos <= end && pos < len(s) {
		r, size := utf8.DecodeRuneInString(s[pos:])
		if r == utf8.RuneError {
			return -1
		}

		count++
		pos += size
	}

	return count
}
