package main

import (
	"bytes"
	"image"
	"image/png"
	"net/url"
	"path/filepath"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func makeWordArtPng(browser *rod.Browser, text string) image.Image {
	wordartPath, _ := filepath.Abs("./assets/wordart.html")

	url := url.URL{
		Scheme: "file",
		Path:   wordartPath,
	}

	page, _ := browser.Page(proto.TargetCreateTarget{
		URL:        url.String(),
		Background: false,
	})

	defer page.MustClose()

	transparentPage(page)

	page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  2560,
		Height: 1440,
	})

	page.MustWaitDOMStable()

	page.MustEval("makeText", text)

	imagePngBytes, _ := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})

	wordArtImage, _ := png.Decode(bytes.NewReader(imagePngBytes))
	wordArtImage = trimPngByTransparency(wordArtImage)

	// var outputPng bytes.Buffer
	// _ = png.Encode(&outputPng, wordArtImage)

	// return outputPng.Bytes()

	return wordArtImage
}
