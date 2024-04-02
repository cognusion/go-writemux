package writemux

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cognusion/go-sequence"
)

var (
	// ErrorNoWriters is returned by a WriteMux created by NewWriteMuxWithErrors,
	// when there are zero writers in the mux.
	ErrorNoWriters = errors.New("there are no writer in the mux")
)

// WriteMux is a WriteCloser that can have WriteClosers added and removed
// from the group, so that Write() and Close() calls can be mirrored to all
// mux members. All operations after proper initialization are goro-safe.
type WriteMux struct {
	mutex        sync.Mutex
	writers      map[string]io.WriteCloser
	seq          *sequence.Seq
	handleErrors bool
}

// NewWriteMux returns an initialized WriteMux. Writes that generate errors
// are ignored.
func NewWriteMux() WriteMux {
	return WriteMux{
		writers: make(map[string]io.WriteCloser),
		seq:     sequence.New(1),
	}
}

// NewWriteMuxWithErrors returns an initialized WriteMux. Writes that generate errors
// cause the Write to fail, returning a wrapped error with the index of the member causing
// the error. Subsequent members are not written to.
func NewWriteMuxWithErrors() WriteMux {
	return WriteMux{
		writers:      make(map[string]io.WriteCloser),
		seq:          sequence.New(1),
		handleErrors: true,
	}
}

// Add will add the WriteCloser to the mux, returning an index that can be used
// in the future to remove the WriteCloser from the mux. The index should not
// be intuited as any usable value beyond the future parameter to Remove().
func (w *WriteMux) Add(wc io.WriteCloser) (index string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	index = w.seq.NextHashID()
	w.writers[index] = wc
	return index
}

// Remove takes the specified index, and removes it from the mux.
func (w *WriteMux) Remove(index string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.writers, index)
}

// Write sequentially writes len(p) bytes to each WriteCloser in the mux.
// Depending which New... method was used to create the mux, the writes are blind
// or are passed back, halting the range.
func (w *WriteMux) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.handleErrors && len(w.writers) == 0 {
		n = 0
		err = ErrorNoWriters
		return
	}

	for index, writer := range w.writers {
		_, rerr := writer.Write(p)
		if rerr != nil && w.handleErrors {
			err = fmt.Errorf("error during mux write to '%s': %w", index, rerr)
			return
		}
	}
	return
}

// Close sequentially closes all members of the mux.
func (w *WriteMux) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for index, writer := range w.writers {
		writer.Close()           // close the writer
		delete(w.writers, index) // remove the writer from the mux
	}
	return nil
}
