package sillyputty

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"plugin"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

const (
	ephemeral = "ephemeral"
	inchannel = "in_channel"
)

// SillyPutty is a plugin server for incoming slack slash commands
type SillyPutty struct {
	AllowedHost    string
	SupportEmail   string
	PluginRoot     string
	PluginFuncName string
	Path           string // path to server from
	DataDir        string // data directory to store SSL certs
}

// response is the json struct for a slack response
type response struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

// Run starts the silly puttyserver
func (s *SillyPutty) Run(p int, d bool) {
	if d {
		http.HandleFunc(s.Path, s.handler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", p), nil))
	} else {
		mux := &http.ServeMux{}
		mux.HandleFunc(s.Path, s.handler)
		hostPolicy := func(ctx context.Context, host string) error {
			if host == s.AllowedHost {
				return nil
			}
			return fmt.Errorf("acme/autocert: only %s allowed", s.AllowedHost)
		}
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache(s.DataDir),
			Email:      s.SupportEmail,
		}
		srv := &http.Server{
			Handler: mux,
			Addr:    fmt.Sprintf(":%v", p),
			TLSConfig: &tls.Config{
				GetCertificate: m.GetCertificate,
			},
		}
		go http.ListenAndServe(":80", m.HTTPHandler(nil))
		log.Fatal(srv.ListenAndServeTLS("", ""))
	}
}

func (s *SillyPutty) handler(resp http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	// Use the command as the plugin name i.e /weather will be found as /plugins/weather.so
	p := strings.TrimLeft(req.Form["command"][0], "/")

	// Load the plugin file
	plug, err := plugin.Open(s.getPlugin(p))
	if err != nil {
		newReponse(resp, "", fmt.Errorf("Command not found"))
		return
	}

	// Lookup the plugin handler
	f, err := plug.Lookup(s.PluginFuncName)
	if err != nil {
		newReponse(resp, "", fmt.Errorf("Command not found"))
		return
	}

	// Handle the request
	msg, err := f.(func(url.Values) (string, error))(req.Form)
	newReponse(resp, msg, err)
}

func (s *SillyPutty) getPlugin(p string) string {
	return filepath.Join(s.PluginRoot, p)
}

func newReponse(resp http.ResponseWriter, message string, err error) {
	r := &response{
		ResponseType: inchannel,
		Text:         message,
	}

	// Switch to an ephemeral message
	if err != nil {
		r.ResponseType = ephemeral
		r.Text = err.Error()
	}

	b, err := json.Marshal(r)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.Write(b)
	return
}

func (s *SillyPutty) hostPolicy(ctx context.Context, host string) error {
	if host == s.AllowedHost {
		return nil
	}

	return fmt.Errorf("acme/autocert: only %s hist is allowed", s.AllowedHost)
}
