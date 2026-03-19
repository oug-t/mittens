#!/bin/bash
set -e

ROOTFS="rootfs.ext4"
MOUNT_DIR="./mnt"

echo "Cleaning up previous build state..."
if mountpoint -q "$MOUNT_DIR"; then
    sudo umount "$MOUNT_DIR"
fi
rm -f "$ROOTFS" 

echo "Building guest listener binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o guest ./cmd/guest/main.go

echo "Building base docker image..."
docker build -t mittens-guest .

echo "Initializing 5GB ext4 filesystem..."
truncate -s 5G "$ROOTFS"
# Allocate higher inode ratio to accommodate heavy node_modules directories
mkfs.ext4 -F -i 4096 "$ROOTFS" >/dev/null 2>&1

echo "Extracting rootfs payload..."
mkdir -p "$MOUNT_DIR"
sudo mount "$ROOTFS" "$MOUNT_DIR"
docker run --rm mittens-guest tar c --exclude=sys --exclude=proc --exclude=dev -C / . | sudo tar x --overwrite -C "$MOUNT_DIR"

sudo mkdir -p "$MOUNT_DIR"/{proc,sys,dev}

echo "Generating guest init script..."
sudo tee "$MOUNT_DIR"/sbin/init > /dev/null << 'EOF'
#!/bin/bash
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

mount -t proc proc /proc
mount -t sysfs sys /sys
mount -t devtmpfs dev /dev

echo "Guest OS booting..." > /dev/console

ip addr add 172.16.0.2/24 dev eth0
ip link set eth0 up
ip route add default via 172.16.0.1
echo "nameserver 8.8.8.8" > /etc/resolv.conf

if [ -f /root/.env ]; then 
    set -a
    source /root/.env
    set +a
fi

echo "Starting vsock listener..." > /dev/console
sleep 1
/usr/local/bin/guest > /dev/console 2>&1

echo "Fatal: Guest listener terminated." > /dev/console
sleep 30
poweroff -f
EOF

sudo chmod +x "$MOUNT_DIR"/sbin/init

echo "Injecting host environment variables..."
if [ -f .host.env ]; then
    sudo cp .host.env "$MOUNT_DIR"/root/.env
else
    echo "Warning: .host.env not found. Guest will boot without API keys." >&2
fi

sudo umount "$MOUNT_DIR"
echo "Image build complete: $ROOTFS"
