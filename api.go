// Copyright 2015 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
)

var config struct {
	DockerHost string
	Username   string
	Password   string
}

func handler(fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Username != "" && config.Password != "" {
			if username, password, ok := r.BasicAuth(); !ok || username != config.Username || password != config.Password {
				w.Header().Add("WWW-Authenticate", `Basic realm="diaats"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		fn.ServeHTTP(w, r)
	})
}

func main() {
}
