// +build windows,amd64

package win

import (
	"fmt"
	"reflect"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modKernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procCloseHandle               = modKernel32.NewProc("CloseHandle")
	procOpenProcess               = modKernel32.NewProc("OpenProcess")
	procCreateToolhelp32Snapshot  = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First            = modKernel32.NewProc("Process32FirstW")
	procProcess32Next             = modKernel32.NewProc("Process32NextW")
	procGetCurrentProcess         = modKernel32.NewProc("GetCurrentProcess")
	procQueryFullProcessImageName = modKernel32.NewProc("QueryFullProcessImageNameW")

	modAdvapi32                   = syscall.NewLazyDLL("advapi32.dll")
	procOpenProcessToken          = modAdvapi32.NewProc("OpenProcessToken")
	procLookupPrivilegeValue      = modAdvapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges     = modAdvapi32.NewProc("AdjustTokenPrivileges")
	procGetTokenInformation       = modAdvapi32.NewProc("GetTokenInformation")

	modSecur32                    = syscall.NewLazyDLL("secur32.dll")
	sessLsaFreeReturnBuffer       = modSecur32.NewProc("LsaFreeReturnBuffer")
	sessLsaEnumerateLogonSessions = modSecur32.NewProc("LsaEnumerateLogonSessions")
	sessLsaGetLogonSessionData    = modSecur32.NewProc("LsaGetLogonSessionData")
)

//goland:noinspection GoSnakeCaseUsage
const (
	MAX_PATH                      = 260
	MAX_FULL_PATH                 = 4096
	PROC_SE_DEBUG_NAME            = "SeDebugPrivilege"

	PROCESS_QUERY_INFORMATION     = 0x0400
	PROC_TOKEN_QUERY              = 0x0008
	PROC_TOKEN_ADJUST_PRIVILEGES  = 0x0020

	PROC_SE_PRIVILEGE_ENABLED     = 0x00000002
)

// PROCESSENTRY32 is the Windows API structure that contains a process's information.
type PROCESSENTRY32 struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	CntThreads        uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [MAX_PATH]uint16
}

type Process struct {
	Pid        int    `json:"pid"`
	Ppid       int    `json:"parentpid"`
	Executable string `json:"exeName"`
	Fullpath   string `json:"fullPath"`
	Username   string `json:"username"`
}

type LUID struct {
	LowPart  uint32
	HighPart int32
}

//goland:noinspection GoSnakeCaseUsage
type TOKEN_PRIVILEGES struct {
	PrivilegeCount uint32
	Privileges     [1]LUID_AND_ATTRIBUTES
}

//goland:noinspection GoSnakeCaseUsage
type LUID_AND_ATTRIBUTES struct {
	LUID       LUID
	Attributes uint32
}

//goland:noinspection GoSnakeCaseUsage
type LSA_UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	buffer        uintptr
}

//goland:noinspection GoSnakeCaseUsage
type SECURITY_LOGON_SESSION_DATA struct {
	Size                  uint32
	LogonId               LUID
	UserName              LSA_UNICODE_STRING
	LogonDomain           LSA_UNICODE_STRING
	AuthenticationPackage LSA_UNICODE_STRING
	LogonType             uint32
	Session               uint32
	Sid                   uintptr
	LogonTime             uint64
	LogonServer           LSA_UNICODE_STRING
	DnsDomainName         LSA_UNICODE_STRING
	Upn                   LSA_UNICODE_STRING
}

//goland:noinspection GoSnakeCaseUsage
type TOKEN_STATISTICS struct {
	TokenId            LUID
	AuthenticationId   LUID
	ExpirationTime     uint64
	TokenType          uint32
	ImpersonationLevel uint32
	DynamicCharged     uint32
	DynamicAvailable   uint32
	GroupCount         uint32
	PrivilegeCount     uint32
	ModifiedId         LUID
}


func ProcessList() ([]Process, error) {
	err := procAssignCorrectPrivs(PROC_SE_DEBUG_NAME)
	if err != nil {
		return nil, fmt.Errorf("error assigning privs... %s", err.Error())
	}

	lList, err := sessUserLUIDs()
	if err != nil {
		return nil, fmt.Errorf("error getting LUIDs... %s", err.Error())
	}

	handle, _, _ := procCreateToolhelp32Snapshot.Call(0x00000002, 0)
	if handle < 0 {
		return nil, syscall.GetLastError()
	}
	//goland:noinspection GoUnhandledErrorResult
	defer procCloseHandle.Call(handle)

	var entry PROCESSENTRY32
	entry.Size = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("error retrieving process info")
	}

	results := make([]Process, 0)
	for {
		path, ll, _ := getProcessFullPathAndLUID(entry.ProcessID)

		var user string
		for k, l := range lList {
			if reflect.DeepEqual(k, ll) {
				user = l
				break
			}
		}

		results = append(results, newProcessData(&entry, path, user))

		ret, _, _ := procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return results, nil
}


