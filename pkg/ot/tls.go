package ot

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
)

func getTls() (*tls.Config, error) {
	cwd, _ := os.Getwd()
	fmt.Println("checking certs folder in working directory:", cwd)
	if _, err := os.Stat("./certs"); os.IsNotExist(err) {
		log.Fatal("certs folder does not exist!")
	}

	clientAuth, err := tls.LoadX509KeyPair("./certs/client.crt", "./certs/client.key")
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
