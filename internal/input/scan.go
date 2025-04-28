package input

import (
	"bufio"
	"bytes"
	"io"
)

func ScanLines(r io.Reader, callback func([]byte) bool) error {
	var (
		err      error
		line     []byte
		isPrefix bool
	)

	reader := bufio.NewReader(r)
	buf := new(bytes.Buffer)

	for {
		line, isPrefix, err = reader.ReadLine()
		if err != nil {
			break
		}

		buf.Write(line)
		if !isPrefix {
			if !callback(buf.Bytes()) {
				return nil
			}
			buf.Reset()
		}
	}

	if err != io.EOF && err != io.ErrClosedPipe {
		return err
	}

	return nil
}
