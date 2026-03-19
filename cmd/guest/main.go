package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/mdlayher/vsock"
)

func main() {
	listener, err := vsock.Listen(1024, nil)
	if err != nil {
		log.Fatalf("Failed to bind vsock: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err == nil {
			go handleRequest(conn)
		}
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("FATAL: Failed to read from vsock: %v\n", err)))
		return
	}

	prompt := strings.TrimSpace(string(buf[:n]))
	binaryPath := "/usr/bin/openclaw"
	envPath := "/root/.env"

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		conn.Write([]byte(fmt.Sprintf("FATAL: Binary missing at %s\n", binaryPath)))
		return
	}

	// 'yes' automatically answers 'y' to unexpected prompts
	// 'timeout 45s' forces the OS to aggressively kill openclaw if it hangs
	script := fmt.Sprintf(`
		export HOME=/root
		export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
		
		echo "--- PRE-FLIGHT CHECKS ---"
		echo "1. Node Version: $(node -v)"
		
		# Check if the internet bridge is actually working
		if ping -c 1 -W 2 8.8.8.8 >/dev/null 2>&1; then
			echo "2. Internet: CONNECTED"
		else
			echo "2. Internet: DEAD (Failed to reach 8.8.8.8)"
		fi

		# Check if Anthropic's API is reachable
		if curl -I -m 3 https://api.anthropic.com >/dev/null 2>&1; then
			echo "3. Anthropic API: REACHABLE"
		else
			echo "3. Anthropic API: UNREACHABLE (DNS or Firewall blocking)"
		fi

		echo "--- LAUNCHING AGENT ---"
		set -a
		[ -f %s ] && source %s
		set +a
		
		# Use stdbuf to force Node to stream logs instantly without buffering
		yes | timeout 40s stdbuf -oL -eL %s --headless --yes --prompt "$PROMPT" 2>&1
	`, envPath, envPath, binaryPath)

	cmd := exec.Command("/bin/bash", "-c", script)
	cmd.Env = append(os.Environ(), "PROMPT="+prompt)

	output, err := cmd.CombinedOutput()

	if err != nil {
		// If timeout kills it, the exit code is usually 124
		conn.Write([]byte(fmt.Sprintf("=== EXECUTION STOPPED ===\nError: %v\nOutput:\n%s", err, string(output))))
		return
	}

	conn.Write([]byte(fmt.Sprintf("=== AGENT OUTPUT ===\n%s", string(output))))
}