func procAssignCorrectPrivs(name string) error {
	handle, _, _ := procGetCurrentProcess.Call()
	if handle == uintptr(0) {
		return fmt.Errorf("unable to get current process handle")
	}
	//goland:noinspection GoUnhandledErrorResult
	defer procCloseHandle.Call(handle)

	var tHandle uintptr
	opRes, _, _ := procOpenProcessToken.Call(
		handle,
		uintptr(uint32(PROC_TOKEN_ADJUST_PRIVILEGES)),
		uintptr(unsafe.Pointer(&tHandle)),
	)
	if opRes != 1 {
		return fmt.Errorf("unable to open current process token")
	}
	//goland:noinspection GoUnhandledErrorResult
	defer procCloseHandle.Call(tHandle)

	nPointer, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return fmt.Errorf("unable to encode SE_DEBUG_NAME to UTF16")
	}
	var pValue LUID
	lpRes, _, _ := procLookupPrivilegeValue.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(nPointer)),
		uintptr(unsafe.Pointer(&pValue)),
	)
	if lpRes != 1 {
		return fmt.Errorf("unable to lookup priv value")
	}

	iVal := TOKEN_PRIVILEGES{
		PrivilegeCount: 1,
	}
	iVal.Privileges[0] = LUID_AND_ATTRIBUTES{
		LUID:       pValue,
		Attributes: PROC_SE_PRIVILEGE_ENABLED,
	}
	ajRes, _, _ := procAdjustTokenPrivileges.Call(
		tHandle,
		uintptr(uint32(0)),
		uintptr(unsafe.Pointer(&iVal)),
		uintptr(uint32(0)),
		uintptr(0),
		uintptr(0),
	)
	if ajRes != 1 {
		return fmt.Errorf("error while adjusting process token")
	}
	return nil
}

func sessUserLUIDs() (map[LUID]string, error) {
	var (
		logonSessionCount uint64
		loginSessionList  uintptr
		sizeTest          LUID
		uList             = make(map[LUID]string)
	)

	_, _, _ = sessLsaEnumerateLogonSessions.Call(
		uintptr(unsafe.Pointer(&logonSessionCount)),
		uintptr(unsafe.Pointer(&loginSessionList)),
	)
	//goland:noinspection GoUnhandledErrorResult
	defer sessLsaFreeReturnBuffer.Call(uintptr(unsafe.Pointer(&loginSessionList)))

	//goland:noinspection GoRedundantConversion,GoVetUnsafePointer
	var iter = uintptr(unsafe.Pointer(loginSessionList))

	for i := uint64(0); i < logonSessionCount; i++ {
		var sessionData uintptr
		_, _, _ = sessLsaGetLogonSessionData.Call(iter, uintptr(unsafe.Pointer(&sessionData)))
		if sessionData != uintptr(0) {
			//goland:noinspection GoRedundantConversion,GoVetUnsafePointer
			var data = (*SECURITY_LOGON_SESSION_DATA)(unsafe.Pointer(sessionData))

			if data.Sid != uintptr(0) {
				uList[data.LogonId] = fmt.Sprintf("%s\\%s", strings.ToUpper(LsatoString(data.LogonDomain)), strings.ToLower(LsatoString(data.UserName)))
			}
		}

		//goland:noinspection GoRedundantConversion,GoVetUnsafePointer
		iter = uintptr(unsafe.Pointer(iter + unsafe.Sizeof(sizeTest)))
		//goland:noinspection GoRedundantConversion,GoVetUnsafePointer
		_, _, _ = sessLsaFreeReturnBuffer.Call(uintptr(unsafe.Pointer(sessionData)))
	}

	return uList, nil
}

func getProcessFullPathAndLUID(pid uint32) (string, LUID, error) {
	var fullpath string

	handle, _, _ := procOpenProcess.Call(uintptr(uint32(PROCESS_QUERY_INFORMATION)), uintptr(0), uintptr(pid))
	if handle < 0 {
		return "", LUID{}, syscall.GetLastError()
	}
	//goland:noinspection GoUnhandledErrorResult
	defer procCloseHandle.Call(handle)

	var pathName [MAX_FULL_PATH]uint16
	pathLength := uint32(MAX_FULL_PATH)
	ret, _, _ := procQueryFullProcessImageName.Call(handle, uintptr(0), uintptr(unsafe.Pointer(&pathName)), uintptr(unsafe.Pointer(&pathLength)))

	if ret > 0 {
		fullpath = syscall.UTF16ToString(pathName[:pathLength])
	}

	var tHandle uintptr
	opRes, _, _ := procOpenProcessToken.Call(
		handle,
		uintptr(uint32(PROC_TOKEN_QUERY)),
		uintptr(unsafe.Pointer(&tHandle)),
	)
	if opRes != 1 {
		return fullpath, LUID{}, fmt.Errorf("unable to open process token")
	}
	//goland:noinspection GoUnhandledErrorResult
	defer procCloseHandle.Call(tHandle)

	var sData TOKEN_STATISTICS
	var sLength uint32
	tsRes, _, _ := procGetTokenInformation.Call(
		tHandle,
		uintptr(uint32(10)), // TOKEN_STATISTICS
		uintptr(unsafe.Pointer(&sData)),
		uintptr(uint32(unsafe.Sizeof(sData))),
		uintptr(unsafe.Pointer(&sLength)),
	)
	if tsRes != 1 {
		return fullpath, LUID{}, fmt.Errorf("error fetching token information (LUID)")
	}

	return fullpath, sData.AuthenticationId, nil
}

func newProcessData(e *PROCESSENTRY32, path string, user string) Process {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return Process{
		Pid:        int(e.ProcessID),
		Ppid:       int(e.ParentProcessID),
		Executable: syscall.UTF16ToString(e.ExeFile[:end]),
		Fullpath:   path,
		Username:   user,
	}
}

func LsatoString(p LSA_UNICODE_STRING) string {
	//goland:noinspection GoVetUnsafePointer
	return syscall.UTF16ToString((*[4096]uint16)(unsafe.Pointer(p.buffer))[:p.Length])
}