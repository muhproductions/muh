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
)

// NotFound - Generic resource 404
func NotFound(resource string, c *gin.Context) {
	c.JSON(404, gin.H{
		"message": resource + " not found",
	})
}

// InternalError - Generic error
func InternalError(c *gin.Context) {
	c.JSON(422, gin.H{
		"message": "Internal error occured.",
	})
}
