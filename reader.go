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

type Superblock struct {
	// TODO: too long variable name
	SuperblockVersin                 int
	FreeSpaceInformationVersion      int
	RootGroupSymbolTableEntryVersion int
	SharedHeaderMessageFormatVersion int
	offset                           int
	length                           int
	leafNodeK                        int
	InternalNodeK                    int
}

type Reader struct {
	input    FileReader
	fileName string

	superblock Superblock
}

func (r *Reader) readChunk(size int, offset int64) ([]byte, error) {
	r.input.Seek(offset, os.SEEK_SET)

	chunk := make([]byte, size)
	if err := binary.Read(r.input, binary.BigEndian, chunk); err != nil {
		return nil, err
	}
	return chunk, nil
}

// parse Superblock field
func (r *Reader) parseSuperblock() error {
	// Decimal:            137   72   68   70   13   10    26   10
	// Hexadecimal:         89   48   44   46   0d   0a    1a   0a
	// ASCII C Notation:  \211    H    D    F   \r   \n  \032   \n
	// https://support.hdfgroup.org/HDF5/doc/H5.format.html

	// get format signature field
	chunk, err := r.readChunk(signatureSize, 0)
	if err != nil {
		return err
	}

	// check format signature field
	if bytes.Compare(chunk, signatureValue) != 0 {
		return fmt.Errorf("Input File is Not hdf5 Format : %s", r.fileName)
	}

	var hoge int

	// get version number of the Superblock
	chunk, err = r.readChunk(2, 8)
	if err != nil {
		return err
	}
	binary.Read(bytes.NewBuffer(chunk), binary.BigEndian, &hoge)
	r.superblock.SuperblockVersin = hoge

	// get version number of file's free space storage
	chunk, err = r.readChunk(2, 10)
	if err != nil {
		return err
	}
	binary.Read(bytes.NewBuffer(chunk), binary.BigEndian, &r.superblock.FreeSpaceInformationVersion)

	// get version number of root group symbol table entry
	chunk, err = r.readChunk(2, 12)
	if err != nil {
		return err
	}
	binary.Read(bytes.NewBuffer(chunk), binary.BigEndian, &r.superblock.RootGroupSymbolTableEntryVersion)

	// get version number of share header mesage format
	chunk, err = r.readChunk(2, 16)
	if err != nil {
		return err
	}
	binary.Read(bytes.NewBuffer(chunk), binary.BigEndian, &r.superblock.SharedHeaderMessageFormatVersion)

	// get size of offsets
	chunk, err = r.readChunk(2, 18)
	if err != nil {
		return err
	}
	binary.Read(bytes.NewBuffer(chunk), binary.BigEndian, &r.superblock.offset)

	// get size of lengths
	chunk, err = r.readChunk(2, 20)
	if err != nil {
		return err
	}
	binary.Read(bytes.NewBuffer(chunk), binary.BigEndian, &hoge)
	r.superblock.length = hoge

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
	if err := reader.parseSuperblock(); err != nil {
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
