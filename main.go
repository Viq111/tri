package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Viq111/tri/storage"
)

// common

var options struct {
	Verbose bool
	Version string
}

// cli examples
// tri <source> <dst> should backup all the files in source to dst.

var syncOptions struct {
	srcPath string
	dstPath string
}

func init() {
	options.Version = "0.1.0"
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(log.WarnLevel)
}

func main() {
	syncCommand := flag.NewFlagSet("sync", flag.ExitOnError)
	syncCommand.BoolVar(&options.Verbose, "v", false, "Display INFO level log")
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
		if options.Verbose {
			log.SetLevel(log.InfoLevel)
		}
		nbArgs := syncCommand.NArg()
		if nbArgs < 2 {
			log.Fatal("sync should be followed by <src> <dst>")
		}
		srcs := syncCommand.Args()[:nbArgs-1]
		dst := syncCommand.Args()[nbArgs-1]
		dstStorage, err := storage.NewLocalStorage(dst)
		if err != nil {
			log.Fatalf("Failed to read destination %s: %s\n", dst, err)
		}
		log.Infof("Syncing %s to %s...\n", strings.Join(srcs, ","), dst)
		for _, src := range srcs {
			srcStorage, err := storage.NewLocalStorage(src)
			if err != nil {
				log.Fatalf("Failed to read source %s: %s\n", src, err)
				continue
			}
			err = storage.Sync(srcStorage, ".", dstStorage, ".")
			if err != nil {
				log.Fatalf("Failed to sync source %s: %s\n", src, err)
			}
		}

	default:
		log.Fatalf("%s is not valid command.\n", os.Args[1])
	}
}
