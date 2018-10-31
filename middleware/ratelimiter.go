package middleware

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// RateKeyFunc genarate key from gin.Context, the key is an unique identification of the access you want to restrict, etc. remote IP, path, methods, custom headers and basic auth usernames
type RateKeyFunc func(ctx *gin.Context) (string, error)

// RateFormattedFunc genarate limiter rateFormatted from gin.Context
type RateFormattedFunc func(ctx *gin.Context) (string, error)

type RateLimiterMiddleware struct {
	intervalDefault  time.Duration
	capacityDefault  int64
	rateKeyCen       RateKeyFunc
	rateFormattedGen RateFormattedFunc
	limiters         map[string]*ratelimit.Bucket
}

// get Bucket from gin.Context, if the key is existed then return corresponding Bucket else create a new key and corresponding Bucket and return the Bucket
// Note. if rateFormattedGen return empty string, use default interval and capacity
func (r *RateLimiterMiddleware) get(ctx *gin.Context) (*ratelimit.Bucket, error) {
	key, err := r.rateKeyCen(ctx)

	if err != nil {
		return nil, err
	}

	if limiter, existed := r.limiters[key]; existed {
		return limiter, nil
	}

	if r.rateFormattedGen == nil {
		r.rateFormattedGen = func(ctx *gin.Context) (string, error) {
			return "", nil
		}
	}

	rateFormatted, err := r.rateFormattedGen(ctx)

	if err != nil {
		return nil, err
	}

	var limiter *ratelimit.Bucket
	if rateFormatted == "" {
		limiter = ratelimit.NewBucketWithQuantum(r.intervalDefault, r.capacityDefault, r.capacityDefault)
	} else {
		interval, capacity, err := ParseFormatted(rateFormatted)
		if err != nil {
			return nil, err
		}
		limiter = ratelimit.NewBucketWithQuantum(interval, capacity, capacity)
	}
	r.limiters[key] = limiter
	return limiter, nil
}

func (r *RateLimiterMiddleware) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limiter, err := r.get(ctx)
		if err != nil || limiter.TakeAvailable(1) == 0 {
			if err == nil {
				err = errors.New("Too many requests")
			}
			ctx.AbortWithError(429, err)
		} else {
			ctx.Writer.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limiter.Available()))
			ctx.Writer.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.Capacity()))
			ctx.Next()
		}
	}
}

func NewRateLimiter(interval time.Duration, capacity int64, rateKeyCen RateKeyFunc, rateFormattedGen RateFormattedFunc) *RateLimiterMiddleware {
	limiters := make(map[string]*ratelimit.Bucket)
	return &RateLimiterMiddleware{
		interval,
		capacity,
		rateKeyCen,
		rateFormattedGen,
		limiters,
	}
}

// You can also use the simplified format "<limit>-<period>"", with the given
// periods:
//
// * "S": second
// * "M": minute
// * "H": hour
//
// Examples:
//
// * 5 reqs/second: "5-S"
// * 10 reqs/minute: "10-M"
// * 1000 reqs/hour: "1000-H"
//
func ParseFormatted(rateFormatted string) (time.Duration, int64, error) {
	values := strings.Split(rateFormatted, "-")
	if len(values) != 2 {
		return time.Second * 0, 0, fmt.Errorf("incorrect format '%s'", rateFormatted)

	}

	periods := map[string]time.Duration{
		"S": time.Second, // Second
		"M": time.Minute, // Minute
		"H": time.Hour,   // Hour
	}

	capacityStr, periodStr := values[0], strings.ToUpper(values[1])

	interval, ok := periods[periodStr]
	if !ok {
		return time.Second * 0, 0, fmt.Errorf("incorrect period '%s'", periodStr)

	}

	capacity, err := strconv.ParseInt(capacityStr, 10, 64)
	if err != nil {
		return time.Second * 0, 0, fmt.Errorf("incorrect limit '%s'", capacityStr)
	}
	return interval, capacity, nil
}

func NewRateLimiterFromFormatted(rateFormatted string, rateKeyCen RateKeyFunc, rateFormattedGen RateFormattedFunc) (*RateLimiterMiddleware, error) {
	rate := &RateLimiterMiddleware{}

	interval, capacity, err := ParseFormatted(rateFormatted)
	if err != nil {
		return rate, err
	}

	limiters := make(map[string]*ratelimit.Bucket)
	rate = &RateLimiterMiddleware{
		interval,
		capacity,
		rateKeyCen,
		rateFormattedGen,
		limiters,
	}

	return rate, nil
}

func RateKeyCenByUser(ctx *gin.Context) (string, error) {
	claims := jwt.ExtractClaims(ctx)
	username := claims["Username"].(string)
	return username, nil
}

func RateFormattedGenByUser(ctx *gin.Context) (string, error) {
	claims := jwt.ExtractClaims(ctx)
	rateFormatted := claims["RateFormatted"].(string)
	return rateFormatted, nil
}

var RateLimiterMiddlewareByUser, _ = NewRateLimiterFromFormatted("1000-M", RateKeyCenByUser, RateFormattedGenByUser)
