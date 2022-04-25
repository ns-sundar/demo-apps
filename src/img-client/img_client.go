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
    initialWaitSec = 2
    imageWaitSec = 2
    dnsResolverTimeoutMs = 5000 // Timeout (ms) for the DNS resolver
)

func main() {
    serverURL, dnsServer := getArgs()
    client := getHttpClient(dnsServer)

    myApp := app.New()
    win := myApp.NewWindow("Demo")
    myCanvas := win.Canvas()

    go updateWin(myCanvas, client, serverURL)

    // A dummy image to show initially
    m := image.NewRGBA(image.Rect(0, 0, 240, 240))
    blue := color.RGBA{0, 0, 255, 255}
    draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
    raster := canvas.NewRasterFromImage(m)

    win.Resize(fyne2.NewSize(240, 240))
    myCanvas.SetContent(raster)

    win.ShowAndRun() // blocking event loop, killed by CTRL-C
    myApp.Quit()
}

//func updateWin(win fyne2.Window, client *http.Client, url string) {
func updateWin(myCanvas fyne2.Canvas, client *http.Client, url string) {
    time.Sleep(initialWaitSec * time.Second)

    counter := 0

    for {
        time.Sleep(imageWaitSec * time.Second)
        name := fmt.Sprintf("image-%d", counter)
        url_with_counter := fmt.Sprintf("%s?image=%d", url, counter)
        fmt.Printf("Calling URL %s\n", url_with_counter)
        image, err := httpGetImage(client, url_with_counter, name)
        if err != nil {
           fmt.Printf("Cannot get image %s. Error: %s\n", name, err.Error())
           return
        }
        //fmt.Printf("Got image %s\n", name)
        raster := canvas.NewRasterFromImage(image)
        myCanvas.SetContent(raster)
        counter++
    }
}

func getArgs()  (string, string) {
    var serverName, serverPort, dnsServer string

    flag.StringVar(&serverName, "server", "webserver.demo.com", "HTTP server name")
    flag.StringVar(&serverPort, "port", "32612", "HTTP server port")
    flag.StringVar(&dnsServer, "dns", "192.168.0.199:53", "Custom DNS server with port")
    flag.Parse()

    serverURL := "http://" + serverName + ":" + serverPort
    fmt.Printf("Will connect to %s using DNS %s\n", serverURL, dnsServer)

    return serverURL, dnsServer
}

func getHttpClient(dnsResolverIP string) (*http.Client) {
    //var dnsResolverTimeoutMs = 5000 // Timeout (ms) for the DNS resolver (optional)

    dialer := &net.Dialer{
        Resolver: &net.Resolver{
            PreferGo: true,
            Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                d := net.Dialer{
                  Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
                }
                return d.DialContext(ctx, network, dnsResolverIP)
            },
        },
    }

    dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
        return dialer.DialContext(ctx, network, addr)
    }

    http.DefaultTransport.(*http.Transport).DialContext = dialContext

    return &http.Client{}
}

func httpGet(client *http.Client, url string) ([]byte, error) {
    resp, err := client.Get(url)
    if err != nil {
        fmt.Printf("unable to connect to http server - %s\n", err.Error())
        return []byte{}, err
    }
    if resp != nil && resp.Body != nil {
        defer resp.Body.Close()
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("unable to read http server response - %s\n", err.Error())
        return []byte{}, err
    }

    //fmt.Println(string(body))
    return body, err
}

func httpGetImage(client *http.Client, url, name string) (image.Image, error) {
   getRespBody, err := httpGet(client, url)
   if err != nil {
      fmt.Printf("HTTP GET failed. Error: (%s)\n", err.Error())
      return nil, err
   }

   reader := bytes.NewReader(getRespBody)
   // 2nd arg below is image type (e.g. jpeg)
   // goImage, _, err := image.Decode(reader) // image of type image.Image

   goImage, err := jpeg.Decode(reader)
   if err != nil {
      fmt.Printf("Image decode failed. Error: (%s)\n", err.Error())
      return nil, err
   }
   return goImage, nil
}

//func getImage() (image.Image) {
func getImage() (*canvas.Image) {
    m := image.NewRGBA(image.Rect(0, 0, 240, 240))
    blue := color.RGBA{0, 0, 255, 255}
    draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

    img := canvas.NewImageFromImage(m)
    return img
}

func getNewImage() (*canvas.Image) {
    m := image.NewRGBA(image.Rect(0, 0, 240, 240))
    blue := color.RGBA{255, 0, 0, 255}
    draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

    //var img image.Image = m
    img := canvas.NewImageFromImage(m)
    return img
}

//func showWindow(img image.Image) {
func showWindow(image *canvas.Image) {
    myApp := app.New()
    win := myApp.NewWindow("Image")

    // image := canvas.NewImageFromResource(theme.FyneLogo())
    // image := canvas.NewImageFromURI(uri)
    // image := canvas.NewImageFromImage(src)
    // image := canvas.NewImageFromReader(reader, name)
    // image := canvas.NewImageFromFile(fileName)
    image.FillMode = canvas.ImageFillOriginal
    win.SetContent(image)

    win.ShowAndRun()
}
