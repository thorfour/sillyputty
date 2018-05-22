package main

import (
	"flag"
	"log"
	"os"

	"github.com/thorfour/sillyputty/pkg/sillyputty"
)

var (
	port         = flag.Int("p", 443, "port to serve on")
	debug        = flag.Bool("d", false, "turn TLS off")
	allowedHost  = flag.String("host", "", "ACME allowed FQDN")
	supportEmail = flag.String("email", "", "ACME support email")
)

func main() {
	flag.Parse()
	if *allowedHost == "" && !*debug {
		log.Printf("AllowedHost required for production server")
		os.Exit(1)
	}
	log.Printf("%s", *allowedHost)
	log.Printf("Starting server on port %v", *port)
	s := &sillyputty.SillyPutty{
		AllowedHost:    *allowedHost,
		SupportEmail:   *supportEmail,
		PluginRoot:     "/plugins",
		PluginFuncName: "Handler",
		Path:           "/v1",
		DataDir:        ".",
	}
	s.Run(*port, *debug)
}
