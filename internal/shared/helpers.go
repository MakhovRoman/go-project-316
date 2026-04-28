package shared

import (
	"context"
	"time"
)

// BaseDelay — задержка по умолчанию между повторными запросами,
// если в CrawlParams.Delay не задано иное.
const BaseDelay = 100 * time.Millisecond

// SleepContext ждёт указанную задержку или прерывается при отмене ctx,
// возвращая в этом случае ctx.Err().
func SleepContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RetryDelay делает паузу между повторными попытками запроса. Перед первой
// попыткой (attempt == 0) не ждёт. Использует params.Delay, либо BaseDelay по умолчанию.
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
