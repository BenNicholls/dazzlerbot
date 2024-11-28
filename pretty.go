package main

import (
	"math/rand"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

var allowed_endings []string
var disallowed_endings []string

func init() {
	allowed_endings = []string{"!", ".", "?"}
	disallowed_endings = []string{","}
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

	//ensure last word has ending punctuation, unless sentence is a single word long
	if len(pretty) > 1 {
		for i, word := range slices.Backward(pretty) {
			if isWord(word) {
				pretty[i] = punctuate(word)
				break
			}
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

	for _, letter := range slices.Backward(letters) {
		//if already punctuated, return
		if slices.Contains(allowed_endings, letter) {
			break
		}

		//trim weird punctuation?
		if slices.Contains(disallowed_endings, letter) {
			letters = letters[:len(letters)-1]
		} else if unicode.IsPunct(stringAsRune(letter)) {
			//println("unicode thinks this is punctuation: ", letter)
			//letters = letters[:len(letters)-1]
			//TODO: handle random weirdo punctuation here
		} else {
			letters = append(letters, allowed_endings[rand.Intn(len(allowed_endings))])
			break
		}
	}

	punc = strings.Join(letters, "")

	return
}

// returns the rune for the first character in string s.
// does NOT check to see if s is a proper string, so be careful
func stringAsRune(s string) (r rune) {
	r, _ = utf8.DecodeRuneInString(s)
	return
}
