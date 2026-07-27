package worktree

import (
	"reflect"
	"runtime"
	"testing"

	"golang.org/x/term"
)

func normalizedTerminalState(state *term.State) []uint64 {
	var values []uint64
	var collect func(reflect.Value, string)
	collect = func(value reflect.Value, name string) {
		switch value.Kind() {
		case reflect.Pointer:
			collect(value.Elem(), name)
		case reflect.Struct:
			valueType := value.Type()
			for i := 0; i < value.NumField(); i++ {
				collect(value.Field(i), valueType.Field(i).Name)
			}
		case reflect.Array:
			for i := 0; i < value.Len(); i++ {
				collect(value.Index(i), name)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			fieldValue := value.Uint()
			if runtime.GOOS == "darwin" && name == "Lflag" {
				fieldValue &^= 0x20000000 // PENDIN is a transient input-reprint flag.
			}
			values = append(values, fieldValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values = append(values, uint64(value.Int()))
		}
	}
	collect(reflect.ValueOf(state), "")
	return values
}

func TestParseListOptionsRecognizesPlain(t *testing.T) {
	options, err := parseListOptions([]string{"--sort", "head", "--root-position", "bottom", "--reverse", "--plain"})
	if err != nil {
		t.Fatal(err)
	}
	if options.sortBy != "head" || options.rootPosition != "bottom" || !options.reverse || !options.plain {
		t.Fatalf("options = %#v", options)
	}
}
