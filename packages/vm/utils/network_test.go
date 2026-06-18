package utils

import "testing"

func TestParseDefaultNetworkInterface(t *testing.T) {
	iface, err := parseDefaultNetworkInterface("default via 172.31.0.1 dev ens5 proto dhcp src 172.31.1.2 metric 100\n")
	if err != nil {
		t.Fatalf("parse default interface: %v", err)
	}
	if iface != "ens5" {
		t.Fatalf("expected ens5, got %s", iface)
	}
}

func TestParseDefaultNetworkInterfaceMissingDev(t *testing.T) {
	if _, err := parseDefaultNetworkInterface("default via 172.31.0.1\n"); err == nil {
		t.Fatal("expected error for missing dev field")
	}
}
