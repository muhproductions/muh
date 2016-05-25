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

package main

import (
  "testing"
  "github.com/stretchr/testify/assert"
  "github.com/appleboy/gofight"
  "github.com/gin-gonic/gin"
)

func Test404(t *testing.T) {
  gin.SetMode(gin.TestMode)
  r := gofight.New()
  r.GET("/").
  SetDebug(true).
  Run(GetEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
    assert.Equal(t, 404, r.Code, "ResponseCode should be 404")
  })
}
