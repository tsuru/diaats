// Copyright 2015 diaats authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsouza/go-dockerclient"
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
	config.Username = ""
	config.Password = ""
	config.MongoURL = "mongodb://127.0.0.1:27017/diaats"
	config.DBName = "diaats"
	s.server, err = dtesting.NewServer("127.0.0.1:0", nil, nil)
	c.Assert(err, check.IsNil)
	config.DockerHost = s.server.URL()
	coll, err := connect()
	c.Assert(err, check.IsNil)
	coll.RemoveAll(nil)
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

func (*S) TestCreateInstanceHandler(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	body := strings.NewReader("name=mycache&plan=supermemcached")
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
	defer DestroyInstance("mycache")
	_, err = GetInstance("mycache")
	c.Assert(err, check.IsNil)
}

func (*S) TestCreateInstanceHandlerDuplicate(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.IsNil)
	defer DestroyInstance("mycache")
	body := strings.NewReader("name=mycache&plan=supermemcached")
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusConflict)
}

func (*S) TestCreateInstanceHandlerInvalidPlan(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	body := strings.NewReader("name=mycache&plan=hipermemcached")
	request, err := http.NewRequest("POST", "/resources", body)
	c.Assert(err, check.IsNil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, "plan not found\n")
}

func (*S) TestBindAppHandler(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.IsNil)
	defer DestroyInstance("mycache")
	request, err := http.NewRequest("POST", "/resources/mycache/bind-app", strings.NewReader(""))
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Log(recorder.Body.String())
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	expected := map[string]string{"DIAATS_SUPERMEMCACHED_INSTANCE": "[]"}
	var result map[string]string
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	c.Assert(err, check.IsNil)
	c.Assert(result, check.DeepEquals, expected)
}

func (*S) TestBindAppHandlerNotFound(c *check.C) {
	request, err := http.NewRequest("POST", "/resources/mycache/bind-app", strings.NewReader(""))
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Log(recorder.Body.String())
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
}

func (*S) TestRemoveInstanceHandler(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.IsNil)
	defer DestroyInstance("mycache")
	request, err := http.NewRequest("DELETE", "/resources/mycache", strings.NewReader(""))
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	_, err = GetInstance("mycache")
	c.Assert(err, check.Equals, ErrInstanceNotFound)
}

func (*S) TestRemoveInstanceHandlerNotFound(c *check.C) {
	request, err := http.NewRequest("DELETE", "/resources/mycache", strings.NewReader(""))
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Log(recorder.Body.String())
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
}

func (*S) TestInstanceStatusHandler(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.IsNil)
	defer DestroyInstance("mycache")
	request, err := http.NewRequest("GET", "/resources/mycache/status", strings.NewReader(""))
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNoContent)
}

func (*S) TestInstanceStatusHandlerNotFound(c *check.C) {
	request, err := http.NewRequest("GET", "/resources/mycache/status", strings.NewReader(""))
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
}

func (*S) TestPlansHandler(c *check.C) {
	config.Plans = []Plan{
		{Name: "supermemcached", Image: "memcached"},
		{Name: "hipermemcached", Image: "memcached:powerful"},
		{Name: "memcached-legacy", Image: "memcached:0.4"},
	}
	request, err := http.NewRequest("GET", "/resources/plans", strings.NewReader(""))
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	handler := buildMuxer()
	handler.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	expected := []map[string]string{
		{"name": "supermemcached", "description": "Run containers of the image memcached"},
		{"name": "hipermemcached", "description": "Run containers of the image memcached:powerful"},
		{"name": "memcached-legacy", "description": "Run containers of the image memcached:0.4"},
	}
	var got []map[string]string
	err = json.NewDecoder(recorder.Body).Decode(&got)
	c.Assert(err, check.IsNil)
	c.Assert(got, check.DeepEquals, expected)
}
