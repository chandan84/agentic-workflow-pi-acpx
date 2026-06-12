// Package testutil holds tiny helpers shared by service tests.
package testutil

import (
	"net"
	"testing"
)

// FreePort returns an OS-allocated free TCP port. The listener is closed
// before returning, so the port may briefly be in TIME_WAIT — tests should
// retry binding if they care.
func FreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}
