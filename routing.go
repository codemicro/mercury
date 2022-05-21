package mercury

import "strings"

func (app *App) getHandlerForPath(path string) *handler {
	path = strings.ToLower(path)
	for _, h := range app.callstack {
		if h.path == path {
			return h
		}
	}
	return nil
}
