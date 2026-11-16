package board

import (
	"embed"
	"fmt"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed assets/fonts/*.ttf
var assetsFS embed.FS

type fontWeight uint8

const (
	weightRegular fontWeight = iota
	weightSemiBold
	weightBold
)

var (
	loadOnce     sync.Once
	loadErr      error
	regularFont  *opentype.Font
	semiBoldFont *opentype.Font
	boldFont     *opentype.Font
)

func ensureAssetsLoaded() error {
	loadOnce.Do(func() {
		var err error
		regularFont, err = loadTTF("assets/fonts/Rubik-Regular.ttf")
		if err != nil {
			loadErr = err
			return
		}
		semiBoldFont, err = loadTTF("assets/fonts/Rubik-SemiBold.ttf")
		if err != nil {
			loadErr = err
			return
		}
		boldFont, err = loadTTF("assets/fonts/Rubik-Bold.ttf")
		if err != nil {
			loadErr = err
			return
		}
	})
	return loadErr
}

func loadTTF(path string) (*opentype.Font, error) {
	data, err := assetsFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read font %s: %w", path, err)
	}
	f, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse font %s: %w", path, err)
	}
	return f, nil
}

func newFontFace(w fontWeight, size float64) (font.Face, error) {
	if err := ensureAssetsLoaded(); err != nil {
		return nil, err
	}
	f := regularFont
	switch w {
	case weightSemiBold:
		f = semiBoldFont
	case weightBold:
		f = boldFont
	}
	return opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}
