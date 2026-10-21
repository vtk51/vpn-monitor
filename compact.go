package main

type compactSize struct{ Width, Height int }
type controlRect struct{ X, Y, W, H int }

func compactDimensions(expanded bool) compactSize {
	h := 310
	if expanded {
		h += 184
	}
	return compactSize{380, h}
}
func compactControls() []controlRect {
	return []controlRect{{274, 6, 28, 27}, {305, 6, 28, 27}, {336, 6, 28, 27}, {13, 282, 143, 23}}
}
