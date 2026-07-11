package middleware

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	RateLimitLimitHeader     = "X-RateLimit-Limit"
	RateLimitRemainingHeader = "X-RateLimit-Remaining"
	RateLimitResetHeader     = "X-RateLimit-Reset"
)

var (
	ErrRateLimitMaximumInvalid = errors.New(
		"rate limit maximum must be greater than zero",
	)
	ErrRateLimitWindowInvalid = errors.New(
		"rate limit window must be greater than zero",
	)
)

type RateLimiterConfig struct {
	MaxRequests int
	Window      time.Duration

	KeyGenerator func(
		c *fiber.Ctx,
	) string

	Next func(
		c *fiber.Ctx,
	) bool

	LimitReached fiber.Handler

	Now func() time.Time
}

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mutex sync.Mutex

	maxRequests int
	window      time.Duration

	keyGenerator func(
		c *fiber.Ctx,
	) string

	next func(
		c *fiber.Ctx,
	) bool

	limitReached fiber.Handler
	now          func() time.Time

	entries     map[string]rateLimitEntry
	lastCleanup time.Time
}

func NewRateLimiter(
	config RateLimiterConfig,
) (fiber.Handler, error) {
	if config.MaxRequests <= 0 {
		return nil, ErrRateLimitMaximumInvalid
	}

	if config.Window <= 0 {
		return nil, ErrRateLimitWindowInvalid
	}

	keyGenerator := config.KeyGenerator
	if keyGenerator == nil {
		keyGenerator = func(
			c *fiber.Ctx,
		) string {
			return c.IP()
		}
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	limiter := &rateLimiter{
		maxRequests:  config.MaxRequests,
		window:       config.Window,
		keyGenerator: keyGenerator,
		next:         config.Next,
		limitReached: config.LimitReached,
		now:          now,
		entries: make(
			map[string]rateLimitEntry,
		),
	}

	return limiter.handle, nil
}

func (
	limiter *rateLimiter,
) handle(
	c *fiber.Ctx,
) error {
	if limiter.next != nil &&
		limiter.next(
			c,
		) {
		return c.Next()
	}

	key := strings.TrimSpace(
		limiter.keyGenerator(
			c,
		),
	)
	if key == "" {
		key = "unknown"
	}

	now := limiter.now().UTC()

	entry, remaining, limited := limiter.acquire(
		key,
		now,
	)

	c.Set(
		RateLimitLimitHeader,
		strconv.Itoa(
			limiter.maxRequests,
		),
	)
	c.Set(
		RateLimitRemainingHeader,
		strconv.Itoa(
			remaining,
		),
	)
	c.Set(
		RateLimitResetHeader,
		strconv.FormatInt(
			entry.resetAt.Unix(),
			10,
		),
	)

	if !limited {
		return c.Next()
	}

	retryAfterSeconds := int64(
		math.Ceil(
			entry.resetAt.Sub(
				now,
			).Seconds(),
		),
	)
	if retryAfterSeconds < 1 {
		retryAfterSeconds = 1
	}

	c.Set(
		fiber.HeaderRetryAfter,
		strconv.FormatInt(
			retryAfterSeconds,
			10,
		),
	)

	if limiter.limitReached != nil {
		return limiter.limitReached(
			c,
		)
	}

	return fiber.ErrTooManyRequests
}

func (
	limiter *rateLimiter,
) acquire(
	key string,
	now time.Time,
) (
	rateLimitEntry,
	int,
	bool,
) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	limiter.cleanupExpiredEntries(
		now,
	)

	entry := limiter.entries[key]
	if entry.resetAt.IsZero() ||
		!now.Before(
			entry.resetAt,
		) {
		entry = rateLimitEntry{
			resetAt: now.Add(
				limiter.window,
			),
		}
	}

	entry.count++
	limiter.entries[key] = entry

	remaining := limiter.maxRequests - entry.count
	if remaining < 0 {
		remaining = 0
	}

	return entry,
		remaining,
		entry.count > limiter.maxRequests
}

func (
	limiter *rateLimiter,
) cleanupExpiredEntries(
	now time.Time,
) {
	if !limiter.lastCleanup.IsZero() &&
		now.Sub(
			limiter.lastCleanup,
		) < limiter.window {
		return
	}

	for key, entry := range limiter.entries {
		if !now.Before(
			entry.resetAt,
		) {
			delete(
				limiter.entries,
				key,
			)
		}
	}

	limiter.lastCleanup = now
}
