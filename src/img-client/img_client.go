// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	fyne2 "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	//"fyne.io/fyne/v2/theme"

	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
)

const (
	initialWaitSec     = 2
	imageWaitSec       = 1
	dnsServerTimeoutMs = 5000 // Name resolution timeout in millisec
	winWidthPixels     = 240
	winHeightPixels    = 240
)

func main() {
	serverName, serverPort, dnsServer := getArgs()

	// Create a HTTP client with a custom DNS resolver
	client := getHttpClient(dnsServer)

	// Init graphics window
	myApp := app.New()
	win := myApp.NewWindow("Demo")
	myCanvas := win.Canvas()

	// Launch the app logic loop
	go updateWin(myCanvas, client, serverName, serverPort, dnsServer)

	// A dummy image to show initially
	m := image.NewRGBA(image.Rect(0, 0, winWidthPixels, winHeightPixels))
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
	raster := canvas.NewRasterFromImage(m)

	win.Resize(fyne2.NewSize(winWidthPixels, winHeightPixels))
	myCanvas.SetContent(raster)

	win.ShowAndRun() // blocking event loop, killed by CTRL-C
}

// updateWin fetches a new image from the specified server endpoint using the custom
// HTTP client and displays it in the app window.
func updateWin(myCanvas fyne2.Canvas, client *http.Client,
	serverName, serverPort, dnsServer string) {

	counter := 0
	time.Sleep(initialWaitSec * time.Second)

	for {
		counter++
		time.Sleep(imageWaitSec * time.Second)

		fmt.Printf("%d: %s ", counter, serverName)

		// We resolve the server Name to an IP only for display purpose.
		// The httpGet below also does a DNS lookup.
		ip, lookupErr := lookupHost(serverName, dnsServer)
		if lookupErr != nil {
			fmt.Printf("Error in DNS resolution: %s\n", lookupErr.Error())
			continue
		}
		fmt.Printf("%s ", ip)

		url := fmt.Sprintf("http://%s:%s?image=%d", serverName, serverPort, counter)
		image, err := httpGetImage(client, url)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}

		fmt.Printf("Got image\n")
		raster := canvas.NewRasterFromImage(image)
		myCanvas.SetContent(raster)
	}
}

// getHttpClient returns a HTTP client configured with a custom DNS resolver
func getHttpClient(dnsServer string) *http.Client {
	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(dnsServerTimeoutMs) * time.Millisecond,
				}
				return d.DialContext(ctx, network, dnsServer)
			},
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	http.DefaultTransport.(*http.Transport).DialContext = dialContext

	return &http.Client{}
}

// httpGet returns the GET response from given URL using the given HTTP client
func httpGet(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		wrapErr := fmt.Errorf("unable to connect to http server: %w", err)
		return []byte{}, wrapErr
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		wrapErr := fmt.Errorf("unable to read http server response: %w", err)
		return []byte{}, wrapErr
	}

	return body, err
}

// httpGetImage decodes the HTTP GET response as a JPEG image
func httpGetImage(client *http.Client, url string) (image.Image, error) {
	getRespBody, err := httpGet(client, url)
	if err != nil {
		wrapErr := fmt.Errorf("HTTP GET failed: %w", err)
		return nil, wrapErr
	}

	reader := bytes.NewReader(getRespBody)

	img, err := jpeg.Decode(reader)
	if err != nil {
		wrapErr := fmt.Errorf("Image decode failed: %w", err)
		return nil, wrapErr
	}

	// Instead we can handle non-jpeg images too, as below.
	// 2nd arg below is image type (e.g. jpeg)
	// img, _, err := image.Decode(reader) // image of type image.Image

	return img, nil
}

// lookupHost does DNS resolution of given host using given DNS server
func lookupHost(hostName, dnsServer string) (string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(dnsServerTimeoutMs),
			}
			return d.DialContext(ctx, network, dnsServer)
		},
	}
	ip, err := r.LookupHost(context.Background(), hostName)
	if err != nil {
		return "", err
	}

	//fmt.Printf("First IP address: %s\n", ip[0])
	return ip[0], nil
}

func getArgs() (string, string, string) {
	var serverName, serverPort, dnsServer string

	flag.StringVar(&serverName, "server", "webserver.demo.com", "HTTP server name")
	flag.StringVar(&serverPort, "port", "32612", "HTTP server port")
	flag.StringVar(&dnsServer, "dns", "192.168.0.199:53", "Custom DNS server with port")
	flag.Parse()

	serverURL := "http://" + serverName + ":" + serverPort
	fmt.Printf("Will connect to %s using DNS %s\n", serverURL, dnsServer)

	return serverName, serverPort, dnsServer
}
