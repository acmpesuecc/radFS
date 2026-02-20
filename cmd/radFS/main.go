package main

import (
	"flag"
	"fmt"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	radFS "github.com/acmpesuecc/radFS/fs"
)

type config struct {
	Mount string
	Debug bool
}

func main() {
	debug := flag.Bool("d", false, "enable debug mode")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("use: go run main.go -d <mountpoint>")
		return
	}

	cfg := &config{
		Debug: *debug,
		Mount: flag.Arg(0),
	}
	if cfg.Debug {
		fmt.Println("Debug mode enabled")
	}

	//c is a fuse connection to dev/fuse
	c, err := fuse.Mount(cfg.Mount)

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
