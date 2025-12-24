package utils

import "strings"

// Check if string is part of the array
func CheckStringInArray(value string, values []string) bool {
	for _, entry := range values {
		if strings.Compare(entry, value) == 0 {
			return true
		}
	}
	return false
}
