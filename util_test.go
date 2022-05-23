package mercury

import (
	"reflect"
	"testing"
)

func Test_splitPath(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want []string
	}{
		{"normal", "/hello/world", []string{"", "hello", "world"}},
		{"singleBackslash", "/", []string{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitPath(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
