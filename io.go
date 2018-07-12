package ptk

import (
	"io"
)

// Pipe is an io.Pipe helper
func Pipe(rfn func(io.Reader) error, wfn func(io.Writer) error) error {
	rd, rw := io.Pipe()
	go func() { rw.CloseWithError(wfn(rw)) }()
	return rfn(rd)
}
