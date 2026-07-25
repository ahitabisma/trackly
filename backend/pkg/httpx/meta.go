package httpx

import (
	"time"

	"github.com/google/uuid"
)

func newMeta() *Meta {
	return &Meta{
		RequestID: uuid.NewString(),
		Timestamp: time.Now(),
	}
}
