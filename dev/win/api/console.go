package winapi

import (
	"fmt"
	"log"
	"os"
	"syscall"
)

//goland:noinspection GoSnakeCaseUsage
const (
	ATTACH_PARENT_PROCESS = ^uint32(0) // (DWORD)-1
)

func attachConsole(dwParentProcess uint32) (ok bool, err error) {
	r1, _, err := syscall.SyscallN(procAttachConsole.Addr(), 1, uintptr(dwParentProcess), 0, 0)
	ok = r1 != 0 && err == nil
	return
}

// GuiToConsole attach the GUI process to the parent console, so we can send "logs" & "prints" if needed
//goland:noinspection GoUnusedExportedFunction
func GuiToConsole() (err error) {
	ok, err := attachConsole(ATTACH_PARENT_PROCESS)
	if ok {
		hOut, err1 := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
		if err1 != nil {
			return fmt.Errorf("stdout connection error : %v", err1)
		}
		hErr, err2 := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
		if err2 != nil {
			return fmt.Errorf("stderr connection error : %v", err2)
		}
		os.Stdout = os.NewFile(uintptr(hOut), "/dev/stdout")
		os.Stderr = os.NewFile(uintptr(hErr), "/dev/stderr")
		log.SetOutput(os.Stderr)
		return
	}
	return fmt.Errorf("attachconsole error : %v", err)
}
