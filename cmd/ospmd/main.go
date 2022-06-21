package main

import (
	"crypto/tls"
	"log"
	"ospm/internal/pkg/otp"
	"ospm/internal/pkg/proto"
)

var bindTo string = ":7733"

func loadCerts() []tls.Certificate {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
	return []tls.Certificate{cert}
}

func main() {
	proto.Debug = true

	auth := otp.InitRatelimited("TESTTESTTESTTESTTESTTEST")
	log.Printf("listening on %s", bindTo)

	l, err := tls.Listen("tcp", bindTo, &tls.Config{Certificates: loadCerts()})
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go proto.HandleClient(conn, auth)
	}
}
