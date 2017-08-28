package main

import (
	"fmt"
	"github.com/musl/libgofr"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const Version = "0.0.2"

var id_chan = make(chan int, 1)

func finish(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	fmt.Fprintf(w, message)
}

func logDuration(message string, start time.Time) {
	end := time.Now()
	log.Printf("%s: %v\n", message, end.Sub(start))
}

func route_png(w http.ResponseWriter, r *http.Request) {
	id := <-id_chan
	start := time.Now()
	defer logDuration(fmt.Sprintf("%08d %s", id, r.URL.Path), start)

	if r.Method != "GET" {
		finish(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	q := r.URL.Query()

	s, err := strconv.Atoi(q.Get("s"))
	if err != nil {
		s = 1
	}

	width, err := strconv.Atoi(q.Get("w"))
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid width")
		return
	}

	height, err := strconv.Atoi(q.Get("h"))
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid height")
		return
	}

	iterations, err := strconv.Atoi(q.Get("i"))
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid i")
		return
	}

	er, err := strconv.ParseFloat(q.Get("e"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid e")
		return
	}

	rmin, err := strconv.ParseFloat(q.Get("rmin"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid rmin")
		return
	}

	imin, err := strconv.ParseFloat(q.Get("imin"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid rmin")
		return
	}

	rmax, err := strconv.ParseFloat(q.Get("rmax"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid rmax")
		return
	}

	imax, err := strconv.ParseFloat(q.Get("imax"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid rmin")
		return
	}

	c := q.Get("c")
	hex := q.Get("m")

	p := gofr.Parameters{
		ImageWidth:   width * s,
		ImageHeight:  height * s,
		MaxI:         iterations,
		EscapeRadius: er,
		Min:          complex(rmin, imin),
		Max:          complex(rmax, imax),
		ColorFunc:    c,
		MemberColor:  hex,
	}

	log.Printf("%08d rendering\n", id)

	// TODO: Check parameters and set reasonable bounds on what we can
	// quickly calculate.

	img := image.NewNRGBA64(image.Rect(0, 0, p.ImageWidth, p.ImageHeight))
	n := runtime.NumCPU()
	contexts := gofr.MakeContexts(img, n, &p)
	gofr.Render(n, contexts, gofr.Mandelbrot)

	scaled_img := resize.Resize(uint(width), uint(height), image.Image(img), resize.Lanczos3)

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, scaled_img)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	bind_addr := "0.0.0.0:8000"

	go func() {
		for i := 0; ; i++ {
			id_chan <- i
		}
	}()

	http.Handle("/", fs)
	http.HandleFunc("/png", route_png)
	log.Printf("Listening on: %s\n", bind_addr)
	log.Fatal(http.ListenAndServe(bind_addr, nil))
}
