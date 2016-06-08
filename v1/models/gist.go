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

package models

import (
	"encoding/json"
	"github.com/muhproductions/muh/helper"
	"github.com/satori/go.uuid"
	"gopkg.in/redis.v3"

	log "github.com/Sirupsen/logrus"
)

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
func (g *Gist) AddSnippets(snippets []map[string]string, userid string) bool {
	pipe := helper.RedisClient().Pipeline()
	defer pipe.Close()
	if g.UUID == "" {
		g.UUID = uuid.NewV4().String()
	}
	for _, v := range snippets {
		tempuuid := uuid.NewV4().String()
		pipe.SAdd("gists::"+g.UUID, tempuuid)
		if userid != "" {
			pipe.SAdd("users::"+userid+"::gists", g.UUID)
		}
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
