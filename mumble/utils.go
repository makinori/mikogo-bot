package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"net/url"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"
	"golang.org/x/image/draw"
	"layeh.com/gumble/gumble"
)

// TODO: should we use *image.Image instead

func setTransparentPage(page *rod.Page) error {
	return (proto.EmulationSetDefaultBackgroundColorOverride{
		Color: &proto.DOMRGBA{
			R: 0,
			G: 0,
			B: 0,
			A: gson.Num(0),
		},
	}).Call(page)
}

func setHTML(page *rod.Page, html string) error {
	return (proto.PageSetDocumentContent{
		HTML: html,
	}.Call(page))
}

func cropImage(src image.Image, bounds image.Rectangle) image.Image {
	return src.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(bounds)
}

func trimPngByTransparency(src image.Image) image.Image {
	imageBounds := src.Bounds()

	minX := imageBounds.Max.X
	minY := imageBounds.Max.Y
	maxX := imageBounds.Min.X
	maxY := imageBounds.Min.Y

	for y := imageBounds.Min.Y; y < imageBounds.Max.Y; y++ {
		for x := imageBounds.Min.X; x < imageBounds.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			if a > 0 {

				if x > maxX {
					maxX = x
				}

				if y > maxY {
					maxY = y
				}
			}
		}
	}

	for y := imageBounds.Max.Y; y >= imageBounds.Min.Y; y-- {
		for x := imageBounds.Max.X; x >= imageBounds.Min.X; x-- {
			_, _, _, a := src.At(x, y).RGBA()
			if a > 0 {
				if x < minX {
					minX = x
				}

				if y < minY {
					minY = y
				}

			}
		}
	}

	cropBounds := image.Rect(minX, minY, maxX, maxY)

	return cropImage(src, cropBounds)
}

func resizeImage(src image.Image, scale float64) image.Image {
	srcWidth := src.Bounds().Max.X - src.Bounds().Min.X
	srcHeight := src.Bounds().Max.Y - src.Bounds().Min.Y

	newWidth := int(math.Round(float64(srcWidth) * scale))
	newHeight := int(math.Round(float64(srcHeight) * scale))

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	return dst
}

func resizeImageKeepAspectRatio(src image.Image, maxWidth int, maxHeight int) image.Image {
	bounds := src.Bounds()
	srcWidth := bounds.Max.X - bounds.Min.X
	srcHeight := bounds.Max.Y - bounds.Min.Y

	widthRatio := float64(maxWidth) / float64(srcWidth)
	heightRatio := float64(maxHeight) / float64(srcHeight)

	scale := widthRatio
	if heightRatio < widthRatio {
		scale = heightRatio
	}

	return resizeImage(src, scale)
}

type MumbleImageOptions struct {
	Transparent bool
	MaxWidth    int
	MaxHeight   int
}

func imageForMumble(src image.Image, options *MumbleImageOptions) (string, error) {
	// Log::imageToImg(QImage img, int maxSize)
	// for defaults: max width, max height and quality

	maxWidth := 600
	if options.MaxWidth != 0 {
		maxWidth = options.MaxWidth
	}

	maxHeight := 400
	if options.MaxHeight != 0 {
		maxHeight = options.MaxHeight
	}

	dst := resizeImageKeepAspectRatio(src, maxWidth, maxHeight)

	var dataUri string
	var bytes bytes.Buffer

	if options.Transparent {
		dataUri = "image/png"

		err := (&png.Encoder{
			CompressionLevel: png.BestCompression,
		}).Encode(&bytes, dst)

		if err != nil {
			return "", err
		}
	} else {
		dataUri = "image/jpeg"

		err := jpeg.Encode(&bytes, dst, &jpeg.Options{
			Quality: 100,
		})

		if err != nil {
			return "", err
		}
	}

	encoded := url.QueryEscape(base64.StdEncoding.EncodeToString(bytes.Bytes()))

	return fmt.Sprintf(
		`<br /><img src="data:%s;base64,%s" />`, dataUri, encoded,
	), nil
}

func getRootChannel(client *gumble.Client) *gumble.Channel {
	var rootChannel *gumble.Channel

	for _, channel := range client.Channels {
		if channel.ID == 0 {
			rootChannel = channel
		}
	}

	return rootChannel
}

func sendToAll(client *gumble.Client, message string) {
	rootChannel := getRootChannel(client)

	if rootChannel == nil {
		return
	}

	rootChannel.Send(message, true)
}

func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	} else {
		return fallback
	}
}

func getEnvExists(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}
