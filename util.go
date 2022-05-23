package mercury

import "strings"

func splitPath(path string) []string {
	// if we have a single backslash and nothing else, we'll get
	// []string{"", ""} instead of []string{""}, which is what we want for
	// proper path matching.
	if path == "/" {
		return []string{""}
	}
	return strings.Split(path, "/")
}

func sliceHasPrefix[T comparable](s []T, prefix []T) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := 0; i < len(prefix); i += 1 {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

func sliceEqual[T comparable](s []T, v []T) bool {
	if len(s) != len(v) {
		return false
	}
	for i := 0; i < len(s); i += 1 {
		if s[i] != v[i] {
			return false
		}
	}
	return true
}
