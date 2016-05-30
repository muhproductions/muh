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
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"gopkg.in/redis.v3"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

// GistResource - Gists API Endpoint
type GistResource struct {
	Engine *gin.RouterGroup
}

// Routes - Setup gists resource routes
func (g GistResource) Routes() {
	g.Engine.GET("/gists/:uuid", g.Get)
	g.Engine.POST("/gists/:uuid", g.CreateSnippets)
	g.Engine.POST("/gists", g.CreateSnippets)
}

/*
Get - gist by id

{
	"gist": {
		"uuid": <UUID>
	},
	"snippets": [
		{
			"paste": "moo",
			"lang": "ruby"
		}
	]
}
*/
func (g GistResource) Get(c *gin.Context) {
	gist := Gist{
		UUID: c.Param("uuid"),
	}
	if gist.Exists() == false {
		NotFound("Gist", c)
	} else {
		c.JSON(200, gin.H{
			"gist": map[string]string{
				"uuid": gist.UUID,
			},
			"snippets": gist.GetSnippets(),
		})
	}
}

/*
CreateSnippets - Create or add new Snippets

POST /gists - Create new gist with (n) new snippets.
POST /gists/<UUID> - Create or update gist with (n) new snippets.
{
	"gist": {
		"uuid": <UUID>
	}
}
*/
func (g GistResource) CreateSnippets(c *gin.Context) {
	gist := Gist{}
	if c.Param("uuid") != "" {
		gist.UUID = c.Param("uuid")
	}
	snippets := []map[string]string{}
	c.Request.ParseForm()
	for i := 0; i < len(c.Request.Form)/2; i++ {
		si := strconv.Itoa(i)
		snippets = append(snippets, map[string]string{
			"paste": string(c.Request.PostFormValue("snippet[" + si + "]paste")),
			"lang":  string(c.Request.PostFormValue("snippet[" + si + "]lang")),
		})
	}
	gist.AddSnippets(snippets)
	c.JSON(201, gin.H{
		"gist": map[string]string{
			"uuid": gist.UUID,
		},
	})
}

//Gist model
type Gist struct {
	UUID string
}

//Exists verifies the persistence level.
func (g *Gist) Exists() bool {
	val, err := RedisClient().Exists("gists::" + g.UUID).Result()
	if err != nil {
		return false
	}
	return val
}

//AddSnippets appends new compressed snippets.
func (g *Gist) AddSnippets(snippets []map[string]string) bool {
	pipe := RedisClient().Pipeline()
	defer pipe.Close()
	if g.UUID == "" {
		g.UUID = uuid.NewV4().String()
	}
	for _, v := range snippets {
		tempuuid := uuid.NewV4().String()
		pipe.SAdd("gists::"+g.UUID, tempuuid)
		json, _ := json.Marshal(v)
		pipe.Set("snippets::"+tempuuid, Zip(string(json)), 0)
	}
	_, err := pipe.Exec()
	if err != nil {
		log.Error(err, "Error on setting snippets")
		return false
	}
	return true
}

//GetSnippets returns all uncompressed snippets which are associated to self.
func (g *Gist) GetSnippets() map[string]map[string]string {
	snippets, err := RedisClient().SMembers("gists::" + g.UUID).Result()
	snippetsprecollection := map[string]*redis.StringCmd{}
	snippetscollection := map[string]map[string]string{}
	if err != nil {
		log.Error(err, "Gist not found")
	} else {
		pipe := RedisClient().Pipeline()
		defer pipe.Close()
		for _, snipp := range snippets {
			snippetsprecollection[snipp] = pipe.Get("snippets::" + snipp)
		}
		pipe.Exec()
		for k, v := range snippetsprecollection {
			var dat map[string]string
			if err := json.Unmarshal([]byte(Unzip(v.Val())), &dat); err != nil {
				log.Error(err, "Snippet loading failed - "+k)
			} else {
				snippetscollection[k] = dat
			}
		}
	}
	return snippetscollection
}
