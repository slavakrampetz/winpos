package cmd

import (
	"log"
	"time"

	"winpos/pkg/win"
	winapi "winpos/pkg/win/api"
)

//goland:noinspection GoSnakeCaseUsage
const (
	NANO2SEC = 1000 * 1000 * 1000

	IDLE_DELAY = time.Minute * 10
	SAVE_DELAY = time.Second * 60
)

func Watch() {

	// Save now
	Save()
	lastRun := time.Now()

	t := time.NewTicker(15 * time.Second)
	for range t.C {

		idle, err := winapi.GetLastInputTime()
		if err != nil {
			idle = 0
		}

		now := time.Now()
		fromLast := now.Sub(lastRun)
		isSaver := win.IsScreenSaverRunning()
		// log.Printf("Saver?: %v\n", isSaver)

		// Time to save?
		if (idle > IDLE_DELAY && !isSaver) || fromLast > SAVE_DELAY {
			log.Printf("Idle?: %.2f sec, delay %.2f s\n", float64(idle/NANO2SEC), float64(IDLE_DELAY/NANO2SEC))
			log.Printf("Last: %s, diff %.2f s, delay %.2f s\n",
				lastRun.Format("15:04:05"),
				float32(fromLast/NANO2SEC),
				float32(SAVE_DELAY/NANO2SEC),
			)
			Save()
			lastRun = time.Now()
		}
	}
}
