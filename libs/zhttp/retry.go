package zhttp

import (
	"context"
	"math/rand"
	"net/http"
	"time"
)

var (
	MaxRetry  = 3
	RetryUnit = 100 * time.Millisecond
	RetryCap  = time.Second
	Random    = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
)

const (
	MaxJitter = 1.0
	NoJitter  = 0.0
)

type RetryOption struct {
	random    *rand.Rand
	maxRetry  int
	retryUnit time.Duration
	retryCap  time.Duration
	jitter    float64
}

type Retry func(option *RetryOption)

func defaultRetryOption() *RetryOption {
	opt := &RetryOption{
		random:    Random,
		maxRetry:  MaxRetry,
		retryCap:  RetryCap,
		retryUnit: RetryUnit,
		jitter:    NoJitter,
	}
	return opt
}

func WithRandom(random *rand.Rand) Retry {
	return func(option *RetryOption) {
		option.random = random
	}
}

func WithMaxRetry(maxRetry int) Retry {
	return func(option *RetryOption) {
		option.maxRetry = maxRetry
	}
}

func WithRetryCap(retryCap time.Duration) Retry {
	return func(option *RetryOption) {
		option.retryCap = retryCap
	}
}

func WithRetryUnit(retryUnit time.Duration) Retry {
	return func(option *RetryOption) {
		option.retryUnit = retryUnit
	}
}

func WithJitter(jitter float64) Retry {
	return func(option *RetryOption) {
		option.jitter = jitter
	}
}

func NewRetryTimer(ctx context.Context, opts ...Retry) <-chan int {
	retry := defaultRetryOption()
	for _, option := range opts {
		option(retry)
	}
	attemptCh := make(chan int)
	exponentialBackoffWait := func(attempt int) time.Duration {
		if retry.jitter < NoJitter {
			retry.jitter = NoJitter
		}
		if retry.jitter > MaxJitter {
			retry.jitter = MaxJitter
		}
		sleep := retry.retryUnit * time.Duration(1<<uint(attempt))
		if sleep > retry.retryCap {
			sleep = retry.retryCap
		}
		if retry.jitter != NoJitter {
			sleep -= time.Duration(retry.random.Float64() * float64(sleep) * retry.jitter)
		}
		return sleep
	}

	go func() {
		defer close(attemptCh)
		for i := 0; i < retry.maxRetry; i++ {
			select {
			case attemptCh <- i + 1:
			case <-ctx.Done():
				return
			}
			select {
			case <-time.After(exponentialBackoffWait(i)):
			case <-ctx.Done():
				return
			}
		}
	}()
	return attemptCh
}

var RetryableHTTPStatusCodes = map[int]struct{}{
	http.StatusInternalServerError:        {},
	http.StatusBadGateway:                 {},
	http.StatusServiceUnavailable:         {},
	http.StatusGatewayTimeout:             {},
	http.StatusTooManyRequests:            {},
	http.StatusUnavailableForLegalReasons: {},
}

func IsHTTPStatusRetryable(httpStatusCode int) (ok bool) {
	_, ok = RetryableHTTPStatusCodes[httpStatusCode]
	return ok
}

var SleepHTTPStatusCodes = map[int]struct{}{
	http.StatusTooManyRequests:            {},
	http.StatusUnavailableForLegalReasons: {},
}

func IsHTTPStatusSleep(httpStatusCode int) (ok bool) {
	_, ok = SleepHTTPStatusCodes[httpStatusCode]
	return ok
}

var SuccessStatusCodes = map[int]struct{}{
	http.StatusOK:             {},
	http.StatusNoContent:      {},
	http.StatusPartialContent: {},
}

func IsHTTPSuccessStatus(httpStatusCode int) (ok bool) {
	_, ok = SuccessStatusCodes[httpStatusCode]
	return
}
