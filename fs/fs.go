package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	radFS "github.com/acmpesuecc/radFS/internal/fs"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid usage. Use go run main.go <mntpoint>")
		return
	}

	mount_point := os.Args[1]

	c, err := fuse.Mount(mount_point)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	serv := make(chan error, 1)
	go func() {
		serv <- fs.Serve(c, radFS.FS{})
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	<-signals
	fmt.Println("Interrupt received: shutting down.")
	unmount_err := fuse.Unmount(mount_point)

	if unmount_err != nil {
		fmt.Println("Lazy Unmounting")
		command := exec.Command("fusermount", "-u", "-z", mount_point)
		cmd_err := command.Run()
		if cmd_err != nil {
			fmt.Println(cmd_err)
			return
		}
		return
	}

	if err := <-serv; err != nil {
		fmt.Println("Serve error:", err)
	}
}
