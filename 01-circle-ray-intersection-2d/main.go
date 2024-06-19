package main

import "fmt"
import "math"
import "math/rand"
import "syscall/js"

const WIDTH = 300
const HEIGHT = 600

type Vec2 struct {
    X int32
    Y int32
}

func (v Vec2) Sub(b Vec2) Vec2 {
    return Vec2 {
        X: v.X - b.X,
        Y: v.Y - b.Y,
    }
}

func (v Vec2) Add(b Vec2) Vec2 {
    return Vec2 {
        X: v.X + b.X,
        Y: v.Y + b.Y,
    }
}

func (v Vec2) Scale(t float64) Vec2 {
    return Vec2 {
        X: int32(float64(v.X)*t),
        Y: int32(float64(v.Y)*t),
    }
}

type Circle struct {
    Origin Vec2
    R int32
}

type Ray struct {
    Origin Vec2
    D Vec2
}

var circle *Circle = &Circle { Origin: Vec2 { X: 0, Y: 0 }, R: 10 }
var ray *Ray = &Ray { Origin: Vec2 { X: 0, Y: 0 }, D: Vec2 { X: 10, Y: 10 } }

var canvasOffset *Vec2 = &Vec2 {}

func (r *Ray) intersectCircle(c *Circle) (bool, float64, float64) {
    /*
    circumference:
    x**2 + y**2 = r**2

    ray as origin + delta*coef:
    x = Ox + Dx * t
    y = Oy + Dy * t

    eq:
    (Ox + Dx * t)**2 + (Oy + Dy * t)**2 = r**2

    (Dx^2 + Dy^2)*t^2 + 2*(Ox*Dx+Oy*Dy)*t + (Ox^2+Oy^2) - r^2 = 0
    ^^^^^^^^^^^^^       ^^^^^^^^^^^^^^^     ^^^^^^^^^^^^^^^^^
    A                   B                   C

    */

    // ray shifted so that the circle is at the zero origin
    or := r.Origin.Sub(c.Origin)
    OX := float64(or.X)
    OY := float64(or.Y)

    DX := float64(r.D.X)
    DY := float64(r.D.Y)

    DLEN := math.Sqrt(DX*DX + DY*DY)
    DX = DX / DLEN
    DY = DY / DLEN

    R := float64(c.R)

    // quadratic equation coefficients
    A := DX*DX + DY*DY
    B := 2.0*(DX*OX+DY*OY)
    C := OX*OX + OY*OY - R*R

    // discriminant
    D := B*B - 4.0*A*C

    // no intersection
    if D < 0 {
        return false, 0, 0
    }

    t1 := ( -B + math.Sqrt(D) ) / (2.0*A)
    t2 := ( -B - math.Sqrt(D) ) / (2.0*A)

    if t1 < 0 && t2 < 0 {
        // backward hit
        return false, 0, 0
    }

    return true, t1 / DLEN, t2 / DLEN
}

func moveRay() js.Func {
    return js.FuncOf(func (this js.Value, args []js.Value) any {
        mouseX := int32(args[0].Get("clientX").Int()) - canvasOffset.X
        mouseY := int32(args[0].Get("clientY").Int()) - canvasOffset.Y

        ray.D.X = mouseX - ray.Origin.X
        ray.D.Y = mouseY - ray.Origin.Y

        return nil
    })
}

func randomizeCircle() js.Func {
    return js.FuncOf(func (this js.Value, args []js.Value) any {
        circle.Origin.X = rand.Int31n(WIDTH)
        circle.Origin.Y = rand.Int31n(HEIGHT)
        circle.R = 10 + rand.Int31n(WIDTH / 3)
        return nil
    })
}

func randomizeRay() js.Func {
    return js.FuncOf(func (this js.Value, args []js.Value) any {
        ray.Origin.X = rand.Int31n(WIDTH)
        ray.Origin.Y = rand.Int31n(HEIGHT)
        return nil
    })
}

func drawCircle(ctx js.Value, x int32, y int32, r int32, color string) {
    ctx.Set("strokeStyle", color)
    ctx.Call("beginPath")
    ctx.Call("arc", x, y, r, 0, 2.0 * math.Pi);
    ctx.Call("stroke")
}

func main() {
	fmt.Println("Hello, WebAssembly!")

    document := js.Global().Get("document")
    window := js.Global().Get("window")
    // get the dom element
	canvas := document.Call("getElementById", "tCanvas")
	tFrameCounter := document.Call("getElementById", "tFrameCounter")
    tRandomCircle := document.Call("getElementById", "tRandomCircle")
    tRandomRay := document.Call("getElementById", "tRandomRay")
    tDebug := document.Call("getElementById", "tDebug")

    tRect := canvas.Call("getBoundingClientRect")
    canvasOffset.X = int32(tRect.Get("left").Int())
    canvasOffset.Y = int32(tRect.Get("top").Int())

    // initialize 2d context
    ctx := canvas.Call("getContext", "2d")

    var frameCounter int32 = 0
    var timePrev float64 = 0

    var drawFrame js.Func
    drawFrame = js.FuncOf(func (this js.Value, args []js.Value) any {
        timeNow := args[0].Float()
        if timeNow - timePrev > 1000 {
            fps := float64(frameCounter * 1000) / (timeNow - timePrev)
            timePrev = timeNow
            tFrameCounter.Set("innerHTML", math.Floor(fps))
            frameCounter = 0
        }
        frameCounter++

        // clear canvas
        ctx.Call("clearRect", 0, 0, WIDTH, HEIGHT)

        intersection, d1, d2 := ray.intersectCircle(circle)
        tDebug.Set("innerHTML", intersection)

        circleColor := "blue"
        if intersection {
            circleColor = "green"
        }
        drawCircle(ctx, circle.Origin.X, circle.Origin.Y, circle.R, circleColor)

        // draw intersections
        if intersection {
            p1 := ray.Origin.Add( ray.D.Scale(d1) )
            drawCircle(ctx, p1.X, p1.Y, 3, "purple")

            p2 := ray.Origin.Add( ray.D.Scale(d2) )
            drawCircle(ctx, p2.X, p2.Y, 3, "purple")
        }

        ctx.Set("fillStyle", "red")
        drawCircle(ctx, ray.Origin.X, ray.Origin.Y, 5, "red")
        ctx.Call("fill")

        ctx.Call("beginPath")
        ctx.Call("moveTo", ray.Origin.X, ray.Origin.Y)
        ctx.Call("lineTo", ray.Origin.X + ray.D.X*100, ray.Origin.Y + ray.D.Y*100)
        ctx.Call("stroke")

        window.Call("requestAnimationFrame", drawFrame)

        return nil
    })

    tRandomCircle.Set("onmousedown", randomizeCircle())
    tRandomRay.Set("onmousedown", randomizeRay())
    canvas.Set("onmousedown", moveRay())
    window.Call("requestAnimationFrame", drawFrame)

    done := make(chan bool)
    <-done
}
