package main

import (
	"log"
	"os"

	"winpos/dev/cmd"
)

func main() {

	// Logging init
	flog, err := os.OpenFile("winpos.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
	if err != nil {
		os.Exit(1)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer flog.Close()
	log.SetOutput(flog)

	// Arguments
	args := os.Args[1:]
	var command string
	if len(args) != 1 {
		command = "help"
	} else {
		command = args[0]
	}

	// Choose command
	switch command {
	case "help", "h":
		fallthrough
	default:
		cmd.Help()
	case "list", "l":
		cmd.List()
	case "save", "s":
		cmd.Save()
	case "restore", "r":
		cmd.Restore()
	case "reset", "z":
		cmd.Reset()
	case "watch", "w":
		cmd.Watch()
	}

}