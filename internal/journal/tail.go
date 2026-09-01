package journal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"
)

const maxTailRead = 1024 * 1024

type Tail struct {
	path    string
	offset  int64
	partial []byte
}

func NewTail(path string) *Tail { return &Tail{path: path} }

// Poll reads at most one bounded chunk and never holds the journal descriptor
// between calls, so appenders are not blocked by a long-lived reader.
func (t *Tail) Poll() ([]string, error) {
	fd, err := syscall.Open(t.path, syscall.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), t.path)
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < t.offset {
		t.offset = 0
		t.partial = t.partial[:0]
	}
	if _, err := file.Seek(t.offset, io.SeekStart); err != nil {
		return nil, err
	}
	buffer := make([]byte, maxTailRead)
	read, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if read == 0 {
		return []string{}, nil
	}
	t.offset += int64(read)
	t.partial = append(t.partial, buffer[:read]...)
	if len(t.partial) > maxRecordSize+maxTailRead {
		return nil, fmt.Errorf("journal tail contains an oversized partial record")
	}
	lines := []string{}
	for {
		index := bytes.IndexByte(t.partial, '\n')
		if index < 0 {
			break
		}
		if index > maxRecordSize {
			return nil, fmt.Errorf("journal record exceeds size limit")
		}
		lines = append(lines, string(t.partial[:index]))
		t.partial = t.partial[index+1:]
	}
	return lines, nil
}
