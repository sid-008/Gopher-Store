package main

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"
)

type AOF struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

func NewAOF(path string) (*AOF, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &AOF{
		file: file,
		rd:   bufio.NewReader(file),
	}

	go func() {
		for {
			aof.mu.Lock()

			aof.file.Sync()

			aof.mu.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (aof *AOF) Close() error {
	aof.mu.Lock()

	defer aof.mu.Unlock()

	return aof.file.Close()
}

func (aof *AOF) Write(value Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (aof *AOF) Read(fn func(value Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	aof.file.Seek(0, io.SeekStart)

	reader := NewResp(aof.file)

	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		fn(value)
	}

	return nil
}
