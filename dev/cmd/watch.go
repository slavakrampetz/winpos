package cmd

import (
	"time"

	"winpos/dev/lib"
)

func Watch() {
	// pr, err := win.ProcessList()
	// if err != nil {
	// 	log.Printf("Error fetching process list... %s\r\n", err.Error())
	// 	return
	// }
	//
	// for _, p := range pr {
	// 	log.Printf("%8d - %-30s - %-30s - %s\r\n", p.Pid, p.Username, p.Executable, p.Fullpath)
	// }
	lib.ShowIdleStatus()
	lib.ShowWindowsList()

	t := time.NewTicker(1 * time.Second)
	for range t.C {
		lib.ShowIdleStatus()
		lib.ShowWindowsList()
	}
}
