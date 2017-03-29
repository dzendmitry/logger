package logger

import (
	"io"
	"sync"
	"fmt"
)

const (
	RINGBUF_SIZE = 1 * 1024 * 1024
	CHUNK_SIZE = 512
)

type RingBufWriter struct {
	writer io.Writer
	head, tail int
	data []byte
	cv *sync.Cond
	closeC chan bool
	droppedLines int
	empty bool
	close bool
}

func NewRingBufWriter(w io.Writer) *RingBufWriter {
	rb := &RingBufWriter{
		writer: w,
		data: make([]byte, RINGBUF_SIZE),
		cv: sync.NewCond(&sync.Mutex{}),
		closeC: make(chan bool),
		empty: true,
	}
	go rb.write()
	return rb
}

func (rbw *RingBufWriter) WriteLine(data ...[]byte) int {
	var totalLen int
	for i := range data {
		totalLen += len(data[i])
	}
	rbw.cv.L.Lock()
	head := rbw.head
	tail := rbw.tail
	var freeSpace int
	if rbw.empty {
		freeSpace = len(rbw.data)
	} else {
		freeSpace = tail - head
		if freeSpace < 0 {
			freeSpace += len(rbw.data)
		}
	}
	var droppedMarker []byte
	if rbw.droppedLines > 0 {
		droppedMarker = []byte(fmt.Sprintf("... %d lines dropped\n", rbw.droppedLines))
		totalLen += len(droppedMarker)
	}
	if freeSpace < totalLen {
		rbw.droppedLines += 1
		rbw.cv.L.Unlock()
		return 0
	}
	newHead := head + len(droppedMarker) + totalLen
	if newHead > len(rbw.data) {
		newHead -= len(rbw.data)
	}
	n := copy(rbw.data[head:], droppedMarker)
	if n < len(droppedMarker) {
		head = copy(rbw.data, droppedMarker[n:])
	} else {
		head += n
	}
	for i := range data {
		n := copy(rbw.data[head:], data[i])
		if n < len(data[i]) {
			head = copy(rbw.data, data[i])
		} else {
			head += n
		}
	}
	rbw.droppedLines = 0
	rbw.head = newHead
	rbw.empty = false
	rbw.cv.L.Unlock()
	rbw.cv.Signal()
	return totalLen
}

func (rbw *RingBufWriter) write() {
	for {
		rbw.cv.L.Lock()
		if rbw.empty && rbw.close {
			rbw.closeC <- true
			break
		}
		for rbw.head == rbw.tail {
			rbw.cv.Wait()
		}
		head := rbw.head
		tail := rbw.tail
		rbw.cv.L.Unlock()
		newTail := tail + CHUNK_SIZE
		if head < tail {
			if newTail > len(rbw.data) {
				n, err := rbw.writer.Write(rbw.data[tail:])
				if err != nil {
					rbw.cv.L.Unlock()
					continue
				}
				tail = 0
				newTail = CHUNK_SIZE - n
			}
			if newTail > head {
				newTail = head
			}
			_, err := rbw.writer.Write(rbw.data[tail:newTail])
			if err != nil {
				continue
			}
		} else {
			if newTail > head {
				newTail = head
			}
			_, err := rbw.writer.Write(rbw.data[tail:newTail])
			if err != nil {
				continue
			}
		}
		rbw.cv.L.Lock()
		rbw.tail = newTail
		if rbw.tail == rbw.head {
			rbw.head = 0
			rbw.tail = 0
			rbw.empty = true
		}
		rbw.cv.L.Unlock()
	}
}

func (rbw *RingBufWriter) Close() error {
	rbw.cv.L.Lock()
	rbw.close = true
	rbw.cv.L.Unlock()
	rbw.cv.Signal()
	<-rbw.closeC
	return nil
}