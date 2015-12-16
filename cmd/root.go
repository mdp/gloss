package cmd

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mdp/gloss/certs"
	"github.com/spf13/cobra"
)

// StdLog is our default logger
var StdLog = log.New(os.Stdout, "", 0)

var mappings string
var certPath string
var keyPath string
var sport int
var port int

// Config for the reverse proxy
type Config struct {
	mappings map[string]int
}

func init() {
	RootCmd.Flags().StringVarP(&mappings, "map", "m", "", "Source directory to read from")
	RootCmd.Flags().StringVar(&certPath, "cert", os.Getenv("HOME")+"/.gloss/cert.pem", "Path to cert")
	RootCmd.Flags().StringVar(&keyPath, "key", os.Getenv("HOME")+"/.gloss/key.pem", "Path to cert key")
	RootCmd.Flags().IntVar(&sport, "sport", 4443, "SSL listening port")
	RootCmd.Flags().IntVar(&port, "port", 0, "HTTP listening port")
}

func (c *Config) setupMapping(mappingStr *string) {
	c.mappings = make(map[string]int)
	for _, mapping := range strings.Split(*mappingStr, ",") {
		mapping = strings.TrimSpace(mapping)
		s := strings.Split(mapping, ":")
		proxyPort, err := strconv.Atoi(s[1])
		if err != nil {
			log.Fatalf("Mapping error: %s", err)
		}
		if s[0] == "*" {
			StdLog.Printf("Mapping * to %d\n", proxyPort)
		} else {
			StdLog.Printf("Mapping %s.* to %d\n", s[0], proxyPort)
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
	StdLog.Printf("%s %s %s [%s] \"%s %s %s\" %d %d\n", req.RemoteAddr, req.Host, user, timeFmtd, req.Method, req.RequestURI, req.Proto, res.StatusCode, res.ContentLength)
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

type httpHandler struct {
	proxy httputil.ReverseProxy
}

func (c *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS != nil {
		r.Header.Add("X-Forwarded-Proto", "https")
	}
	r.Header.Add("X-Forwarded-For", r.RemoteAddr)
	c.proxy.ServeHTTP(w, r)
}

func printPortRedirHelp(port int) {
	StdLog.Printf("*Helpful hint on how to redirect port %d -> 443*\n", port)
	if runtime.GOOS == "windows" {
		StdLog.Println("Windows instuctions")
		StdLog.Printf("\tnetsh interface portproxy add v4tov4 connectport=%d listenport=443 connectaddress=127.0.0.1 listenaddress=127.0.0.1\n", port)
	} else if runtime.GOOS == "darwin" {
		StdLog.Println("OSX instuctions")
		StdLog.Printf("\techo \"rdr pass on lo0 inet proto tcp from any to any port 443 -> 127.0.0.1 port %d\" | sudo pfctl -ef -\n", port)
	} else if runtime.GOOS == "linux" {
		StdLog.Println("Linux instuctions")
		StdLog.Printf("\tsudo iptables -t nat -A PREROUTING -p tcp --dport 443 -j REDIRECT --to-port %d", port)
	}
}

// RootCmd is what runs by default
var RootCmd = &cobra.Command{
	Use:   "gloss",
	Short: "gloss is a very simple https reverse proxy",
	Long:  `more information about gloss can be found at https://github.com/mdp/gloss`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cert, err := certs.GetCerts(&certPath, &keyPath)
		if err != nil {
			return errors.New("Unable to find SSL cert, make sure you run setup first\ne.g.\tgloss setup --host='*.local.dev,local.dev'")
		}
		config := Config{}
		if len(mappings) < 1 {
			return errors.New("What ports do you want to map to?\ne.g.\t`gloss --map '*:3000,someapp:4000'`\n")
		}
		config.setupMapping(&mappings)
		tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
		tlsConfig.Rand = rand.Reader
		printPortRedirHelp(sport)
		proxy := multipleHostReverseProxy(&config)
		s := &http.Server{
			Handler:        &httpHandler{proxy: *proxy},
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		if port > 0 {
			httpService := ":" + strconv.Itoa(port)
			httpListener, err := net.Listen("tcp", httpService)
			if err != nil {
				return err
			}
			go s.Serve(httpListener)
		}
		tlsService := ":" + strconv.Itoa(sport)
		tlsListener, err := tls.Listen("tcp", tlsService, &tlsConfig)
		if err != nil {
			return err
		}
		s.Serve(tlsListener)
		return nil
	},
}
