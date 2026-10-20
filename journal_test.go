package main

import "testing"

func TestJournalCaretUsesEndOfUTF16Text(t *testing.T) {
	if n := journalCaret("А\r\n😀"); n != 5 {
		t.Fatalf("caret = %d, want UTF16 text end 5", n)
	}
}
