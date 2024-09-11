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
	"math/rand"
)

// GenerateCaptchaImage takes a verification code string and generates an image
// to use for verifying non-email users.
func GenerateCaptchaImage(code string, isCLI bool) (string, error) {
	// Add spaces to code for legibility
	//code = strings.Join(strings.Split(code, ""), " ")
	img := image.NewRGBA(image.Rect(0, 0, 45, 25))

	addNoise(img, isCLI)
	addLabel(img, 2, 17, code)
	addNoise(img, isCLI)

	// Resize
	dst := image.NewRGBA(image.Rect(0, 0, 250, 100))
	draw.NearestNeighbor.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, dst, &jpeg.Options{Quality: 100}); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{G: 255, A: 255}
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func addNoise(img *image.RGBA, isCLI bool) {
	if isCLI {
		return
	}

	bounds := img.Bounds()
	draw.Draw(img, bounds, img, bounds.Min, draw.Src)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if rand.Float64() < 0.07 {
				noiseColor := color.RGBA{G: 100, B: 255, A: 255}
				if rand.Intn(2) == 0 {
					noiseColor = color.RGBA{G: 100, R: 255, A: 255}
				}

				img.Set(x, y, noiseColor)
			}
		}
	}
}
