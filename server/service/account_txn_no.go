package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func newTransactionNo() string {
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return fmt.Sprintf("TXN%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("TXN%d%s", time.Now().UnixNano(), hex.EncodeToString(suffix[:]))
}
