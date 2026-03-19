# Mittens

I love exploring new AI tools and agents, but I am extremely hesitant to give utilities like `openclaw` unrestricted access to my host machine. I built Mittens as a secure, local sandbox to run these agents safely.

Mittens isolates headless AI agents inside a Firecracker microVM, controlled via a Go-based Bubble Tea Terminal User Interface (TUI).

**Status: Work in Progress / Help Wanted**
The core infrastructure (VM lifecycle, TUI, Vsock communication, automated rootfs building) is fully functional. However, I have hit a wall with the guest Node.js process hanging. I don't have the time to continue debugging the KVM/Node edge cases right now, but if this architecture interests you, feel free to take the code, fork it, and build something great.

---

## Architecture

1. **Host TUI (`mittens_cli`)**: A Go application using Charmbracelet's Bubble Tea. It manages the Firecracker process, renders the UI, and sends user prompts via a UNIX socket.
2. **Vsock Bridge**: Facilitates secure, local communication between the host and the guest OS without exposing SSH.
3. **Guest Listener (`guest`)**: A static Go binary running inside the microVM. It listens on the vsock, receives the prompt, and executes the Node.js agent.
4. **Image Builder (`setup_image.sh`)**: Automates the creation of an `ext4` filesystem by building a Docker image, extracting its contents, injecting API keys, and writing initialization scripts.

## Features

* **Automated OS Building**: `setup_image.sh` seamlessly converts a Dockerfile into a bootable `rootfs.ext4` drive in seconds.
* **Dynamic Networking**: `setup_network.sh` configures `tap0`, handles NAT routing, and bypasses aggressive host firewalls (like Fedora's `firewalld`).
* **Asynchronous TUI**: The UI remains responsive while the microVM boots and processes data in the background.
* **Fast Boot Sequence**: Custom `/sbin/init` script mounts required filesystems, configures `eth0`, and starts the vsock listener in under 2 seconds.

## Known Blockers

The microVM boots, configures the network, and receives prompts over vsock perfectly. However, the `openclaw` Node.js process hangs indefinitely when executed by the guest listener.

* **Symptoms**: The guest Go script triggers a 45-second timeout kill (`exit status 124`) because the Node process never returns output.
* **Suspected Causes**: A subtle DNS/routing failure bridging `tap0` to the internet causing Anthropic API calls to silently hang, or Node.js attempting to write cache to a read-only/missing directory inside the minimal `ext4` filesystem.

If you know how to trace headless Node.js processes inside minimal Linux environments, PRs are welcome! Otherwise, please use this repository as a boilerplate for your own Go/Firecracker projects.

## Getting Started

### Prerequisites
* Linux Host (Tested on Fedora/Ubuntu)
* KVM enabled (`/dev/kvm`)
* Docker
* Go 1.21+
* Firecracker binary installed and in your `$PATH`

### 1. Build the MicroVM Image
Compiles the guest listener, builds the base container, creates a 5GB `ext4` disk, and injects your `.host.env` file (for API keys).
```bash
./setup_image.sh
