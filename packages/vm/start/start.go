package start

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/omjikush09/sandboxing-infra/packages/vm/client"
)

// type Start interface {

// }
type VM struct {
	Id         string
	Name       string
	SocketPath string
	TapName    string // This is cable
	ipAddress  string
	HostTapIP  string
	GuestIP    string
	HostIP     string
	KernelPath string
	BaseRootfs string
	RootfsPath string
	MAC        string
}

func run(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

func newVM(id int) VM {

	name := fmt.Sprintf("vm%d", id)
	hostIP := fmt.Sprintf("172.16.%d.1", id)
	guestIP := fmt.Sprintf("172.16.%d.2", id)

	vm := VM{
		Id:         strconv.Itoa(id),
		SocketPath: fmt.Sprintf("/tmp/%s.socket", name),
		Name:       name,
		TapName:    fmt.Sprintf("tap-%s", name),
		GuestIP:    guestIP,
		HostIP:     hostIP,
		MAC:        fmt.Sprintf("AA:FC:00:00:00:%02X", id),
		KernelPath: "/home/ubuntu/firecracker-lab/vmlinux.bin",
		BaseRootfs: "/home/ubuntu/firecracker-lab/rootfs.ext4",
		RootfsPath: fmt.Sprintf("/home/ubuntu/firecracker-lab/%s-rootfs.ext4", name),
	}
	return vm
}

func (vm *VM) prepareRootfs() error {

	if _, err := os.Stat(vm.RootfsPath); err == nil {
		return nil
	}

	return run("cp", vm.BaseRootfs, vm.RootfsPath)

}

func (vm *VM) SetUpNework() error {
	//create tap device
	if err := run("sudo", "ip", "tuntap", "add", vm.TapName, "mode", "tap"); err != nil {
		return err
	}

	// assign host-side IP to tap
	if err := run("sudo", "ip", "addr", "add", vm.HostIP+"/24", "dev", vm.TapName); err != nil {
		return err
	}

	// bring tap up
	if err := run("sudo", "ip", "link", "set", vm.TapName, "up"); err != nil {
		return err
	}

	// enable forwarding
	// if err := run("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
	// 	return err
	// }

	// NAT traffic from VM subnet to host internet interface

	// if err := run("sudo", "iptables", "-t", "nat", "-A", "POSTROUTING",
	// 	"-s", "172.16.0.0/24", "-o", "eth0", "-j", "MASQUERADE"); err != nil {
	// 	return err
	// }
	// if err := run("sudo", "iptables", "-A", "FORWARD",
	// 	"-i", vm.TapName, "-o", "eth0", "-j", "ACCEPT"); err != nil {
	// 	return err
	// }

	// if err := run("sudo", "iptables", "-A", "FORWARD",
	// 	"-i", "eth0", "-o", vm.TapName,
	// 	"-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
	// 	return err
	// }

	return nil
}

func StartVM() {

}

func (vm *VM) startFirecraker() (*exec.Cmd, error) {
	_ = os.Remove(vm.SocketPath)

	cmd := exec.Command("firecracker", "--api-sock", vm.SocketPath)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil

}

func (vm *VM) waitForSocket(timeout time.Duration) error {
	deadline := time.Now().Add((timeout))

	for time.Now().Before(deadline) {
		if _, err := os.Stat(vm.SocketPath); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("socket not created: %s", vm.SocketPath)
}

func put(client *http.Client, path string, body string) error {
	req, err := http.NewRequest(http.MethodPut, "http://localhost"+path, bytes.NewBufferString(body))

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.Status < "200" || resp.Status >= "300" {
		return fmt.Errorf("Put is failed for path %s with  status %s", path, resp.Status)
	}

	return nil

}

func (vm *VM) ConfigVm() error {

	client := client.FirecrakerClient(vm.SocketPath)
	if err := put(client, "/machine-config", `{
			"vcpu_count": 1,
			"mem_size_mib": 512,
			"smt": false
		}`); err != nil {
		return err
	}

	bootArgs := fmt.Sprintf(
		"console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw ip=%s::%s:255.255.255.0::eth0:off",
		vm.GuestIP,
		vm.HostIP,
	)

	if err := put(client, "/boot-source", fmt.Sprintf(`{
			"kernel_image_path": "%s",
			"boot_args": "%s"
		}`, vm.KernelPath, bootArgs)); err != nil {
		return err
	}

	if err := put(client, "/drives/rootfs", fmt.Sprintf(`{
			"drive_id": "rootfs",
			"path_on_host": "%s",
			"is_root_device": true,
			"is_read_only": false
		}`, vm.RootfsPath)); err != nil {
		return err
	}
	if err := put(client, "/network-interfaces/eth0", fmt.Sprintf(`{
			"iface_id": "eth0",
			"host_dev_name": "%s",
			"guest_mac": "%s"
		}`, vm.TapName, vm.MAC)); err != nil {
		return err
	}
	if err := put(client, "/actions", `{
			"action_type": "InstanceStart"
		}`); err != nil {
		return err
	}
	return nil

}

var count atomic.Int32

func CreateVm() (*VM, *exec.Cmd, error) {
	if int(count.Load()) >= 254 {
		return nil, nil, fmt.Errorf("We are out of capacity")
	}
	counter := count.Add(1)
	vm := newVM(int(counter))
	log.Printf("creating vm name=%s guest_ip=%s host_ip=%s socket=%s tap=%s rootfs=%s", vm.Name, vm.GuestIP, vm.HostIP, vm.SocketPath, vm.TapName, vm.RootfsPath)

	if err := vm.prepareRootfs(); err != nil {
		log.Printf("failed to prepare rootfs for vm=%s rootfs=%s base_rootfs=%s err=%v", vm.Name, vm.RootfsPath, vm.BaseRootfs, err)
		return nil, nil, fmt.Errorf("prepare rootfs: %w", err)
	}
	log.Printf("prepared rootfs for vm=%s rootfs=%s", vm.Name, vm.RootfsPath)

	if err := vm.SetUpNework(); err != nil {
		log.Printf("failed to setup network for vm=%s tap=%s host_ip=%s err=%v", vm.Name, vm.TapName, vm.HostIP, err)
		return nil, nil, fmt.Errorf("setup network: %w", err)
	}
	log.Printf("configured network for vm=%s tap=%s host_ip=%s", vm.Name, vm.TapName, vm.HostIP)

	cmd, err := vm.startFirecraker()

	if err != nil {
		log.Printf("failed to start firecracker for vm=%s socket=%s err=%v", vm.Name, vm.SocketPath, err)
		return nil, nil, fmt.Errorf("start firecracker: %w", err)
	}
	log.Printf("started firecracker process for vm=%s pid=%d socket=%s", vm.Name, cmd.Process.Pid, vm.SocketPath)

	if err := vm.waitForSocket(2 * time.Second); err != nil {
		log.Printf("failed waiting for firecracker socket for vm=%s socket=%s err=%v", vm.Name, vm.SocketPath, err)
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return nil, cmd, fmt.Errorf("wait for socket: %w", err)
	}
	log.Printf("firecracker socket ready for vm=%s socket=%s", vm.Name, vm.SocketPath)

	if err := vm.ConfigVm(); err != nil {
		log.Printf("failed to configure/start firecracker vm=%s socket=%s tap=%s err=%v", vm.Name, vm.SocketPath, vm.TapName, err)
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return nil, cmd, fmt.Errorf("configure vm: %w", err)
	}

	log.Printf("started vm=%s guest_ip=%s socket=%s tap=%s", vm.Name, vm.GuestIP, vm.SocketPath, vm.TapName)

	time.Sleep(100 * time.Millisecond)
	return &vm, cmd, nil

}
