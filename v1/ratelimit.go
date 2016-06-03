// Copyright 2016 Tim Foerster <github@mailserver.1n3t.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/muhproductions/muh-api/helper"
	"os"
	"strconv"
	"strings"
	"time"
)

// Ratelimit - Middleware to handle ratelimiting.
// - Rejecting reqests...
// - Bumping request/ip count ...
func Ratelimit() gin.HandlerFunc {
	return func(c *gin.Context) {

		t := time.Now()

		pipe := helper.RedisClient().Pipeline()
		hits := pipe.Incr("ratelimit::hits::" + c.ClientIP())
		bytes := pipe.IncrBy("ratelimit::bytes::"+c.ClientIP(), c.Request.ContentLength)
		defer pipe.Close()
		pipe.Exec()

		ratelimitcheck("Hits", hits.Val(), c)
		ratelimitcheck("Bytes", bytes.Val(), c)

		c.Header("X-Ratelimit-Latency", time.Since(t).String())

		c.Next()
	}
}

func ratelimitcheck(name string, value int64, c *gin.Context) {
	converted := strconv.Itoa(int(value))
	c.Header("X-Ratelimit-"+name, converted)
	env := os.Getenv("LIMIT_" + strings.ToUpper(name))
	if env != "" {
		c.Header("X-Ratelimit-"+name+"-Limit", env)
		limit, _ := strconv.Atoi(env)
		if value > int64(limit) {
			c.Header("X-Ratelimit-"+name+"-State", "BLOCKED")
			c.AbortWithStatus(429)
		} else {
			c.Header("X-Ratelimit-"+name+"-State", "OK")
		}
	}

}
