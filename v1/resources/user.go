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

//UserResource - Users API endpoint
type UserResource struct {
	Engine *gin.RouterGroup
}

//Login - keeps user data.
type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

//Routes - Users routing definition
func (u UserResource) Routes() {
	u.Engine.GET("/users/:userid/profile", u.Get)
	u.Engine.PUT("/users/:userid/gists", GistResource{Engine: u.Engine}.CreateSnippets)
	u.Engine.PUT("/users/:userid/uuid", u.ResetUUID)
	u.Engine.POST("/users/authorize", u.Authorize)
	u.Engine.POST("/users", u.Create)
}

func checkUserExists(c *gin.Context) models.User {
	user := models.User{
		UUID: c.Param("userid"),
	}
	if user.GetUsername() == "" {
		NotFound("User", c)
		c.AbortWithStatus(404)
	}
	return user
}

//Authorize - authorize users by json response
func (u UserResource) Authorize(c *gin.Context) {
	var login Login
	if (c.PostForm("username") == "" && c.BindJSON(&login) == nil) || c.Bind(&login) == nil {
		user := models.User{Username: login.Username}
		if user.EqualsPassword(login.Password) {
			c.JSON(200, gin.H{
				"user": map[string]string{
					"uuid": user.GetUUID(),
				},
			})
			return
		}
		c.AbortWithStatus(403)
	}
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
		"gists": map[string][]string{
			"created": user.CreatedGists(),
			"marked":  user.MarkedGists(),
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
	var login Login
	err := c.PostForm("username") == "" && c.BindJSON(&login) == nil || c.Bind(&login) == nil
	if !err {
		c.AbortWithStatus(400)
		return
	}
	newuser := models.NewUser(login.Username, login.Password)
	if newuser.Available() {
		c.AbortWithStatus(405)
		return
	}
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
