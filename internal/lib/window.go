package lib

import (
	"log"

	"winpos/pkg/win"
	api "winpos/pkg/win/api"
)

func ShowWindowsList() {
	list, err := win.EnumAllWindows()
	if err != nil {
		log.Printf("Error fetching list of windows... %s\n", err.Error())
		return
	}
	for idx, wd := range list {
		log.Printf("%d: %s\n", idx, wd.String())
	}
}

//goland:noinspection GoUnusedExportedFunction
func ShowIdleStatus() {
	idle, err := api.GetLastInputTime()
	if err != nil {
		log.Printf("Error getting idle time: %v\n", err)
	} else {
		log.Printf("Idle time is: %f s\n", float64(idle/(1000*1000*1000)))
	}
	isSaver := win.IsScreenSaverRunning()
	log.Printf("Screen saver: %v\n", isSaver)
}
