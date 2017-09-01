package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/musl/libgofr"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const Version = "0.0.2"

var id_chan = make(chan uuid.UUID, 100)

type LogResponseWriter struct {
	http.ResponseWriter
	Status int
	Start  time.Time
	End    time.Time
}

func NewLogResponseWriter(w http.ResponseWriter) *LogResponseWriter {
	return &LogResponseWriter{w, 0, time.Now(), time.Now()}
}

func (self *LogResponseWriter) WriteHeader(code int) {
	self.Status = code
	self.ResponseWriter.WriteHeader(code)
}

func (self LogResponseWriter) Log(message string) {
	self.End = time.Now()
	log.Printf("%s %v\n", message, self.End.Sub(self.Start))
}

func wrapHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := NewLogResponseWriter(w)

		h.ServeHTTP(lrw, r)
		lrw.Log(fmt.Sprintf("%d %s %s %s", lrw.Status, r.Method, r.URL.Path, r.RemoteAddr))
	})
}

func wrapHandlerFunc(h http.HandlerFunc) http.Handler {
	return wrapHandler(http.Handler(h))
}

func finish(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	io.WriteString(w, message)
}

func route_png(w http.ResponseWriter, r *http.Request) {
	id := <-id_chan

	if r.Method != "GET" {
		finish(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}

	q := r.URL.Query()

	s, err := strconv.Atoi(q.Get("s"))
	if err != nil {
		s = 1
	}

	e, err := strconv.Atoi(q.Get("p"))
	if err != nil {
		e = 1
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
		Power:        e,
	}

	// TODO: Check parameters and set reasonable bounds on what we can
	// quickly calculate.
	//
	// Create a pool of goroutines that process render jobs, with a
	// time-out for accepting render jobs. Have UI support for the "try
	// again later" response.

	img := image.NewNRGBA64(image.Rect(0, 0, p.ImageWidth, p.ImageHeight))
	n := runtime.NumCPU()
	contexts := gofr.MakeContexts(img, n, &p)
	gofr.Render(n, contexts, gofr.Mandelbrot)

	scaled_img := resize.Resize(uint(width), uint(height), image.Image(img), resize.Lanczos3)

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("X-Render-Job-ID", id.String())
	w.WriteHeader(http.StatusOK)
	png.Encode(w, scaled_img)
}

func route_status(w http.ResponseWriter, r *http.Request) {
	finish(w, http.StatusOK, "OK")
}

func main() {
	var value string

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
	log.Printf("gofrd v%s", Version)
	log.Printf("libgofrd v%s", gofr.Version)

	static_dir := "./static"
	if value = os.Getenv("GOFR_STATIC_DIR"); value != "" {
		static_dir = value
	}
	static_dir, err := filepath.Abs(static_dir)
	if err != nil {
		panic(err)
	}
	log.Printf("Serving from: %s\n", static_dir)

	bind_addr := "0.0.0.0:8000"
	if value = os.Getenv("GOFR_BIND_ADDR"); value != "" {
		bind_addr = value
	}
	log.Printf("Listening on: %s\n", bind_addr)

	go func() {
		for i := 0; ; i++ {
			id_chan <- uuid.New()
		}
	}()

	http.Handle("/", wrapHandler(http.FileServer(http.Dir(static_dir))))
	http.Handle("/png", wrapHandlerFunc(route_png))
	http.Handle("/status", wrapHandlerFunc(route_status))

	/* Run the thing. */
	log.Fatal(http.ListenAndServe(bind_addr, nil))
}
