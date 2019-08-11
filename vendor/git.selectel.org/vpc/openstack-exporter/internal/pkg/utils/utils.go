package utils

import (
	"context"
	"fmt"
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"github.com/gophercloud/gophercloud/pagination"
)

// Retry implements a simple retry wrapper over functions.
// It tries to call a function several times with a provided interval.
func Retry(attempts int, interval time.Duration, f func() error) error {
	var err error

	for i := 0; i < attempts; i++ {
		if err = f(); err == nil {
			return nil
		}
		log.Debugf("retrying after error: %s", err)
		time.Sleep(interval)
	}

	return fmt.Errorf("after %d attempts, last error was: %s", attempts, err)
}

// RetryForPager implements a simple retry wrapper over functions that returns
// Gophercloud's pagination page.
// It tries to call a function several times with a provided interval.
func RetryForPager(attempts int, interval time.Duration, f func() (pagination.Page, error)) (pagination.Page, error) {
	var (
		err  error
		page pagination.Page
	)

	// extend last error with context about attempts
	err = Retry(attempts, interval, func() error {
		page, err = f()
		// return error in base Retry func
		// to try next attempt.
		return err
	})

	return page, err
}

// DoEvery runs provided function every specified interval and can be gracefully
// shut down via provided context.
func DoEvery(ctx context.Context, interval time.Duration, f func() error) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if err := f(); err != nil {
				log.Error(err)
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}
