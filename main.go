// Copyright 2015 diaats authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
)

var listen string

func init() {
	flag.StringVar(&listen, "l", "0.0.0.0:8080", "Address to bind")
}

func main() {
	flag.Parse()
	loadConfig()
	handler := buildMuxer()
	log.Printf("Binding on %q", listen)
	http.ListenAndServe(listen, handler)
}
