// Copyright (c) 2022-2023, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package web_site

import (
	"html"
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sync"

	"github.com/simonlingoogle/go-simplelogger"
)

var httpServer *http.Server = nil
var canServe bool = true
var httpServerMutex sync.Mutex
var Started = make(chan struct{})

func Serve(listenAddr string) error {
	defer simplelogger.Debugf("webserver exit.")

	assetDir := os.Getenv("HOME")
	if assetDir == "" {
		assetDir = "/tmp"
	}
	assetDir = filepath.Join(assetDir, ".otns-web")
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return err
	}

	for _, name := range AssetNames() {
		data, err := Asset(name)
		if err != nil {
			return err
		}

		fp := filepath.Join(assetDir, name)
		if err := os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
			return err
		}

		f, err := os.OpenFile(fp, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			return err
		}

		if _, err := f.Write(data); err != nil {
			return err
		}
	}

	templates := template.Must(template.ParseGlob(filepath.Join(assetDir, "templates", "*.html")))

	fs := http.FileServer(http.Dir(filepath.Join(assetDir, "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/visualize", func(writer http.ResponseWriter, request *http.Request) {
		addr := html.EscapeString(request.URL.Query()["addr"][0])
		simplelogger.Debugf("visualizing addr=%+v", addr)
		err := templates.ExecuteTemplate(writer, "visualize.html", map[string]interface{}{
			"addr": addr,
		})
		if err != nil {
			writer.WriteHeader(501)
		}
	})

	http.HandleFunc("/energyViewer", func(writer http.ResponseWriter, request *http.Request) {
		addr := html.EscapeString(request.URL.Query()["addr"][0])
		simplelogger.Debugf("energy charts visualizing addr=%+v", addr)
		err := templates.ExecuteTemplate(writer, "energyViewer.html", map[string]interface{}{
			"addr": addr,
		})
		if err != nil {
			writer.WriteHeader(501)
		}
	})

	httpServerMutex.Lock()
	if !canServe {
		httpServer = nil
		httpServerMutex.Unlock()
		close(Started)
		return http.ErrServerClosed
	}
	httpServer = &http.Server{Addr: listenAddr, Handler: nil}
	simplelogger.Infof("OTNS webserver now serving on %s ...", listenAddr)
	defer simplelogger.Debugf("webserver: httpServer.ListenAndServe() done")
	httpServerMutex.Unlock()
	close(Started)
	return httpServer.ListenAndServe()
}

func StopServe() {
	simplelogger.Debugf("requesting webserver to exit ...")
	httpServerMutex.Lock()
	if httpServer != nil {
		_ = httpServer.Close()
	}
	canServe = false // prevent serving again in same execution.
	httpServerMutex.Unlock()
}
