package main

import (
	"bytes"
	"image"
	"image/png"
	"net/url"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func makeWordArtPng(browser *rod.Browser, text string) (image.Image, error) {
	wordartPath, err := filepath.Abs("./assets/wordart.html")
	if err != nil {
		return image.Black, err
	}

	url := url.URL{
		Scheme: "file",
		Path:   wordartPath,
	}

	page, err := browser.Page(proto.TargetCreateTarget{
		URL: url.String(),
	})

	if err != nil {
		return image.Black, err
	}

	defer page.MustClose()

	err = setTransparentPage(page)
	if err != nil {
		return image.Black, err
	}

	page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  2560,
		Height: 1440,
	})

	err = page.WaitDOMStable(time.Second, 0)
	if err != nil {
		return image.Black, err
	}

	_, err = page.Eval("makeText", text)
	if err != nil {
		return image.Black, err
	}

	imagePngBytes, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})

	if err != nil {
		return image.Black, err
	}

	wordArtImage, err := png.Decode(bytes.NewReader(imagePngBytes))
	if err != nil {
		return image.Black, err
	}

	wordArtImage = trimPngByTransparency(wordArtImage)

	// var outputPng bytes.Buffer
	// _ = png.Encode(&outputPng, wordArtImage)

	// return outputPng.Bytes()

	return wordArtImage, nil
}
