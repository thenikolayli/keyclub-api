package sync

import (
	"fmt"
	"strconv"
)

// Parsers are the parse functions passed to Normalize for each column type.
var Parsers = struct {
	String func(string) (string, error)
	Float  func(string) (float64, error)
	Int    func(string) (int, error)
	Bool   func(string) (bool, error)
}{
	String: func(s string) (string, error) { return s, nil },
	Float:  func(s string) (float64, error) { return strconv.ParseFloat(s, 64) },
	Int: func(s string) (int, error) {
		v, err := strconv.ParseInt(s, 0, 64)
		return int(v), err
	},
	Bool: strconv.ParseBool,
}

// Normalize turns a ragged sheet column into a fixed-length slice, using zero values for blanks and unparseable cells.
func Normalize[T any](values [][]any, length int, parse func(string) (T, error)) []T {
	out := make([]T, length)
	for i := range values {
		if len(values[i]) == 0 {
			continue
		}
		v, err := parse(fmt.Sprintf("%v", values[i][0]))
		if err == nil {
			out[i] = v
		}
	}
	return out
}
