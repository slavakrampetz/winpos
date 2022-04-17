package cmd

import (
	"log"

	"winpos/internal/lib"
	"winpos/pkg/win"
)

func Save() {

	list, err := win.EnumAllWindows()
	if err != nil {
		log.Printf("Error fetching list of windows... %s\n", err.Error())
		return
	}

	err = lib.SaveWindows(list)
	if err != nil {
		log.Printf("Error saving list of windows... %s\n", err.Error())
	}

}
