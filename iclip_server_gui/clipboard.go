package main

import (
	"context"

	"golang.design/x/clipboard"
)

// ClipboardManager handles clipboard operations
type ClipboardManager struct {
	initialized bool
}

// NewClipboardManager creates a new clipboard manager
func NewClipboardManager() *ClipboardManager {
	return &ClipboardManager{initialized: false}
}

// Init initializes the clipboard
func (c *ClipboardManager) Init() error {
	if c.initialized {
		return nil
	}
	err := clipboard.Init()
	if err != nil {
		return err
	}
	c.initialized = true
	return nil
}

// GetText gets text from clipboard
func (c *ClipboardManager) GetText() string {
	if !c.initialized {
		if err := c.Init(); err != nil {
			return ""
		}
	}
	return string(clipboard.Read(clipboard.FmtText))
}

// SetText sets text to clipboard
func (c *ClipboardManager) SetText(text string) error {
	if !c.initialized {
		if err := c.Init(); err != nil {
			return err
		}
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}

// Watch watches for clipboard changes
func (c *ClipboardManager) Watch(ctx context.Context) <-chan []byte {
	if !c.initialized {
		if err := c.Init(); err != nil {
			return nil
		}
	}
	return clipboard.Watch(ctx, clipboard.FmtText)
}
