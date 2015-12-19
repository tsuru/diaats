// Copyright 2015 diaats authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/pat"
)

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

func createInstance(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "please provide the name of the instance", http.StatusBadRequest)
		return
	}
	planName := r.FormValue("plan")
	if planName == "" {
		http.Error(w, "please provide the name of the plan", http.StatusBadRequest)
		return
	}
	plan, err := getPlan(planName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = CreateInstance(name, plan)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrInstanceAlreadyExists {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func bindApp(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	instance, err := GetInstance(name)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrInstanceNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	encodedEndpoints, _ := json.Marshal(instance.Endpoints())
	encodedEnvs, _ := json.Marshal(instance.EnvMap())
	envVarName := fmt.Sprintf("DIAATS_%s_INSTANCE", strings.ToUpper(instance.Plan.Name))
	dockerEnvVarName := fmt.Sprintf("DIAATS_%s_DOCKER_ENVS", strings.ToUpper(instance.Plan.Name))
	data := map[string]string{
		envVarName:       string(encodedEndpoints),
		dockerEnvVarName: string(encodedEnvs),
	}
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("ERROR - failed to encode JSON: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func removeInstance(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	err := DestroyInstance(name)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrInstanceNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
}

func instanceStatus(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	_, err := GetInstance(name)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrInstanceNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func listPlans(w http.ResponseWriter, r *http.Request) {
	result := make([]map[string]string, len(config.Plans))
	for i, plan := range config.Plans {
		result[i] = plan.ToMap()
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildMuxer() http.Handler {
	m := pat.New()
	m.Post("/resources/{name}/bind-app", handler(bindApp))
	m.Delete("/resources/{name}/bind-app", handler(func(http.ResponseWriter, *http.Request) {}))
	m.Post("/resources/{name}/bind", handler(func(http.ResponseWriter, *http.Request) {}))
	m.Delete("/resources/{name}/bind", handler(func(http.ResponseWriter, *http.Request) {}))
	m.Get("/resources/{name}/status", handler(instanceStatus))
	m.Delete("/resources/{name}", handler(removeInstance))
	m.Get("/resources/plans", handler(listPlans))
	m.Post("/resources", handler(createInstance))
	return m
}
