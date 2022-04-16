package cmd

import (
	"log"

	"winpos/dev/lib"
	"winpos/dev/win"
)

func Restore() {

	saved, err := lib.LoadWindows()
	if err != nil {
		log.Printf("Error reading list of windows... %s\n", err.Error())
		return
	}

	windows, err := win.EnumAllWindows()
	if err != nil {
		log.Printf("Error fetching list of windows... %s\n", err.Error())
		return
	}

	for _, wd := range windows {

		sw, ok := saved[wd.Handle]
		if !ok {
			// no such window in stored
			log.Printf("Skip window:\n"+
				"  %s\n",
				wd.String(),
			)
			continue
		}

		log.Printf("Restoring window:\n"+
			"  > %s\n"+
			"  < %s\n",
			sw.String(),
			wd.String(),
		)

		err = wd.RestorePosition(sw)
	}

}
