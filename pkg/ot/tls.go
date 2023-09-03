package ot

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"os"
)

func getTls() (*tls.Config, error) {
	clientAuth, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		return nil, errors.Wrap(err, "failed to load client key pair")
	}

	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read ca cert")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{clientAuth},
		RootCAs:      caCertPool,
	}, nil
}
