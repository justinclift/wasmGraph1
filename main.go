//Wasming
// compile: GOOS=js GOARCH=wasm go build -o main.wasm ./main.go
package main

import (
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
)

type matrix []float64

type Point struct {
	Colour [3]int
	Num    int
	X      float64
	Y      float64
	Z      float64
}

type Object []Point

var (
	// The empty world space
	worldSpace   map[string]Object
	pointCounter = int(1)

	// The point objects
	object1 = Object{
		{X: 0, Y: 1.75, Z: 1.0},
		{X: 1.5, Y: -1.75, Z: 1.0},
		{X: -1.5, Y: -1.75, Z: 1.0},
		{X: 0, Y: 0, Z: 1.75},
	}

	width  float64
	height float64
	//debug  = true // If true, some debugging info is printed to the javascript console
	rCall              js.Callback
	ctx, doc, canvasEl js.Value
)

func main() {
	// Initialise canvas
	doc = js.Global().Get("document")
	canvasEl = doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Float()
	height = doc.Get("body").Get("clientHeight").Float()
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx = canvasEl.Call("getContext", "2d")

	// Seed the random generator (only used for colour generation), with a value that generates a "known ok" colour set
	rand.Seed(7)

	// Add some objects to the world space
	worldSpace = make(map[string]Object, 1)
	worldSpace["ob1"] = importObject(object1, 3.0, 3.0, 0.0)

	//// Simple keyboard handler for catching the arrow, WASD, and numpad keys
	//// Key value info can be found here: https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/key/Key_Values
	//keypressEvt := js.NewCallback(func(args []js.Value) {
	//	event := args[0]
	//	key := event.Get("key").String()
	//	if debug {
	//		fmt.Printf("Key is: %v\n", key)
	//	}
	//	// TODO: Use key presses to rotate the view around the world space origin
	//	//switch key {
	//	//case "ArrowLeft", "a", "A", "4":
	//	//	rectX -= step
	//	//case "ArrowRight", "d", "D", "6":
	//	//	rectX += step
	//	//case "ArrowUp", "w", "W", "8":
	//	//	rectY -= step
	//	//case "ArrowDown", "s", "S", "2":
	//	//	rectY += step
	//	//case "7", "Home":
	//	//	rectX -= step
	//	//	rectY -= step
	//	//case "9", "PageUp":
	//	//	rectX += step
	//	//	rectY -= step
	//	//case "1", "End":
	//	//	rectX -= step
	//	//	rectY += step
	//	//case "3", "PageDown":
	//	//	rectX += step
	//	//	rectY += step
	//	//}
	//})
	//defer keypressEvt.Release()
	//doc.Call("addEventListener", "keydown", keypressEvt)

	// Set the frame renderer going
	rCall = js.NewCallback(renderFrame)
	js.Global().Call("requestAnimationFrame", rCall)
	defer rCall.Release()

	// Keeps the application running
	done := make(chan struct{}, 0)
	<-done
}

