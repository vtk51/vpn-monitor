package main

import "unicode/utf16"

func journalCaret(s string) int { return len(utf16.Encode([]rune(s))) }
