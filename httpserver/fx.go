// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package httpserver

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New(lc fx.Lifecycle, handler http.Handler, cfg Config, log *zap.Logger) (*http.Server, error) {
	srv, err := cfg.Handler(handler)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lc := net.ListenConfig{
				KeepAlive: cfg.KeepAlive,
			}
			ln, err := lc.Listen(ctx, "tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("Starting HTTP server", zap.String("addr", srv.Addr))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping HTTP server", zap.String("addr", srv.Addr))
			return srv.Shutdown(ctx)
		},
	})
	return srv, nil
}
