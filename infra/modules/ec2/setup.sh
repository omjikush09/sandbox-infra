mkdir -p ~/firecracker-lab
cd ~/firecracker-lab

#KERNEL
curl -L -o vmlinux.bin \
https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin

ls -lh vmlinux.bin

#rootfs
curl -L -o rootfs.ext4 \
https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/rootfs/bionic.rootfs.ext4

ls -lh rootfs.ext4

sudo ip tuntap add dev tap0 mode tap
sudo ip link set tap0 up
sudo ip addr add 172.16.0.1/24 dev tap0
sudo sysctl -w net.ipv4.ip_forward=1

cat > vm-config.json <<EOF
{
  "boot-source": {
    "kernel_image_path": "./vmlinux.bin",
    "boot_args": "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw"
  },

  "drives": [
    {
      "drive_id": "rootfs",
      "path_on_host": "./rootfs.ext4",
      "is_root_device": true,
      "is_read_only": false
    }
  ],

  "machine-config": {
    "vcpu_count": 1,
    "mem_size_mib": 512
  },

  "network-interfaces": [
    {
      "iface_id": "eth0",
      "host_dev_name": "tap0",
      "guest_mac": "AA:FC:00:00:00:01"
    }
  ]
}
EOF


firecracker --no-api --config-file vm-config.json
#SOKET start
firecracker --api-sock /tmp/firecracker.socket
#VM command
ip addr add 172.16.0.2/24 dev eth0
ip link set eth0 up


