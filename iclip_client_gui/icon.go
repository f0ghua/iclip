package main

import "fyne.io/fyne/v2"

// Clipboard icon SVG for client (upload arrow)
var iconData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="64" height="64" viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg">
  <!-- Clipboard background -->
  <rect x="12" y="8" width="40" height="52" rx="4" fill="#4CAF50"/>
  <!-- Clipboard clip -->
  <rect x="22" y="4" width="20" height="8" rx="2" fill="#388E3C"/>
  <!-- Paper -->
  <rect x="16" y="16" width="32" height="40" rx="2" fill="white"/>
  <!-- Upload arrow -->
  <path d="M32 24 L32 46 M24 32 L32 24 L40 32" stroke="#4CAF50" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
</svg>`)

// AppIcon returns the application icon
var AppIcon = &fyne.StaticResource{
	StaticName:    "icon.svg",
	StaticContent: iconData,
}
