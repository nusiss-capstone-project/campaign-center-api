package service

import (
	"fmt"
	"time"
)

func newTransactionNo() string {
	return fmt.Sprintf("TXN%d", time.Now().UnixNano())
}
