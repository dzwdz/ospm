package main

import (
	"crypto/tls"
	"flag"
	"log"
	"ospm/internal/pkg/otp"
	"ospm/internal/pkg/proto"
	"ospm/internal/pkg/storage"
)

var bindTo string
var certPath, keyPath string

func loadCerts() []tls.Certificate {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatal(err)
	}
	return []tls.Certificate{cert}
}

func main() {
	flag.StringVar(&bindTo, "bind", ":7733", "Address to listen on.")
	flag.StringVar(&certPath, "cert", "./cert.pem", "")
	flag.StringVar(&keyPath, "key", "./key.pem", "")
	flag.BoolVar(&proto.Debug, "debug-insecure", false, "Don't.")
	flag.Parse()

	db := storage.Init("./data")

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
		go proto.HandleClient(conn, auth, db)
	}
}
