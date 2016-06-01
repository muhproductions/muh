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

package helper

import (
	"bytes"
	"compress/gzip"
	"github.com/golang/snappy"
	"gopkg.in/redis.v3"
	"io/ioutil"
	"os"
)

var redisconn *redis.Client

// RedisClient - Get new redis connection.
func RedisClient() *redis.Client {
	if redisconn == nil {
		redisconn = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: "",
			DB:       0,
			PoolSize: 100,
			Network:  os.Getenv("REDIS_NETWORK"),
		})
	}
	return redisconn
}

/*
Zip - Generic compression layer.
It provides supports multiple compression layers
 * gzip
 * snappy
 * uncompress (empty - default)
This option should be set by system environment var "COMPRESSION"
*/
func Zip(str string) string {
	if os.Getenv("COMPRESSION") == "snappy" {
		encoded := snappy.Encode(nil, []byte(str))
		return string(encoded)
	} else if os.Getenv("COMPRESSION") == "gzip" {
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		w.Write([]byte(str))
		w.Close()
		return b.String()
	}
	return str
}

// Unzip - reverse method to Zip()
func Unzip(str string) string {
	if os.Getenv("COMPRESSION") == "snappy" {
		decoded, _ := snappy.Decode(nil, []byte(str))
		return string(decoded)
	} else if os.Getenv("COMPRESSION") == "gzip" {
		readbuf := new(bytes.Buffer)
		readbuf.WriteString(str)

		r, _ := gzip.NewReader(readbuf)
		defer r.Close()
		unzip, _ := ioutil.ReadAll(r)

		return string(unzip)
	}
	return str
}
