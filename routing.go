package mercury

import (
	"fmt"
)

func doesHandlerMatchPath(path []string, h *handler) bool {
	fmt.Printf("%#v %#v\n", path, h)
	if h.isMiddleware {
		return sliceHasPrefix(path, h.pathComponents)
	}
	return sliceEqual(path, h.pathComponents)
}
