package cmd

import (
	"log"
)

const usage = `
winpos: save/load window positions

Usage: winpos command [flags]

List of commands:

help:    show this screen
         short: h 

list:    show list of windows
         short: l, ls 

save:    save window positions to file
         short: s 

restore: restore window positions from file
         short: r

reset:   cleanup stored window positions
         short: z

watch:   TO BE IMPLEMENTED
         watch for changed positions and auto-save it
         short: w
`

func Help() {
	log.Print(usage)
}
