package main

import (
	"os"
	"testing"
)

func TestPrioritizeLANServersPrefersRemoteLANAndDedupesLocalNoise(t *testing.T) {
	localName, _ := os.Hostname()
	if localName == "" {
		localName = "LOCALHOST"
	}

	servers := map[string]LANServer{
		"localhost:8080":      {Name: "本机服务器", IP: "localhost:8080"},
		"127.0.0.1:8080":      {Name: localName, IP: "127.0.0.1:8080"},
		"169.254.181.1:8080":  {Name: localName, IP: "169.254.181.1:8080"},
		"198.18.0.1:8080":     {Name: localName, IP: "198.18.0.1:8080"},
		"192.168.31.92:8080":  {Name: localName, IP: "192.168.31.92:8080"},
		"192.168.31.100:8080": {Name: "ROOM-HOST", IP: "192.168.31.100:8080"},
		"169.254.10.100:8080": {Name: "ROOM-HOST", IP: "169.254.10.100:8080"},
		"198.18.10.100:8080":  {Name: "ROOM-HOST", IP: "198.18.10.100:8080"},
		"192.168.31.101:8080": {Name: "ANOTHER-HOST", IP: "192.168.31.101:8080"},
	}

	got := prioritizeLANServers(servers)
	if len(got) != 3 {
		t.Fatalf("expected 3 deduped servers, got %d: %#v", len(got), got)
	}

	if got[0].Name != "ANOTHER-HOST" || got[0].IP != "192.168.31.101:8080" || !got[0].Recommended || got[0].Tag != "推荐" {
		t.Fatalf("first server should be a recommended remote LAN host, got %#v", got[0])
	}
	if got[1].Name != "ROOM-HOST" || got[1].IP != "192.168.31.100:8080" || !got[1].Recommended || got[1].Tag != "推荐" {
		t.Fatalf("second server should keep ROOM-HOST's best LAN address, got %#v", got[1])
	}
	if got[2].IP != "192.168.31.92:8080" || got[2].Recommended || got[2].Tag != "本机" {
		t.Fatalf("local server should keep the best local LAN address and be tagged as local, got %#v", got[2])
	}
}
