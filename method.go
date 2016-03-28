// Copyright 2014 martini-contrib/method Authors
// Copyright 2014 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	headerHTTPMethodOverride = "X-HTTP-Method-Override"
	paramHTTPMethodOverride  = "_method"
)

var httpMethods = []string{"PUT", "PATCH", "DELETE"}
var errInvalidOverrideMethod = errors.New("invalid override method")

func isValidOverrideMethod(method string) bool {
	for _, m := range httpMethods {
		if m == method {
			return true
		}
	}
	return false
}

func overrideRequestMethod(r *http.Request, method string) error {
	if !isValidOverrideMethod(method) {
		return errInvalidOverrideMethod
	}
	r.Header.Set(headerHTTPMethodOverride, method)
	return nil
}

func methodOverride() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request

		if r.Method == "POST" {
			m := r.FormValue(paramHTTPMethodOverride)
			if isValidOverrideMethod(m) {
				overrideRequestMethod(r, m)
			}
			m = r.Header.Get(headerHTTPMethodOverride)
			if isValidOverrideMethod(m) {
				r.Method = m
			}
		}
	}
}
