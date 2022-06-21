package proto

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"ospm/internal/pkg/otp"
	"ospm/internal/pkg/storage"
	"strings"
)

var Debug = false

func requireTOTP(c net.Conn, reader io.Reader, totp otp.Ratelimited) bool {
	fmt.Fprintf(c, "NEED TOTP\n")
	if Debug {
		for i := -1; i <= 1; i++ {
			fmt.Fprintf(c, "DEBUG %d\n", totp.Peek(i))
		}
	}

	var pass int
	_, err := fmt.Fscanf(reader, "TOTP %d\n", &pass)
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

	reader := bufio.NewReader(c)
	cmd, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(c, "IERROR ?\n")
		return
	}
	cmd = strings.TrimSpace(cmd)
	verb, noun, _ := strings.Cut(cmd, " ")

	switch {
	case verb == "LIST":
		for _, file := range db.List() {
			fmt.Fprintf(c, "LIST %s\n", file)
		}
	case verb == "PING":
		fmt.Fprintf(c, "PONG %s\n", noun)
	case Debug && verb == "TOTP_TEST":
		if !requireTOTP(c, reader, totp) {
			return
		}
		fmt.Fprintf(c, "DEBUG passed auth\n")
	default:
		fmt.Fprintf(c, "IERROR unknown command\n")
	}
}
