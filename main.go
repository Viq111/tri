package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Viq111/tri/storage"
)

// cli examples
// tri <source> <dst> should backup all the files in source to dst.

var syncOptions struct {
	srcPath string
	dstPath string
}

func main() {
	syncCommand := flag.NewFlagSet("sync", flag.ExitOnError)
	if len(os.Args) == 1 {
		fmt.Printf(`usage: %s <command> [<args>]
		Available commands:
		  - sync <src> <dst> - Sync folder dst to mirror folder src
		`, os.Args[0])
		return
	}
	switch os.Args[1] {
	case "sync":
		syncCommand.Parse(os.Args[2:])
		nbArgs := syncCommand.NArg()
		if nbArgs < 2 {
			fmt.Printf("sync should be followed by <src> <dst>")
			os.Exit(2)
		}
		srcs := syncCommand.Args()[:nbArgs-1]
		dst := syncCommand.Args()[nbArgs-1]
		dstStorage, err := storage.NewLocalStorage(dst)
		if err != nil {
			fmt.Printf("Failed to read destination %s: %s\n", dst, err)
			os.Exit(5)
		}
		fmt.Printf("Syncing %s to %s...\n", srcs, dst)
		for _, src := range srcs {
			srcStorage, err := storage.NewLocalStorage(src)
			if err != nil {
				fmt.Printf("Failed to read source %s: %s\n", src, err)
				continue
			}
			err = storage.Sync(srcStorage, ".", dstStorage, ".")
			if err != nil {
				fmt.Printf("Failed to sync source %s: %s\n", src, err)
			}
		}

	default:
		fmt.Printf("%s is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}
