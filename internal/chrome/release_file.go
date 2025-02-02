package chrome

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	rstrtmgr                = syscall.NewLazyDLL("rstrtmgr.dll")
	procRmStartSession      = rstrtmgr.NewProc("RmStartSession")
	procRmRegisterResources = rstrtmgr.NewProc("RmRegisterResources")
	procRmGetList           = rstrtmgr.NewProc("RmGetList")
	procRmShutdown          = rstrtmgr.NewProc("RmShutdown")
	procRmEndSession        = rstrtmgr.NewProc("RmEndSession")
)

const (
	CCH_RM_SESSION_KEY  = 64
	ERROR_SUCCESS       = 0
	ERROR_MORE_DATA     = 234
	CCH_RM_MAX_APP_NAME = 255
	CCH_RM_MAX_SVC_NAME = 63
	CCH_RM_MAX_APP_ID   = 255
)

type RM_UNIQUE_PROCESS struct {
	ProcessId uint32
	StartTime FILETIME
}

type RM_PROCESS_INFO struct {
	AppType          uint32
	Process          RM_UNIQUE_PROCESS
	strAppName       [CCH_RM_MAX_APP_NAME + 1]uint16
	strServiceShort  [CCH_RM_MAX_SVC_NAME + 1]uint16
	strApplicationID [CCH_RM_MAX_APP_ID + 1]uint16
}

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

func ReleaseFileLock(filePath string) bool {
	var sessionHandle uint32
	var sessionKeyBuffer [CCH_RM_SESSION_KEY]uint16            // Adjust the buffer size as needed
	copy(sessionKeyBuffer[:], syscall.StringToUTF16(filePath)) // Set your session key here
	sessionKey := &sessionKeyBuffer[0]

	err := RmStartSession(&sessionHandle, 0, sessionKey)
	if err == nil {
		err = RmRegisterResources(sessionHandle, []string{filePath})
		if err == nil {
			processInfo, err := RmGetList(sessionHandle)
			if err == nil {
				for _, info := range processInfo {
					// Shutdown the process (forceful termination)
					shutdownResult := RmShutdown(sessionHandle, info.Process.ProcessId, 0)
					if shutdownResult != nil {
						RmEndSession(sessionHandle)
						return false
					} else {
						RmEndSession(sessionHandle)
						return true
					}
				}
			}
		}
	}
	RmEndSession(sessionHandle)
	return false
}

func RmStartSession(pSessionHandle *uint32, dwSessionFlags uint32, sessionKey *uint16) error {
	ret, _, _ := procRmStartSession.Call(
		uintptr(unsafe.Pointer(pSessionHandle)),
		uintptr(dwSessionFlags),
		uintptr(unsafe.Pointer(sessionKey)),
	)

	if ret != ERROR_SUCCESS {
		return fmt.Errorf("RmStartSession failed with error code %d", ret)
	}

	return nil
}

func RmRegisterResources(sessionHandle uint32, filePaths []string) error {
	var filePointers []*uint16
	for _, filePath := range filePaths {
		filePointer, err := syscall.UTF16PtrFromString(filePath)
		if err != nil {
			return err
		}
		filePointers = append(filePointers, filePointer)
	}

	ret, _, _ := procRmRegisterResources.Call(
		uintptr(sessionHandle),
		uintptr(len(filePaths)),
		uintptr(unsafe.Pointer(&filePointers[0])),
		0,
		0,
		0,
	)

	if ret != ERROR_SUCCESS {
		return fmt.Errorf("RmRegisterResources failed with error code %d", ret)
	}

	return nil
}

func RmGetList(sessionHandle uint32) ([]RM_PROCESS_INFO, error) {
	var (
		pnProcInfoNeeded, pnProcInfo, lpdwRebootReasons uint32
		rgAffectedApps                                  []RM_PROCESS_INFO
	)

	// First call to get the required buffer size
	result, _, _ := procRmGetList.Call(
		uintptr(sessionHandle),
		uintptr(unsafe.Pointer(&pnProcInfoNeeded)),
		uintptr(unsafe.Pointer(&pnProcInfo)),
		uintptr(0), // Pass nil since we're only interested in the required size
		uintptr(unsafe.Pointer(&lpdwRebootReasons)),
	)

	if result != ERROR_SUCCESS && result != ERROR_MORE_DATA {
		return nil, fmt.Errorf("RmGetList failed with error code %d", result)
	}

	if pnProcInfoNeeded == 0 {
		return nil, nil // No process information available
	}

	// Temporary buffer to retrieve the process information
	tempBuffer := make([]RM_PROCESS_INFO, pnProcInfoNeeded)

	// Second call to get the actual process information
	result, _, _ = procRmGetList.Call(
		uintptr(sessionHandle),
		uintptr(unsafe.Pointer(&pnProcInfoNeeded)),
		uintptr(unsafe.Pointer(&pnProcInfo)),
		uintptr(unsafe.Pointer(&(tempBuffer[0]))), // Pass the address of the first element
		uintptr(unsafe.Pointer(&lpdwRebootReasons)),
	)

	if result != ERROR_SUCCESS {
		return nil, fmt.Errorf("RmGetList failed with error code %d", result)
	}

	// Resize the slice to the actual number of elements
	rgAffectedApps = tempBuffer[:pnProcInfo]

	return rgAffectedApps, nil
}

func RmShutdown(sessionHandle uint32, flags uint32, fnStatusCallback uintptr) error {
	ret, _, _ := procRmShutdown.Call(
		uintptr(sessionHandle),
		uintptr(flags),
		fnStatusCallback,
	)

	if ret != ERROR_SUCCESS {
		return fmt.Errorf("RmShutdown failed with error code %d", ret)
	}

	return nil
}

func RmEndSession(sessionHandle uint32) error {
	ret, _, _ := procRmEndSession.Call(uintptr(sessionHandle))

	if ret != ERROR_SUCCESS {
		return fmt.Errorf("RmEndSession failed with error code %d", ret)
	}

	return nil
}
