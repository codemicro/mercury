package mercury

import (
	"reflect"
	"testing"
)

func TestApp_getHandlerForPath(t *testing.T) {
	csh := &handler{nil, "/b"}

	tests := []struct {
		name      string
		callstack []*handler
		args      string
		want      *handler
	}{
		{"normal", []*handler{{nil, "/a"}, csh, {nil, "/c"}}, "/b", csh},
		{"none", []*handler{{nil, "/a"}, csh, {nil, "/c"}}, "/d", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				callstack: tt.callstack,
			}
			if got := app.getHandlerForPath(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getHandlerForPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
