// SPDX-License-Identifier: Apache-2.0

package goiter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunePositionIter(t *testing.T) {
	var (
		text        = "line 1\rline 2\nline3\r\nline44"
		lines       = []string{"line 1", "line 2", "line3", "line44"}
		maxPos      = []int{6, 6, 5, 6}
		iter        = NewRunePositionIter(strings.NewReader(text))
		char        rune
		lineNum     = 0
		lastCharPos = 0
	)

	var lineText strings.Builder
	for iter.Next() {
		if char = iter.Value(); char == '\n' {
			assert.Equal(t, lines[lineNum], lineText.String())
			assert.Equal(t, maxPos[lineNum], lastCharPos)
			lineNum++
			assert.Equal(t, lineNum+1, iter.Line())

			lineText.Reset()
		} else {
			lineText.WriteRune(char)
			lastCharPos = iter.Position()
		}
	}

	assert.Equal(t, len(lines)-1, lineNum)
	assert.Equal(t, len(lines), iter.Line())
	assert.Equal(t, maxPos[len(maxPos)-1], iter.Position())
}
