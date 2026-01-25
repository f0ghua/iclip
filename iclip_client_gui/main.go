package main

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

// Application holds the main application state
type Application struct {
	ui             *AppUI
	config         *Config
	lastClipboard  string
	paused         bool
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.Mutex
}

func main() {
	// Check dependencies
	if !CheckXclipAvailable() {
		fmt.Println("Error: xclip is not installed. Please install it with: sudo apt-get install xclip")
		return
	}
	if !CheckSmbclientAvailable() {
		fmt.Println("Error: smbclient is not installed. Please install it with: sudo apt-get install smbclient")
		return
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		// Create default config if not found
		config = &Config{
			SMB: SMBConfig{
				Server:    "192.168.1.1",
				Share:     "Share",
				Path:      "",
				Username:  "user",
				Password:  "password",
				Workgroup: "",
			},
			Monitor: MonitorConfig{
				Interval: 1.0,
			},
		}
	}

	// Create Fyne app
	fyneApp := app.New()

	// Create application
	application := &Application{
		config:   config,
		paused:   false,
		stopChan: make(chan struct{}),
	}

	// Create UI
	application.ui = NewAppUI(fyneApp, config)
	application.ui.OnPauseToggle = application.togglePause
	application.ui.OnRefresh = application.forceRefresh

	// Handle window close
	application.ui.Window.SetOnClosed(func() {
		application.stop()
	})

	// Show config error dialog if needed
	if err != nil {
		application.ui.Window.Show()
		dialog.ShowError(fmt.Errorf("Config error: %v\nPlease configure settings.", err), application.ui.Window)
	}

	// Start clipboard monitoring
	application.startMonitoring()

	// Run UI
	application.ui.Run()
}

// startMonitoring starts the clipboard monitoring goroutine
func (a *Application) startMonitoring() {
	a.wg.Add(1)
	go a.monitorLoop()
}

// stop stops the application
func (a *Application) stop() {
	close(a.stopChan)
	a.wg.Wait()
}

// togglePause toggles the pause state
func (a *Application) togglePause() {
	a.mu.Lock()
	a.paused = !a.paused
	paused := a.paused
	a.mu.Unlock()

	a.ui.SetPaused(paused)
}

// forceRefresh forces a clipboard check
func (a *Application) forceRefresh() {
	go a.checkAndTransfer(true)
}

// monitorLoop is the main monitoring loop
func (a *Application) monitorLoop() {
	defer a.wg.Done()

	// Initial status
	a.ui.UpdateStatus("Running", true)

	ticker := time.NewTicker(time.Duration(a.config.Monitor.Interval * float64(time.Second)))
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.mu.Lock()
			paused := a.paused
			a.mu.Unlock()

			if !paused {
				a.checkAndTransfer(false)
			}
		}
	}
}

// checkAndTransfer checks clipboard and transfers if changed
func (a *Application) checkAndTransfer(force bool) {
	// Get current clipboard
	clipboard, err := GetClipboardText()
	if err != nil {
		a.ui.UpdateError(fmt.Sprintf("Clipboard error: %v", err))
		return
	}

	// Update display
	a.ui.UpdateClipboardContent(clipboard)

	// Check if changed
	a.mu.Lock()
	changed := clipboard != a.lastClipboard
	if changed || force {
		a.lastClipboard = clipboard
	}
	a.mu.Unlock()

	if (changed || force) && clipboard != "" {
		// Transfer clipboard
		a.ui.UpdateStatus("Syncing...", true)
		err := TransferClipboard(clipboard, a.config)
		if err != nil {
			a.ui.UpdateError(fmt.Sprintf("Transfer error: %v", err))
			a.ui.UpdateStatus("Error", false)
		} else {
			a.ui.UpdateError("")
			a.ui.UpdateLastSync(time.Now())
			a.ui.UpdateStatus("Running", true)
		}
	}
}
