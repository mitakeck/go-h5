package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/k0kubun/pp"
)

type FileReader interface {
	io.Reader
	io.Seeker
	io.ReaderAt
}

type Reader struct {
	input    FileReader
	fileName string
}

// parse format signature field
func (r *Reader) parseSignature() error {
	// Decimal:        	   137	72	68	70	13	10  	26	10
	// Hexadecimal:	        89	48	44	46	0d	0a	  1a	0a
	// ASCII C Notation:	\211	H		D		F	  \r  \n  \032	\n
	// https://support.hdfgroup.org/HDF5/doc/H5.format.html

	// get format signature field
	chunk := make([]byte, signatureSize)
	if err := binary.Read(r.input, binary.BigEndian, chunk); err != nil {
		return err
	}

	// check format signature field
	if bytes.Compare(chunk, signatureValue) != 0 {
		return fmt.Errorf("Input File is Not hdf5 Format : %s", r.fileName)
	}

	return nil
}

func newH5Reader(fileName string) (*Reader, error) {
	// check file exist
	_, err := os.Stat(fileName)
	if err != nil {
		return &Reader{}, err
	}

	// file open
	file, err := os.Open(fileName)
	if err != nil {
		return &Reader{}, err
	}
	defer file.Close()

	// data read
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return &Reader{}, err
	}

	reader := new(Reader)
	reader.fileName = fileName
	reader.input = bytes.NewReader(data)

	// check format signature
	if err := reader.parseSignature(); err != nil {
		return reader, err
	}

	return reader, nil
}

func main() {
	reader, err := newH5Reader("sample/e300.h5")
	if err != nil {
		fmt.Println(err)
		return
	}

	pp.Println(reader)
}
