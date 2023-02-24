package main

import (
	"math/rand"
	"strings"
	"unicode"

	"golang.org/x/exp/slices"
)

var endings []string

func init() {
	endings = []string{"!", ".", "?"}
}

// makes the words pretty! cleans up formatting errors, etc.
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

	//ensure last word has ending punctuation
	for i := len(pretty) - 1; i >= 0; i-- {
		if isWord(pretty[i]) {
			pretty[i] = punctuate(pretty[i])
			break
		}
	}

	return
}

// returns true if word contains any letters. otherwise i assume we're
// dealing with an emoji or a single exclamation point or something else.
func isWord(word string) bool {
	for _, rune := range word {
		if unicode.IsLetter(rune) {
			//urls are not words. TODO: this could be much more comprehensive and actually work, but whatever.
			if strings.HasPrefix(word, "http") || strings.HasPrefix(word, "www") || strings.Contains(word, ".com") {
				return false
			} else {
				return true
			}
		}
	}

	return false
}

// returns the capitalized word.
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

// adds sentence-ending punctuation to a word
func punctuate(word string) (punc string) {
	letters := strings.Split(word, "")

	//trim current trailing symbols
	for i := len(letters) - 1; i >= 0; i-- {
		//if already punctuated, return
		if slices.Contains(endings, letters[len(letters)-1]) {
			break
		}

		var letterAsRune rune
		for _, letter := range letters[len(letters)-1] { //trick to convert string character to rune.
			letterAsRune = letter
		}
		if unicode.IsPunct(letterAsRune) {
			letters = letters[:len(letters)-1]
		} else {
			letters = append(letters, endings[rand.Intn(len(endings))])
			break
		}
	}

	punc = strings.Join(letters, "")
	return
}
