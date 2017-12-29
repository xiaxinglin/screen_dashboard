// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/screen_dashboard/src/backend/handler"
	"github.com/screen_dashboard/src/backend/ini"
	"github.com/spf13/pflag"
)

var (
	argInsecurePort        = pflag.Int("insecure-port", 9090, "The port to listen to for incoming HTTP requests.")
	argInsecureBindAddress = pflag.IP("insecure-bind-address", net.IPv4(0, 0, 0, 0), "The IP address on which to serve the --port (set to 0.0.0.0 for all interfaces).")
)

func main() {
	// Set logging output to standard console out
	log.SetOutput(os.Stdout)

	iniPortal, err := iniparser.LoadFile("./portal.ini", "utf-8")
	if err != nil {
		//
		print(err)
	}
	port, ok := iniPortal.GetString("monitor", "port")
	if !ok {
		//
	}

	iport := *argInsecurePort
	iport, err = strconv.Atoi(port)
	if err != nil {
		panic(err)
	}


	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	flag.CommandLine.Parse(make([]string, 0)) // Init for glog calls in kubernetes packages

	apiHandler, err := handler.CreateHTTPAPIHandler()
	if err != nil {
		//
	}
	// Run a HTTP server that serves static public files from './public' and handles API calls.
	// TODO(bryk): Disable directory listing.
	http.Handle("/", handler.MakeGzipHandler(handler.CreateLocaleHandler()))
	http.Handle("/monitor_api/", apiHandler)
	// TODO(maciaszczykm): Move to /appConfig.json as it was discussed in #640.
	http.Handle("/monitor_api/appConfig.json", handler.AppHandler(handler.ConfigHandler))
	//http.Handle("/api/sockjs/", handler.CreateAttachHandler("/api/sockjs"))
	//http.Handle("/metrics", prometheus.Handler())

	// Listen for http or https
	log.Printf("Serving insecurely on HTTP port: %d", iport)
	addr := fmt.Sprintf("%s:%d", *argInsecureBindAddress, iport)
	go func() { log.Fatal(http.ListenAndServe(addr, nil)) }()

	select {}
}
