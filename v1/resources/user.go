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
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/redis.v3"

	log "github.com/Sirupsen/logrus"
)

//UserResource - Users API endpoint
type UserResource struct {
	Redis  *redis.Client
	Engine *gin.RouterGroup
}

//Routes - Users routing definition
func (u UserResource) Routes() {
	u.Engine.GET("/users/:uuid", u.Get)
	u.Engine.POST("/users", u.Create)
	u.Engine.POST("/users/:uuid/uuid", u.ResetUUID)
}

/*
Get - Fetch user by id

	{
		"user": {
			"uuid": <UUID>,
			"username": <Name of user>
		}
	}
*/
func (u UserResource) Get(c *gin.Context) {
	user := User{
		UserResource: &u,
		UUID:         c.Param("uuid"),
	}
	if user.GetUsername() == "" {
		NotFound("User", c)
	} else {
		c.JSON(200, gin.H{
			"user": map[string]string{
				"uuid":     user.GetUUID(),
				"username": user.GetUsername(),
			},
		})
	}
}

/*
Create - User by username and password

  # curl $API/users -d 'username=moo' -d 'password=swordfish'
	{
		"user":
			"uuid": ...,
			"username": ...
		}
	}

	# curl $API/users -d 'username=moo' -d 'password=swordfish'
	=> HTTP 405
	{
		"message": "User already available"
	}
*/
func (u UserResource) Create(c *gin.Context) {
	user := base64.StdEncoding.EncodeToString([]byte(c.PostForm("username")))
	_, err := u.Redis.Get("user::name::" + user).Result()
	if err != redis.Nil {
		c.JSON(405, gin.H{
			"message": "User already available",
		})
	} else {
		newuser := NewUser(c.PostForm("username"), c.PostForm("password"), &u)
		if newuser.Save() {
			c.JSON(201, gin.H{
				"user": map[string]string{
					"uuid":     newuser.UUID,
					"username": newuser.Username,
				},
			})
		} else {
			c.JSON(422, gin.H{
				"message": "Createing new user failed.",
			})
		}
	}
}

/*
ResetUUID - reset users uuid.

	# curl -X POST $API/users/<uuid>
	{
		"user": {
			"uuid": <UUID>,
			"username": <Name of user>
		}
	}
*/
func (u UserResource) ResetUUID(c *gin.Context) {
	user := User{
		UserResource: &u,
		UUID:         c.Param("uuid"),
	}
	if user.GetUsername() == "" {
		NotFound("User", c)
	} else {
		c.JSON(200, gin.H{
			"user": map[string]string{
				"uuid":     user.ResetUUID(),
				"username": user.GetUsername(),
			},
		})
	}
}

// User model
type User struct {
	UserResource   *UserResource
	UUID           string
	Username       string
	Password       string
	PasswordDigest string
}

// NewUser returns new user instance
func NewUser(username string, password string, ur *UserResource) User {
	newuser := User{
		UserResource: ur,
		Username:     username,
		Password:     password,
		UUID:         uuid.NewV4().String(),
	}
	v, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	newuser.PasswordDigest = string(v)
	return newuser
}

// Save user by using current redis session to database.
func (u *User) Save() bool {
	pipe := u.UserResource.Redis.Pipeline()
	defer pipe.Close()
	pipe.Set("user::id::"+u.GetUUID(), u.EncodedUsername(), 0)
	pipe.Set("user::name::"+u.EncodedUsername(), u.GetUUID(), 0)
	pipe.Set("user::pass::"+u.EncodedUsername(), u.GetPasswordDigest(), 0)
	_, err := pipe.Exec()
	if err != nil {
		log.Error(err, "Error during User.Save().")
		return false
	}
	return true
}

// EncodedUsername encodes the username into base64 to prevent
// to handle each kind of username.
func (u *User) EncodedUsername() string {
	return base64.StdEncoding.EncodeToString([]byte(u.GetUsername()))
}

// GetUUID returns the objects internal UUID or prefetch them from datastore
func (u *User) GetUUID() string {
	if u.UUID == "" {
		val, err := u.UserResource.Redis.Get("user::name::" + u.EncodedUsername()).Result()
		if err != nil {
			log.Error(err, "Fetching UUID failed")
		} else {
			u.UUID = val
		}
	}
	return u.UUID
}

// ResetUUID sets a new UUID to current user.
func (u *User) ResetUUID() string {
	id := uuid.NewV4().String()
	pipe := u.UserResource.Redis.Pipeline()
	defer pipe.Close()
	pipe.Set("user::id::"+id, u.EncodedUsername(), 0)
	pipe.Set("user::name::"+u.EncodedUsername(), id, 0)
	pipe.Del("user::id::" + u.GetUUID())
	_, err := pipe.Exec()
	if err != nil {
		log.Error(err, "Error on resetting UUID.")
	} else {
		u.UUID = id
	}
	return id
}

// GetUsername returns the objects internal username or prefetch them from datastore
func (u *User) GetUsername() string {
	if u.Username == "" {
		val, err := u.UserResource.Redis.Get("user::id::" + u.UUID).Result()
		if err != nil {
			log.Error(err, "Fetching Username failed")
		} else {
			str, _ := base64.StdEncoding.DecodeString(val)
			u.Username = string(str)
		}
	}
	return u.Username
}

// GetPasswordDigest returns the objects internal passwordDigest or prefetch them from datastore
func (u *User) GetPasswordDigest() string {
	if u.PasswordDigest == "" {
		val, err := u.UserResource.Redis.Get("user::pass::" + u.EncodedUsername()).Result()
		if err != nil {
			log.Error(err, "Fetching PasswordDigest failed")
		} else {
			u.PasswordDigest = val
		}
	}
	return u.PasswordDigest
}
