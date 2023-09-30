package parser

import (
	"strings"
	"text/scanner"
	"unicode"
)

var reservedCharacters map[rune]struct{} = map[rune]struct{}{
	'+': struct{}{},
	'-': struct{}{},
	'*': struct{}{},
	'/': struct{}{},
	'(': struct{}{},
	')': struct{}{},
}

func NewScanner(src string, filename string) scanner.Scanner {
	var s scanner.Scanner

	s.Init(strings.NewReader(src))
	s.Filename = filename
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats

	s.IsIdentRune = func(ch rune, i int) bool {
		if unicode.IsLetter(ch) {
			return true
		}

		if i > 0 {
			_, isReservedCharacted := reservedCharacters[ch]
			return i > 0 && !isReservedCharacted && !unicode.IsSpace(ch) && unicode.IsPrint(ch)
		}

		return false
	}

	return s
}
