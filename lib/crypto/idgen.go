package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"io"
	"os"
	"sync"
	"time"
)

var (
	counterMap map[string]uint16
	m          sync.Mutex
)

func init() {
	m = sync.Mutex{}
	counterMap = make(map[string]uint16)
}

func RefreshCounterMap() {
	for {
		m.Lock()
		counterMap = make(map[string]uint16)
		m.Unlock()
		time.Sleep(time.Second)
	}
}

func NewID(id string) (string, error) {
	d := []byte(id)
	dLen := len(d)

	bufSize := dLen
	bufSize += 4 // time stamp utc is uint32
	bufSize += 2 // pid is uint16
	bufSize += 2 // counter variable is uint16
	bufSize += 2 // random bytes array

	buf := make([]byte, bufSize)
	copy(buf, d)

	t := time.Now().UTC().Unix()
	binary.BigEndian.PutUint32(buf[dLen:dLen+4], uint32(t))

	pid := os.Getpid()
	binary.BigEndian.PutUint16(buf[dLen+4:dLen+6], uint16(pid))

	counter := counterMap[id]
	m.Lock()
	counterMap[id] = counter + 1
	m.Unlock()
	binary.BigEndian.PutUint16(buf[dLen+6:dLen+8], uint16(counter))

	_, err := io.ReadFull(rand.Reader, buf[dLen+8:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}
