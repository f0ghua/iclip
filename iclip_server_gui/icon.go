package main

import "fyne.io/fyne/v2"

// Clipboard icon SVG for server (download arrow)
var iconData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="64" height="64" viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg">
  <!-- Clipboard background -->
  <rect x="12" y="8" width="40" height="52" rx="4" fill="#2196F3"/>
  <!-- Clipboard clip -->
  <rect x="22" y="4" width="20" height="8" rx="2" fill="#1976D2"/>
  <!-- Paper -->
  <rect x="16" y="16" width="32" height="40" rx="2" fill="white"/>
  <!-- Download arrow -->
  <path d="M32 24 L32 46 M24 38 L32 46 L40 38" stroke="#2196F3" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
</svg>`)

// AppIcon returns the application icon
var AppIcon = &fyne.StaticResource{
	StaticName:    "icon.svg",
	StaticContent: iconData,
}
