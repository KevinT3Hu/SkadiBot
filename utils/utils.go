package utils

import "strings"

func stripeString(s string) string {
	// remove leading and trailing whitespaces
	return strings.TrimSpace(s)
}
