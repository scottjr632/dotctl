package utils

import "strings"

func FilterStrings(arr []string, filter string) []string {
	if filter == "" {
		return arr
	}

	filtered := []string{}
	for _, s := range arr {
		if strings.Contains(s, filter) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func WithoutStrings(arr []string, filter []string) []string {
	if len(filter) == 0 {
		return arr
	}

	filtered := []string{}
	for _, s := range arr {
		shouldInclude := true
		for _, f := range filter {
			if s == f {
				shouldInclude = false
			}
		}
		if shouldInclude {
			filtered = append(filtered, s)
		}
		return filtered
	}
	return filtered
}
