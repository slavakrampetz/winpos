// +build windows,amd64

package winapi

type (
	HANDLE        uintptr
	HWND          HANDLE
)

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

//goland:noinspection GoNameStartsWithPackageName
type WINDOWPLACEMENT struct {
	Length           uint32
	Flags            uint32
	ShowCmd          uint32
	PtMinPosition    POINT
	PtMaxPosition    POINT
	RcNormalPosition RECT
}

type LASTINPUTINFO struct {
	cbSize uint32
	dwTime uint32
}