package testx

import (
	"reflect"
	"testing"
)

func ResultOf[T any](tb testing.TB, f func() (T, error)) T {
	tb.Helper()

	t, err := f()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}

	return t
}

func ResultOfA[T any](tb testing.TB, f any, args ...any) (t T) {
	tb.Helper()

	//nolint:exhaustive // we are only interested in functions
	switch reflect.TypeOf(f).Kind() {
	case reflect.Func:
		// do nothing
	default:
		tb.Fatalf("expected a function, got %T", f)
	}

	rargs := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		rargs = append(rargs, reflect.ValueOf(arg))
	}

	rf := reflect.ValueOf(f)

	result := rf.Call(rargs)

	switch len(result) {
	case 0:
		tb.Fatalf("expected at least one return value, got none")
	case 1:
		return result[0].Interface().(T)
	case 2:
		if result[1].IsNil() {
			return result[0].Interface().(T)
		}
		tb.Fatalf("unexpected error: %v", result[1].Interface().(error))
	default:
		tb.Fatalf("expected at most two return values, got %d", len(result))
	}

	return t
}
