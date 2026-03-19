#!/bin/bash
set -e

# Auto-detect default outbound interface
HOST_IFACE=$(ip route list default | awk '/default/ {print $5}' | head -1)

if [ -z "$HOST_IFACE" ]; then
    echo "Error: Could not detect default route interface." >&2
    echo "Please set HOST_IFACE manually in this script." >&2
    exit 1
fi

echo "Cleaning up existing tap0 interface..."
sudo ip link set tap0 down 2>/dev/null || true
sudo ip tuntap del tap0 mode tap 2>/dev/null || true

echo "Configuring tap0 with routing via $HOST_IFACE..."
sudo ip tuntap add tap0 mode tap
sudo nmcli device set tap0 managed no || true
sudo ip addr add 172.16.0.1/24 dev tap0
sudo ip link set tap0 up

sudo sysctl -w net.ipv4.ip_forward=1 >/dev/null

# Configure firewalld (Fedora/RHEL environments)
sudo firewall-cmd --zone=trusted --add-interface=tap0 >/dev/null 2>&1 || true
sudo firewall-cmd --add-masquerade >/dev/null 2>&1 || true

# Flush existing iptables rules for tap0
sudo iptables -t nat -D POSTROUTING -o "$HOST_IFACE" -j MASQUERADE 2>/dev/null || true
sudo iptables -D FORWARD -i tap0 -o "$HOST_IFACE" -j ACCEPT 2>/dev/null || true
sudo iptables -D FORWARD -o tap0 -i "$HOST_IFACE" -j ACCEPT 2>/dev/null || true

# Establish NAT routing
sudo iptables -t nat -A POSTROUTING -o "$HOST_IFACE" -j MASQUERADE
sudo iptables -A FORWARD -i tap0 -o "$HOST_IFACE" -j ACCEPT
sudo iptables -A FORWARD -o tap0 -i "$HOST_IFACE" -j ACCEPT

echo "Network bridge configured: tap0 (172.16.0.1/24) -> $HOST_IFACE"
