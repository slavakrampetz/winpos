package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"winpos/dev/win"
)

type WndMap map[syscall.Handle]win.Wnd

var (
	statePath string
)

func getRootPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	exePath := filepath.Dir(exe)
	return exePath, nil
}

func init() {
	rootPath, err := getRootPath()
	if err != nil {
		panic(err)
	}

	statePath = filepath.Join(rootPath, "windows.txt")
}

func SaveWindows(data []win.Wnd) error {

	if len(data) < 1 {
		log.Println("No windows to save")
		return nil
	}

	// Open file
	fp, err := os.OpenFile(statePath, os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 644)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fp.Close()

	for _, wd := range data {
		state := wd.Save()
		_, err := fp.WriteString(state + "\n")
		if err != nil {
			return err
		}
	}

	log.Printf("Saved %d windows to %s\n", len(data), statePath)
	return nil
}


func LoadWindows() (WndMap, error) {

	wMap := make(WndMap)

	// Open file
	fp, err := os.OpenFile(statePath, os.O_RDONLY, 644)
	if err != nil {
		return wMap, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		state := scanner.Text()
		wp := win.Wnd {}
		if !wp.Load(state) {
			return nil, fmt.Errorf("error reading window data")
		}
		wMap[wp.Handle] = wp
	}
	log.Printf("Loaded %d windows to %s\n", len(wMap), statePath)
	return wMap, nil
}

func Cleanup() {
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		log.Printf("State already deleted: %s\n", statePath)
		return
	}
	if err := os.Remove(statePath); err != nil {
		log.Printf("Error deleting state %s: %v\n", statePath, err.Error())
	} else {
		log.Printf("State deleted: %s\n", statePath)
	}
}