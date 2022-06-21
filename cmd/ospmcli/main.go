package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"
)

func appendDefaultPort(addr string, port string) string {
	if strings.Index(addr, ":") > -1 {
		return addr
	}
	return addr + ":" + port
}

func main() {
	tlsInsecure := false

	flag.BoolVar(&tlsInsecure, "tls-insecure", false, "Don't check the server certificate.")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ospmcli server[:port] ...")
		os.Exit(1)
	}

	addr := appendDefaultPort(flag.Arg(0), "7733")
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: tlsInsecure,
	})
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	reader := bufio.NewReader(conn)
	defer conn.Close()

	verb := strings.ToLower(flag.Arg(1))
	switch {
	case verb == "list":
		fmt.Fprintf(conn, "LIST\n")
		// TODO error checking
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimPrefix(line, "LIST ")
			fmt.Fprintf(os.Stdout, "%s", line)
		}
	case verb == "get":
		fmt.Fprintf(conn, "GET %s\n", flag.Arg(2))
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}
			switch {
			case line == "NEED TOTP\n":
				fmt.Fprintf(os.Stderr, "totp: ")
				var code string
				fmt.Scan(&code)
				fmt.Fprintf(conn, "TOTP %s\n", code)
			case line == "SUCCESS\n":
				reader.WriteTo(os.Stdout)
				// TODO get rid of the weird EOF string at the end
			// TODO parse UERROR
			default:
				fmt.Fprintf(os.Stderr, "unrecognized server message\n\t%s\n", line)
				os.Exit(1)
			}
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command. available: list, get")
		os.Exit(1)
	}
}
