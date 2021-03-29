// +build windows,amd64

package win

import (
	"fmt"
	"syscall"
	"unsafe"

	api "winpos/dev/win/api"
)

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

type Wnd struct {
	Title string
	Handle syscall.Handle
	Style  int32
	Flags  uint32
	ShowCmd uint32
	MinPosition api.POINT
	MaxPosition api.POINT
	NormalPosition api.RECT
}

func EnumAllWindows() ([]Wnd, error) {
	list := make([]Wnd, 0)

	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {

		style := api.GetWindowStyle(h)
		if !isGuiWindow(style) {
			return 1 // incorrect window
		}

		var wp api.WINDOWPLACEMENT
		wp.Length = uint32(unsafe.Sizeof(wp))
		if !api.GetWindowPlacement(api.HWND(h), &wp) {
			return 1
			// return lastError("GetWindowPlacement")
		}

		b := make([]uint16, 200)
		_, err := api.GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			return 1 // ignore the error, continue enumeration
		}

		data := Wnd{
			Title:          syscall.UTF16ToString(b),
			Handle:         h,
			Style:          style,
			Flags:          wp.Flags,
			ShowCmd:        wp.ShowCmd,
			MinPosition:    api.POINT {
				X: wp.PtMinPosition.X,
				Y: wp.PtMinPosition.Y},
			MaxPosition:    api.POINT {
				X: wp.PtMaxPosition.X,
				Y: wp.PtMaxPosition.Y},
			NormalPosition: api.RECT {
				Left: wp.RcNormalPosition.Left,
				Top: wp.RcNormalPosition.Top,
				Right: wp.RcNormalPosition.Right,
				Bottom: wp.RcNormalPosition.Bottom,
			},
		}
		list = append(list, data)
		return 1 // continue enumeration
	})
	ew := api.EnumWindows(cb, 0)
	if ew != nil {
		return nil, fmt.Errorf("error finding window: %v", ew)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("cannot enumerate windows, empty list")
	}
	return list, nil
}

func IsScreenSaverRunning() bool {
	var flag uint32 = 0
	err := api.SystemParametersInfo(api.SPI_GETSCREENSAVERRUNNING, 0, uintptr(unsafe.Pointer(&flag)), 0)
	if err == nil {
		return flag == 1
	}
	return false
}

func (w *Wnd) String() string {
	return fmt.Sprintf("%x, %s, %d:%d-%d:%d, %s",
		w.Handle, w.ShowText(),
		w.NormalPosition.Left, w.NormalPosition.Top,
		w.NormalPosition.Right, w.NormalPosition.Bottom,
		w.Title)
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-windowplacement
func (w *Wnd) ShowText() string {
	switch w.ShowCmd {
	case 1:
		return "normal"
	case 2:
		return "min"
	case 3:
		return "max"
	}
	return fmt.Sprintf("%d: unknown", w.ShowCmd)
}

// Save window positioning status to string
func (w *Wnd) Save() string {
	state := fmt.Sprint(
		w.Handle,
		w.Flags, w.ShowCmd,
		w.MinPosition.X, w.MinPosition.Y,
		w.MaxPosition.X, w.MaxPosition.Y,
		w.NormalPosition.Left, w.NormalPosition.Top,
		w.NormalPosition.Right, w.NormalPosition.Bottom)
	return state
}

// Load window positioning from string
func (w *Wnd) Load(state string) bool {
	w.Title = "?"
	_, err := fmt.Sscan(state,
		&w.Handle,
		&w.Flags, &w.ShowCmd,
		&w.MinPosition.X, &w.MinPosition.Y,
		&w.MaxPosition.X, &w.MaxPosition.Y,
		&w.NormalPosition.Left, &w.NormalPosition.Top,
		&w.NormalPosition.Right, &w.NormalPosition.Bottom)
	return err == nil
}

func (w *Wnd) RestorePosition(saved Wnd) error {

	var wp api.WINDOWPLACEMENT
	wp.Length = uint32(unsafe.Sizeof(wp))
	wp.Flags = saved.Flags
	wp.ShowCmd = saved.ShowCmd
	wp.PtMinPosition.X = saved.MinPosition.X
	wp.PtMinPosition.Y = saved.MinPosition.Y
	wp.PtMaxPosition.X = saved.MaxPosition.X
	wp.PtMaxPosition.Y = saved.MaxPosition.Y
	wp.RcNormalPosition.Left = saved.NormalPosition.Left
	wp.RcNormalPosition.Top = saved.NormalPosition.Top
	wp.RcNormalPosition.Right = saved.NormalPosition.Right
	wp.RcNormalPosition.Bottom = saved.NormalPosition.Bottom

	if !api.SetWindowPlacement(api.HWND(w.Handle), &wp) {
		return fmt.Errorf("cannot set window position, %x", w.Handle)
	}
	return nil
}

// Tools

func isGuiWindow(style int32) bool {

	// Child windows
	// Disabled windows
	mask := int32(api.WS_DISABLED | api.WS_CHILD)
	if (style & mask) > 0 {
		return false
	}

	mask = int32(api.WS_SYSMENU | api.WS_VISIBLE)
	if (style & mask) != mask {
		return false
	}
	return true
}

