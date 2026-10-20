package main

// HWND_TOPMOST/NOTOPMOST; SWP_NOSIZE | SWP_NOMOVE | SWP_NOACTIVATE.
func pinRequest(enable bool) (uintptr, uintptr) {
	if enable {
		return ^uintptr(0), 0x13
	}
	return ^uintptr(1), 0x13
}
