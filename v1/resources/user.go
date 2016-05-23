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
  "github.com/satori/go.uuid"
  "golang.org/x/crypto/bcrypt"
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
  u.Engine.POST("/users", u.Create)
  u.Engine.POST("/users/:uuid/uuid", u.ResetUuid)
}

func (u UserResource) Get(c *gin.Context) {
  user := User{
    UserResource: &u,
    Uuid: c.Param("uuid"),
  }
  if user.GetUsername() == "" {
    NotFound("User", c)
  } else {
    c.JSON(200, gin.H{
      "user": map[string]string{
        "uuid": user.GetUuid(),
        "username": user.GetUsername(),
      },
    })
  }
}

func (u UserResource) Create(c *gin.Context) {
  user := base64.StdEncoding.EncodeToString([]byte(c.PostForm("username")))
  _, err := u.Redis.Get("user::name::"+user).Result()
  if err != redis.Nil {
    c.JSON(405, gin.H{
      "message": "User already available",
    })
  } else {
    new_user := NewUser(c.PostForm("username"),c.PostForm("password"), &u)
    if new_user.Save() {
      c.JSON(201, gin.H{
        "user":  map[string]string{
          "uuid": new_user.Uuid,
          "username": new_user.Username,
        },
      })
    } else { 
      c.JSON(422, gin.H{
        "message": "Createing new user failed.",
      })
    }
  }
}

func (u UserResource) ResetUuid(c *gin.Context) {
  user := User{
    UserResource: &u,
    Uuid: c.Param("uuid"),
  }
  if user.GetUsername() == "" {
    NotFound("User", c)
  } else {
    c.JSON(200, gin.H{
      "user": map[string]string{
        "uuid": user.ResetUuid(),
        "username": user.GetUsername(),
      },
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

func NewUser(username string, password string, ur *UserResource) User {
  new_user := User{
    UserResource: ur,
    Username: username,
    Password: password,
    Uuid: uuid.NewV4().String(),
  }
  v, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
  new_user.PasswordDigest = string(v)
  return new_user
}

func (u *User) Save() bool {
  pipe := u.UserResource.Redis.Pipeline()
  defer pipe.Close()
  pipe.Set("user::id::"+u.GetUuid(), u.EncodedUsername(), 0)
  pipe.Set("user::name::"+u.EncodedUsername(), u.GetUuid(), 0)
  pipe.Set("user::pass::"+u.EncodedUsername(), u.GetPasswordDigest(), 0)
  _, err := pipe.Exec()
  if err != nil {
    log.Error(err, "Error during User.Save().")
    return false
  } else {
    return true
  }
}

func (u *User) EncodedUsername() string {
  return base64.StdEncoding.EncodeToString([]byte(u.GetUsername()))
}

func (u *User) GetUuid() string {
  if u.Uuid == "" {
    val, err := u.UserResource.Redis.Get("user::name::"+u.EncodedUsername()).Result()
    if err != nil {
      log.Error(err, "Fetching Uuid failed")
    } else {
      u.Uuid = val
    }
  }
  return u.Uuid
}

func (u *User) ResetUuid() string {
  id := uuid.NewV4().String()
  pipe := u.UserResource.Redis.Pipeline()
  defer pipe.Close()
  pipe.Set("user::id::"+id, u.EncodedUsername(), 0)
  pipe.Set("user::name::"+u.EncodedUsername(), id, 0)
  pipe.Del("user::id::"+u.GetUuid())
  _, err := pipe.Exec()
  if err != nil {
    log.Error(err, "Error on resetting Uuid.")
  } else {
    u.Uuid = id
  }
  return id
}

func (u *User) GetUsername() string {
  if u.Username == "" {
    val, err := u.UserResource.Redis.Get("user::id::"+u.Uuid).Result()
    if err != nil {
      log.Error(err, "Fetching Username failed")
    } else {
      str, _ := base64.StdEncoding.DecodeString(val)
      u.Username = string(str)
    }
  }
  return u.Username
}

func (u *User) GetPasswordDigest() string {
  if u.PasswordDigest == "" {
    val, err := u.UserResource.Redis.Get("user::pass::"+u.EncodedUsername()).Result()
    if err != nil {
      log.Error(err, "Fetching PasswordDigest failed")
    } else {
      u.PasswordDigest = val
    }
  }
  return u.PasswordDigest
}
