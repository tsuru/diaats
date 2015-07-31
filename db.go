// Copyright 2015 diaats authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "gopkg.in/mgo.v2"

type collection struct {
	*mgo.Collection
	*mgo.Session
}

func connect() (*collection, error) {
	session, err := mgo.DialWithTimeout(config.MongoURL, 30e9)
	if err != nil {
		return nil, err
	}
	coll := session.DB(config.DBName).C("instances")
	return &collection{coll, session}, nil
}
