package test

import (
	"strings"
	"testing"
)

func AssertLineCountContains(
	t *testing.T,
	output string,
	expected int,
	contains ...string,
) {
	t.Helper()

	count := 0

	for _, line := range strings.Split(output, "\n") {
		matches := true

		for _, s := range contains {
			if !strings.Contains(line, s) {
				matches = false
				break
			}
		}

		if matches {
			count++
		}
	}

	if count != expected {
		t.Errorf(
			"expected %d lines containing %v, got %d",
			expected,
			contains,
			count,
		)
	}
}
