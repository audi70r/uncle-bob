package checker

import (
	"github.com/audi70r/uncle-bob/utilities/clog"
)

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
