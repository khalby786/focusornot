package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/getlantern/systray"
	hook "github.com/robotn/gohook"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	// Tells Windows we want to read the current state of active window tracking
	SPI_GETACTIVEWINDOWTRACKING = 0x1000

	// Tells Windows we want to change the state of active window tracking
	SPI_SETACTIVEWINDOWTRACKING = 0x1001

	// Flag instructing Windows to write this change permanently to the user profile registry so it stays even after restart
	SPIF_UPDATEINIFILE = 0x01

	// Flag instructing Windows to broadcast the change to every open window when changes happen
	SPIF_SENDCHANGE = 0x02

	// Combine both the flags with a bitwise OR
	SPIF_FLAGS = SPIF_UPDATEINIFILE | SPIF_SENDCHANGE

	AppName = "FocusOrNot?"
)

var (
	// Access user32 cause it has SystemParametersInfo that we need
	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfow
	modUser32               = windows.NewLazySystemDLL("user32.dll")
	procSystemParamtersInfo = modUser32.NewProc("SystemParametersInfoW")
)

//go:embed active.ico
var activeWindowTrackingEnabledIcon []byte

//go:embed disabled.ico
var activeWindowTrackingDisabledIcon []byte

func getActiveWindowTrackingStatus() uint32 {
	var state uint32

	ret, _, err := procSystemParamtersInfo.Call(
		// The setting we wanna get, so the active window tracking
		uintptr(SPI_GETACTIVEWINDOWTRACKING),

		// uiParam, no clue what this is but apparently it should be 0
		0,

		// Pointer to our state
		uintptr(unsafe.Pointer(&state)),

		// Change propagation, we don't want anyone to know when we get info
		0,
	)

	if ret == 0 {
		fmt.Printf("Error reading the current system state: %v\n", err)
	}

	return state
}

func setActiveWindowTrackingStatus(state uint32) {
	ret, _, err := procSystemParamtersInfo.Call(
		// The setting we wanna get, so the active window tracking
		uintptr(SPI_SETACTIVEWINDOWTRACKING),

		// Again no clue what this is, some configuration input
		0,

		// Our state
		uintptr(state),

		// Write to disk and notify all windows asap
		uintptr(SPIF_FLAGS),
	)

	if ret == 0 {
		fmt.Printf("Error setting system status: %v\n", err)
	}
}

func updateTrayIcon(state uint32) {
	if state == 1 {
		systray.SetIcon(activeWindowTrackingEnabledIcon)
	} else {
		systray.SetIcon(activeWindowTrackingDisabledIcon)
	}
}

func toggleActiveWindowTrackingStatus() {
	currentState := getActiveWindowTrackingStatus()
	var newState uint32 = 0
	if currentState == 0 {
		newState = 1
		fmt.Println("Active window tracking enabled!")
	} else {
		newState = 0
		fmt.Println("Active window tracking disabled")
	}

	setActiveWindowTrackingStatus(newState)
	updateTrayIcon(newState)
}

func isStartupEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)

	if err != nil {
		return false
	}

	defer k.Close()

	_, _, err = k.GetStringValue(AppName)
	return err == nil
}

func toggleStartup(menuStartup *systray.MenuItem) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE|registry.QUERY_VALUE)

	if err != nil {
		fmt.Printf("Error opening registry key: %v\n", err)
	}

	defer k.Close()

	if isStartupEnabled() {
		// Disable startup then
		err = k.DeleteValue(AppName)
		if err == nil {
			menuStartup.Uncheck()
		}
	} else {
		// Enable startup otherwise

		// Find the path of this program
		execPath, err := os.Executable()

		if err != nil {
			return
		}

		// Write the path of the program to the Windows startup registry
		err = k.SetStringValue(AppName, fmt.Sprintf(`"%s"`, filepath.Clean(execPath)))

		if err == nil {
			menuStartup.Check()
		}
	}
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("FocusOrNot?")
	systray.SetTooltip("Toggle Active Window Tracking")
	updateTrayIcon(getActiveWindowTrackingStatus())

	menuToggle := systray.AddMenuItem("Toggle tracking", "Enable/disable window tracking manually")
	menuStartup := systray.AddMenuItemCheckbox("Run at startup", "Launch this app when Window starts", isStartupEnabled())
	systray.AddSeparator()
	menuQuit := systray.AddMenuItem("Exit", "Close the application completely")

	go func() {
		fmt.Println("Listening for keybind...")

		hook.Register(
			hook.KeyDown,
			[]string{"ctrl", "alt", "x"}, func(e hook.Event) {
				toggleActiveWindowTrackingStatus()
			},
		)

		s := hook.Start()
		<-hook.Process(s)
	}()

	go func() {
		for {
			select {
			case <-menuToggle.ClickedCh:
				toggleActiveWindowTrackingStatus()
			case <-menuStartup.ClickedCh:
				toggleStartup(menuStartup)
			case <-menuQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	hook.End()
	fmt.Println("Application exited.")
}
