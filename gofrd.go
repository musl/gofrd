package main

import (
	"fmt"
	"github.com/golang/net/websocket"
	"github.com/musl/libgofr"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

func finish(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	fmt.Fprintf(w, message)
}

func logDuration(message string, start time.Time) {
	end := time.Now()
	log.Printf("%s: %v\n", message, end.Sub(start))
}

func route_png(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer logDuration(fmt.Sprintf("%s", r.URL.Path), start)

	if r.Method != "GET" {
		finish(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	q := r.URL.Query()

	s, err := strconv.Atoi(q.Get("s"))
	if err != nil {
		s = 1
	}

	iw, err := strconv.Atoi(q.Get("w"))
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid w")
		return
	}

	ih, err := strconv.Atoi(q.Get("h"))
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid h")
		return
	}

	i, err := strconv.Atoi(q.Get("i"))
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
	color_func := gofr.ColorMono
	switch c {
	case "mono":
		color_func = gofr.ColorMono
	case "stripe":
		color_func = gofr.ColorMonoStripe
	case "bands":
		color_func = gofr.ColorBands
	case "smooth":
		color_func = gofr.ColorSmooth
	default:
		finish(w, http.StatusUnprocessableEntity, "Invalid c")
		return
	}

	hex := q.Get("m")
	member_color := color.NRGBA64{0, 0, 0, 0xffff}
	if len(hex) > 2 {
		m, err := strconv.ParseInt(hex[1:len(hex)], 16, 32)
		if err != nil {
			finish(w, http.StatusUnprocessableEntity, "Invalid m")
			return
		}
		member_color = color.NRGBA64{
			uint16(((m >> 16) & 0xff) * 0x101),
			uint16(((m >> 8) & 0xff) * 0x101),
			uint16((m & 0xff) * 0x101),
			0xffff,
		}
	}

	p := gofr.Parameters{
		ImageWidth:   iw * s,
		ImageHeight:  ih * s,
		MaxI:         i,
		EscapeRadius: er,
		Min:          complex(rmin, imin),
		Max:          complex(rmax, imax),
		ColorFunc:    color_func,
		MemberColor:  member_color,
	}

	// TODO: Check parameters and set reasonable bounds on what we can
	// quickly calculate.

	img := image.NewNRGBA64(image.Rect(0, 0, p.ImageWidth, p.ImageHeight))
	n := runtime.NumCPU()
	contexts := gofr.MakeContexts(img, n, &p)
	gofr.Render(n, contexts, gofr.Mandelbrot)

	scaled_img := resize.Resize(uint(iw), uint(ih), image.Image(img), resize.Lanczos3)

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, scaled_img)
}

/*
 * TODO: need a protocol/lifecycle
 */
func handler_png(ws *websocket.Conn) {
	fmt.Fprintf(ws, "Hello websockets.\n")
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	bind_addr := "0.0.0.0:8000"

	http.Handle("/", fs)
	http.HandleFunc("/png", route_png)
	http.Handle("/png-socket", websocket.Handler(handler_png))
	log.Printf("Listening on: %s\n", bind_addr)
	log.Fatal(http.ListenAndServe(bind_addr, nil))
}
