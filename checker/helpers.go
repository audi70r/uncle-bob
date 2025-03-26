package checker

import (
	"github.com/audi70r/uncle-bob/utilities/clog"
)

// contains checks if a string is present in a slice
func contains(s []string, searchterm string) bool {
	for _, x := range s {
		if x == searchterm {
			return true
		}
	}

	return false
}

// containsInCheckResults checks if a message is already in the results
func containsInCheckResults(s []clog.CheckResult, searchterm string) bool {
	for _, x := range s {
		if x.Message == searchterm {
			return true
		}
	}

	return false
}
