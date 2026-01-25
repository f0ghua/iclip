package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Status colors
var (
	colorRunning  = color.NRGBA{R: 76, G: 175, B: 80, A: 255}  // Green
	colorSyncing  = color.NRGBA{R: 255, G: 193, B: 7, A: 255}  // Yellow/Amber
	colorError    = color.NRGBA{R: 244, G: 67, B: 54, A: 255}  // Red
	colorPaused   = color.NRGBA{R: 158, G: 158, B: 158, A: 255} // Gray
)

// UIState holds the current UI state
type UIState struct {
	Status        string
	LastSync      time.Time
	ClipboardText string
	ErrorMessage  string
	Paused        bool
}

// AppUI holds all UI components
type AppUI struct {
	Window         fyne.Window
	StatusIndicator *canvas.Circle
	StatusLabel    *widget.Label
	LastSyncLabel  *widget.Label
	ClipboardText  *widget.Label
	ErrorText      *widget.RichText
	PauseButton    *widget.Button
	RefreshButton  *widget.Button
	SettingsButton *widget.Button

	State         *UIState
	Config        *Config
	OnPauseToggle func()
	OnRefresh     func()
}

// NewAppUI creates a new application UI
func NewAppUI(app fyne.App, config *Config) *AppUI {
	window := app.NewWindow("iclip Client")
	window.Resize(fyne.NewSize(420, 380))

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
	// Status indicator (colored circle)
	ui.StatusIndicator = canvas.NewCircle(colorPaused)
	ui.StatusIndicator.StrokeWidth = 0
	indicatorContainer := container.NewWithoutLayout(ui.StatusIndicator)
	indicatorContainer.Resize(fyne.NewSize(12, 12))
	ui.StatusIndicator.Resize(fyne.NewSize(12, 12))
	ui.StatusIndicator.Move(fyne.NewPos(0, 4))

	// Status label
	ui.StatusLabel = widget.NewLabel("Starting")
	ui.StatusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Last sync time
	ui.LastSyncLabel = widget.NewLabel("--:--:--")

	// Status row with indicator
	statusRow := container.NewHBox(
		container.NewWithoutLayout(ui.StatusIndicator),
		container.NewPadded(ui.StatusLabel),
		layout.NewSpacer(),
		widget.NewLabel("Last:"),
		ui.LastSyncLabel,
	)
	// Make indicator container proper size
	statusRow.Objects[0].(*fyne.Container).Resize(fyne.NewSize(16, 20))

	// Status card content
	statusContent := container.NewVBox(statusRow)

	// Status card
	statusCard := widget.NewCard("Status", "", statusContent)

	// Clipboard content
	ui.ClipboardText = widget.NewLabel("")
	ui.ClipboardText.Wrapping = fyne.TextWrapWord
	ui.ClipboardText.TextStyle = fyne.TextStyle{}

	clipboardScroll := container.NewScroll(ui.ClipboardText)
	clipboardScroll.SetMinSize(fyne.NewSize(380, 120))

	// Content card
	contentCard := widget.NewCard("Clipboard Content", "", clipboardScroll)

	// Error text (using RichText for colored text)
	ui.ErrorText = widget.NewRichText()

	// Buttons with icons
	ui.PauseButton = widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {
		if ui.OnPauseToggle != nil {
			ui.OnPauseToggle()
		}
	})
	ui.PauseButton.Importance = widget.MediumImportance

	ui.RefreshButton = widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		if ui.OnRefresh != nil {
			ui.OnRefresh()
		}
	})
	ui.RefreshButton.Importance = widget.HighImportance

	ui.SettingsButton = widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
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
	content := container.NewVBox(
		statusCard,
		ui.ErrorText,
		contentCard,
		layout.NewSpacer(),
		buttonContainer,
	)

	ui.Window.SetContent(container.NewPadded(content))
}

// UpdateStatus updates the status display
func (ui *AppUI) UpdateStatus(status string, isRunning bool) {
	ui.State.Status = status
	ui.StatusLabel.SetText(status)

	if ui.State.Paused {
		ui.StatusIndicator.FillColor = colorPaused
	} else if ui.State.ErrorMessage != "" {
		ui.StatusIndicator.FillColor = colorError
	} else if isRunning {
		ui.StatusIndicator.FillColor = colorRunning
	} else {
		ui.StatusIndicator.FillColor = colorPaused
	}
	ui.StatusIndicator.Refresh()
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
	if err != "" {
		ui.ErrorText.Segments = []widget.RichTextSegment{
			&widget.TextSegment{
				Text: err,
				Style: widget.RichTextStyle{
					ColorName: theme.ColorNameError,
					TextStyle: fyne.TextStyle{Bold: true},
				},
			},
		}
		ui.StatusIndicator.FillColor = colorError
		ui.StatusIndicator.Refresh()
	} else {
		ui.ErrorText.Segments = nil
		if !ui.State.Paused {
			ui.StatusIndicator.FillColor = colorRunning
			ui.StatusIndicator.Refresh()
		}
	}
	ui.ErrorText.Refresh()
}

// SetPaused updates the paused state
func (ui *AppUI) SetPaused(paused bool) {
	ui.State.Paused = paused
	if paused {
		ui.PauseButton.SetText("Resume")
		ui.PauseButton.SetIcon(theme.MediaPlayIcon())
		ui.StatusIndicator.FillColor = colorPaused
		ui.StatusIndicator.Refresh()
		ui.UpdateStatus("Paused", false)
	} else {
		ui.PauseButton.SetText("Pause")
		ui.PauseButton.SetIcon(theme.MediaPauseIcon())
		if ui.State.ErrorMessage == "" {
			ui.StatusIndicator.FillColor = colorRunning
			ui.StatusIndicator.Refresh()
		}
		ui.UpdateStatus("Running", true)
	}
}

// SetSyncing shows syncing status
func (ui *AppUI) SetSyncing(syncing bool) {
	if syncing && !ui.State.Paused && ui.State.ErrorMessage == "" {
		ui.StatusIndicator.FillColor = colorSyncing
		ui.StatusIndicator.Refresh()
		ui.StatusLabel.SetText("Syncing...")
	} else if !ui.State.Paused && ui.State.ErrorMessage == "" {
		ui.StatusIndicator.FillColor = colorRunning
		ui.StatusIndicator.Refresh()
		ui.StatusLabel.SetText("Running")
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
