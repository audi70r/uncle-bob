package checker

import "github.com/audi70r/go-archangel/utilities/clog"

func contains(s []string, searchterm string) bool {
	for _, x := range s {
		if x == searchterm {
			return true
		}
	}

	return false
}

func containsInCheckResults(s []clog.CheckResult, searchterm string) bool {
	for _, x := range s {
		if x.Message == searchterm {
			return true
		}
	}

	return false
}

func remove(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
