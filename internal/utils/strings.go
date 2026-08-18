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

func WithoutStrings(arr []string, excluded []string) []string {
	if len(excluded) == 0 {
		return arr
	}

	filtered := make([]string, 0, len(arr))
	for _, value := range arr {
		include := true
		for _, excludedValue := range excluded {
			if value == excludedValue {
				include = false
				break
			}
		}
		if include {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
