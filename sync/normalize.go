package sync

import (
	"fmt"
	"strconv"
)

// the following functions create normalized lists with blanks from the originals

func NormalizeBoolValues(values [][]any, length int) []bool {
	normalizedStringValues := make([]bool, length)

	for i := range values {
		if len(values[i]) != 0 { // if the cell isn't blank
			value, err := strconv.ParseBool(values[i][0].(string))
			if err == nil {
				normalizedStringValues[i] = bool(value)
			}
		}
	}

	return normalizedStringValues
}

func NormalizeFloatValues(values [][]any, length int) []float64 {
	normalizedFloatValues := make([]float64, length)

	for i := range values {
		if len(values[i]) != 0 { // if the cell isn't blank
			value, err := strconv.ParseFloat(fmt.Sprintf("%v", values[i][0]), 64)
			if err == nil {
				normalizedFloatValues[i] = value
			}
		}
	}

	return normalizedFloatValues
}

func NormalizeIntValues(values [][]any, length int) []int {
	normalizedIntValues := make([]int, length)

	for i := range values {
		if len(values[i]) != 0 { // if the cell isn't blank
			value, err := strconv.ParseInt(fmt.Sprintf("%v", values[i][0]), 0, 64)
			if err == nil {
				normalizedIntValues[i] = int(value)
			}
		}
	}

	return normalizedIntValues
}

func NormalizeStringValues(values [][]any, length int) []string {
	normalizedStringValues := make([]string, length)

	for i := range values {
		if len(values[i]) != 0 { // if the cell isn't blank
			normalizedStringValues[i] = fmt.Sprintf("%v", values[i][0])
		}
	}

	return normalizedStringValues
}