// Renders one frame of the animation
func renderFrame(args []js.Value) {
	{
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
		graphHeight := height - 1
		centerX := graphWidth / 2
		centerY := graphHeight / 2
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
		ctx.Call("moveTo", left, centerY)
		ctx.Call("lineTo", graphWidth-gap, centerY)
		ctx.Call("stroke")

		// Draw vertical axis
		ctx.Call("beginPath")
		ctx.Set("strokeStyle", "black")
		ctx.Set("lineWidth", "2")
		ctx.Call("moveTo", centerX, top)
		ctx.Call("lineTo", centerX, graphHeight-gap)
		ctx.Call("stroke")

		// Draw horizontal markers
		step := math.Min(width, height) / 30
		markerPoint := graphHeight/2 - 5
		for i := graphWidth / 2; i < graphWidth-step; i += step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", i+step, markerPoint)
			ctx.Call("lineTo", i+step, markerPoint+10)
			ctx.Call("stroke")
		}
		for i := graphWidth / 2; i > left; i -= step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", i-step, markerPoint)
			ctx.Call("lineTo", i-step, markerPoint+10)
			ctx.Call("stroke")
		}

		// Draw vertical markers
		markerPoint = graphWidth/2 - 5
		for i := graphHeight / 2; i < graphHeight-step; i += step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", markerPoint, i+step)
			ctx.Call("lineTo", markerPoint+10, i+step)
			ctx.Call("stroke")
		}
		for i := graphHeight / 2; i > top; i -= step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", markerPoint, i-step)
			ctx.Call("lineTo", markerPoint+10, i-step)
			ctx.Call("stroke")
		}

		// TODO: Maybe add axis label text every few points?

		// Draw the points
		var pointNum int
		for _, o := range worldSpace {
			for k, l := range o {

				// Draw the coloured dot for the point
				px := centerX + (l.X * step)
				py := centerY + ((l.Y * step) * -1)
				ctx.Set("fillStyle", fmt.Sprintf("rgb(%d, %d, %d)", l.Colour[0], l.Colour[1], l.Colour[2]))
				ctx.Call("beginPath")
				ctx.Call("arc", px, py, 5, 0, 2*math.Pi)
				ctx.Call("fill")
				ctx.Set("fillStyle", "black")
				ctx.Set("font", "12px sans-serif")
				ctx.Call("fillText", fmt.Sprintf("Point %d", l.Num), px+5, py+15)

				// Draw lines between the points
				var tx, ty float64
				if k == 0 {
					if len(o) >= 4 {
						// TODO: This is just a dodgy workaround while testing.  Would be good to figure out a better approach.
						//       Maybe instead of using just Points, use some kind of structure that also defines edges?  Or would
						//       code to automatically work out the edges between points be better instead?
						tx = centerX + (o[len(o)-2].X * step)
						ty = centerY + ((o[len(o)-2].Y * step) * -1)
						ctx.Call("beginPath")
						ctx.Set("strokeStyle", "black")
						ctx.Set("lineWidth", "1")
						ctx.Call("moveTo", px, py)
						ctx.Call("lineTo", tx, ty)
						ctx.Call("stroke")

						px := centerX + (o[1].X * step)
						py := centerY + ((o[1].Y * step) * -1)
						tx = centerX + (o[len(o)-1].X * step)
						ty = centerY + ((o[len(o)-1].Y * step) * -1)
						ctx.Call("beginPath")
						ctx.Set("strokeStyle", "black")
						ctx.Set("lineWidth", "1")
						ctx.Call("moveTo", px, py)
						ctx.Call("lineTo", tx, ty)
						ctx.Call("stroke")
					}
					if len(o) == 5 {
						px := centerX + (o[2].X * step)
						py := centerY + ((o[2].Y * step) * -1)
						tx = centerX + (o[len(o)-1].X * step)
						ty = centerY + ((o[len(o)-1].Y * step) * -1)
						ctx.Call("beginPath")
						ctx.Set("strokeStyle", "black")
						ctx.Set("lineWidth", "1")
						ctx.Call("moveTo", px, py)
						ctx.Call("lineTo", tx, ty)
						ctx.Call("stroke")
					}
					tx = centerX + (o[len(o)-1].X * step)
					ty = centerY + ((o[len(o)-1].Y * step) * -1)
				} else {
					tx = centerX + (o[k-1].X * step)
					ty = centerY + ((o[k-1].Y * step) * -1)
				}
				ctx.Call("beginPath")
				ctx.Set("strokeStyle", "black")
				ctx.Set("lineWidth", "1")
				// setLineDash doesn't seem to work.  TODO: Figure out the right way to call this, as ctx.Call() outright fails
				// https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/setLineDash
				//ctx.Set("setLineDash", "{5, 15}")
				//ctx.Set("setLineDash", "5, 15")
				//ctx.Set("setLineDash", "(5, 15)")
				//ctx.Set("setLineDash", "[5, 15]")
				//ctx.Call("setLineDash", 5, 15)
				ctx.Call("moveTo", px, py)
				ctx.Call("lineTo", tx, ty)
				ctx.Call("stroke")

				// Draw darker coloured legend text
				ctx.Set("fillStyle", "black")
				ctx.Set("font", "bold 14px serif")
				ctx.Call("fillText", fmt.Sprintf("Point %d:", l.Num), graphWidth+20, l.Num*25)

				// Draw lighter coloured legend text
				ctx.Set("font", "12px sans-serif")
				ctx.Call("fillText", fmt.Sprintf("(%0.1f, %0.1f, %0.1f)", l.X, l.Y, l.Z), graphWidth+85, l.Num*25)
				pointNum++
			}
		}

		// It seems kind of weird (to me) to recursively call itself here, instead of using a timer approach, but
		// apparently this is best practise (at least in web environments: https://css-tricks.com/using-requestanimationframe)
		js.Global().Call("requestAnimationFrame", rCall)
	}
}

// Returns an object who's points have been transformed into 3D world space XYZ co-ordinates.  Also assigns numbers
// and colours to each point
func importObject(ob Object, x float64, y float64, z float64) (translatedObject Object) {
	// X and Y translation matrix.  Translates the objects into the world space at the given X and Y co-ordinates
	translateMatrix := matrix{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}

	for _, j := range ob {
		translatedObject = append(translatedObject, Point{
			Colour: [3]int{rand.Intn(255), rand.Intn(255), rand.Intn(255)},
			Num:    pointCounter,
			X:      (translateMatrix[0] * j.X) + (translateMatrix[1] * j.Y) + (translateMatrix[2] * j.Z) + (translateMatrix[3] * 1),   // 1st col, top
			Y:      (translateMatrix[4] * j.X) + (translateMatrix[5] * j.Y) + (translateMatrix[6] * j.Z) + (translateMatrix[7] * 1),   // 1st col, upper middle
			Z:      (translateMatrix[8] * j.X) + (translateMatrix[9] * j.Y) + (translateMatrix[10] * j.Z) + (translateMatrix[11] * 1), // 1st col, lower middle
		})
		pointCounter++
	}
	return translatedObject
}
