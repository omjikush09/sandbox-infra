package start

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/omjikush09/sandboxing-infra/packages/vm/client"
)

// type Start interface {

// }
type VM struct {
	Id         string
	SocketPath string
	TapName    string // This is cable
	ipAddress  string
	HostTapIP  string
	GuestIP    string
	HostIP     string
}

func run(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdin
	c.Stdout = os.Stdout

	return c.Run()
}

// func StartFirecracker() (string, error) {
// 	id, err := uuid.NewV7()
// 	if err != nil {
// 		return "", err
// 	}
// 	path := fmt.Sprintf("/tmp/%s.socket", id.String())

// 	cmd := exec.Command("firecracker", "--api-sock", path)

// 	cmd.Stdout = os.Stdout
// 	cmd.Stdin = os.Stdin

// 	err = cmd.Start()
// 	if err != nil {
// 		return "", err
// 	}

// 	// Sleep for two second
// 	time.Sleep(50 * time.Millisecond)
// 	return path, nil
// }

func (vm *VM) SetUpNework() error {
	//create tap device
	if err := run("sudo", "ip", "tuntap", "add", vm.TapName, "mode", "tap"); err != nil {
		return err
	}

	// assign host-side IP to tap
	if err := run("sudo", "ip", "addr", "add", vm.HostTapIP+"/24", "dev", vm.TapName); err != nil {
		return err
	}

	// bring tap up
	if err := run("sudo", "ip", "link", "set", vm.TapName, "up"); err != nil {
		return err
	}

	// enable forwarding
	if err := run("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}

	// NAT traffic from VM subnet to host internet interface
	// Replace eth0 if your default interface is different.
	if err := run("sudo", "iptables", "-t", "nat", "-A", "POSTROUTING",
		"-s", "172.16.0.0/24", "-o", "eth0", "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := run("sudo", "iptables", "-A", "FORWARD",
		"-i", vm.TapName, "-o", "eth0", "-j", "ACCEPT"); err != nil {
		return err
	}

	if err := run("sudo", "iptables", "-A", "FORWARD",
		"-i", "eth0", "-o", vm.TapName,
		"-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}

	return nil
}

func StartVM() {

}

func (vm *VM) startFirecraker() error {
	_ = os.Remove(vm.SocketPath)

	cmd := exec.Command("firecracker", "--api-sock", vm.SocketPath)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}
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

func put(client *http.Client, path string, body string) error {
	req, err := http.NewRequest(http.MethodPut, "http://localhost/"+path, bytes.NewBufferString(body))

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
	if err := put(client, "/machine-config", ``); err != nil {
		return err
	}

	// bootArgs := fmt.Sprintf(
	// 	"console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw ip=%s::%s:255.255.255.0::eth0:off",
	// 	vm.GuestIP,
	// 	vm.HostIP,
	// )
	//
	return nil

}
