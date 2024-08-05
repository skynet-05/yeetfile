package auth

import (
	"bytes"
	"encoding/base64"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/jpeg"
)

// GenerateCaptchaImage takes a verification code string and generates an image
// to use for verifying non-email users.
func GenerateCaptchaImage(code string) string {
	// Add spaces to code for legibility
	//code = strings.Join(strings.Split(code, ""), " ")
	img := image.NewRGBA(image.Rect(0, 0, 45, 25))
	addLabel(img, 2, 17, code)

	// Resize
	dst := image.NewRGBA(image.Rect(0, 0, 250, 100))
	draw.NearestNeighbor.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, dst, &jpeg.Options{Quality: 100}); err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}
