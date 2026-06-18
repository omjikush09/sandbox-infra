package start

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/omjikush09/sandboxing-infra/packages/vm/firecraker"
	"github.com/omjikush09/sandboxing-infra/packages/vm/ippool"
	"github.com/omjikush09/sandboxing-infra/packages/vm/utils"
)

// type Start interface {

// }
type VM struct {
	Id         string
	Name       string
	SocketPath string
	LogPath    string
	TapName    string // This is cable
	ipAddress  string
	HostTapIP  string
	GuestIP    string
	HostIP     string
	KernelPath string
	BaseRootfs string
	RootfsPath string
	MAC        string
	Cmd        *exec.Cmd
	ipLease    *ippool.IP
}

func newVM(id int, ip *ippool.IP) VM {

	name := fmt.Sprintf("vm%d", id)
	hostIP := ip.HostIP
	guestIP := ip.VmIP

	vm := VM{
		Id:         strconv.Itoa(id),
		SocketPath: fmt.Sprintf("/tmp/%s.socket", name),
		LogPath:    fmt.Sprintf("/tmp/%s.log", name),
		Name:       name,
		TapName:    fmt.Sprintf("tap-%s", name),
		GuestIP:    guestIP,
		HostIP:     hostIP,
		MAC:        fmt.Sprintf("AA:FC:00:00:00:%02X", id),
		KernelPath: "/home/ubuntu/firecracker-lab/vmlinux.bin",
		BaseRootfs: "/home/ubuntu/firecracker-lab/rootfs.ext4",
		RootfsPath: fmt.Sprintf("/home/ubuntu/firecracker-lab/%s-rootfs.ext4", name),
		ipLease:    ip,
	}
	return vm
}

func (vm *VM) prepareRootfs() error {

	if _, err := os.Stat(vm.RootfsPath); err == nil {
		if err := os.Remove(vm.RootfsPath); err != nil {
			return err
		}
	}

	return utils.Run("cp", vm.BaseRootfs, vm.RootfsPath)

}

