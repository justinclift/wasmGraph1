//Wasming
// compile: GOOS=js GOARCH=wasm go build -o main.wasm ./main.go
package main

import (
	"fmt"
	"math"
	"syscall/js"
)

var (
	width  float64
	height float64
	operationQueue = []string{"stuff"}
	debug  = true // If true, some debugging info is printed to the javascript console
)

func main() {
	// Initialise canvas
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Float()
	height = doc.Get("body").Get("clientHeight").Float()
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx := canvasEl.Call("getContext", "2d")

	// Simple keyboard handler for catching the arrow, WASD, and numpad keys
	// Key value info can be found here: https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/key/Key_Values
	keypressEvt := js.NewCallback(func(args []js.Value) {
		event := args[0]
		key := event.Get("key").String()
		if debug {
			fmt.Printf("Key is: %v\n", key)
		}
		//switch key {
		//case "ArrowLeft", "a", "A", "4":
		//	rectX -= step
		//case "ArrowRight", "d", "D", "6":
		//	rectX += step
		//case "ArrowUp", "w", "W", "8":
		//	rectY -= step
		//case "ArrowDown", "s", "S", "2":
		//	rectY += step
		//case "7", "Home":
		//	rectX -= step
		//	rectY -= step
		//case "9", "PageUp":
		//	rectX += step
		//	rectY -= step
		//case "1", "End":
		//	rectX -= step
		//	rectY += step
		//case "3", "PageDown":
		//	rectX += step
		//	rectY += step
		//}
	})
	defer keypressEvt.Release()
	doc.Call("addEventListener", "keydown", keypressEvt)

	done := make(chan struct{}, 0)

	var renderFrame js.Callback
	renderFrame = js.NewCallback(func(args []js.Value) {
		// Handle window resizing
		curBodyW := doc.Get("body").Get("clientWidth").Float()
		curBodyH := doc.Get("body").Get("clientHeight").Float()
		if curBodyW != width || curBodyH != height {
			width, height = curBodyW, curBodyH
			canvasEl.Set("width", width)
			canvasEl.Set("height", height)
		}

		// ** Draw the graph area **

		// Grey background
		ctx.Set("fillStyle", "lightgrey")
		ctx.Call("fillRect", 1, 1, width-1, height-1)

		// Draw border around the graph area
		border := float64(2)
		gap := float64(3)
		left := border + gap
		top := border + gap
		graphWidth := width * 0.75
		graphHeight := height -1
		ctx.Call("beginPath")
		ctx.Set("strokeStyle", "black")
		ctx.Set("lineWidth", "2")
		ctx.Call("moveTo", border, border)
		ctx.Call("lineTo", graphWidth, border)
		ctx.Call("lineTo", graphWidth, graphHeight)
		ctx.Call("lineTo", border, graphHeight)
		ctx.Call("closePath")
		ctx.Call("stroke")

		// Draw horizontal axis
		ctx.Call("beginPath")
		ctx.Set("strokeStyle", "black")
		ctx.Set("lineWidth", "2")
		ctx.Call("moveTo", left, graphHeight/2)
		ctx.Call("lineTo", graphWidth-gap, graphHeight/2)
		ctx.Call("stroke")

		// Draw vertical axis
		ctx.Call("beginPath")
		ctx.Set("strokeStyle", "black")
		ctx.Set("lineWidth", "2")
		ctx.Call("moveTo", graphWidth/2, top)
		ctx.Call("lineTo", graphWidth/2, graphHeight - gap)
		ctx.Call("stroke")

		// Draw horizontal markers
		step := math.Min(width, height) / 30
		markerPoint := graphHeight/2 - 5
		for i := graphWidth/2; i < graphWidth - step; i += step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", i + step, markerPoint)
			ctx.Call("lineTo", i + step, markerPoint + 10)
			ctx.Call("stroke")
		}
		for i := graphWidth/2; i > left; i -= step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", i - step, markerPoint)
			ctx.Call("lineTo", i - step, markerPoint + 10)
			ctx.Call("stroke")
		}

		// Draw vertical markers
		markerPoint = graphWidth/2 - 5
		for i := graphHeight/2; i < graphHeight - step; i += step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", markerPoint, i + step)
			ctx.Call("lineTo", markerPoint + 10, i + step)
			ctx.Call("stroke")
		}
		for i := graphHeight/2; i > top; i -= step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", markerPoint, i - step)
			ctx.Call("lineTo", markerPoint + 10, i - step)
			ctx.Call("stroke")
		}

		// TODO: Maybe add axis label text every few points?


		// TODO: Whatever else

		// It seems kind of weird (to me) to recursively call itself here, instead of using a timer approach, but
		// apparently this is best practise (at least in web environments: https://css-tricks.com/using-requestanimationframe)
		js.Global().Call("requestAnimationFrame", renderFrame)
	})
	defer renderFrame.Release()

	// Start running
	js.Global().Call("requestAnimationFrame", renderFrame)

	// Keeps the application running
	<-done
}
