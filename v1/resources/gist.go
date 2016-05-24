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

package resources

import (
  "github.com/gin-gonic/gin"
  "gopkg.in/redis.v3"
  "github.com/satori/go.uuid"
  "encoding/json"
  "strconv"

  log "github.com/Sirupsen/logrus"
)

type GistResource struct {
  Redis *redis.Client
  Engine *gin.RouterGroup
}

func (g GistResource) Routes() {
  g.Engine.GET("/gists/:uuid", g.Get)
  g.Engine.POST("/gists/:uuid", g.CreateSnippets)
  g.Engine.POST("/gists", g.CreateSnippets)
}

func (g GistResource) Get(c *gin.Context) {
  gist := Gist{
    GistResource: &g,
    Uuid: c.Param("uuid"),
  }
  if gist.Exists() == false {
    NotFound("Gist", c)
  } else {
    c.JSON(200, gin.H{
      "gist": map[string]string{
        "uuid": gist.Uuid,
      },
      "snippets": gist.GetSnippets(),
    })
  }
}

func (g GistResource) CreateSnippets(c *gin.Context) {
  gist := Gist{
    GistResource: &g,
  }
  if c.Param("uuid") != "" {
    gist.Uuid = c.Param("uuid")
  } 
  snippets := []map[string]string{}
  c.Request.ParseForm()
  for i := 0; i < len(c.Request.Form)/2; i++ {
    si := strconv.Itoa(i)
    snippets = append(snippets, map[string]string{
      "paste": string(c.Request.PostFormValue("snippet["+si+"]paste")),
      "lang": string(c.Request.PostFormValue("snippet["+si+"]lang")),
    })
  }
  gist.AddSnippets(snippets)
  c.JSON(201, gin.H{
    "gist": map[string]string{
      "uuid": gist.Uuid,
    },
  })
}

type Gist struct {
  GistResource *GistResource
  Uuid string
}

func (g *Gist) Exists() bool {
  val, err := g.GistResource.Redis.Exists("gists::"+g.Uuid).Result()
  if err != nil {
    return false
  } else {
    return val
  }
}

func (g *Gist) AddSnippets(snippets []map[string]string) bool {
  pipe := g.GistResource.Redis.Pipeline()
  defer pipe.Close()
  if g.Uuid == "" {
    g.Uuid = uuid.NewV4().String()
  }
  for _, v := range snippets {
    temp_uuid := uuid.NewV4().String()
    pipe.SAdd("gists::"+g.Uuid, temp_uuid)
    json, _ := json.Marshal(v)
    pipe.Set("snippets::"+temp_uuid, string(json), 0)
  }
  _, err := pipe.Exec()
  if err != nil {
    log.Error(err, "Error on setting snippets")
    return false
  } else {
    return true
  }
}

func (g *Gist) GetSnippets() map[string]map[string]string {
  snippets, err := g.GistResource.Redis.SMembers("gists::"+g.Uuid).Result()
  snippets_pre_collection := map[string]*redis.StringCmd{}
  snippets_collection := map[string]map[string]string{}
  if err != nil {
    log.Error(err, "Gist not found")
  } else {
    pipe := g.GistResource.Redis.Pipeline()
    defer pipe.Close()
    for _, snipp := range snippets {
      snippets_pre_collection[snipp] = pipe.Get("snippets::"+snipp)
    }
    pipe.Exec()
    for k, v := range snippets_pre_collection {
      var dat map[string]string
      if err := json.Unmarshal([]byte(v.Val()), &dat); err != nil {
        log.Error(err, "Snippet loading failed - "+k)
      } else {
        snippets_collection[k] = dat
      }
    }
  }
  return snippets_collection
}
