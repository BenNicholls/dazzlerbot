package main

import (
	"unicode"
)

//makes the words pretty! cleans up formatting errors, etc.
func prettify(words []string) (pretty []string) {
	pretty = make([]string, len(words))
	for i, word := range words {
		pretty[i] = word
	}

	//ensure first word is capitalized
	for i, word := range words {
		if isWord(word) {
			pretty[i] = capitalize(word)
			break
		}
	}

	return
}

//returns true if word contains any letters. otherwise i assume we're
//dealing with an emoji or a single exclamation point or something else.
func isWord(word string) bool {
	for _, rune := range word {
		if unicode.IsLetter(rune) {
			return true
		}
	}

	return false
}

//returns the capitalized word.
func capitalize(word string) (cap string) {
	capped := false
	for i, rune := range word {
		if capped {
			cap += word[i:]
			break
		}
		if !unicode.IsLetter(rune) {
			cap += string(rune)
		} else {
			cap += string(unicode.ToUpper(rune))
			capped = true
		}
	}
	return
}
