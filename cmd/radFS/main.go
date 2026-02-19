package main

import (
	"fmt"
	"os"

	radFS "github.com/acmpesuecc/radFS/fs"
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
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
