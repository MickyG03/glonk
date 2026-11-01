package engine

import (
	"bytes"
	"strings"
	"testing"
	"testing/iotest"
)

func TestReadInput(t *testing.T) {
	r := iotest.DataErrReader(iotest.OneByteReader(strings.NewReader("ab\x03")))
	var chunks [][]byte
	for chunk := range readInput(r) {
		chunks = append(chunks, chunk)
	}
	got := bytes.Join(chunks, nil)

	if !bytes.Equal([]byte("ab\x03"), got) {
		t.Errorf("wanted %q, got %q", "ab\x03", got)
	}
}
