// Package app ...
package app

// <---------------------------------------------------------------------------------------------------->

import (
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/hotkey"
	"golang.org/x/sys/windows"

	"github.com/skillptm/Bolt/internal/searchhandler"
)

// <---------------------------------------------------------------------------------------------------->

// App holds all the main data and functions relevant to the front- and backend.
type App struct {
	CTX           context.Context
	hotkey        *hotkey.Hotkey
	images        map[string]string
	SearchHandler *searchhandler.SearchHandler
}

// NewApp creates a new App struct with all it's values.
func NewApp(images embed.FS) (*App, error) {
	imagePaths := map[string]string{
		"cross":            "frontend/src/assets/images/cross.png",
		"google":           "frontend/src/assets/images/google.png",
		"file":             "frontend/src/assets/images/file.png",
		"folder":           "frontend/src/assets/images/folder.png",
		"left":             "frontend/src/assets/images/left.png",
		"magnifying_glass": "frontend/src/assets/images/magnifying_glass.png",
		"not-left":         "frontend/src/assets/images/not_left.png",
		"right":            "frontend/src/assets/images/right.png",
		"not-right":        "frontend/src/assets/images/not_right.png",
		"tick":             "frontend/src/assets/images/tick.png",
	}

	imageMap := make(map[string]string)

	for name, path := range imagePaths {
		imageData, err := images.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("couldn't get image %s from embed: %s", path, err.Error())
		}

		imageMap[name] = "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData)
	}

	return &App{
		hotkey:        hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyS),
		images:        imageMap,
		SearchHandler: searchhandler.New(),
	}, nil
}

/*
Startup is called when the app starts. The context is saved so we can call the runtime methods.

It also starts the goroutine for emiting the search results to the frontend.
*/
func (a *App) Startup(CTX context.Context) {
	a.CTX = CTX
	go a.emitSearchResult()
	go a.openOnHotKey()
	go a.windowHideOnUnselected()
}

// <---------------------------------------------------------------------------------------------------->

// GetImageData receives a key (being the name of an image) and returns the base64 string data of that image
func (a *App) GetImageData(name string) string {
	if data, ok := a.images[name]; ok {
		return data
	}

	return ""
}

// LaunchSearch starts a search on the SearchHandler of the app
func (a *App) LaunchSearch(input string) {
	if len(input) < 1 {
		a.SearchHandler.ResultsChan <- []string{}
		return
	}

	a.SearchHandler.StartSearch(input)
}

// OpenFileExplorer allows you to open the file explorer at any entry's location
func (a *App) OpenFileExplorer(filePath string) {
	cmd := exec.Command("explorer", "/select,", strings.ReplaceAll(strings.TrimSuffix(filePath, "/"), "/", "\\"))
	cmd.Run()
}

// <---------------------------------------------------------------------------------------------------->

// emitSearchResult runs continuously and emits the search results with the "searchResult" event to the frontend
func (a *App) emitSearchResult() {
	for result := range a.SearchHandler.ResultsChan {
		runtime.EventsEmit(a.CTX, "searchResult", result)
	}
}

// openOnHotKey will unhide and reload the app when ctrl+shift+s is pressed
func (a *App) openOnHotKey() {
	err := a.hotkey.Register()
	if err != nil {
		log.Fatalf("main hotkey failed to register: %s", err)
		return
	}

	for range a.hotkey.Keydown() {
		runtime.WindowShow(a.CTX)
	}
}

// windowHideOnUnselected will hide the window once you unselected it, by clicking somewhere else
func (a *App) windowHideOnUnselected() {
	recheckTicker := time.NewTicker(100 * time.Millisecond)

	for range recheckTicker.C {
		// The functonality here was copied from: https://gist.github.com/obonyojimmy/d6b263212a011ac7682ac738b7fb4c70
		mod := windows.NewLazyDLL("user32.dll")

		proc := mod.NewProc("GetForegroundWindow")
		hwnd, _, _ := proc.Call()

		proc = mod.NewProc("GetWindowTextLengthW")
		ret, _, _ := proc.Call(hwnd)

		buf := make([]uint16, int(ret)+1)
		proc = mod.NewProc("GetWindowTextW")
		proc.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(int(ret)+1))

		title := syscall.UTF16ToString(buf)

		if title != "Quick Search" {
			runtime.WindowHide(a.CTX)
			runtime.EventsEmit(a.CTX, "hidApp")
		}
	}
}
