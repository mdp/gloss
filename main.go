package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	mappings   = flag.String("map", "", "Mappings of hostnames to ports")
	path       = flag.String("cert_path", os.Getenv("HOME")+"/.gloss", "Path to certification location")
	port       = flag.Int("port", 4443, "TLS/SSL port to listen on")
	host       = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	validFrom  = flag.String("start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	validFor   = flag.Duration("duration", 365*24*time.Hour, "Duration that certificate is valid for")
	isCA       = flag.Bool("ca", false, "whether this cert should be its own Certificate Authority")
	rsaBits    = flag.Int("rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
	ecdsaCurve = flag.String("ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256, P384, P521")
)

var hostPortMapping = make(map[string]int)

func multipleHostReverseProxy(hostMapping *map[string]int) *httputil.ReverseProxy {
	// Fairly simple right now:
	// use the subdomain to route to a specific port
	director := func(req *http.Request) {
		domains := strings.Split(req.Host, ".")
		fmt.Printf("%v", req.Host)
		topSubdomain := domains[0]
		port := hostPortMapping["*"]
		if hostPortMapping[topSubdomain] > 0 {
			port = hostPortMapping[topSubdomain]
			fmt.Printf("%v", port)
		}
		req.URL.Scheme = "http"
		req.URL.Host = "localhost:" + strconv.Itoa(port)
	}
	return &httputil.ReverseProxy{Director: director}
}

func generateCertificate() {
	cert := Certificate{
		host:       host,
		path:       path,
		validFrom:  validFrom,
		validFor:   validFor,
		isCA:       isCA,
		rsaBits:    rsaBits,
		ecdsaCurve: ecdsaCurve,
	}

	cert.Generate()
}

func usage(msgType string) {
	switch msgType {
	case "setup":
		fmt.Println("Unable to find keys, make sure you run setup first")
		fmt.Println("e.g.\tgloss setup --host='*.local.dev,local.dev'")
	case "mapping":
		fmt.Println("What ports do you want to map to?")
		fmt.Println("e.g.\tgloss --map '*:3000,someapp:4000")
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		flag.CommandLine.Parse(os.Args[2:])
		fmt.Println("Generating certificates")
		generateCertificate()
		return
	}
	flag.CommandLine.Parse(os.Args[1:])
	if len(*mappings) < 1 {
		usage("mapping")
		return
	}
	for _, mapping := range strings.Split(*mappings, ",") {
		mapping = strings.TrimSpace(mapping)
		s := strings.Split(mapping, ":")
		proxyPort, err := strconv.Atoi(s[1])
		if err != nil {
			log.Fatalf("Mapping error: %s", err)
		}
		if s[0] == "*" {
			fmt.Printf("Mapping * to %d\n", proxyPort)
		} else {
			fmt.Printf("Mapping %s.* to %d\n", s[0], proxyPort)
		}
		hostPortMapping[s[0]] = proxyPort
	}
	cert, err := GetCerts(path)
	if err != nil {
		fmt.Printf("server: loadkeys: %s\n", err)
		usage("setup")
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	service := "0.0.0.0:" + strconv.Itoa(*port)
	fmt.Printf("Listening for SSL on %s\n", service)
	listener, err := tls.Listen("tcp", service, &config)
	proxy := multipleHostReverseProxy(&hostPortMapping)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Forwarded-Proto", "https")
		w.Header().Add("X-Forwarded-For", r.RemoteAddr)
		proxy.ServeHTTP(w, r)
	})
	log.Fatal(http.Serve(listener, nil))
}
