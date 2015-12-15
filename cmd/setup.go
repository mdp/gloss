package cmd

import (
	"os"
	"time"

	"github.com/mdp/gloss/certs"
	"github.com/spf13/cobra"
)

var host string
var path string
var validFrom string
var validFor time.Duration
var isCA bool
var rsaBits int
var ecdsaCurve string

func init() {
	setupCmd.Flags().StringVar(&host, "host", "local.dev,*.local.dev", "Comma-separated hostnames and IPs to generate a certificate for")
	setupCmd.Flags().StringVar(&path, "path", os.Getenv("HOME")+"/.gloss", "Comma-separated hostnames and IPs to generate a certificate for")
	setupCmd.Flags().StringVar(&validFrom, "start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	setupCmd.Flags().DurationVar(&validFor, "duration", 365*24*time.Hour, "Duration that certificate is valid for")
	setupCmd.Flags().BoolVar(&isCA, "ca", false, "whether this cert should be its own Certificate Authority")
	setupCmd.Flags().IntVar(&rsaBits, "rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
	setupCmd.Flags().StringVar(&ecdsaCurve, "ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256, P384, P521")
	RootCmd.AddCommand(setupCmd)
}

func generateCertificate() {
	c := certs.Certificate{
		Host:       &host,
		Path:       &path,
		ValidFrom:  &validFrom,
		ValidFor:   &validFor,
		IsCA:       &isCA,
		RsaBits:    &rsaBits,
		EcdsaCurve: &ecdsaCurve,
	}
	c.Generate()
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "create certificates for gloss",
	Long:  `create certificates for use with the gloss proxy`,
	Run: func(cmd *cobra.Command, args []string) {
		generateCertificate()
		return
	},
}
