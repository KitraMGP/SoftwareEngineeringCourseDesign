package model

import (
	"strconv"
	"strings"
)

func FormatVector(values []float32) string {
	if len(values) == 0 {
		return "[]"
	}

	var builder strings.Builder
	builder.Grow(len(values) * 10)
	builder.WriteByte('[')
	for idx, value := range values {
		if idx > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.FormatFloat(float64(value), 'f', -1, 32))
	}
	builder.WriteByte(']')
	return builder.String()
}
