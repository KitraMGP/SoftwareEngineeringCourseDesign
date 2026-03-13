package httpx

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusRecorderImplementsFlusher(t *testing.T) {
	t.Parallel()

	base := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	recorder := &statusRecorder{ResponseWriter: base, status: http.StatusOK}

	flusher, ok := any(recorder).(http.Flusher)
	if !ok {
		t.Fatal("statusRecorder should implement http.Flusher")
	}

	flusher.Flush()
	if !base.flushed {
		t.Fatal("Flush() should delegate to the wrapped response writer")
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (r *flushRecorder) Flush() {
	r.flushed = true
}

func (r *flushRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, http.ErrNotSupported
}
