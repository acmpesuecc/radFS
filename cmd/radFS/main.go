package main

import (
	"flag"
	"fmt"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	radFS "github.com/acmpesuecc/radFS/fs"
)

type config struct {
	debug bool
	mount string
}

func main() {

	debug := flag.Bool("d", false, "enable debug mode")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("use: go run main.go -d <mountpoint>")
		return
	}

	cfg := &config{
		debug: *debug,
		mount: flag.Arg(0),
	}

	if cfg.debug {
		fmt.Println("debug mode enabled")
	}

	//c is a fuse connection to dev/fuse
	c, err := fuse.Mount(cfg.mount)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer c.Close() //delay execution of Close

	err = fs.Serve(c, radFS.FS{}) //starts listening for FS reqs

	if err != nil {
		fmt.Println(err)
	}

	// <-c.Ready
	// if err := c.MountError; err != nil {
	// 	fmt.Println(err)
	// }
}
