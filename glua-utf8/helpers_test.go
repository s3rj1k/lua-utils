package gluautf8_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	gluautf8 "github.com/projectsveltos/lua-utils/glua-utf8"
)

func TestGetRuneValues(t *testing.T) { // https://cryptii.com/pipes/text-decimal
	tests := []struct {
		input    string
		expected []int
	}{
		{
			input:    "",
			expected: []int{},
		},
		{
			input:    "hello",
			expected: []int{104, 101, 108, 108, 111},
		},
		{
			input:    "hi👋",
			expected: []int{104, 105, 128075},
		},
		{
			input:    "café",
			expected: []int{99, 97, 102, 233},
		},
		{
			input:    "你好",
			expected: []int{20320, 22909},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, tt.input), func(t *testing.T) {
			got := gluautf8.GetRuneValues(tt.input)
			require.Equal(t, tt.expected, got, "runes should match for input %q", tt.input)
		})
	}
}

func TestGetRuneValuesAsStrings(t *testing.T) { // https://cryptii.com/pipes/text-decimal
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "hello",
			expected: []string{"104", "101", "108", "108", "111"},
		},
		{
			input:    "hi👋",
			expected: []string{"104", "105", "128075"},
		},
		{
			input:    "café",
			expected: []string{"99", "97", "102", "233"},
		},
		{
			input:    "你好",
			expected: []string{"20320", "22909"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, tt.input), func(t *testing.T) {
			got := gluautf8.GetRuneValuesAsStrings(tt.input)
			require.Equal(t, tt.expected, got, "runes should match for input %q", tt.input)
		})
	}
}

func TestGetRunePositions(t *testing.T) {
	tests := []struct {
		input    string
		lax      bool
		offset   int
		expected []gluautf8.RunePosition
	}{
		{
			input:    "",
			lax:      false,
			offset:   0,
			expected: []gluautf8.RunePosition{},
		},
		{
			input:  "hello",
			lax:    false,
			offset: 0,
			expected: []gluautf8.RunePosition{
				{Position: 0, CodePoint: 104},
				{Position: 1, CodePoint: 101},
				{Position: 2, CodePoint: 108},
				{Position: 3, CodePoint: 108},
				{Position: 4, CodePoint: 111},
			},
		},
		{
			input:  "hi👋",
			lax:    false,
			offset: 0,
			expected: []gluautf8.RunePosition{
				{Position: 0, CodePoint: 104},
				{Position: 1, CodePoint: 105},
				{Position: 2, CodePoint: 128075},
			},
		},
		{
			input:  "café",
			lax:    false,
			offset: 0,
			expected: []gluautf8.RunePosition{
				{Position: 0, CodePoint: 99},
				{Position: 1, CodePoint: 97},
				{Position: 2, CodePoint: 102},
				{Position: 3, CodePoint: 233},
			},
		},
		{
			input:  "你好",
			lax:    false,
			offset: 0,
			expected: []gluautf8.RunePosition{
				{Position: 0, CodePoint: 20320},
				{Position: 3, CodePoint: 22909},
			},
		},
		{
			input:  "hello",
			lax:    false,
			offset: 10,
			expected: []gluautf8.RunePosition{
				{Position: 10, CodePoint: 104},
				{Position: 11, CodePoint: 101},
				{Position: 12, CodePoint: 108},
				{Position: 13, CodePoint: 108},
				{Position: 14, CodePoint: 111},
			},
		},
		{
			input:  "hello\xFF\xFEworld",
			lax:    true,
			offset: 0,
			expected: []gluautf8.RunePosition{
				{Position: 0, CodePoint: 104},
				{Position: 1, CodePoint: 101},
				{Position: 2, CodePoint: 108},
				{Position: 3, CodePoint: 108},
				{Position: 4, CodePoint: 111},
				{Position: 5, CodePoint: 0xFFFD}, // RuneError
				{Position: 6, CodePoint: 0xFFFD}, // RuneError
				{Position: 7, CodePoint: 119},
				{Position: 8, CodePoint: 111},
				{Position: 9, CodePoint: 114},
				{Position: 10, CodePoint: 108},
				{Position: 11, CodePoint: 100},
			},
		},
		{
			input:  "hello\xFF\xFEworld",
			lax:    false,
			offset: 0,
			expected: []gluautf8.RunePosition{
				{Position: 0, CodePoint: 104},
				{Position: 1, CodePoint: 101},
				{Position: 2, CodePoint: 108},
				{Position: 3, CodePoint: 108},
				{Position: 4, CodePoint: 111},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, tt.input), func(t *testing.T) {
			got := gluautf8.GetRunePositions(tt.input, tt.lax, tt.offset)
			require.Equal(t, tt.expected, got, "rune positions should match for input %q", tt.input)
		})
	}
}

func TestCountChars(t *testing.T) {
	tests := []struct {
		input    string
		start    int
		end      int
		expected int
	}{
		{
			input:    "",
			start:    0,
			end:      0,
			expected: 0,
		},
		{
			input:    "hello",
			start:    0,
			end:      4,
			expected: 5,
		},
		// { // ToDo: FixMe
		// 	input:    "héllo世界",
		// 	start:    0,
		// 	end:      6,
		// 	expected: 7,
		// },
		{
			input:    "hello",
			start:    1,
			end:      3,
			expected: 3,
		},
		// { // ToDo: FixMe
		// 	input:    "héllo世界",
		// 	start:    1,
		// 	end:      4,
		// 	expected: 4,
		// },
		{
			input:    "hello",
			start:    -1,
			end:      4,
			expected: 5,
		},
		{
			input:    "hello",
			start:    0,
			end:      -1,
			expected: 5,
		},
		{
			input:    "hello",
			start:    3,
			end:      1,
			expected: 0,
		},
		{
			input:    "hello",
			start:    10,
			end:      15,
			expected: 0,
		},
		{
			input:    string([]byte{0xFF, 0xFE, 0xFD}),
			start:    0,
			end:      2,
			expected: -1,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case=%d/string=%q", i, tt.input), func(t *testing.T) {
			got := gluautf8.CountChars(tt.input, tt.start, tt.end)
			require.Equal(t, tt.expected, got, "rune count should match for input %q", tt.input)
		})
	}
}
