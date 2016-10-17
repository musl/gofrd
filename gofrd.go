package main

import (
	"fmt"
	"github.com/musl/libgofr"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"runtime"
	"strconv"
)

func finish(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	fmt.Fprintf(w, message)
}

func route_index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		finish(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	q := r.URL.Query()

	iw, err := strconv.Atoi(q.Get("w"))
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid w")
	}

	ih, err := strconv.Atoi(q.Get("h"))
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid h")
	}

	i, err := strconv.Atoi(q.Get("i"))
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid i")
	}

	er, err := strconv.ParseFloat(q.Get("e"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid e")
	}

	rmin, err := strconv.ParseFloat(q.Get("rmin"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid rmin")
	}

	imin, err := strconv.ParseFloat(q.Get("imin"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid rmin")
	}

	rmax, err := strconv.ParseFloat(q.Get("rmax"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid rmax")
	}

	imax, err := strconv.ParseFloat(q.Get("imax"), 64)
	if err != nil {
		finish(w, http.StatusUnprocessableEntity, "Invalid rmin")
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
	}

	hex := q.Get("m")
	member_color := color.NRGBA64{0, 0, 0, 0xffff}
	if len(hex) > 2 {
		m, err := strconv.ParseInt(hex[1:len(hex)], 16, 32)
		if err != nil {
			finish(w, http.StatusUnprocessableEntity, "Invalid m")
		}
		member_color = color.NRGBA64{
			uint16(((m >> 16) & 0xff) * 0x101),
			uint16(((m >> 8) & 0xff) * 0x101),
			uint16((m & 0xff) * 0x101),
			0xffff,
		}
	}

	p := gofr.Parameters{
		ImageWidth:   iw,
		ImageHeight:  ih,
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

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/png", route_index)
	http.ListenAndServe(":8000", nil)
}
