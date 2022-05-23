package mercury

import (
	"fmt"
	"strings"
)

func doesHandlerMatchPath(path []string, h *handler) bool {
	fmt.Printf("%#v %#v\n", path, h)
	if h.isMiddleware {
		if len(h.pathComponents) > len(path) {
			return false
		}
		return doPathComponentsMatch(path[:len(h.pathComponents)], h.pathComponents)
	}

	return doPathComponentsMatch(path, h.pathComponents)
}

func doPathComponentsMatch(input, source []string) bool {
	if len(input) != len(source) {
		return false
	}
	for i := 0; i < len(input); i += 1 {
		if !(input[i] == source[i] || strings.HasPrefix(source[i], ":")) {
			return false
		}
	}
	return true
}
