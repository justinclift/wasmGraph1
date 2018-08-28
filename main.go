//Wasming
// compile: GOOS=js GOARCH=wasm go build -o main.wasm ./main.go
package main

import (
	"math"
	"syscall/js"

	"github.com/lucasb-eyer/go-colorful"
)

var (
	width  float64
	height float64
	mouseX float64
	mouseY float64
	size   = 25
)

func main() {
	// Initialise canvas
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Float()
	height = doc.Get("body").Get("clientHeight").Float()
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	ctx := canvasEl.Call("getContext", "2d")

	done := make(chan struct{}, 0)

	mouseMoveEvt := js.NewCallback(func(args []js.Value) {
		e := args[0]
		mouseX = e.Get("clientX").Float()
		mouseY = e.Get("clientY").Float()
		//fmt.Printf("Mouse position - X: %v, Y: %v\n", mouseX, mouseY)
	})
	defer mouseMoveEvt.Release()

	// Handle mouse
	doc.Call("addEventListener", "mousemove", mouseMoveEvt)

	var renderFrame js.Callback
	colorRot := float64(0)
	currentX := float64(100)
	currentY := float64(75)

	renderFrame = js.NewCallback(func(args []js.Value) {
		// Handle window resizing
		curBodyW := doc.Get("body").Get("clientWidth").Float()
		curBodyH := doc.Get("body").Get("clientHeight").Float()
		if curBodyW != width || curBodyH != height {
			width, height = curBodyW, curBodyH
			canvasEl.Set("width", width)
			canvasEl.Set("height", height)
		}

		// Work out where to draw the next shape
		moveX := (mouseX - currentX) * 0.02
		moveY := (mouseY - currentY) * 0.02
		currentX += moveX
		currentY += moveY

		// Draw the shape at the new position
		colorRot = float64(int(colorRot+1) % 360)
		ctx.Set("fillStyle", colorful.Hsv(colorRot, 1, 1).Hex())
		ctx.Call("beginPath")
		ctx.Call("arc", currentX, currentY, size, 0, 2*math.Pi) // Radians
		ctx.Call("fill")

		js.Global().Call("requestAnimationFrame", renderFrame)
	})
	defer renderFrame.Release()

	// Start running
	js.Global().Call("requestAnimationFrame", renderFrame)

	// Keeps the application running
	<-done
}
