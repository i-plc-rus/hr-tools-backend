package lock

import (
	"context"
	"sync"
	"time"
)

var (
	lockMap sync.Map
)

func WithDelay(ctx context.Context, key string, wait time.Duration, safeCode func() error) (success bool, err error) {
	isLocked := false
	isTimeout := time.After(wait)
	for {
		if _, loaded := lockMap.LoadOrStore(key, true); !loaded {
			isLocked = true
			break
		}
		select {
		case <-isTimeout:
			return false, nil
		case <-ctx.Done():
			return false, nil
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
	if isLocked {
		defer lockMap.Delete(key)
		return true, safeCode()
	}
	return false, nil
}
