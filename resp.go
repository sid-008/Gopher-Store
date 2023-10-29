package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
) // types we will be handling

type Value struct {
	typ string // evaluate data type
	// str   string
	// num   int
	bulk  string
	array []Value
} // this struct is used in the serialisation/deserialisation process

type Resp struct {
	reader *bufio.Reader
} // This is a reader we can spawn to parse the buffer

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader((rd))}
} // spawn a new reader

func (r *Resp) readline() (line []byte, n int, err error) { // basically reads all bytes till a crlf is encountered
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}

		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' { // stop parsing when crlf is encountered
			break
		}
	}
	return line[:len(line)-2], n, nil // return contents without the crlf
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readline() // read the byte
	if err != nil {
		return 0, 0, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64) // convert from string to int64
	if err != nil {
		return 0, n, err
	}

	return int(i64), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte() // parse type

	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = "array"

	len, _, err := r.readInteger() // len = number of elements in array
	if err != nil {
		return v, err
	}

	v.array = make([]Value, 0)
	for i := 0; i < len; i++ {
		val, err := r.Read() // read that particular element

		if err != nil {
			return v, err
		}

		v.array = append(v.array, val)
	}
	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, len)

	r.reader.Read(bulk)

	v.bulk = string(bulk)

	r.readline()

	return v, nil
}
