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
	LastSync        time.Time
	ClipboardText   string
	ErrorMessage    string
	Paused          bool
}

// AppUI holds all UI components
type AppUI struct {
	Window          fyne.Window
	StatusLabel     *widget.Label
	StatusIcon      *widget.Icon
	LastSyncLabel   *widget.Label
	ClipboardText   *widget.Label
	ErrorLabel      *widget.Label
	PauseButton     *widget.Button
	RefreshButton   *widget.Button
	SettingsButton  *widget.Button

	State           *UIState
	Config          *Config
	OnPauseToggle   func()
	OnRefresh       func()
}

// NewAppUI creates a new application UI
func NewAppUI(app fyne.App, config *Config) *AppUI {
	window := app.NewWindow("iclip Client")
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

	// Last sync time
	ui.LastSyncLabel = widget.NewLabel("Never")
	lastSyncContainer := container.NewHBox(
		widget.NewLabel("Last sync:"),
		ui.LastSyncLabel,
	)

	// Error label
	ui.ErrorLabel = widget.NewLabel("")
	ui.ErrorLabel.Wrapping = fyne.TextWrapWord

	// Top status section
	topSection := container.NewVBox(
		statusContainer,
		lastSyncContainer,
		ui.ErrorLabel,
	)

	// Clipboard content section
	clipboardHeader := widget.NewLabel("Clipboard Content:")
	ui.ClipboardText = widget.NewLabel("")
	ui.ClipboardText.Wrapping = fyne.TextWrapWord

	clipboardScroll := container.NewScroll(ui.ClipboardText)
	clipboardScroll.SetMinSize(fyne.NewSize(380, 150))

	clipboardSection := container.NewBorder(
		clipboardHeader, nil, nil, nil,
		clipboardScroll,
	)

	// Buttons
	ui.PauseButton = widget.NewButton("Pause", func() {
		if ui.OnPauseToggle != nil {
			ui.OnPauseToggle()
		}
	})

	ui.RefreshButton = widget.NewButton("Refresh", func() {
		if ui.OnRefresh != nil {
			ui.OnRefresh()
		}
	})

	ui.SettingsButton = widget.NewButton("Settings", func() {
		ui.showSettingsDialog()
	})

	buttonContainer := container.NewHBox(
		layout.NewSpacer(),
		ui.PauseButton,
		ui.RefreshButton,
		ui.SettingsButton,
		layout.NewSpacer(),
	)

	// Main layout
	content := container.NewBorder(
		topSection,
		buttonContainer,
		nil,
		nil,
		clipboardSection,
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

// UpdateLastSync updates the last sync time display
func (ui *AppUI) UpdateLastSync(t time.Time) {
	ui.State.LastSync = t
	ui.LastSyncLabel.SetText(t.Format("15:04:05"))
}

// UpdateClipboardContent updates the clipboard content display
func (ui *AppUI) UpdateClipboardContent(text string) {
	ui.State.ClipboardText = text
	// Truncate for display
	displayText := truncateText(text, 10, 50)
	ui.ClipboardText.SetText(displayText)
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
		ui.UpdateStatus("Running", true)
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
	// SMB settings
	serverEntry := widget.NewEntry()
	serverEntry.SetText(ui.Config.SMB.Server)

	shareEntry := widget.NewEntry()
	shareEntry.SetText(ui.Config.SMB.Share)

	pathEntry := widget.NewEntry()
	pathEntry.SetText(ui.Config.SMB.Path)

	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(ui.Config.SMB.Username)

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(ui.Config.SMB.Password)

	workgroupEntry := widget.NewEntry()
	workgroupEntry.SetText(ui.Config.SMB.Workgroup)

	intervalEntry := widget.NewEntry()
	intervalEntry.SetText(fmt.Sprintf("%.1f", ui.Config.Monitor.Interval))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Server", Widget: serverEntry},
			{Text: "Share", Widget: shareEntry},
			{Text: "Path", Widget: pathEntry},
			{Text: "Username", Widget: usernameEntry},
			{Text: "Password", Widget: passwordEntry},
			{Text: "Workgroup", Widget: workgroupEntry},
			{Text: "Interval (s)", Widget: intervalEntry},
		},
	}

	dialog.ShowCustomConfirm("Settings", "Save", "Cancel", form, func(save bool) {
		if save {
			ui.Config.SMB.Server = serverEntry.Text
			ui.Config.SMB.Share = shareEntry.Text
			ui.Config.SMB.Path = pathEntry.Text
			ui.Config.SMB.Username = usernameEntry.Text
			ui.Config.SMB.Password = passwordEntry.Text
			ui.Config.SMB.Workgroup = workgroupEntry.Text

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
