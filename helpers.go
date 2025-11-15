package main

import (
	"strings"
)

func replaceBadWords(msg string) string {
	words := strings.Split(msg, " ")

	for i, word := range words {
		lowerCaseWord := strings.ToLower(word)
		if lowerCaseWord == "kerfuffle" || lowerCaseWord == "sharbert" || lowerCaseWord == "fornax" {
			words[i] = "****"
		}
	}

	result := strings.Join(words, " ")
	return result
}