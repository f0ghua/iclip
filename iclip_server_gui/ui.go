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
	colorMonitoring = color.NRGBA{R: 76, G: 175, B: 80, A: 255}  // Green
	colorReceiving  = color.NRGBA{R: 255, G: 193, B: 7, A: 255}  // Yellow/Amber
	colorError      = color.NRGBA{R: 244, G: 67, B: 54, A: 255}  // Red
	colorPaused     = color.NRGBA{R: 158, G: 158, B: 158, A: 255} // Gray
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
	StatusIndicator *canvas.Circle
	StatusLabel     *widget.Label
	FromLabel       *widget.Label
	LastRecvLabel   *widget.Label
	ContentText     *widget.Label
	ErrorText       *widget.RichText
	PauseButton     *widget.Button
	SettingsButton  *widget.Button

	State         *UIState
	Config        *Config
	OnPauseToggle func()
}

// NewAppUI creates a new application UI
func NewAppUI(app fyne.App, config *Config) *AppUI {
	window := app.NewWindow("iclip Server")
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

	// Last received time
	ui.LastRecvLabel = widget.NewLabel("--:--:--")

	// Status row with indicator
	statusRow := container.NewHBox(
		container.NewWithoutLayout(ui.StatusIndicator),
		container.NewPadded(ui.StatusLabel),
		layout.NewSpacer(),
		widget.NewLabel("Last:"),
		ui.LastRecvLabel,
	)
	// Make indicator container proper size
	statusRow.Objects[0].(*fyne.Container).Resize(fyne.NewSize(16, 20))

	// From label
	ui.FromLabel = widget.NewLabel("-")
	fromRow := container.NewHBox(
		widget.NewLabel("From:"),
		ui.FromLabel,
	)

	// Status card content
	statusContent := container.NewVBox(statusRow, fromRow)

	// Status card
	statusCard := widget.NewCard("Status", "", statusContent)

	// Received content
	ui.ContentText = widget.NewLabel("")
	ui.ContentText.Wrapping = fyne.TextWrapWord
	ui.ContentText.TextStyle = fyne.TextStyle{}

	contentScroll := container.NewScroll(ui.ContentText)
	contentScroll.SetMinSize(fyne.NewSize(380, 110))

	// Content card
	contentCard := widget.NewCard("Received Content", "", contentScroll)

	// Error text (using RichText for colored text)
	ui.ErrorText = widget.NewRichText()

	// Buttons with icons
	ui.PauseButton = widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {
		if ui.OnPauseToggle != nil {
			ui.OnPauseToggle()
		}
	})
	ui.PauseButton.Importance = widget.MediumImportance

	ui.SettingsButton = widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		ui.showSettingsDialog()
	})

	buttonContainer := container.NewHBox(
		layout.NewSpacer(),
		ui.PauseButton,
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
		ui.StatusIndicator.FillColor = colorMonitoring
	} else {
		ui.StatusIndicator.FillColor = colorPaused
	}
	ui.StatusIndicator.Refresh()
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
			ui.StatusIndicator.FillColor = colorMonitoring
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
			ui.StatusIndicator.FillColor = colorMonitoring
			ui.StatusIndicator.Refresh()
		}
		ui.UpdateStatus("Monitoring", true)
	}
}

// SetReceiving shows receiving status
func (ui *AppUI) SetReceiving(receiving bool) {
	if receiving && !ui.State.Paused && ui.State.ErrorMessage == "" {
		ui.StatusIndicator.FillColor = colorReceiving
		ui.StatusIndicator.Refresh()
		ui.StatusLabel.SetText("Receiving...")
	} else if !ui.State.Paused && ui.State.ErrorMessage == "" {
		ui.StatusIndicator.FillColor = colorMonitoring
		ui.StatusIndicator.Refresh()
		ui.StatusLabel.SetText("Monitoring")
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
