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
	ui        *AppUI
	config    *Config
	clipboard *ClipboardManager
	paused    bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
	mu        sync.Mutex
}

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		// Create default config if not found
		config = &Config{
			Monitor: MonitorConfig{
				WatchPath: `\\192.168.1.1\Share`,
				Interval:  1.0,
				Pattern:   "clipboard_*.tar.gz",
			},
		}
	}

	// Create Fyne app
	fyneApp := app.New()
	fyneApp.SetIcon(AppIcon)

	// Create clipboard manager
	clipboardMgr := NewClipboardManager()

	// Create application
	application := &Application{
		config:    config,
		clipboard: clipboardMgr,
		paused:    false,
		stopChan:  make(chan struct{}),
	}

	// Create UI
	application.ui = NewAppUI(fyneApp, config)
	application.ui.OnPauseToggle = application.togglePause

	// Handle window close
	application.ui.Window.SetOnClosed(func() {
		application.stop()
	})

	// Initialize clipboard
	if err := clipboardMgr.Init(); err != nil {
		application.ui.Window.Show()
		dialog.ShowError(fmt.Errorf("Clipboard init error: %v", err), application.ui.Window)
	}

	// Show config error dialog if needed
	if err != nil {
		application.ui.Window.Show()
		dialog.ShowError(fmt.Errorf("Config error: %v\nPlease configure settings.", err), application.ui.Window)
	}

	// Start monitoring
	application.startMonitoring()

	// Run UI
	application.ui.Run()
}

// startMonitoring starts the directory monitoring goroutine
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

// monitorLoop is the main monitoring loop
func (a *Application) monitorLoop() {
	defer a.wg.Done()

	// Initial status
	a.ui.UpdateStatus("Monitoring", true)

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
				a.checkAndProcess()
			}
		}
	}
}

// checkAndProcess checks for new files and processes them
func (a *Application) checkAndProcess() {
	// Find clipboard files
	files, err := FindClipboardFiles(a.config.Monitor.WatchPath, a.config.Monitor.Pattern)
	if err != nil {
		a.ui.UpdateError(fmt.Sprintf("Monitor error: %v", err))
		return
	}

	// Clear error if successful
	if a.ui.State.ErrorMessage != "" && err == nil {
		a.ui.UpdateError("")
	}

	// Process each file
	for _, file := range files {
		a.processFile(file)
	}
}

// processFile processes a single clipboard file
func (a *Application) processFile(filePath string) {
	// Extract data
	data, err := ExtractClipboardData(filePath)
	if err != nil {
		a.ui.UpdateError(fmt.Sprintf("Extract error: %v", err))
		return
	}

	// Set clipboard
	if err := a.clipboard.SetText(data.Content); err != nil {
		a.ui.UpdateError(fmt.Sprintf("Clipboard error: %v", err))
		return
	}

	// Update UI
	a.ui.UpdateFrom(data.Username)
	a.ui.UpdateReceivedContent(data.Content)
	a.ui.UpdateLastReceived(time.Now())
	a.ui.UpdateError("")

	// Delete processed file
	if err := DeleteFile(filePath); err != nil {
		a.ui.UpdateError(fmt.Sprintf("Delete error: %v", err))
	}
}
