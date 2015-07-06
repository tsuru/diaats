// Copyright 2015 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"gopkg.in/check.v1"
)

func (*S) TestLoadConfig(c *check.C) {
	os.Setenv("DOCKER_HOST", "tcp://192.168.50.4:2375")
	os.Setenv("API_USERNAME", "root")
	os.Setenv("API_PASSWORD", "r00t")
	os.Setenv("DOCKER_CONFIG", `{"Memory":268435456}`)
	os.Setenv("IMAGE_PLANS", `[{"image":"memcached:1","plan":"memcached_1"},{"image":"memcached:1.3","plan":"memcached_1_3"}]`)
	os.Setenv("MONGODB_URL", "mongodb://user:password@host:27017/diaats")
	os.Setenv("MONGODB_DB_NAME", "diaaats")
	loadConfig()
	c.Assert(config.DockerHost, check.Equals, "tcp://192.168.50.4:2375")
	c.Assert(config.Username, check.Equals, "root")
	c.Assert(config.Password, check.Equals, "r00t")
	c.Assert(config.HostConfig.Memory, check.Equals, int64(268435456))
	c.Assert(config.Plans, check.HasLen, 2)
	c.Assert(config.Plans[0].Name, check.Equals, "memcached_1")
	c.Assert(config.Plans[0].Image, check.Equals, "memcached:1")
	c.Assert(config.Plans[1].Name, check.Equals, "memcached_1_3")
	c.Assert(config.Plans[1].Image, check.Equals, "memcached:1.3")
	c.Assert(config.MongoURL, check.Equals, "mongodb://user:password@host:27017/diaats")
	c.Assert(config.DBName, check.Equals, "diaaats")
}

func (*S) TestLoadConfigNoDockerConfigNoDBName(c *check.C) {
	os.Setenv("DOCKER_HOST", "tcp://192.168.50.4:2375")
	os.Setenv("API_USERNAME", "root")
	os.Setenv("API_PASSWORD", "r00t")
	os.Setenv("IMAGE_PLANS", `[{"image":"memcached:1","plan":"memcached_1"},{"image":"memcached:1.3","plan":"memcached_1_3"}]`)
	os.Setenv("MONGODB_URL", "mongodb://user:password@host:27017/diaatss")
	os.Unsetenv("MONGODB_DB_NAME")
	os.Unsetenv("DOCKER_CONFIG")
	loadConfig()
	c.Assert(config.DockerHost, check.Equals, "tcp://192.168.50.4:2375")
	c.Assert(config.Username, check.Equals, "root")
	c.Assert(config.Password, check.Equals, "r00t")
	c.Assert(config.HostConfig, check.IsNil)
	c.Assert(config.Plans, check.HasLen, 2)
	c.Assert(config.Plans[0].Name, check.Equals, "memcached_1")
	c.Assert(config.Plans[0].Image, check.Equals, "memcached:1")
	c.Assert(config.Plans[1].Name, check.Equals, "memcached_1_3")
	c.Assert(config.Plans[1].Image, check.Equals, "memcached:1.3")
	c.Assert(config.MongoURL, check.Equals, "mongodb://user:password@host:27017/diaatss")
	c.Assert(config.DBName, check.Equals, "diaatss")
}
