// Copyright 2019 Weston Schmidt
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type webHandler struct {
	logic           *Logic
	tempSensors     *TempSensors
	main_page       string
	post_page       string
	ctlRoute        *mux.Router
	ctlEndpoint     *http.Server
	metricsRoute    *mux.Router
	metricsEndpoint *http.Server
}

func (wh *webHandler) handleControl(w http.ResponseWriter, r *http.Request) {
	fan_goal := r.URL.Query().Get("fan_goal_state")
	if "run" == fan_goal {
		fan_duration, err := time.ParseDuration(r.URL.Query().Get("fan_duration"))
		fmt.Printf("Duration: %v\n", fan_duration)
		if nil == err {
			wh.logic.Fan(time.Now().Add(fan_duration))
		}
	}
	preheat := r.URL.Query().Get("preheat_domestic")
	if "preheat" == preheat {
		wh.logic.Preheat()
	}
	heat_goal := r.URL.Query().Get("heat_goal_state")
	if "run" == heat_goal {
		heat_down_duration, err := time.ParseDuration(r.URL.Query().Get("heat_down_duration"))
		if nil == err {
			fmt.Printf("Duration: %v\n", heat_down_duration)
			wh.logic.HeatDownstairs(time.Now().Add(heat_down_duration))
		}
		heat_up_duration, err := time.ParseDuration(r.URL.Query().Get("heat_up_duration"))
		if nil == err {
			fmt.Printf("Duration: %v\n", heat_up_duration)
			wh.logic.HeatUpstairs(time.Now().Add(heat_up_duration))
		}
	}
	if "maintain" == heat_goal {
		if goal, err := strconv.ParseFloat(r.URL.Query().Get("heat_target"), 64); nil == err {
			wh.logic.SetDownstairsTarget(goal)
		}
	}

	buf, _ := ioutil.ReadFile(wh.post_page)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write(buf)
}

func (wh *webHandler) page(w http.ResponseWriter, r *http.Request) {
	buf, _ := ioutil.ReadFile(wh.main_page)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write(buf)
}

func (wh *webHandler) Start() {
	wh.ctlEndpoint = &http.Server{
		Handler:      wh.ctlRoute,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	wh.metricsEndpoint = &http.Server{
		Handler:      wh.metricsRoute,
		Addr:         "127.0.0.1:8001",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go wh.ctlEndpoint.ListenAndServe()
	go wh.metricsEndpoint.ListenAndServe()
}

func (wh *webHandler) Stop() {
	wh.ctlEndpoint.Shutdown(context.Background())
}

func NewWeb(l *Logic, ts *TempSensors, v *viper.Viper) *webHandler {
	wh := &webHandler{
		logic:       l,
		tempSensors: ts,
		main_page:   "index.html",
		post_page:   "post.html",
	}

	wh.ctlRoute = mux.NewRouter()
	wh.ctlRoute.HandleFunc("/", wh.page)
	wh.ctlRoute.HandleFunc("/control", wh.handleControl)

	wh.metricsRoute = mux.NewRouter()
	wh.metricsRoute.Handle("/metrics", promhttp.Handler())

	return wh
}
