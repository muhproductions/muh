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
	"github.com/muhproductions/muh-api/helper"
	"github.com/muhproductions/muh-api/v1/models"
	"gopkg.in/redis.v3"
)

//UserResource - Users API endpoint
type UserResource struct {
	Engine *gin.RouterGroup
}

//Routes - Users routing definition
func (u UserResource) Routes() {
	u.Engine.GET("/users/:uuid", u.Get)
	u.Engine.POST("/users", u.Create)
	u.Engine.POST("/users/:uuid/uuid", u.ResetUUID)
}

func checkUserExists(c *gin.Context) models.User {
	user := models.User{
		UUID: c.Param("uuid"),
	}
	if user.GetUsername() == "" {
		NotFound("User", c)
		c.AbortWithStatus(404)
	}
	return user
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
	user := checkUserExists(c)
	c.JSON(200, gin.H{
		"user": map[string]string{
			"uuid":     user.GetUUID(),
			"username": user.GetUsername(),
		},
	})
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
	_, err := helper.RedisClient().Get("user::name::" + user).Result()
	if err != redis.Nil {
		c.JSON(405, gin.H{
			"message": "User already available",
		})
	} else {
		newuser := models.NewUser(c.PostForm("username"), c.PostForm("password"))
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
	user := checkUserExists(c)
	c.JSON(200, gin.H{
		"user": map[string]string{
			"uuid":     user.ResetUUID(),
			"username": user.GetUsername(),
		},
	})
}
