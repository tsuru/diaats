// Copyright 2016 diaats authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	ErrInstanceAlreadyExists = errors.New("instance already exists")
	ErrInstanceNotFound      = errors.New("instance not found")
)

type Instance struct {
	Name        string
	DockerHost  string
	ContainerID string
	HostPorts   []string
	Envs        []string
	Plan        Plan
}

// Endpoints returns a list of endpoints to this instance.
func (i *Instance) Endpoints() []string {
	url_, err := url.Parse(i.DockerHost)
	if err != nil {
		log.Printf("ERROR - failed to parse instance Docker host: %s", err)
		return nil
	}
	host, _, err := net.SplitHostPort(url_.Host)
	if err != nil {
		host = url_.Host
	}
	result := make([]string, len(i.HostPorts))
	for i, port := range i.HostPorts {
		result[i] = host + ":" + port
	}
	return result
}

// EnvMap returns the set of environment variables in the instance. The API
// gets these environment variables from Docker when creating the instance.
func (i *Instance) EnvMap() map[string]string {
	result := make(map[string]string, len(i.Envs))
	for _, env := range i.Envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// CreateInstance creates a new instance using the given name and plan.
func CreateInstance(name string, plan *Plan) error {
	instance := Instance{
		Name:       name,
		Plan:       *plan,
		DockerHost: config.DockerHost,
	}
	if _, err := GetInstance(name); err == nil {
		return ErrInstanceAlreadyExists
	}
	collection, err := connect()
	if err != nil {
		return err
	}
	defer collection.Close()
	opts := docker.CreateContainerOptions{
		Name:       fmt.Sprintf("diaats-%s-%s", plan.Name, name),
		Config:     &docker.Config{Cmd: plan.Args, Image: plan.Image},
		HostConfig: config.HostConfig,
	}
	client, err := docker.NewClient(instance.DockerHost)
	if err != nil {
		return err
	}
	container, err := client.CreateContainer(opts)
	if err != nil {
		return err
	}
	err = client.StartContainer(container.ID, config.HostConfig)
	if err != nil {
		return err
	}
	container, err = client.InspectContainer(container.ID)
	if err != nil {
		return err
	}
	instance.ContainerID = container.ID
	instance.HostPorts = make([]string, 0, len(container.NetworkSettings.Ports))
	for _, ports := range container.NetworkSettings.Ports {
		for _, p := range ports {
			instance.HostPorts = append(instance.HostPorts, p.HostPort)
		}
	}
	instance.Envs = container.Config.Env
	err = collection.Insert(instance)
	if err != nil {
		client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID, Force: true})
	}
	return err
}

// DestroyInstance destroys the instance identified by the given name.
func DestroyInstance(name string) error {
	instance, err := GetInstance(name)
	if err != nil {
		return err
	}
	client, err := docker.NewClient(instance.DockerHost)
	if err != nil {
		return err
	}
	coll, err := connect()
	if err != nil {
		return err
	}
	defer coll.Close()
	opts := docker.RemoveContainerOptions{ID: instance.ContainerID, Force: true}
	err = client.RemoveContainer(opts)
	if err != nil {
		log.Printf("ERROR - failed to remove Docker container %q: %s", instance.ContainerID, err)
	}
	return coll.Remove(bson.M{"name": instance.Name})
}

// GetInstance returns the instance identified by the given name.
func GetInstance(name string) (*Instance, error) {
	var instance Instance
	coll, err := connect()
	if err != nil {
		return nil, err
	}
	defer coll.Close()
	err = coll.Find(bson.M{"name": name}).One(&instance)
	if err == mgo.ErrNotFound {
		return nil, ErrInstanceNotFound
	}
	return &instance, err
}
