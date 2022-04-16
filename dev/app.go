package main

import (
	"log"
	"os"
	"strings"

	"winpos/dev/cmd"
)

func main() {

	// Logging init
	flog, err := os.OpenFile("winpos.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
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
	switch strings.ToLower(command) {
	default:
		cmd.Help()
	case "help", "h", "--help", "-h":
		cmd.Help()
	case "list", "l", "ls", "--list", "-l":
		cmd.List()
	case "save", "s", "--save", "-s":
		cmd.Save()
	case "restore", "r", "--restore", "-r":
		cmd.Restore()
	case "reset", "z", "--reset", "-z":
		cmd.Reset()
	case "watch", "w", "--watch", "-w":
		cmd.Watch()
	}

}
