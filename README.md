# GLoSS (Go SSL)
Develop on Localhost with SSL

## Why

Your webapp runs on HTTPS, but you develop locally on HTTP. That's not right.

GLoSS is a very simple HTTPS reverse proxy.

Highlights:
- No dependencies. Just one self-contained single executable
- Works on a variety of platforms and architectures: Mac, Linux - Arm7(Raspi2)/Arm6(Raspi)/Amd64/386

## Downloads

[Grab the latest release for your platform - https://github.com/mdp/gloss/releases](https://github.com/mdp/gloss/releases)

## Usage

1. Pick a hostname for your local development
  - `echo "127.0.0.1   local.dev" | sudo tee /etc/hosts > /dev/null`
  - `echo "127.0.0.1   foo.local.dev" | sudo tee /etc/hosts > /dev/null`
1. Setup the certificate
  - `gossl setup --host "local.dev,*.local.dev"`
1. Import the certificate to your keychain (Mac specific instructions below)
  - `open ~/.gloss/cert.pem`
  - Find the GLoSS cert and make it "Trusted"
1. Start using GLoSS
  - `gloss --map "*:3000,foo:4000"` Maps foo.local.dev to port 4000, everything else to 3000
1. Visit https://foo.local.dev:4443
  - Will return the content at localhost:4000 via HTTPS


### Setup redirection from port 443

*Mac*

    echo "rdr pass on lo0 inet proto tcp from any to any port 443 -> 127.0.0.1 port 4443" | sudo pfctl -ef -

## Build from source

`go get github.com/mdp/gloss`

## Notes

- Passes the same headers you'd expect with reverse proxy ssl
  - "X-Forwarded-Proto": "https"
  - "X-Forwarded-For": "the.clients.real.ip"
- Doesn't require trusting a CA cert, only valid signing for the \*.local.dev hostname for example.
- Works on a variety of architectures with zero dependencies thanks to Golang (ARM, x86)

### License: MIT

