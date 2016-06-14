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
	"os"
	"time"

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

type snippet struct {
	UUID  string
	Value map[string]string
}

func (snippet *snippet) cacheSnippet(r *redis.Pipeline) {
	json, _ := json.Marshal(snippet.Value)
	expire, _ := time.ParseDuration("1h")
	if os.Getenv("CACHING_TIME") != "" {
		t, err := time.ParseDuration(os.Getenv("CACHING_TIME"))
		if err != nil {
			expire = t
		}
	}
	r.Set("shadow::snippets::"+snippet.UUID, "", expire)
	r.Set("snippets::"+snippet.UUID, helper.Zip(string(json)), 0)
}

func getSnippet(r *redis.Pipeline, key string, value *redis.StringCmd) snippet {
	snipp := snippet{
		UUID: key,
	}
	val, err := value.Result()
	var dat map[string]string
	if err != nil {
		json.Unmarshal([]byte(helper.Unzip(helper.BoltGet("snippets::"+key))), &dat)
		snipp.Value = dat
		snipp.cacheSnippet(r)
		helper.BoltDel("snippets::" + key)
	} else {
		json.Unmarshal([]byte(helper.Unzip(val)), &dat)
	}
	snipp.Value = dat
	return snipp
}

func (g *Gist) initSnippet(r *redis.Pipeline, snippet snippet, userid string) {
	r.SAdd("gists::"+g.UUID, snippet.UUID)
	if userid != "" {
		r.SAdd("users::"+userid+"::gists", g.UUID)
	}
}

// SetupUUID defines a new UUID unless set.
func (g *Gist) SetupUUID() {
	if g.UUID == "" {
		g.UUID = uuid.NewV4().String()
	}
}

//AddSnippets appends new compressed snippets.
func (g *Gist) AddSnippets(snippets []map[string]string, userid string) bool {
	pipe := helper.RedisClient().Pipeline()
	defer pipe.Close()
	g.SetupUUID()
	for _, v := range snippets {
		s := snippet{UUID: uuid.NewV4().String(), Value: v}
		g.initSnippet(pipe, s, userid)
		s.cacheSnippet(pipe)
	}
	pipe.Exec()
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
			snippetscollection[k] = getSnippet(pipe, k, v).Value
		}
		pipe.Exec()
	}
	return snippetscollection
}
