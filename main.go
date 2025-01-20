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
		w.Write([]byte("Queue ist leer"))

		return
	}

	data := a.queue[0]

	w.Header().Set("Content-Type", "image/gif")

	gif.EncodeAll(w, data)

}

func (a *QueueApp) Serve() error {
	http.HandleFunc("/add", a.handleAdd)
	http.HandleFunc("/show", a.handleShow)

	err := http.ListenAndServe(":8090", nil)
	return err
}

func ledWorker(app *QueueApp) {

	for {
		//fmt.Printf("Current queue: %v\n", len(app.queue))
		if len(app.queue) > 0 {

		}
		time.Sleep(200 * time.Millisecond)
	}
}

func main() {
	app := &QueueApp{}

	go ledWorker(app)
	if err := app.Serve(); err != nil {
		panic(err)
	}

}
