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
  "encoding/base64"

  log "github.com/Sirupsen/logrus"
)

type UserResource struct {
  Redis *redis.Client
  Engine *gin.RouterGroup
} 

func (u UserResource) Routes() {
  u.Engine.GET("/users/:uuid", u.Get)
}

func (u UserResource) Get(c *gin.Context) {
  val, err := u.Redis.Get("user::id::"+c.Param("uuid")).Result()
  if err == redis.Nil {
    NotFound("User", c)
  } else if err != nil {
    log.Error(err)
    InternalError(c)
  } else {
    c.JSON(200, gin.H{
      "message": "User fetched.",
      "value": val,
    })
  }
}

type User struct {
  UserResource *UserResource
  Uuid string
  Username string
  Password string
  PasswordDigest string
} 

func (u *User) EncodedUsername() string {
  return base64.StdEncoding.EncodeToString([]byte(u.Username))
}

func (u *User) GetUuid() string {
  val, err := u.UserResource.Redis.Get("user::name::"+u.EncodedUsername()).Result()
  if err != nil {
    log.Error(err)
    return ""
  } else {
    return val
  }
}

func (u *User) GetUsername() string {
  val, err := u.UserResource.Redis.Get("user::id::"+u.Uuid).Result()
  if err != nil {
    log.Error(err)
    return ""
  } else {
    str, _ := base64.StdEncoding.DecodeString(val)
    return string(str)
  }
}
