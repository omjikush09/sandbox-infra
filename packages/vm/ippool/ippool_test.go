package ippool

import "testing"

func TestInitCreatesDistinctHostAndGuestIPs(t *testing.T) {
	pool := IpPool{}
	pool.Init()

	if len(pool.Ideal) != MaxSlots {
		t.Fatalf("expected %d free IP slots, got %d", MaxSlots, len(pool.Ideal))
	}

	first := pool.Ideal[0]
	if first.HostIP != "172.16.1.1" {
		t.Fatalf("expected first host IP 172.16.1.1, got %s", first.HostIP)
	}
	if first.VmIP != "172.16.1.2" {
		t.Fatalf("expected first VM IP 172.16.1.2, got %s", first.VmIP)
	}
	if first.HostIP == first.VmIP {
		t.Fatal("host and VM IPs must not be the same")
	}
}

func TestLeaseAndRelease(t *testing.T) {
	pool := IpPool{}
	pool.Init()

	ip, err := pool.LeaseIp()
	if err != nil {
		t.Fatalf("lease IP: %v", err)
	}
	if ip.state != IPLeased {
		t.Fatalf("expected leased state, got %s", ip.state)
	}
	if len(pool.InUse) != 1 {
		t.Fatalf("expected one in-use IP, got %d", len(pool.InUse))
	}
	if len(pool.Ideal) != MaxSlots-1 {
		t.Fatalf("expected %d free IP slots, got %d", MaxSlots-1, len(pool.Ideal))
	}

	if err := pool.ReleaseIP(ip); err != nil {
		t.Fatalf("release IP: %v", err)
	}
	if ip.state != IPFree {
		t.Fatalf("expected free state, got %s", ip.state)
	}
	if len(pool.InUse) != 0 {
		t.Fatalf("expected no in-use IPs, got %d", len(pool.InUse))
	}
	if len(pool.Ideal) != MaxSlots {
		t.Fatalf("expected %d free IP slots, got %d", MaxSlots, len(pool.Ideal))
	}

	if err := pool.ReleaseIP(ip); err != nil {
		t.Fatalf("second release should be ignored, got: %v", err)
	}
	if len(pool.Ideal) != MaxSlots {
		t.Fatalf("double release added duplicate IP, free slots=%d", len(pool.Ideal))
	}
}
