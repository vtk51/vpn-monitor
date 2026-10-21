package main

// Win32 alpha is opacity (255 opaque), while the setting is transparency.
func transparencyAlpha(percent int) byte { return byte((255*(100-percent) + 50) / 100) }
func layeredStyle(style uintptr) uintptr { return style | 0x80000 }
