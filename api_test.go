// Copyright 2015 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	dtesting "github.com/fsouza/go-dockerclient/testing"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

var _ = check.Suite(&S{})

type S struct {
	server *dtesting.DockerServer
}

func (s *S) SetUpTest(c *check.C) {
	var err error
	config.MongoURL = "mongodb://127.0.0.1:27017/diaats"
	config.DBName = "diaats"
	s.server, err = dtesting.NewServer("127.0.0.1:0", nil, nil)
	c.Assert(err, check.IsNil)
	config.DockerHost = s.server.URL()
}

func (s *S) TearDownTest(c *check.C) {
	s.server.Stop()
}

func (*S) TestHandlerSuccess(c *check.C) {
	var called bool
	config.Username = "admin"
	config.Password = "admin123"
	h := handler(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/", nil)
	c.Assert(err, check.IsNil)
	request.SetBasicAuth(config.Username, config.Password)
	h.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(called, check.Equals, true)
}

func (*S) TestHandlerNoUser(c *check.C) {
	var called bool
	config.Username = ""
	config.Password = "admin123"
	h := handler(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/", nil)
	c.Assert(err, check.IsNil)
	request.SetBasicAuth(config.Username, config.Password)
	h.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(called, check.Equals, true)
}

func (*S) TestHandlerNoPassword(c *check.C) {
	var called bool
	config.Username = "admin"
	config.Password = ""
	h := handler(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/", nil)
	c.Assert(err, check.IsNil)
	request.SetBasicAuth(config.Username, config.Password)
	h.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(called, check.Equals, true)
}

func (*S) TestHandlerNoAuthenticationHeader(c *check.C) {
	var called bool
	config.Username = "admin"
	config.Password = "admin123"
	h := handler(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/", nil)
	c.Assert(err, check.IsNil)
	h.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusUnauthorized)
	c.Assert(called, check.Equals, false)
}

func (*S) TestHandlerWrongCredentials(c *check.C) {
	var called bool
	config.Username = "admin"
	config.Password = "admin123"
	h := handler(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/", nil)
	c.Assert(err, check.IsNil)
	request.SetBasicAuth("admin123", "admin")
	h.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusUnauthorized)
	c.Assert(called, check.Equals, false)
}
