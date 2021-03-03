// Copyright (c) 2021 AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package peercred

import (
	"fmt"
	"net"
	"os/user"
	"strconv"

	"golang.org/x/sys/unix"
)

func init() {
	osGet = getLinux
}

func getLinux(c net.Conn) (*Creds, error) {
	switch c := c.(type) {
	case *net.UnixConn:
		return getUnix(c)
	case *net.TCPConn:
		// TODO: use /proc tcp info for localhost connections like Windows?
	}
	return nil, ErrUnsupportedConnType
}

func getUnix(c *net.UnixConn) (*Creds, error) {
	raw, err := c.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("SyscallConn: %w", err)
	}

	var cred *unix.Ucred
	cerr := raw.Control(func(fd uintptr) {
		cred, err = unix.GetsockoptUcred(int(fd),
			unix.SOL_SOCKET,
			unix.SO_PEERCRED)
	})
	if cerr != nil {
		return nil, fmt.Errorf("raw.Control: %w", cerr)
	}
	if err != nil {
		return nil, fmt.Errorf("unix.GetsockoptUcred: %w", err)
	}
	creds := &Creds{
		pid: int(cred.Pid),
		uid: strconv.FormatUint(uint64(cred.Uid), 10),
	}

	u, err := user.LookupId(creds.uid)
	if err != nil {
		return nil, err
	}

	creds.gids, err = u.GroupIds()
	if err != nil {
		return nil, err
	}

	return creds, nil
}
