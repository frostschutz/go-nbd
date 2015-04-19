// Copyright (C) 2014 Andreas Klauer <Andreas.Klauer@metamorpher.de>
// License: GPL

// nbdsetup is an alternative to losetup using network block devices.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/frostschutz/nbd"
)

func main() {
	file := flag.String("file", "", "regular file or block device")
	flag.Parse()
	fmt.Printf("Using %s\n", *file)
	device, err := os.Open(*file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	stat, _ := device.Stat()
	dev, err := nbd.Create(device, stat.Size()).Connect()

	fmt.Println(dev)
	fmt.Println(err)
}

// End of file.
