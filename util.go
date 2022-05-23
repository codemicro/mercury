package mercury

import "strings"

func splitPath(path string) []string {
	// if we have a single backslash and nothing else, we'll get
	// []string{"", ""} instead of []string{""}, which is what we want for
	// proper path matching.
	if path == "/" {
		return []string{""}
	}
	path = strings.TrimSuffix(path, "/") // if only some of the paths have
	// this and others don't, they won't match when they potentially should
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
