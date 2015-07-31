// Copyright 2015 diaats authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (*S) TestInstanceEndpoints(c *check.C) {
	instance := Instance{
		DockerHost: "tcp://192.168.50.4:2375",
		HostPorts:  []string{"49159", "49160", "49161"},
	}
	expected := []string{
		"192.168.50.4:49159",
		"192.168.50.4:49160",
		"192.168.50.4:49161",
	}
	c.Assert(instance.Endpoints(), check.DeepEquals, expected)
}

func (s *S) TestCreateInstance(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.IsNil)
	defer DestroyInstance("mycache")
	coll, err := connect()
	c.Assert(err, check.IsNil)
	defer coll.Close()
	var instance Instance
	err = coll.Find(bson.M{"name": "mycache"}).One(&instance)
	c.Assert(err, check.IsNil)
	c.Assert(instance.Name, check.Equals, "mycache")
	c.Assert(instance.Plan, check.DeepEquals, config.Plans[0])
	c.Assert(instance.DockerHost, check.Equals, config.DockerHost)
	container, err := client.InspectContainer(instance.ContainerID)
	c.Assert(err, check.IsNil)
	c.Assert(container.Name, check.Equals, "diaats-supermemcached-mycache")
}

func (s *S) TestCreateInstanceDuplicate(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.IsNil)
	defer DestroyInstance("mycache")
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.Equals, ErrInstanceAlreadyExists)
}

func (s *S) TestDestroyInstance(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.IsNil)
	var instance Instance
	coll, err := connect()
	c.Assert(err, check.IsNil)
	defer coll.Close()
	err = coll.Find(bson.M{"name": "mycache"}).One(&instance)
	c.Assert(err, check.IsNil)
	err = DestroyInstance("mycache")
	c.Assert(err, check.IsNil)
	_, err = client.InspectContainer(instance.ContainerID)
	c.Assert(err, check.NotNil)
	e, ok := err.(*docker.NoSuchContainer)
	c.Assert(ok, check.Equals, true)
	c.Assert(e.ID, check.Equals, instance.ContainerID)
	err = coll.Find(bson.M{"name": "mycache"}).One(&instance)
	c.Assert(err, check.Equals, mgo.ErrNotFound)
}

func (s *S) TestDestroyInstanceNotFound(c *check.C) {
	err := DestroyInstance("watcache")
	c.Assert(err, check.Equals, ErrInstanceNotFound)
}

func (s *S) TestGetInstance(c *check.C) {
	client, err := docker.NewClient(config.DockerHost)
	c.Assert(err, check.IsNil)
	err = client.PullImage(docker.PullImageOptions{Repository: "memcached"}, docker.AuthConfiguration{})
	c.Assert(err, check.IsNil)
	config.Plans = []Plan{{Name: "supermemcached", Image: "memcached"}}
	err = CreateInstance("mycache", &config.Plans[0])
	c.Assert(err, check.IsNil)
	defer DestroyInstance("mycache")
	var dbInstance Instance
	coll, err := connect()
	c.Assert(err, check.IsNil)
	defer coll.Close()
	err = coll.Find(bson.M{"name": "mycache"}).One(&dbInstance)
	c.Assert(err, check.IsNil)
	instance, err := GetInstance("mycache")
	c.Assert(err, check.IsNil)
	c.Assert(*instance, check.DeepEquals, dbInstance)
}

func (s *S) TestGetInstanceNotFound(c *check.C) {
	instance, err := GetInstance("watcache")
	c.Assert(err, check.Equals, ErrInstanceNotFound)
	c.Assert(instance, check.IsNil)
}
