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
	"github.com/muhproductions/muh/helper"
	"github.com/satori/go.uuid"
	"gopkg.in/redis.v3"

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

type rawGist struct {
	Snippets []rawSnippet `json:"snippets"`
}

type rawSnippet struct {
	Paste string `json:"paste"`
	Lang  string `json:"lang"`
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
	var rawgist rawGist
	snippets := []map[string]string{}
	if c.BindJSON(&rawgist) == nil {
		for _, snip := range rawgist.Snippets {
			snippets = append(snippets, map[string]string{
				"paste": snip.Paste,
				"lang":  snip.Lang,
			})
		}
	} else {
		return
	}
	if len(snippets) > 0 {
		gist.AddSnippets(snippets)
		c.JSON(201, gin.H{
			"gist": map[string]string{
				"uuid": gist.UUID,
			},
		})
	} else {
		c.AbortWithStatus(400)
	}
}

//Gist model
type Gist struct {
	UUID string
}

//Exists verifies the persistence level.
func (g *Gist) Exists() bool {
	val, err := helper.RedisClient().Exists("gists::" + g.UUID).Result()
	if err != nil {
		return false
	}
	return val
}

//AddSnippets appends new compressed snippets.
func (g *Gist) AddSnippets(snippets []map[string]string) bool {
	pipe := helper.RedisClient().Pipeline()
	defer pipe.Close()
	if g.UUID == "" {
		g.UUID = uuid.NewV4().String()
	}
	for _, v := range snippets {
		tempuuid := uuid.NewV4().String()
		pipe.SAdd("gists::"+g.UUID, tempuuid)
		json, _ := json.Marshal(v)
		pipe.Set("snippets::"+tempuuid, helper.Zip(string(json)), 0)
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
	snippets, err := helper.RedisClient().SMembers("gists::" + g.UUID).Result()
	snippetsprecollection := map[string]*redis.StringCmd{}
	snippetscollection := map[string]map[string]string{}
	if err != nil {
		log.Error(err, "Gist not found")
	} else {
		pipe := helper.RedisClient().Pipeline()
		defer pipe.Close()
		for _, snipp := range snippets {
			snippetsprecollection[snipp] = pipe.Get("snippets::" + snipp)
		}
		pipe.Exec()
		for k, v := range snippetsprecollection {
			var dat map[string]string
			if err := json.Unmarshal([]byte(helper.Unzip(v.Val())), &dat); err != nil {
				log.Error(err, "Snippet loading failed - "+k)
			} else {
				snippetscollection[k] = dat
			}
		}
	}
	return snippetscollection
}
