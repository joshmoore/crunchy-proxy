/*
 Copyright 2016 Crunchy Data Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package admin

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/crunchydata/crunchy-proxy/config"
	"github.com/golang/glog"
	"net/http"
)

const DEFAULT_ADMIN_HOST_PORT = "127.0.0.1:10000"

var globalconfig *config.Config

func Initialize(config *config.Config) {
	glog.Infoln("[adminserver] ---- Initializing Admin Server ----")

	if config.AdminHostPort == "" {
		config.AdminHostPort = DEFAULT_ADMIN_HOST_PORT
		glog.Infof("[adminserver] Admin Server host and port is not specified, using default: %s\n",
			DEFAULT_ADMIN_IPADDR)
	}
	glog.V(2).Infoln("adminserver: initializing on " + ipaddr)
	globalconfig = config

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/config", GetConfig},
		&rest.Route{"GET", "/stats", GetStats},
		&rest.Route{"GET", "/stream", StreamEvents},
	)
	if err != nil {
		glog.Fatalln(err)
	}
	api.SetApp(router)

	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))

	http.ListenAndServe(config.AdminIPAddr, nil)
}

func GetConfig(w rest.ResponseWriter, r *rest.Request) {
	glog.V(2).Infoln("adminserver: GetConfig called")

	w.Header().Set("Content-Type", "text/json")
	w.WriteJson(globalconfig)
	glog.V(2).Infoln("adminserver: GetConfig report written")
}

type AdminStatsNode struct {
	HostPort string `json:"ipaddr"`
	Healthy  bool   `json:"healthy"`
	Queries  int    `json:"queries"`
}

type AdminStats struct {
	Nodes []AdminStatsNode `json:"nodes"`
}

func GetStats(w rest.ResponseWriter, r *rest.Request) {
	glog.V(2).Infoln("adminserver: GetStats called")

	stats := AdminStats{}
	stats.Nodes = make([]AdminStatsNode, 1+len(globalconfig.Replicas))
	stats.Nodes[0].HostPort = globalconfig.Master.HostPort
	stats.Nodes[0].Queries = globalconfig.Master.Stats.Queries
	stats.Nodes[0].Healthy = globalconfig.Master.Healthy

	for i := 1; i < len(globalconfig.Replicas)+1; i++ {
		stats.Nodes[i].HostPort = globalconfig.Replicas[i-1].HostPort
		stats.Nodes[i].Queries = globalconfig.Replicas[i-1].Stats.Queries
		stats.Nodes[i].Healthy = globalconfig.Replicas[i-1].Healthy
	}

	w.Header().Set("Content-Type", "text/json")
	w.WriteJson(&stats)
	glog.V(2).Infoln("adminserver: GetStatus report written")
}
