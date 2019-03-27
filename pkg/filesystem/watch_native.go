package filesystem

import (
	"time"
)

const (
	watchNativeEventsBufferSize = 25

	watchNativeCoalescingWindow = 10 * time.Millisecond
)
