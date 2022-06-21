package proto

import (
	"fmt"
	"log"
	"net"
	"ospm/internal/pkg/otp"
	"ospm/internal/pkg/storage"
)

var Debug = false

func requireTOTP(c net.Conn, totp otp.Ratelimited) bool {
	fmt.Fprintf(c, "NEED TOTP\n")
	if Debug {
		for i := -1; i <= 1; i++ {
			fmt.Fprintf(c, "DEBUG %d\n", totp.Peek(i))
		}
	}

	var pass int
	_, err := fmt.Fscanf(c, "TOTP %d\n", &pass)
	if err != nil {
		fmt.Fprintf(c, "IERROR expected TOTP\n")
		return false
	}
	if valid, msg := totp.Verify(pass); !valid {
		log.Printf("%s failed auth: %s", c.RemoteAddr(), msg)
		fmt.Fprintf(c, "UERROR %s\n", msg)
		return false
	}
	log.Printf("%s passed auth", c.RemoteAddr())
	return true
}

func HandleClient(c net.Conn, totp otp.Ratelimited, db storage.Storage) {
	defer c.Close()
	if Debug {
		fmt.Fprintf(c, "DEBUG i am insecure\n")
	}

	var cmd string
	if _, err := fmt.Fscanf(c, "%s\n", &cmd); err != nil {
		fmt.Fprintf(c, "IERROR ?\n")
		return
	}
	switch {
	case cmd == "LIST":
		for _, file := range db.List() {
			fmt.Fprintf(c, "LIST %s\n", file)
		}
	case Debug && cmd == "TOTP_TEST":
		if !requireTOTP(c, totp) {
			return
		}
		fmt.Fprintf(c, "DEBUG passed auth\n")
	default:
		fmt.Fprintf(c, "IERROR unknown command\n")
	}
}
