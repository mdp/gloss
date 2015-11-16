package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var stdLog = log.New(os.Stdout, "", 0)

// Config for the reverse proxy
type Config struct {
	mappings map[string]int
}

func (c *Config) setupMapping(mappings *string) {
	c.mappings = make(map[string]int)
	for _, mapping := range strings.Split(*mappings, ",") {
		mapping = strings.TrimSpace(mapping)
		s := strings.Split(mapping, ":")
		proxyPort, err := strconv.Atoi(s[1])
		if err != nil {
			log.Fatalf("Mapping error: %s", err)
		}
		if s[0] == "*" {
			stdLog.Printf("Mapping * to %d\n", proxyPort)
		} else {
			stdLog.Printf("Mapping %s.* to %d\n", s[0], proxyPort)
		}
		c.mappings[s[0]] = proxyPort
	}
}

type upstreamTransport struct {
	config *Config
}

func (trans *upstreamTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		errorMsg := fmt.Sprintf("Gloss proxy error: %v", err)
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       ioutil.NopCloser(strings.NewReader(errorMsg)),
		}, nil
	}
	user, _, ok := req.BasicAuth()
	if !ok {
		user = "-"
	}
	timeFmtd := time.Now().Format("02/Jan/2006 03:04:05")
	stdLog.Printf("%s %s %s [%s] \"%s %s %s\" %d %d\n", req.RemoteAddr, req.Host, user, timeFmtd, req.Method, req.RequestURI, req.Proto, res.StatusCode, res.ContentLength)
	return res, err
}

func multipleHostReverseProxy(config *Config) *httputil.ReverseProxy {
	// Fairly simple right now:
	// use the subdomain to route to a specific port
	director := func(req *http.Request) {
		domains := strings.Split(req.Host, ".")
		topSubdomain := domains[0]
		port := config.mappings["*"]
		if config.mappings[topSubdomain] > 0 {
			port = config.mappings[topSubdomain]
		}
		req.URL.Scheme = "http"
		req.URL.Host = "localhost:" + strconv.Itoa(port)
	}
	return &httputil.ReverseProxy{Director: director, Transport: &upstreamTransport{config: config}}
}

func generateCertificate(c *Args) {
	cert := Certificate{
		host:       c.host,
		path:       c.path,
		validFrom:  c.validFrom,
		validFor:   c.validFor,
		isCA:       c.isCA,
		rsaBits:    c.rsaBits,
		ecdsaCurve: c.ecdsaCurve,
	}
	cert.Generate()
}

func printPortRedirHelp(port int) {
	stdLog.Printf("*Helpful hint on how to redirect port %d -> 443*\n", port)
	if runtime.GOOS == "windows" {
		stdLog.Printf("\tnetsh interface portproxy add v4tov4 connectport=%d listenport=443 connectaddress=127.0.0.1 listenaddress=127.0.0.1\n", port)
	} else if runtime.GOOS == "darwin" {
		stdLog.Printf("\techo \"rdr pass on lo0 inet proto tcp from any to any port 443 -> 127.0.0.1 port %d\" | sudo pfctl -ef -\n", port)
	}
}

func usage(msgType string) {
	switch msgType {
	case "setup":
		stdLog.Println("Unable to find keys, make sure you run setup first")
		stdLog.Println("e.g.\tgloss setup --host='*.local.dev,local.dev'")
	case "mapping":
		stdLog.Println("What ports do you want to map to?")
		stdLog.Println("e.g.\tgloss --map '*:3000,someapp:4000")
	}
}

func main() {
	args := Args{}
	config := Config{}
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		args.parseSetup(os.Args[2:])
		generateCertificate(&args)
		return
	}
	args.parseSetup(os.Args[1:])
	if len(*args.mappings) < 1 {
		usage("mapping")
		return
	}
	config.setupMapping(args.mappings)
	cert, err := GetCerts(args.path)
	if err != nil {
		stdLog.Printf("server: loadkeys: %s\n", err)
		usage("setup")
	}
	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
	tlsConfig.Rand = rand.Reader
	service := "0.0.0.0:" + strconv.Itoa(*args.port)
	stdLog.Printf("Listening for SSL on %s\n", service)
	printPortRedirHelp(*args.port)
	listener, err := tls.Listen("tcp", service, &tlsConfig)
	proxy := multipleHostReverseProxy(&config)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Forwarded-Proto", "https")
		w.Header().Add("X-Forwarded-For", r.RemoteAddr)
		proxy.ServeHTTP(w, r)
	})
	log.Fatal(http.Serve(listener, nil))
}
