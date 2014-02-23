// This file is part of fs1up.
// Copyright (C) 2014 Andreas Klauer <Andreas.Klauer@metamorpher.de>
// License: GPL-2

// Package nbd uses the Linux NBD layer to emulate a block device in user space
package nbd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"syscall"
)

const (
	// Defined in <linux/fs.h>:
	BLKROSET = 4701
	// Defined in <linux/nbd.h>:
	NBD_SET_SOCK        = 43776
	NBD_SET_BLKSIZE     = 43777
	NBD_SET_SIZE        = 43778
	NBD_DO_IT           = 43779
	NBD_CLEAR_SOCK      = 43780
	NBD_CLEAR_QUE       = 43781
	NBD_PRINT_DEBUG     = 43782
	NBD_SET_SIZE_BLOCKS = 43783
	NBD_DISCONNECT      = 43784
	NBD_SET_TIMEOUT     = 43785
	NBD_SET_FLAGS       = 43786
	// enum
	NBD_CMD_READ  = 0
	NBD_CMD_WRITE = 1
	NBD_CMD_DISC  = 2
	NBD_CMD_FLUSH = 3
	NBD_CMD_TRIM  = 4
	// values for flags field
	NBD_FLAG_HAS_FLAGS  = (1 << 0) // nbd-server supports flags
	NBD_FLAG_READ_ONLY  = (1 << 1) // device is read-only
	NBD_FLAG_SEND_FLUSH = (1 << 2) // can flush writeback cache
	// there is a gap here to match userspace
	NBD_FLAG_SEND_TRIM = (1 << 5) // send trim/discard
	// These are sent over the network in the request/reply magic fields
	NBD_REQUEST_MAGIC = 0x25609513
	NBD_REPLY_MAGIC   = 0x67446698
	// Do *not* use magics: 0x12560953 0x96744668.
)

// Device interface is a subset of os.File.
type Device interface {
	ReadAt(b []byte, off int64) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
}

type request struct {
	magic  uint32
	typus  uint32
	handle uint64
	from   uint64
	len    uint32
}

func handle(fd int) {
	buf := make([]byte, 1024)

	var x request

	for {
		n, _ := syscall.Read(fd, buf)
		b := bytes.NewReader(buf[0:n])
		binary.Read(b, binary.BigEndian, &x.magic)
		binary.Read(b, binary.BigEndian, &x.typus)
		binary.Read(b, binary.BigEndian, &x.handle)
		binary.Read(b, binary.BigEndian, &x.from)
		binary.Read(b, binary.BigEndian, &x.len)
		fmt.Println("read", buf[0:n], x)
		// syscall.Write(fd, buf[0:n])
		// fmt.Println("wrote", buf[0:n])
	}
}

func Client(d Device, offset int64, size int64) {
	nbd, _ := os.Open("/dev/nbd0") // TODO: find a free one
	fd, _ := syscall.Socketpair(syscall.SOCK_STREAM, syscall.AF_UNIX, 0)
	go handle(fd[1])
	runtime.LockOSThread()
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_SET_SOCK, uintptr(fd[0]))
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_SET_BLKSIZE, 4096)
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_SET_SIZE_BLOCKS, uintptr(size/4096))
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_SET_FLAGS, 1)
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), BLKROSET, 0)  // || 1
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_DO_IT, 0) // doesn't return
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_DISCONNECT, 0)
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_CLEAR_SOCK, 0)
	runtime.UnlockOSThread()
}
