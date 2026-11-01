//go:build windows

package tui

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	ntdll              = syscall.NewLazyDLL("ntdll.dll")
	procNtSuspend      = ntdll.NewProc("NtSuspendProcess")
	procNtResume       = ntdll.NewProc("NtResumeProcess")
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess    = kernel32.NewProc("OpenProcess")
	procCreateSnapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First = kernel32.NewProc("Process32FirstW")
	procProcess32Next  = kernel32.NewProc("Process32NextW")
	procGetConsoleWnd  = kernel32.NewProc("GetConsoleWindow")
	procGetModuleHdl   = kernel32.NewProc("GetModuleHandleW")
	user32             = syscall.NewLazyDLL("user32.dll")
	procLoadIconW      = user32.NewProc("LoadIconW")
	procSendMessageW   = user32.NewProc("SendMessageW")
)

const (
	processSuspendResume = 0x0800
	processTerminate     = 0x0001
	th32csSnapProcess    = 0x00000002
	wmSetIcon            = 0x0080
	iconSmall            = 0
	iconBig              = 1
)

type processEntry32 struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	Threads           uint32
	ParentProcessID   uint32
	PriClassBase      int32
	Flags             uint32
	ExeFile           [260]uint16
}

func collectProcessTreePIDs(rootPID uint32) []uint32 {
	pids := []uint32{rootPID}
	snap, _, _ := procCreateSnapshot.Call(th32csSnapProcess, 0)
	if snap == 0 || snap == ^uintptr(0) {
		return pids
	}
	defer syscall.CloseHandle(syscall.Handle(snap))

	var entry processEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32First.Call(snap, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return pids
	}

	childrenMap := make(map[uint32][]uint32)
	for {
		childrenMap[entry.ParentProcessID] = append(childrenMap[entry.ParentProcessID], entry.ProcessID)
		entry.Size = uint32(unsafe.Sizeof(entry))
		ret, _, _ = procProcess32Next.Call(snap, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	queue := []uint32{rootPID}
	seen := map[uint32]bool{rootPID: true}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, child := range childrenMap[curr] {
			if !seen[child] {
				seen[child] = true
				pids = append(pids, child)
				queue = append(queue, child)
			}
		}
	}
	return pids
}

func suspendOrResumeProcessTree(cmd *exec.Cmd, suspend bool) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	pids := collectProcessTreePIDs(uint32(cmd.Process.Pid))
	var firstErr error
	for _, pid := range pids {
		h, _, err := procOpenProcess.Call(processSuspendResume, 0, uintptr(pid))
		if h == 0 {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		var status uintptr
		if suspend {
			status, _, err = procNtSuspend.Call(h)
		} else {
			status, _, err = procNtResume.Call(h)
		}
		_ = syscall.CloseHandle(syscall.Handle(h))
		if status != 0 && firstErr == nil {
			firstErr = fmt.Errorf("ntstatus 0x%x: %v", status, err)
		}
	}
	return firstErr
}

func pauseProcess(cmd *exec.Cmd) error {
	return suspendOrResumeProcessTree(cmd, true)
}

func resumeProcess(cmd *exec.Cmd) error {
	return suspendOrResumeProcessTree(cmd, false)
}

func killProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = resumeProcess(cmd)
	pids := collectProcessTreePIDs(uint32(cmd.Process.Pid))
	for i := len(pids) - 1; i >= 0; i-- {
		pid := pids[i]
		h, _, _ := procOpenProcess.Call(processTerminate, 0, uintptr(pid))
		if h != 0 {
			_ = syscall.TerminateProcess(syscall.Handle(h), 1)
			_ = syscall.CloseHandle(syscall.Handle(h))
		}
	}
	return cmd.Process.Kill()
}

func applyPlatformWindowSettings() {
	hwnd, _, _ := procGetConsoleWnd.Call()
	if hwnd == 0 {
		return
	}
	hInstance, _, _ := procGetModuleHdl.Call(0)
	if hInstance == 0 {
		return
	}
	// Resource ID 1 is the embedded RT_GROUP_ICON in rsrc_windows_amd64.syso
	hIcon, _, _ := procLoadIconW.Call(hInstance, uintptr(1))
	if hIcon != 0 {
		procSendMessageW.Call(hwnd, wmSetIcon, iconSmall, hIcon)
		procSendMessageW.Call(hwnd, wmSetIcon, iconBig, hIcon)
	}
}
