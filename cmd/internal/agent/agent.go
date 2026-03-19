package agent

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

func SendCommand(prompt string) (string, error) {
	var conn net.Conn
	var err error

	// Wait up to 5 seconds for the microVM to boot
	for i := 0; i < 30; i++ {
		conn, err = net.Dial("unix", "/tmp/v.sock")
		if err == nil {
			if _, err = conn.Write([]byte("CONNECT 1024\n")); err == nil {
				ackBuf := make([]byte, 32)
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				n, err := conn.Read(ackBuf)

				if err == nil && strings.HasPrefix(string(ackBuf[:n]), "OK") {
					conn.SetReadDeadline(time.Time{})
					goto Connected
				}
			}
			conn.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", fmt.Errorf("handshake failed: guest listener never started (EOF)")

Connected:
	defer conn.Close()

	if _, err := conn.Write([]byte(prompt + "\n")); err != nil {
		return "", err
	}

	outBuf, err := io.ReadAll(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(outBuf), nil
}
