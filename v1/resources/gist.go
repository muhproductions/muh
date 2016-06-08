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
	"github.com/muhproductions/muh/v1/models"
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

// Get - gist by id
func (g GistResource) Get(c *gin.Context) {
	gist := models.Gist{
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

// CreateSnippets - Create or add new Snippets
func (g GistResource) CreateSnippets(c *gin.Context) {
	gist := models.Gist{}
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
		gist.AddSnippets(snippets, c.Param("userid"))
		c.JSON(201, gin.H{
			"gist": map[string]string{
				"uuid": gist.UUID,
			},
		})
	} else {
		c.AbortWithStatus(400)
	}
}
