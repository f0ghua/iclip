package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// UIState holds the current UI state
type UIState struct {
	Status          string
	LastReceived    time.Time
	ReceivedContent string
	ReceivedFrom    string
	ErrorMessage    string
	Paused          bool
}

// AppUI holds all UI components
type AppUI struct {
	Window          fyne.Window
	StatusLabel     *widget.Label
	StatusIcon      *widget.Icon
	FromLabel       *widget.Label
	LastRecvLabel   *widget.Label
	ContentText     *widget.Label
	ErrorLabel      *widget.Label
	PauseButton     *widget.Button
	SettingsButton  *widget.Button

	State           *UIState
	Config          *Config
	OnPauseToggle   func()
}

// NewAppUI creates a new application UI
func NewAppUI(app fyne.App, config *Config) *AppUI {
	window := app.NewWindow("iclip Server")
	window.Resize(fyne.NewSize(400, 350))

	ui := &AppUI{
		Window: window,
		State: &UIState{
			Status: "Starting",
			Paused: false,
		},
		Config: config,
	}

	ui.buildUI()
	return ui
}

// buildUI constructs the UI layout
func (ui *AppUI) buildUI() {
	// Status section
	ui.StatusIcon = widget.NewIcon(theme.MediaRecordIcon())
	ui.StatusLabel = widget.NewLabel("Starting")
	statusContainer := container.NewHBox(
		widget.NewLabel("Status:"),
		ui.StatusIcon,
		ui.StatusLabel,
	)

	// From label
	ui.FromLabel = widget.NewLabel("-")
	fromContainer := container.NewHBox(
		widget.NewLabel("From:"),
		ui.FromLabel,
	)

	// Last received time
	ui.LastRecvLabel = widget.NewLabel("Never")
	lastRecvContainer := container.NewHBox(
		widget.NewLabel("Last received:"),
		ui.LastRecvLabel,
	)

	// Error label
	ui.ErrorLabel = widget.NewLabel("")
	ui.ErrorLabel.Wrapping = fyne.TextWrapWord

	// Top status section
	topSection := container.NewVBox(
		statusContainer,
		fromContainer,
		lastRecvContainer,
		ui.ErrorLabel,
	)

	// Received content section
	contentHeader := widget.NewLabel("Received Content:")
	ui.ContentText = widget.NewLabel("")
	ui.ContentText.Wrapping = fyne.TextWrapWord

	contentScroll := container.NewScroll(ui.ContentText)
	contentScroll.SetMinSize(fyne.NewSize(380, 130))

	contentSection := container.NewBorder(
		contentHeader, nil, nil, nil,
		contentScroll,
	)

	// Buttons
	ui.PauseButton = widget.NewButton("Pause", func() {
		if ui.OnPauseToggle != nil {
			ui.OnPauseToggle()
		}
	})

	ui.SettingsButton = widget.NewButton("Settings", func() {
		ui.showSettingsDialog()
	})

	buttonContainer := container.NewHBox(
		layout.NewSpacer(),
		ui.PauseButton,
		ui.SettingsButton,
		layout.NewSpacer(),
	)

	// Main layout
	content := container.NewBorder(
		topSection,
		buttonContainer,
		nil,
		nil,
		contentSection,
	)

	ui.Window.SetContent(container.NewPadded(content))
}

// UpdateStatus updates the status display
func (ui *AppUI) UpdateStatus(status string, isRunning bool) {
	ui.State.Status = status
	ui.StatusLabel.SetText(status)
	if isRunning {
		ui.StatusIcon.SetResource(theme.MediaRecordIcon())
	} else {
		ui.StatusIcon.SetResource(theme.MediaStopIcon())
	}
}

// UpdateFrom updates the sender display
func (ui *AppUI) UpdateFrom(from string) {
	ui.State.ReceivedFrom = from
	ui.FromLabel.SetText(from)
}

// UpdateLastReceived updates the last received time display
func (ui *AppUI) UpdateLastReceived(t time.Time) {
	ui.State.LastReceived = t
	ui.LastRecvLabel.SetText(t.Format("15:04:05"))
}

// UpdateReceivedContent updates the received content display
func (ui *AppUI) UpdateReceivedContent(text string) {
	ui.State.ReceivedContent = text
	// Truncate for display
	displayText := truncateText(text, 8, 50)
	ui.ContentText.SetText(displayText)
}

// UpdateError updates the error message display
func (ui *AppUI) UpdateError(err string) {
	ui.State.ErrorMessage = err
	ui.ErrorLabel.SetText(err)
}

// SetPaused updates the paused state
func (ui *AppUI) SetPaused(paused bool) {
	ui.State.Paused = paused
	if paused {
		ui.PauseButton.SetText("Resume")
		ui.UpdateStatus("Paused", false)
	} else {
		ui.PauseButton.SetText("Pause")
		ui.UpdateStatus("Monitoring", true)
	}
}

// truncateText truncates text for display
func truncateText(text string, maxLines int, maxWidth int) string {
	lines := strings.Split(text, "\n")
	var result []string

	for i, line := range lines {
		if i >= maxLines {
			result = append(result, "...")
			break
		}
		if len(line) > maxWidth {
			line = line[:maxWidth-3] + "..."
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// showSettingsDialog shows the settings dialog
func (ui *AppUI) showSettingsDialog() {
	watchPathEntry := widget.NewEntry()
	watchPathEntry.SetText(ui.Config.Monitor.WatchPath)

	patternEntry := widget.NewEntry()
	patternEntry.SetText(ui.Config.Monitor.Pattern)

	intervalEntry := widget.NewEntry()
	intervalEntry.SetText(fmt.Sprintf("%.1f", ui.Config.Monitor.Interval))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Watch Path", Widget: watchPathEntry},
			{Text: "Pattern", Widget: patternEntry},
			{Text: "Interval (s)", Widget: intervalEntry},
		},
	}

	dialog.ShowCustomConfirm("Settings", "Save", "Cancel", form, func(save bool) {
		if save {
			ui.Config.Monitor.WatchPath = watchPathEntry.Text
			ui.Config.Monitor.Pattern = patternEntry.Text

			var interval float64
			fmt.Sscanf(intervalEntry.Text, "%f", &interval)
			if interval > 0 {
				ui.Config.Monitor.Interval = interval
			}

			if err := SaveConfig(ui.Config); err != nil {
				dialog.ShowError(err, ui.Window)
			} else {
				dialog.ShowInformation("Settings", "Settings saved successfully", ui.Window)
			}
		}
	}, ui.Window)
}

// Run shows the window and starts the UI
func (ui *AppUI) Run() {
	ui.Window.ShowAndRun()
}

// Show shows the window
func (ui *AppUI) Show() {
	ui.Window.Show()
}
