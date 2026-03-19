# 🧤 Mittens

A secure, local Firecracker microVM sandbox to safely run headless AI agents (like `openclaw`) via a Go-based Bubble Tea TUI.

**Status: Help Wanted 🚧**
The core infrastructure (VM lifecycle, TUI, Vsock, automated rootfs building) is fully functional. However, the guest Node.js process is currently hanging. If this architecture interests you, feel free to fork it and build something great!

## Architecture & Features

* **Host TUI (`mittens_cli`)**: Go/Bubble Tea interface managing the Firecracker lifecycle asynchronously.
* **Vsock Bridge**: Secure, SSH-less host-to-guest communication.
* **Automated OS Build**: `setup_image.sh` seamlessly converts a Dockerfile into a bootable 5GB `rootfs.ext4` drive in seconds.
* **Dynamic Networking**: `setup_network.sh` auto-configures `tap0`, NAT routing, and bypasses host firewalls.

## Known Blocker

The VM boots and receives prompts perfectly, but the `openclaw` Node.js process hangs indefinitely until the host's 45s timeout (`exit status 124`) kills it.

* **Suspected Causes**: A subtle DNS/routing failure bridging `tap0` to the internet, or Node.js struggling with the minimal `ext4` filesystem.

PRs are welcome! Otherwise, please use this repo as a boilerplate for your own Go/Firecracker projects.

## Quick Start

**Prerequisites**: Linux Host, KVM (`/dev/kvm`), Docker, Go 1.21+, Firecracker binary.

```bash
# 1. Build the microVM image and inject API keys (.host.env)
./setup_image.sh

# 2. Configure the network bridge and firewall rules
sudo bash setup_network.sh

# 3. Compile and launch the sandbox
```bash
go build -o mittens_cli ./cmd/mittens
sudo ./mittens_cli
```
Controls: s (Start) | k (Kill) | i (Insert prompt) | esc (Normal Mode) | q (Quit)
