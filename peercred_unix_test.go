// Copyright (c) 2021 AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.15
// +build linux darwin freebsd

package peercred // import "inet.af/peercred"

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUnixSock(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, "foo.sock")
	sock, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer sock.Close()

	clientConnCh := make(chan net.Conn, 1)
	go func() {
		defer close(clientConnCh)
		c, err := net.Dial("unix", path)
		if err != nil {
			t.Error(err)
			return
		}
		clientConnCh <- c
	}()
	clientConn, ok := <-clientConnCh
	if !ok {
		return
	}
	defer clientConn.Close()

	c, err := sock.Accept()
	if err != nil {
		t.Fatalf("Accept: %v", err)
	}
	defer c.Close()

	creds, err := Get(c)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	uid, ok := creds.UserID()
	if !ok {
		t.Errorf("no UID")
	}
	if got, want := uid, fmt.Sprint(os.Getuid()); got != want {
		t.Errorf("UID = %q; want %q", got, want)
	}
	pid, ok := creds.PID()
	if runtime.GOOS == "freebsd" {
		if ok {
			t.Error("PID ok; want !ok. Thank you for fixing FreeBSD, please update the test.")
		}
	} else {
		if !ok {
			t.Errorf("no PID")
		}
		if got, want := pid, os.Getpid(); got != want {
			t.Errorf("PID = %v; want %v", got, want)
		}
	}
}
