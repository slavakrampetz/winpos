// +build windows,amd64

package winapi

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                   = syscall.MustLoadDLL("user32.dll")
	procEnumWindows          = user32.MustFindProc("EnumWindows")
	procGetWindowTextW       = user32.MustFindProc("GetWindowTextW")
	procGetWindowPlacement   = user32.MustFindProc("GetWindowPlacement")
	procSetWindowPlacement   = user32.MustFindProc("SetWindowPlacement")
	procGetWindowLong        = user32.MustFindProc("GetWindowLongW")
	procGetLastInputInfo     = user32.MustFindProc("GetLastInputInfo")
	procSystemParametersInfo = user32.MustFindProc("SystemParametersInfoW")

	kernel32                 = syscall.MustLoadDLL("kernel32.dll")
	procGetTickCount         = kernel32.MustFindProc("GetTickCount")
	procAttachConsole        = kernel32.MustFindProc("AttachConsole")

	// getWindowLongPtr   = user32.MustFindProc("GetWindowLongPtrW")
	// getLastError       = kernel32.MustFindProc("GetLastError")
)

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfoa?redirectedfrom=MSDN
func SystemParametersInfo(uiAction uint32, uiParam uint32, pvParam uintptr, fWinIni uint32) error {
	r1, _, err := procSystemParametersInfo.Call(
		uintptr(uiAction),
		uintptr(uiParam),
		pvParam,
		uintptr(fWinIni),
	)
	// Error
	if r1 == 0 {
		return err
	}
	return nil
}

// Get last user input time in nanoseconds (1/10^9)
// from: https://stackoverflow.com/questions/22949444/using-golang-to-get-windows-idle-time-getlastinputinfo-or-similar
func GetLastInputTime() (t time.Duration, err error) {
	lii := LASTINPUTINFO {
		cbSize: 0,
		dwTime: 0,
	}
	lii.cbSize = uint32(unsafe.Sizeof(lii))
	currentTickCount, _, _ := procGetTickCount.Call()
	r1, _, err := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if r1 == 0 {
		if err != nil {
			return 0, fmt.Errorf("error getting last input info: %v", err.Error())
		}
		return 0, fmt.Errorf("error getting last input info: unknown error")
	}
	return time.Duration(uint32(currentTickCount) - lii.dwTime) * time.Millisecond, nil
}

func EnumWindows(enumFunc uintptr, lparam uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procEnumWindows.Addr(), 2, enumFunc, lparam, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetWindowLong(hWnd HWND, index int32) int32 {
	ret, _, _ := syscall.Syscall(procGetWindowLong.Addr(), 2,
		uintptr(hWnd),
		uintptr(index),
		0)

	return int32(ret)
}

func GetWindowStyle(hWnd syscall.Handle) int32 {
	return GetWindowLong(HWND(hWnd), GWL_STYLE)
}

func GetWindowText(hwnd syscall.Handle, str *uint16, maxCount int32) (len int32, err error) {
	r0, _, e1 := syscall.Syscall(procGetWindowTextW.Addr(), 3, uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	len = int32(r0)
	if len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

/*func findWindow(title string) (syscall.Handle, error) {
	var hwnd syscall.Handle

	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			// ignore the error
			return 1 // continue enumeration
		}
		if syscall.UTF16ToString(b) == title {
			// note the window
			hwnd = h
			return 0 // stop enumeration
		}
		return 1 // continue enumeration
	})
	ew := enumWindows(cb, 0)
	if ew != nil {
		return 0, fmt.Errorf("error finding window: %v", ew)
	}
	if hwnd == 0 {
		return 0, fmt.Errorf("No window with title '%s' found", title)
	}
	return hwnd, nil
}
*/

// func _GetLastError() uint32 {
// 	ret, _, _ := syscall.Syscall(getLastError.Addr(), 0,
// 		0,
// 		0,
// 		0)
//
// 	return uint32(ret)
// }

// func lastError(win32FuncName string) error {
// 	if errno := _GetLastError(); errno != ERROR_SUCCESS {
// 		return fmt.Errorf(fmt.Sprintf("%s: Error %d", win32FuncName, errno))
// 	}
// 	return fmt.Errorf(win32FuncName)
// }

func GetWindowPlacement(hWnd HWND, lpwndpl *WINDOWPLACEMENT) bool {
	ret, _, _ := syscall.Syscall(procGetWindowPlacement.Addr(), 2,
		uintptr(hWnd),
		uintptr(unsafe.Pointer(lpwndpl)),
		0)
	return ret != 0
}

func SetWindowPlacement(hWnd HWND, lpwndpl *WINDOWPLACEMENT) bool {
	ret, _, _ := syscall.Syscall(procSetWindowPlacement.Addr(), 2,
		uintptr(hWnd),
		uintptr(unsafe.Pointer(lpwndpl)),
		0)
	return ret != 0
}