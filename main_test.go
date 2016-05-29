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
  "github.com/timmyArch/muh-api/v1"
  "encoding/json"
)

func conf(t *testing.T) *gofight.RequestConfig {
  err := v1.RedisClient().FlushDb().Err()
  assert.Equal(t, nil, err, "Flushing Redis failed")
  gin.SetMode(gin.TestMode)
  return gofight.New().SetDebug(true)
}

func Test404(t *testing.T) {
  conf(t).GET("/").
    Run(GetEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
      assert.Equal(t, 404, r.Code, "ResponseCode should be 404")
    })
}

func TestGistCreateReturns201(t *testing.T) {
  conf(t).POST("/v1/gists").
    SetFORM(gofight.H{
      "snippet[0]lang": "ruby",
      "snippet[0]paste": "some ruby code",
    }).
    Run(GetEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
      assert.Equal(t, 201, r.Code, "ResponseCode should be 201")
    })
}

func TestGistCreateAndFindGist(t *testing.T) {
  conf := conf(t)
  var firstcall map[string]interface{}
  var secondcall map[string]interface{}

  conf.POST("/v1/gists").
    SetFORM(gofight.H{
      "snippet[0]lang": "ruby",
      "snippet[0]paste": "some ruby code",
      "snippet[1]lang": "go",
      "snippet[1]paste": "some go code",
    }).
    Run(GetEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
      assert.Equal(t, 201, r.Code, "ResponseCode should be 201")
      jsonerror := json.Unmarshal(r.Body.Bytes(), &firstcall)
      assert.Equal(t, nil, jsonerror, "ReponseBody could be parsed.")
    })
    gist := firstcall["gist"].(map[string]interface{})
    assert.NotEmpty(t, gist["uuid"], "Gist id should not be empty")

  conf.GET("/v1/gists/"+gist["uuid"].(string)).
    Run(GetEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
      assert.Equal(t, 200, r.Code, "ResponseCode should be 200")
      jsonerror := json.Unmarshal(r.Body.Bytes(), &secondcall)
      assert.Equal(t, nil, jsonerror, "ReponseBody could be parsed.")
    })

    newgist := secondcall["gist"].(map[string]interface{})
    snippets := secondcall["snippets"].(map[string]interface{})
    assert.Equal(t, newgist["uuid"], gist["uuid"], "Returned uuid is same as fetched.")
    assert.Equal(t, len(snippets), 2, "2 snippets returned")
    for _, snip := range snippets {
      parsedsnippet :=  snip.(map[string]interface{})
      if parsedsnippet["lang"].(string) == "ruby" {
        assert.Contains(t, parsedsnippet["paste"], "ruby")
      } else if parsedsnippet["lang"].(string) == "go" {
        assert.Contains(t, parsedsnippet["paste"], "go")
      } else {
        assert.FailNow(t, "No snippet matched")
      }
    }
}
