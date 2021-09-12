package main

import "testing"

func TestIsWord(t *testing.T) {
	tests := make(map[string]bool)
	tests[""] = false
	tests["1"] = false
	tests["hello"] = true
	tests["\""] = false
	tests["\"hello\""] = true
	tests["!!!!!!!"] = false

	for s, res := range tests {
		if isWord(s) != res {
			t.Fatalf("String %s returned %t", s, !res)
		}
	}
}

func TestCapitalize(t *testing.T) {
	tests := make(map[string]string)
	tests["hello"] = "Hello"
	tests[""] = ""
	tests["123"] = "123"
	tests["\"hello\""] = "\"Hello\""
	tests["Hello"] = "Hello"

	for s, res := range tests {
		if capitalize(s) != res {
			t.Fatalf("String %s returned %s", s, capitalize(s))
		}
	}
}
