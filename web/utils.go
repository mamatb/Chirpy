package web

import "strings"

func cleanProfanities(body string, profanities map[string]bool) string {
	bodySlice := strings.Split(body, space)
	for wordIdx, word := range bodySlice {
		if profanities[strings.ToLower(word)] {
			bodySlice[wordIdx] = profanitiesReplacement
		}
	}
	return strings.Join(bodySlice, space)
}
