package dpkg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"archive/tar"

	"github.com/blakesmith/ar"
	"github.com/mholt/archiver"
)

// NewReader return control and data
func NewReader(r io.Reader) (control, data *TarReader, err error) {
	arr := ar.NewReader(r)
	for {
		var head *ar.Header
		head, err = arr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if strings.HasSuffix(head.Name, "/") {
			head.Name = head.Name[:len(head.Name)-1]
		}
		switch {
		case strings.HasPrefix(head.Name, "control"):
			var controlBuff bytes.Buffer
			io.Copy(&controlBuff, arr)
			if err != nil {
				log.Println("read all", err)
			}
			control, err = NewTarReader(head.Name, &controlBuff)
			if err != nil {
				return
			}
		case strings.HasPrefix(head.Name, "data"):
			data, err = NewTarReader(head.Name, arr)
			if err != nil {
				return
			}
			if control != nil {
				return
			}
		}
	}
	if control == nil || data == nil {
		err = errors.New("missing file")
	}
	return control, data, nil
}

// NewTarReader create tar reader from stream, support tar.gz tar.xz tar.lz
func NewTarReader(name string, r io.Reader) (*TarReader, error) {
	type Archiver interface {
		archiver.Reader
		archiver.Archiver
	}
	as := []Archiver{
		archiver.NewTar(), archiver.NewTarGz(), archiver.NewTarXz(), archiver.NewTarLz4(),
	}
	for i := range as {
		if as[i].CheckExt(name) == nil {
			err := as[i].Open(r, 0)
			if err != nil {
				return nil, err
			}
			return &TarReader{a: as[i]}, nil
		}
	}
	return nil, fmt.Errorf("unknown type (%s)", name)
}

// TarReader implements tar.reader
type TarReader struct {
	a archiver.Reader
	f archiver.File
}

// Next like tar next
func (tr *TarReader) Next() (*tar.Header, error) {
	f, err := tr.a.Read()
	if err != nil {
		return nil, err
	}
	tr.f = f
	return f.Header.(*tar.Header), nil
}

// Read like tar Read
func (tr *TarReader) Read(b []byte) (int, error) {
	n, err := tr.f.Read(b)
	return n, err
}
