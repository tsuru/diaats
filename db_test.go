// Copyright 2015 diaats authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"gopkg.in/check.v1"
)

func (*S) TestConnect(c *check.C) {
	config.MongoURL = "mongodb://127.0.0.1:27017/diaats"
	config.DBName = "diaats"
	coll, err := connect()
	c.Assert(err, check.IsNil)
	defer coll.Close()
	c.Assert(coll.Database.Name, check.Equals, config.DBName)
	c.Assert(coll.Name, check.Equals, "instances")
}
