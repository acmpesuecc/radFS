package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/exec"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	radFS "github.com/acmpesuecc/radFS/fs"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid usage. Use go run main.go <mntpoint>")
		return
	}

	mount_point := os.Args[1]

	//c is a fuse connection to dev/fuse
	c, err := fuse.Mount(mount_point)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	//routine for serve
	serv := make(chan error, 1)
	go func() {	
		serv <- fs.Serve(c, radFS.FS{})
	}()
	
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt) //internal go runtime/routine


	//main() goroutine waits here
	<-signals
	fmt.Println("Interrupt recived : shutting down.")
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
