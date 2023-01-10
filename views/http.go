// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package views

import (
	_ "embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/schmidtw/heaticus-maximus/views/assets"
)

func Handler() (http.Handler, error) {
	mux := http.NewServeMux()

	jsFS, _ := fs.Sub(assets.SiteJS, "site/js")
	cssFS, _ := fs.Sub(assets.SiteCSS, "site/css")

	mux.HandleFunc("/", indexHandler)
	mux.Handle("/css", http.FileServer(http.FS(cssFS)))
	mux.Handle("/js/", http.FileServer(http.FS(jsFS)))
	//mux.HandleFunc("/chart", getChartData())
	//mux.HandleFunc("/control", handleControl())

	return mux, nil
}

//go:embed index.gohtml
var Index string

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Path: '%s'\n", r.URL.Path)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(Index))
}
