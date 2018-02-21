package internal

import (
	"time"
)

type Reporter interface {
	Report(time.Time, string)
	Polling()
}
