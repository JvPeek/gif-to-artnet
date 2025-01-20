package main

import (
	"fmt"
	"image"
	"image/gif"
	"net/http"
	"sync"
	"time"

	"golang.org/x/image/draw"
)

type QueueApp struct {
	queue []*gif.GIF
	mu    sync.RWMutex
}

func (a *QueueApp) handleAdd(w http.ResponseWriter, req *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, url := range req.URL.Query()["image"] {
		resp, err := http.Get(url)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("something went wrong: %s", err)))
			return
		}

		inputGIF, err := gif.DecodeAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("something went wrong: %s", err)))
			return
		}

		newWidth := 32
		newHeight := 32

		outputGIF := &gif.GIF{
			LoopCount: inputGIF.LoopCount,
			Config: image.Config{
				ColorModel: inputGIF.Config.ColorModel,
				Width:      newWidth,
				Height:     newHeight,
			},
		}

		for i, frame := range inputGIF.Image {
			resizedFrame := image.NewPaletted(
				image.Rect(0, 0, newWidth, newHeight),
				frame.Palette,
			)

			draw.NearestNeighbor.Scale(
				resizedFrame,
				resizedFrame.Bounds(),
				frame,
				frame.Bounds(),
				draw.Over,
				nil,
			)

			outputGIF.Image = append(outputGIF.Image, resizedFrame)
			outputGIF.Delay = append(outputGIF.Delay, inputGIF.Delay[i])
		}

		// image, _, err := image.Decode(resp.Body)
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	w.Write([]byte(fmt.Sprintf("something went wrong: %s", err)))
		// 	return
		// }

		// scaledImage := imaging.Resize(image, 32, 32, imaging.Lanczos)

		a.queue = append(a.queue, outputGIF)
	}

	fmt.Fprint(w, "Image added")
}

func (a *QueueApp) handleShow(w http.ResponseWriter, req *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.queue) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Geh weg"))

		return
	}

	data := a.queue[0]
	a.queue = a.queue[1:]

	w.Header().Set("Content-Type", "image/gif")
	// imaging.Encode(w, data, imaging.GIF)
	gif.EncodeAll(w, data)

}

func (a *QueueApp) Serve() error {
	http.HandleFunc("/add", a.handleAdd)
	http.HandleFunc("/show", a.handleShow)

	err := http.ListenAndServe(":8090", nil)
	return err
}

var meinChannel chan int
var zweiterChannel chan int

func irgendEineGoroutine() {
	counter := 0

	for {
		fmt.Println("hier passiert irgendwas")

		<-time.After(2 * time.Second)

		// jetzt schreiben wir einfach alle zwei Sekunden in diesen Channel
		meinChannel <- counter
		fmt.Println("ich habe gerade in den channel geschrieben", counter)
		counter++
	}
}

func zweiteGoRoutine() {
	for {
		select {
		case value := <-meinChannel:
			fmt.Println("Ich bekam aus dem ersten Channel", value)
		case value := <-zweiterChannel:
			fmt.Println("Ich bekam aus dem zweiten Channel", value)
		}
	}
}

func dritteGoRoutine() {
	counter := 0

	for {
		<-time.After(3 * time.Second)
		zweiterChannel <- counter
		counter++
	}
}

func main() {
	app := &QueueApp{}

	meinChannel = make(chan int)
	zweiterChannel = make(chan int)

	go irgendEineGoroutine()
	go dritteGoRoutine()
	go zweiteGoRoutine()

	if err := app.Serve(); err != nil {
		panic(err)
	}

}
