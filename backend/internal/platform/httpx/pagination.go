package httpx

import (
	"net/http"
	"strconv"
)

func ParsePageSize(r *http.Request, defaultSize, maxSize int) (int, int) {
	page := 1
	size := defaultSize

	if rawPage := r.URL.Query().Get("page"); rawPage != "" {
		if parsed, err := strconv.Atoi(rawPage); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if rawSize := r.URL.Query().Get("size"); rawSize != "" {
		if parsed, err := strconv.Atoi(rawSize); err == nil && parsed > 0 {
			size = parsed
		}
	}

	if size > maxSize {
		size = maxSize
	}
	return page, size
}

func Offset(page, size int) int {
	if page <= 1 {
		return 0
	}
	return (page - 1) * size
}
