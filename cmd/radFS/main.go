package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	radFS "github.com/acmpesuecc/radFS/internal/fs"
)

type config struct {
	debug bool
	mount string
}

func main() {
	debug := flag.Bool("d", false, "Enable debug mode")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: go run cmd/radFS/main.go [-d] <mountpoint>")
		return
	}

	cfg := &config{
		debug: *debug,
		mount: flag.Arg(0),
	}

	if cfg.debug {
		log.Println("Debug mode enabled")
	}

	//c is a fuse connection to dev/fuse
	c, err := fuse.Mount(cfg.mount)

	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()

	serv := make(chan error, 1)
	go func() {
		serv <- fs.Serve(c, &radFS.FS{Debug: cfg.debug})
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	<-signals
	log.Println("Interrupt received: shutting down.")
	unmount_err := fuse.Unmount(cfg.mount)

	if unmount_err != nil {
		log.Println("Lazy Unmounting")
		command := exec.Command("fusermount", "-u", "-z", cfg.mount)
		cmd_err := command.Run()

		if cmd_err != nil {
			log.Println(cmd_err)
			return
		}

		return
	}

	if err := <-serv; err != nil {
		log.Println("Serve error:", err)
	}
}
