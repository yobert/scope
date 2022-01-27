package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/jfreymuth/pulse"
)

const (
	rate  = 44100 / 4
	speed = 1000
)

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	c, err := pulse.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	samples := make(chan float32, 1e6)

	debug := func(f []float32) (int, error) {
		for _, ff := range f {
			samples <- ff
		}
		return len(f), nil
	}

	if err := glfw.Init(); err != nil {
		return err
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	window, err := glfw.CreateWindow(480, 480, "scope", nil, nil)
	if err != nil {
		return err
	}

	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		return err
	}

	gl.ClearColor(0.0, 0.3, 0.0, 0)

	var (
		width, height int32
	)

	stream, err := c.NewRecord(pulse.Float32Writer(debug),
		pulse.RecordMono,
		pulse.RecordSampleRate(rate),
		pulse.RecordMediaName("scopeyscopey"),
		pulse.RecordLatency(0.01),
	)
	if err != nil {
		return err
	}
	defer stream.Close()

	stream.Start()
	defer stream.Stop()

	time.Sleep(time.Millisecond)

	sampledur := time.Second / rate
	fmt.Println("sample", sampledur)

	tn := time.Now()
	tl := tn

	x := 0
	xl := x

	//time.Sleep(time.Millisecond * 10)

	for !window.ShouldClose() {
		w, h := window.GetSize()

		if int32(w) != width || int32(h) != height {
			width = int32(w)
			height = int32(h)
			gl.MatrixMode(gl.PROJECTION)
			gl.LoadIdentity()
			//f := ((float64(width) / float64(height)) - 1) / 2
			//gl.Frustum(-1-f, 1+f, -1, 1, 1.0, 10.0)
			gl.Viewport(0, 0, width, height)
			gl.Ortho(0, float64(width), -1, 1, 1, -1)

			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()
		}

		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.Color3f(0, 1, 0)

		tn = time.Now()
		td := tn.Sub(tl)
		tl = tn
		xd := x - xl

		// in that amount of time, how many samples should we read?
		count := int(td / sampledur)

		gl.Begin(gl.LINE_STRIP)
		last := float32(0)
		for i := 0; i < count; i++ {
			f := <-samples

			fx := float32(x) + ((-(float32(i) / float32(count)) + 1) * float32(xd))
			xx := int(fx) % int(width)

			fxx := float32(xx)

			if i > 0 && last < fxx {
				gl.End()
				gl.Begin(gl.LINE_STRIP)
			}

			gl.Vertex3f(
				fxx,
				f,
				0)
			last = fxx
		}
		gl.End()

		window.SwapBuffers()
		glfw.PollEvents()

		xl = x
		x += speed
		//		if x > 1 {
		//			x = x - (2 + speed)
		//			xl = xl - (2 + speed)
		//		}
	}

	return nil
}
