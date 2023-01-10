package httpserver

import (
	"net/http"
	"time"

	"github.com/xmidt-org/arrange/arrangetls"
	"github.com/xmidt-org/httpaux"
	serveraux "github.com/xmidt-org/httpaux/server"
)

type Config struct {
	// ReadTimeout corresponds to http.Server.Addr
	Address string `validate:"empty=false"`

	// Path represents the url path where to locate the metrics listener.
	Path string

	// ReadTimeout corresponds to http.Server.ReadTimeout
	ReadTimeout time.Duration

	// ReadHeaderTimeout corresponds to http.Server.ReadHeaderTimeout
	ReadHeaderTimeout time.Duration

	// WriteTime corresponds to http.Server.WriteTimeout
	WriteTimeout time.Duration

	// IdleTimeout corresponds to http.Server.IdleTimeout
	IdleTimeout time.Duration

	// MaxHeaderBytes corresponds to http.Server.MaxHeaderBytes
	MaxHeaderBytes int

	// KeepAlive corresponds to net.ListenConfig.KeepAlive.  This value is
	// only used for listeners created via Listen.
	KeepAlive time.Duration

	// Header supplies HTTP headers to emit on every response from this server
	Headers http.Header

	// TLS is the optional unmarshaled TLS configuration.  If set, the resulting
	// server will use HTTPS.
	TLS *arrangetls.Config
}

func (c Config) Handler(h http.Handler) (server *http.Server, err error) {
	path := "/"
	if len(c.Path) > 0 {
		path = c.Path
	}

	// This bit converts the headers into the httpaux.Header list then decorates
	// the outgoing headers via a chained http.Handler
	headers := httpaux.NewHeader(c.Headers)
	handler := serveraux.Header(headers.SetTo)(h)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	server = &http.Server{
		Addr:              c.Address,
		Handler:           mux,
		ReadTimeout:       c.ReadTimeout,
		ReadHeaderTimeout: c.ReadHeaderTimeout,
		WriteTimeout:      c.WriteTimeout,
		IdleTimeout:       c.IdleTimeout,
		MaxHeaderBytes:    c.MaxHeaderBytes,
	}

	server.TLSConfig, err = c.TLS.New()

	return
}
