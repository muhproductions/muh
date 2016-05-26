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
  var first_call map[string]interface{}
  var second_call map[string]interface{}

  conf.POST("/v1/gists").
    SetFORM(gofight.H{
      "snippet[0]lang": "ruby",
      "snippet[0]paste": "some ruby code",
      "snippet[1]lang": "go",
      "snippet[1]paste": "some go code",
    }).
    Run(GetEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
      assert.Equal(t, 201, r.Code, "ResponseCode should be 201")
      json_error := json.Unmarshal(r.Body.Bytes(), &first_call)
      assert.Equal(t, nil, json_error, "ReponseBody could be parsed.")
    })
    gist := first_call["gist"].(map[string]interface{})
    assert.NotEmpty(t, gist["uuid"], "Gist id should not be empty")

  conf.GET("/v1/gists/"+gist["uuid"].(string)).
    Run(GetEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
      assert.Equal(t, 200, r.Code, "ResponseCode should be 200")
      json_error := json.Unmarshal(r.Body.Bytes(), &second_call)
      assert.Equal(t, nil, json_error, "ReponseBody could be parsed.")
    })

    new_gist := second_call["gist"].(map[string]interface{})
    snippets := second_call["snippets"].(map[string]interface{})
    assert.Equal(t, new_gist["uuid"], gist["uuid"], "Returned uuid is same as fetched.")
    assert.Equal(t, len(snippets), 2, "2 snippets returned")
    for _, snip := range snippets {
      parsed_snippet :=  snip.(map[string]interface{})
      if parsed_snippet["lang"].(string) == "ruby" {
        assert.Contains(t, parsed_snippet["paste"], "ruby")
      } else if parsed_snippet["lang"].(string) == "go" {
        assert.Contains(t, parsed_snippet["paste"], "go")
      } else {
        assert.FailNow(t, "No snippet matched")
      }
    }
}
