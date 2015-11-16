package main

import (
	"flag"
	"os"
	"time"
)

// Args will be parsed from the flags
type Args struct {
	mappings   *string
	path       *string
	port       *int
	host       *string
	validFrom  *string
	validFor   *time.Duration
	isCA       *bool
	rsaBits    *int
	ecdsaCurve *string
}

func (a *Args) parseSetup(args []string) {
	a.mappings = flag.String("map", "", "Mappings of hostnames to ports")
	a.path = flag.String("cert_path", os.Getenv("HOME")+"/.gloss", "Path to certification location")
	a.port = flag.Int("port", 4443, "TLS/SSL port to listen on")
	a.host = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	a.validFrom = flag.String("start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	a.validFor = flag.Duration("duration", 365*24*time.Hour, "Duration that certificate is valid for")
	a.isCA = flag.Bool("ca", false, "whether this cert should be its own Certificate Authority")
	a.rsaBits = flag.Int("rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
	a.ecdsaCurve = flag.String("ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256, P384, P521")
	flag.CommandLine.Parse(args)
}
