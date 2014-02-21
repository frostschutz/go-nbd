// This file is part of fs1up.
// Copyright (C) 2014 Andreas Klauer <Andreas.Klauer@metamorpher.de>
// License: GPL-2

// Package nbd uses the Linux NBD layer to emulate a block device in user space
package nbd

import (
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
	NBD_SET_SIZE_BLOCKS = 43782
	NBD_DISCONNECT      = 43783
	NBD_SET_TIMEOUT     = 43784
	NBD_SET_FLAGS       = 43785
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

// DeviceInfo interface is a subset of os.FileInfo.
type DeviceInfo interface {
	Size() int64
}

// Device interface is a subset of os.File.
type Device interface {
	Stat() (di DeviceInfo, err error)
	ReadAt(b []byte, off int64) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
}

func Client(b Device) {
	runtime.LockOSThread()
	nbd := os.Open("/dev/nbd0") // TODO: find a free one
	fd, _ := Syscall.Socketpair(SOCK_STREAM, AF_UNIX, 0)
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_SET_SOCK, fd[0])
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_SET_BLKSIZE, 4096)
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_SET_SIZE_BLOCKS, b.Stat().Size()/4096)
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_SET_FLAGS, 0)
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), BLKROSET, 0)  // || 1
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_DO_IT, 0) // doesn't return
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_DISCONNECT, 0)
	syscall.Syscall(syscall.SYS_IOCTL, nbd.Fd(), NBD_CLEAR_SOCK, 0)
	runtime.UnlockOSThread()
}
