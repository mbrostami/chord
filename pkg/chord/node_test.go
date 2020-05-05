package chord

import (
	"crypto/sha256"
	"testing"
)

func TestNewNode(t *testing.T) {
	newNode := NewNode("127.0.0.1", 10001)
	if newNode.IP != "127.0.0.1" {
		t.Errorf("IP address should be %s, got %s", "127.0.0.1", newNode.IP)
	}
	if newNode.Port != 10001 {
		t.Errorf("Port should be %d, got %d", 10001, newNode.Port)
	}
	if len(newNode.Identifier) != sha256.Size {
		t.Errorf("Identifier length should be %d, got %d", sha256.Size, len(newNode.Identifier))
	}
}

func TestFullAddrNode(t *testing.T) {
	newNode := NewNode("127.0.0.1", 10001)
	if newNode.FullAddr() != "127.0.0.1:10001" {
		t.Errorf("Full address should be %s, got %s", "127.0.0.1:10001", newNode.FullAddr())
	}
}
