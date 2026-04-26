package shared

import (
	"context"
	"time"
)

const BaseDelay = 100 * time.Millisecond

func SleepContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func RetryDelay(params CrawlParams, attempt int) error {
	if attempt > 0 {
		delay := params.Delay
		if delay == 0 {
			delay = BaseDelay
		}

		if err := SleepContext(params.CTX, delay); err != nil {
			return err
		}
	}

	return nil
}
