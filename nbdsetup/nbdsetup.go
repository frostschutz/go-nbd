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
	write := flag.Bool("write", false, "use true for read-write mode")
	flag.Parse()
	if *file == "" {
		flag.Usage()
		os.Exit(2)
	}
	fmt.Printf("Using %s in read", *file)
	device, err := os.Open(*file)
	if *write {
		fmt.Printf("-write")
		device, err = os.OpenFile(*file, os.O_RDWR, os.FileMode(0666))
	}
	fmt.Printf(" mode.\n")
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
