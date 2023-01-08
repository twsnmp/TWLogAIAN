package main

import (
	"embed"
	"log"
	"os"
	"os/exec"

	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

var version = "v0.0.0"
var commit = ""

func main() {
	// Create an instance of the app structure
	app := NewApp()
	if len(os.Args) > 1 && os.Args[1] == "debug" {
		debug = true
		log.Println("debug mode on")
	}
	mainMenu := menu.NewMenu()
	mainMenu.Append(menu.AppMenu())
	fileMenu := mainMenu.AddSubmenu("File")
	fileMenu.AddText("&New", keys.CmdOrCtrl("n"), newWindow)
	mainMenu.Append(menu.EditMenu())
	// Create application with options
	err := wails.Run(&options.App{
		Title:             "TWLogAIAN",
		Width:             1024,
		Height:            900,
		MinWidth:          720,
		MinHeight:         570,
		Menu:              mainMenu,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 33, G: 37, B: 43, A: 255},
		Assets:            assets,
		LogLevel:          logger.DEBUG,
		OnStartup:         app.startup,
		OnDomReady:        app.domReady,
		OnShutdown:        app.shutdown,
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			// TitleBar:             mac.TitleBarHiddenInset(),
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       false,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "TWLogAIAN",
				Message: "© 2021 Masayuki Yamai",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func newWindow(cd *menu.CallbackData) {
	if ex, err := os.Executable(); err == nil {
		exec.Command(ex).Start()
	}
}
