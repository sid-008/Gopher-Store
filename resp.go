package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
	str string
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
		log.Println("Step Read()")
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// example: we're parsing '*2\r\nhello\r\nworld\r\n'
// First the Read() method parses *, this tells us it is an array
// now control goes to readArray() method. Here we first parse the 2nd byte
// the 2nd byte tells us exactly how many elements to parse. We encounter a CRLF now.
// now we iterate over the buffer. Since this is an array we call Read per element
// recall; read will read the value using the readBulk() function now.
//

func (r *Resp) readArray() (Value, error) {
	// log.Println("Step readArray")
	v := Value{} // spawn ser/deser struct
	v.typ = "array"

	// keep in mind the first byte has been parsed already, that byte told us what type was sent
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
	// log.Println(v.array)
	return v, nil
}

// $5\r\nHello\r\n

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"

	len, _, err := r.readInteger() // read 5 ($ has alr been read)
	if err != nil {
		return v, err
	}

	bulk := make([]byte, len)

	_, err = r.reader.Read(bulk) // Note that this bufio.Reader.Read(), stores the read bytes into variable 'bulk'
	// read Hello into slice bulk

	if err != nil {
		return Value{}, err
	}

	// log.Println("Contents of bulk: ", bulk)
	v.bulk = string(bulk)
	// log.Println(v.bulk)

	r.readline()
	// log.Println(v.array)
	// log.Println("Step ReadBulk")

	return v, nil
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	len := len(v.array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len; i++ {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}
func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}

func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshallNull()
	case "error":
		return v.marshallError()
	default:
		return []byte{}
	}
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
