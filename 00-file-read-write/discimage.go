package main

import (
	"io"
	"log"
	"os"
)

type Span struct {
	from   uint64
	length uint32 // we have this limitation in the API
}

func (s Span) Max() uint64 {
	return s.from + uint64(s.length)
}

func NewSpan(from uint64, length uint32) Span {
	return Span{from: from, length: length}
}

type ReadRequest struct {
	span     Span
	receiver io.Writer
}

type WriteRequest struct {
	span   Span
	sender io.Reader
}

func NewReadRequest(from uint64, length uint32, receiver io.Writer) ReadRequest {
	return ReadRequest{receiver: receiver, span: NewSpan(from, length)}
}

func NewWriteRequest(from uint64, length uint32, sender io.Reader) WriteRequest {
	return WriteRequest{sender: sender, span: NewSpan(from, length)}
}

type Request interface {
	Span() Span
	Fulfil(file os.File) error
}

func (r ReadRequest) Fulfil(file os.File) error {
	file.Seek(int64(r.span.from), os.SEEK_SET)
	_, err := io.CopyN(r.receiver, &file, int64(r.span.length))
	return err
}

func (r WriteRequest) Fulfil(file os.File) error {
	file.Seek(int64(r.span.from), os.SEEK_SET)
	_, err := io.CopyN(&file, r.sender, int64(r.span.length))
	return err
}

func (r ReadRequest) Span() Span  { return r.span }
func (r WriteRequest) Span() Span { return r.span }

func HandleRequest(file os.File, r Request, size uint64) error {
	if r.Span().Max() > size {
		log.Fatal("Request outside file size.")
	}
	return r.Fulfil(file)
}

func RequestLoop(file os.File, msgs chan Request, errs chan error) {
	stat, err := file.Stat()
	fatalIf(err)
	size := uint64(stat.Size())
	for r := range msgs {
		err := HandleRequest(file, r, size)
		if err == nil {
			errs <- nil
		} else {
			// Obvious thing to do here is try to capture the error
			// and feed it back down errs, but can't be bothered right now
			log.Fatal(err)
		}
	}
}

type DiscImage struct {
	file *os.File
	msgs chan Request
	errs chan error
}

func NewDiscImage(filename string) *DiscImage {
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	fatalIf(err)
	d := &DiscImage{file: file, msgs: make(chan Request), errs: make(chan error)}
	go RequestLoop(*file, d.msgs, d.errs)
	return d
}

func (d *DiscImage) Read(from uint64, length uint32, receiver io.Writer) error {
	d.msgs <- NewReadRequest(from, length, receiver)
	return <-d.errs
}

func (d *DiscImage) Write(from uint64, length uint32, sender io.Reader) error {
	d.msgs <- NewWriteRequest(from, length, sender)
	return <-d.errs
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
