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
var MaxSize = 16 * 1024

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
	case verb == "GET":
		if !requireTOTP(c, reader, totp) {
			return
		}
		buf, err := db.Get(noun)
		if err != nil {
			fmt.Fprintf(c, "UERROR nope\n") // TODO better error msg
		} else {
			fmt.Fprintf(c, "SUCCESS\n")
			c.Write(buf)
		}
	case verb == "ADD":
		var size int
		_, err := fmt.Fscanf(reader, "SIZE %d\n", &size)
		if err != nil || size <= 0 {
			fmt.Fprintf(c, "IERROR expected payload size\n")
			break
		}
		if size >= MaxSize {
			fmt.Fprintf(c, "UERROR payload too big\n")
			break
		}
		buf := make([]byte, size)
		_, err = io.ReadFull(c, buf)
		if err != nil {
			fmt.Fprintf(c, "IERROR partial payload read")
			break
		}
		err = db.Add(noun, buf)
		if err != nil {
			fmt.Fprintf(c, "IERROR couldn't add to database\n")
		} else {
			fmt.Fprintf(c, "SUCCESS\n")
		}
	case verb == "PING":
		fmt.Fprintf(c, "PONG %s\n", noun)
	default:
		fmt.Fprintf(c, "IERROR unknown command\n")
	}
}
