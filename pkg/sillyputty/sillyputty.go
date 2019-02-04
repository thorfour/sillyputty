package sillyputty

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

const (
	ephemeral = "ephemeral"
	inchannel = "in_channel"
)

// SillyPutty is a plugin server for incoming slack slash commands
type SillyPutty struct {
	Port    int
	Path    string // path to serve from
	Handler func(http.ResponseWriter, *http.Request)

	// TLS options
	allowedHost  string
	supportEmail string
	dataDir      string // data directory to store SSL certs if supported

	mux *mux.Router
}

// response is the json struct for a slack response
type response struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

// Option modify a sillyputty server on creation
type Option func(s *SillyPutty)

// WithTLSOpt adds tls options for a sillyputty server
func WithTLSOpt(host, cacheDir, email string) func(s *SillyPutty) {
	return func(s *SillyPutty) {
		s.allowedHost = host
		s.supportEmail = email
		s.dataDir = cacheDir
	}
}

// PluginHandlerOpt adds a plugin handler to the server
func PluginHandlerOpt(path, root, funcName string) func(s *SillyPutty) {
	return func(s *SillyPutty) {
		s.mux.PathPrefix(path).Handler(http.HandlerFunc(pluginHandler(root, funcName)))
	}
}

// HandlerOpt registers a handler route
func HandlerOpt(path string, f func(url.Values) (string, error)) func(s *SillyPutty) {
	return func(s *SillyPutty) {
		s.mux.PathPrefix(path).Handler(http.HandlerFunc(handler(f)))
	}
}

// New returns a new sillyputty server
func New(path string, opts ...Option) *SillyPutty {
	s := &SillyPutty{
		mux:  mux.NewRouter().PathPrefix(path).Subrouter(),
		Port: 80,
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

// Run starts the silly puttyserver
func (s *SillyPutty) Run() {
	if s.allowedHost != "" {
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(s.allowedHost),
			Cache:      autocert.DirCache(s.dataDir),
			Email:      s.supportEmail,
		}
		srv := &http.Server{
			Handler:      s.mux,
			Addr:         fmt.Sprintf(":%v", s.Port),
			TLSConfig:    m.TLSConfig(),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		go http.ListenAndServe(":80", m.HTTPHandler(nil))
		log.Fatal(srv.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", s.Port), s.mux))
	}
}

// pluginHandler performs the same actions as handler but it first will lookup the function to call via a plugin based on the function name
func pluginHandler(root, funcName string) func(resp http.ResponseWriter, req *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}

		// Use the command as the plugin name i.e /weather will be found as /plugins/weather.so
		p := strings.TrimLeft(req.Form["command"][0], "/")

		// Load the plugin file
		plug, err := plugin.Open(filepath.Join(root, p))
		if err != nil {
			newReponse(resp, "", fmt.Errorf("Plugin not found: %v", err))
			return
		}

		// Lookup the plugin handler
		f, err := plug.Lookup(funcName)
		if err != nil {
			newReponse(resp, "", fmt.Errorf("Command not found: %v", err))
			return
		}

		// Handle the request
		msg, err := f.(func(url.Values) (string, error))(req.Form)
		newReponse(resp, msg, err)
	}
}

// handler is a generic handler wrapper. It takes a function that processes url.Values and responds with a string to print.
// In the event of an error the handler writes the error message as an ephemeral response for slack to print
func handler(f func(url.Values) (string, error)) func(resp http.ResponseWriter, req *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}

		msg, err := f(req.Form)
		newReponse(resp, msg, err)
	}
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
