package main

type compactSize struct{ Width, Height int }
type controlRect struct{ X, Y, W, H int }

func compactDimensions(expanded bool) compactSize {
	h := 353
	if expanded {
		h += 184
	}
	return compactSize{380, h}
}
func compactControls() []controlRect {
	return []controlRect{{274, 6, 28, 27}, {305, 6, 28, 27}, {336, 6, 28, 27}, {76, 281, 60, 25}, {161, 281, 31, 25}, {334, 281, 30, 25}, {13, 325, 143, 23}}
}
