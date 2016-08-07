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
	"encoding/base64"
	"errors"
	"github.com/muhproductions/muh/helper"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"string"
)

// User model
type User struct {
	UUID           string
	Username       string
	PasswordDigest string
}

// NewUser returns new user instance
func NewUser(username string, password string) User {
	newuser := User{
		Username: username,
		UUID:     uuid.NewV4().String(),
	}
	newuser.SetPassword(password)
	return newuser
}

// FindUserByName will look for a user and will return
// it available. Otherwise it returns an error.
func FindUserByName(name string) (User, error) {
	return findAtomicUser(User{Username: name})
}

// FindUserByUUID will look for a user and will return
// it available. Otherwise it returns an error.
func FindUserByUUID(uuid string) (User, error) {
	return findAtomicUser(User{UUID: uuid})
}

func findAtomicUser(user User) (User, error) {
	if user.Available() {
		return user, nil
	}
	return user, errors.New("user not existing")
}

//EqualsPassword verifies password
func (u *User) EqualsPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(u.GetPasswordDigest()),
		[]byte(password),
	)
	return (err == nil)
}

func fetchIdsByKey(key string) []string {
	var ids []string
	for _, v := range helper.RedisClient().Keys(key + "::*").Val() {
		ids = append(ids, strings.Replace(v, key, "", -1))
	}
	return ids
}

// CreatedGists returns a list with gist id's
func (u *User) CreatedGists() []string {
	return fetchIdsByKey("users::" + u.GetUUID() + "::gists")
}

// MarkGist marks a new gist
func (u *User) MarkGist(UUID string) bool {
	gist := Gist{UUID: UUID}
	if gist.Exists() {
		helper.RedisClient().SAdd("users::" + u.GetUUID() + "::marked_gists")
		return true
	}
	return false
}

// MarkedGists returns a list of marked gists.
func (u *User) MarkedGists() []string {
	return fetchIdsByKey("users::" + u.GetUUID() + "::marked_gists")
}

//SetPassword calculates a bcrypt hash and updates objects PasswordDigest.
func (u *User) SetPassword(password string) {
	v, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	u.PasswordDigest = string(v)
}

//Available check if user is available.
func (u *User) Available() bool {
	if u.Username != "" {
		return helper.RedisClient().Exists(u.keyName()).Val()
	} else if u.UUID != "" {
		return helper.RedisClient().Exists(u.keyID()).Val()
	}
	return false
}

// Save user by using current redis session to database.
func (u *User) Save() bool {
	pipe := helper.RedisClient().Pipeline()
	defer pipe.Close()
	pipe.Set(u.keyID(), u.EncodedUsername(), 0)
	pipe.Set(u.keyName(), u.GetUUID(), 0)
	pipe.Set(u.keyPass(), u.GetPasswordDigest(), 0)
	_, err := pipe.Exec()
	return (err == nil)
}

func (u *User) keyID() string {
	if u.UUID == "" {
		return "user::id::" + u.GetUUID()
	}
	return "user::id::" + u.UUID
}

func (u *User) keyName() string {
	return "user::name::" + u.EncodedUsername()
}

func (u *User) keyPass() string {
	return "user::pass::" + u.EncodedUsername()
}

// EncodedUsername encodes the username into base64 to prevent
// to handle each kind of username.
func (u *User) EncodedUsername() string {
	username := u.Username
	if username == "" {
		username = u.GetUsername()
	}
	return base64.StdEncoding.EncodeToString([]byte(username))
}

func (u *User) cachedResponse(prop string, key string) string {
	v := reflect.ValueOf(u).Elem().FieldByName(prop)
	if v.String() == "" {
		v.SetString(helper.RedisClient().Get(key).Val())
	}
	return v.String()
}

// GetUUID returns the objects internal UUID or prefetch them from datastore
func (u *User) GetUUID() string {
	return u.cachedResponse("UUID", u.keyName())
}

// ResetUUID sets a new UUID to current user.
func (u *User) ResetUUID() string {
	id := uuid.NewV4().String()
	pipe := helper.RedisClient().Pipeline()
	defer pipe.Close()
	pipe.Set(u.keyID(), u.EncodedUsername(), 0)
	pipe.Set(u.keyName(), id, 0)
	pipe.Del(u.keyID())
	pipe.Exec()
	u.UUID = id
	return id
}

// GetUsername returns the objects internal username or prefetch them from datastore
func (u *User) GetUsername() string {
	return u.cachedResponse("Username", u.keyID())
}

// GetPasswordDigest returns the objects internal passwordDigest or prefetch them from datastore
func (u *User) GetPasswordDigest() string {
	return u.cachedResponse("PasswordDigest", u.keyPass())
}