func tapDeviceExists(name string) (bool, error) {
	cmd := exec.Command("ip", "link", "show", "dev", name)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (vm *VM) SetUpNetwork() error {
	//create tap device
	if err := utils.Run("sudo", "ip", "tuntap", "add", vm.TapName, "mode", "tap"); err != nil {
		return err
	}

	// assign host-side IP to tap
	if err := utils.Run("sudo", "ip", "addr", "add", vm.HostIP+"/24", "dev", vm.TapName); err != nil {
		_ = vm.deleteTapDevice()
		return err
	}

	// bring tap up
	if err := utils.Run("sudo", "ip", "link", "set", vm.TapName, "up"); err != nil {
		_ = vm.deleteTapDevice()
		return err
	}

	// enable forwarding
	if err := utils.Run("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}

	// NAT traffic from VM subnet to host internet interface

	if err := utils.Run("sudo", "iptables", "-t", "nat", "-A", "POSTROUTING",
		"-s", "172.16.0.0/16", "-o", "eth0", "-j", "MASQUERADE"); err != nil { // 172.16.0.0/16 it is broder, so that all the vm can use this
		return err
	}

	if err := utils.Run("sudo", "iptables", "-A", "FORWARD",
		"-i", vm.TapName, "-o", "eth0", "-j", "ACCEPT"); err != nil {
		return err
	}

	if err := utils.Run("sudo", "iptables", "-A", "FORWARD",
		"-i", "eth0", "-o", vm.TapName,
		"-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}

	return nil
}

func (vm *VM) deleteTapDevice() error {
	exists, err := tapDeviceExists(vm.TapName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return utils.Run("sudo", "ip", "link", "delete", vm.TapName)
}

func (vm *VM) Cleanup() {
	if vm.Cmd != nil && vm.Cmd.Process != nil {
		if err := vm.Cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			log.Printf("failed to kill firecracker process for vm=%s pid=%d err=%v", vm.Name, vm.Cmd.Process.Pid, err)
		}
		if err := vm.Cmd.Wait(); err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				log.Printf("failed waiting for firecracker process cleanup vm=%s err=%v", vm.Name, err)
			}
		}
	}

	if err := vm.deleteTapDevice(); err != nil {
		log.Printf("failed to delete tap device for vm=%s tap=%s err=%v", vm.Name, vm.TapName, err)
	}
	if err := os.Remove(vm.SocketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to remove firecracker socket for vm=%s socket=%s err=%v", vm.Name, vm.SocketPath, err)
	}
	if err := os.Remove(vm.RootfsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to remove rootfs for vm=%s rootfs=%s err=%v", vm.Name, vm.RootfsPath, err)
	}
	if vm.ipLease != nil {
		ippool.ReleaseIP(vm.ipLease)
		vm.ipLease = nil
	}
}

func (vm *VM) startFirecraker() error {
	_ = os.Remove(vm.SocketPath)

	logFile, err := os.OpenFile(vm.LogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("firecracker", "--api-sock", vm.SocketPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	vm.Cmd = cmd
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return err
	}
	_ = logFile.Close()
	return nil

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

func (vm *VM) waitForAgent(timeout time.Duration) error {
	address := net.JoinHostPort(vm.GuestIP, "3000")
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}

	return fmt.Errorf("agent not reachable at %s", address)
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

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}

	if resp.Status < "200" || resp.Status >= "300" {
		return fmt.Errorf("PUT %s failed with status %s body=%s", path, resp.Status, string(respBody))
	}

	return nil

}

func (vm *VM) ConfigVm() error {

	client := firecraker.FirecrakerClient(vm.SocketPath)
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

func nextVM() (VM, error) {
	for attempts := 0; attempts < ippool.MaxSlots; attempts++ {
		ip, err := ippool.GetAIP()

		if err != nil {
			return VM{}, fmt.Errorf("no free vm slot available")
		}

		vm := newVM(ip.ID, ip)
		exists, err := tapDeviceExists(vm.TapName)
		if err != nil {
			ippool.ReleaseIP(ip)
			return VM{}, err
		}
		if exists {
			log.Printf("skipping vm name=%s because tap device already exists tap=%s", vm.Name, vm.TapName)
			ippool.ReleaseIP(ip)
			continue
		}
		if _, err := os.Stat(vm.SocketPath); err == nil {
			log.Printf("skipping vm name=%s because socket already exists socket=%s", vm.Name, vm.SocketPath)
			ippool.ReleaseIP(ip)
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			ippool.ReleaseIP(ip)
			return VM{}, err
		}

		return vm, nil
	}

	return VM{}, fmt.Errorf("no free vm slot available")
}

func CreateVm() (*VM, error) {
	vm, err := nextVM()
	if err != nil {
		return nil, err
	}
	log.Printf("creating vm name=%s guest_ip=%s host_ip=%s socket=%s tap=%s rootfs=%s firecracker_log=%s", vm.Name, vm.GuestIP, vm.HostIP, vm.SocketPath, vm.TapName, vm.RootfsPath, vm.LogPath)

	if err := vm.prepareRootfs(); err != nil {
		log.Printf("failed to prepare rootfs for vm=%s rootfs=%s base_rootfs=%s err=%v", vm.Name, vm.RootfsPath, vm.BaseRootfs, err)
		_ = os.Remove(vm.RootfsPath)
		return nil, fmt.Errorf("prepare rootfs: %w", err)
	}
	log.Printf("prepared rootfs for vm=%s rootfs=%s", vm.Name, vm.RootfsPath)

	if err := vm.SetUpNetwork(); err != nil {
		log.Printf("failed to setup network for vm=%s tap=%s host_ip=%s err=%v", vm.Name, vm.TapName, vm.HostIP, err)
		vm.Cleanup()
		return nil, fmt.Errorf("setup network: %w", err)
	}
	log.Printf("configured network for vm=%s tap=%s host_ip=%s", vm.Name, vm.TapName, vm.HostIP)

	err = vm.startFirecraker()

	if err != nil {
		log.Printf("failed to start firecracker for vm=%s socket=%s err=%v", vm.Name, vm.SocketPath, err)
		vm.Cleanup()
		return nil, fmt.Errorf("start firecracker: %w", err)
	}
	log.Printf("started firecracker process for vm=%s pid=%d socket=%s log=%s", vm.Name, vm.Cmd.Process.Pid, vm.SocketPath, vm.LogPath)

	if err := vm.waitForSocket(2 * time.Second); err != nil {
		log.Printf("failed waiting for firecracker socket for vm=%s socket=%s err=%v", vm.Name, vm.SocketPath, err)
		vm.Cleanup()
		return nil, fmt.Errorf("wait for socket: %w", err)
	}
	log.Printf("firecracker socket ready for vm=%s socket=%s", vm.Name, vm.SocketPath)

	if err := vm.ConfigVm(); err != nil {
		log.Printf("failed to configure/start firecracker vm=%s socket=%s tap=%s err=%v", vm.Name, vm.SocketPath, vm.TapName, err)
		vm.Cleanup()
		return nil, fmt.Errorf("configure vm: %w", err)
	}

	if err := vm.waitForAgent(3 * time.Minute); err != nil {
		log.Printf("failed waiting for agent vm=%s guest_ip=%s log=%s err=%v", vm.Name, vm.GuestIP, vm.LogPath, err)
		vm.Cleanup()
		return nil, fmt.Errorf("wait for agent: %w", err)
	}

	log.Printf("started vm=%s guest_ip=%s socket=%s tap=%s", vm.Name, vm.GuestIP, vm.SocketPath, vm.TapName)

	time.Sleep(100 * time.Millisecond)
	return &vm, nil

}
