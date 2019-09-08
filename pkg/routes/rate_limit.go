// Copyright 2019 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package routes

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/red"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"github.com/labstack/echo/v4"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"github.com/ulule/limiter/v3/drivers/store/redis"
	"net/http"
	"strconv"
	"time"
)

// RateLimit is the rate limit middleware
func RateLimit(rateLimiter *limiter.Limiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			var rateLimitKey string
			switch config.RateLimitKind.GetString() {
			case "ip":
				rateLimitKey = c.RealIP()
			case "user":
				auth, err := apiv1.GetAuthFromClaims(c)
				if err != nil {
					log.Errorf("Error getting auth from jwt claims: %v", err)
				}
				rateLimitKey = "user_" + strconv.FormatInt(auth.GetID(), 10)
			default:
				log.Errorf("Unknown rate limit kind configured: %s", config.RateLimitKind.GetString())
			}
			limiterCtx, err := rateLimiter.Get(c.Request().Context(), rateLimitKey)
			if err != nil {
				log.Errorf("IPRateLimit - rateLimiter.Get - err: %v, %s on %s", err, rateLimitKey, c.Request().URL)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": err,
				})
			}

			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
			h.Set("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

			if limiterCtx.Reached {
				log.Infof("Too Many Requests from %s on %s", rateLimitKey, c.Request().URL)
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"message": "Too Many Requests on " + c.Request().URL.String(),
				})
			}

			// log.Printf("%s request continue", c.RealIP())
			return next(c)
		}
	}
}

func setupRateLimit(a *echo.Group) {
	if config.RateLimitEnabled.GetBool() {
		rate := limiter.Rate{
			Period: config.RateLimitPeriod.GetDuration() * time.Second,
			Limit:  config.RateLimitLimit.GetInt64(),
		}
		var store limiter.Store
		var err error
		switch config.RateLimitStore.GetString() {
		case "memory":
			store = memory.NewStore()
		case "redis":
			if !config.RedisEnabled.GetBool() {
				log.Fatal("Redis is configured for rate limiting, but not enabled!")
			}
			store, err = redis.NewStore(red.GetRedis())
			if err != nil {
				log.Fatalf("Error while creating rate limit redis store: %s", err)
			}
		default:
			log.Fatalf("Unknown Rate limit store \"%s\"", config.RateLimitStore.GetString())
		}
		rateLimiter := limiter.New(store, rate)
		log.Debugf("Rate limit configured with %s and %v requests per %v", config.RateLimitStore.GetString(), rate.Limit, rate.Period)
		a.Use(RateLimit(rateLimiter))
	}
}
