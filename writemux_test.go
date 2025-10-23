package writemux

import (
	"errors"
	"io"
	"testing"

	"github.com/cognusion/go-recyclable/v2"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWrite(t *testing.T) {

	rb := recyclable.NewBufferPool()
	hand := "Have a nice day!"

	Convey("When a WriteMux is created, but no buffers are added, and a write is made, it succeeds errantly.", t, func() {
		mux := NewWriteMux()

		// Write to the mux
		n, err := io.WriteString(&mux, hand)
		So(n, ShouldEqual, len(hand))
		So(err, ShouldBeNil)

		mux.Close()
	})

	Convey("When a WriteMux is created WithErrors, but no buffers are added, and a write is made, it returns the appropriate error.", t, func() {
		mux := NewWriteMuxWithErrors()

		// Write to the mux
		n, err := io.WriteString(&mux, hand)
		So(n, ShouldEqual, 0)
		So(err, ShouldEqual, ErrorNoWriters)

		mux.Close()
	})

	Convey("When a WriteMux is created WithErrors, and a write is made, it succeeds and looks correct.", t, func() {
		mux := NewWriteMuxWithErrors()
		buf := rb.Get()
		buf.Reset([]byte{})

		mux.Add(buf)

		// Write to the mux
		n, err := io.WriteString(&mux, hand)
		So(n, ShouldEqual, len(hand))
		So(err, ShouldBeNil)

		// Check the buffer
		So(buf.String(), ShouldEqual, hand)

		Convey("... when the mux is closed (writers are removed), and a write is made, it returns the appropriate error", func() {
			mux.Close()

			n, err := io.WriteString(&mux, hand)
			So(n, ShouldEqual, 0)
			So(err, ShouldEqual, ErrorNoWriters)
		})

	})

	Convey("When a WriteMux is created WithErrors, and a write is made, it succeeds and looks correct.", t, func() {
		mux := NewWriteMuxWithErrors()
		buf := rb.Get()
		buf.Reset([]byte{})

		bindex := mux.Add(buf)

		// Write to the mux
		n, err := io.WriteString(&mux, hand)
		So(n, ShouldEqual, len(hand))
		So(err, ShouldBeNil)

		// Check the buffer
		So(buf.String(), ShouldEqual, hand)

		Convey("... when the writer is explicitly removed, and a write is made, it returns the appropriate error", func() {
			mux.Remove(bindex)

			n, err := io.WriteString(&mux, hand)
			So(n, ShouldEqual, 0)
			So(err, ShouldEqual, ErrorNoWriters)
		})

	})

	Convey("When a WriteMux is created WithErrors, and a bad writer is added, and a write is made, it errors out.", t, func() {
		mux := NewWriteMuxWithErrors()

		mux.Add(&noopCloser{&badWriter{}})

		// Write to the mux
		_, err := io.WriteString(&mux, hand)
		So(err, ShouldNotBeNil)

	})
}

func BenchmarkWrite1(b *testing.B) {
	mux := NewWriteMux()
	mux.Add(&noopCloser{io.Discard})

	var hello = []byte("Hello World")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.Write(hello)
	}
}

func BenchmarkWrite10(b *testing.B) {
	mux := NewWriteMux()
	for range 10 {
		mux.Add(&noopCloser{io.Discard})
	}

	var hello = []byte("Hello World")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.Write(hello)
	}
}

func BenchmarkWrite100(b *testing.B) {
	mux := NewWriteMux()
	for range 100 {
		mux.Add(&noopCloser{io.Discard})
	}

	var hello = []byte("Hello World")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.Write(hello)
	}
}

// noopCloser wraps a Writer and makes it a WriteCloser
type noopCloser struct {
	io.Writer
}

func (n *noopCloser) Close() error {
	return nil
}

// badWriter is a Writer that always returns an error
type badWriter struct {
}

func (b *badWriter) Write(p []byte) (int, error) {
	return 0, errors.New("this is a bad error")
}
