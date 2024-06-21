package main

import "fmt"
import "math"
import "syscall/js"

const WIDTH = 300
const HEIGHT = 600

type Vec3 struct {
	X float32
	Y float32
	Z float32
}

type Number interface {
    ~int | ~int32 | ~int64 | ~byte | ~float32 | ~float64
}

func lerp[T Number](a T, b T, t float32) float32 {
    return float32(a) + float32(b-a)*t
}

func alerp[T Number](a []T, b []T, t float32) []float32 {
    c := make([]float32, len(a))
    for idx := 0; idx < len(a); idx++ {
        c[idx] = lerp(a[idx], b[idx], t)
    }
    return c
}

func (v Vec3) Neg() Vec3 {
	return Vec3{
		X: -v.X,
		Y: -v.Y,
		Z: -v.Z,
	}
}

func (v Vec3) Add(b Vec3) Vec3 {
	return Vec3{
		X: v.X + b.X,
		Y: v.Y + b.Y,
		Z: v.Z + b.Z,
	}
}

func (v Vec3) Unit() Vec3 {
	size := float32(math.Sqrt(float64(v.Dot(v))))
	return Vec3{
		X: v.X / size,
		Y: v.Y / size,
		Z: v.Z / size,
	}
}

func (v Vec3) Dot(b Vec3) float32 {
	return v.X*b.X + v.Y*b.Y + v.Z*b.Z
}

func (v Vec3) Dot64(b Vec3) float64 {
	return float64(v.X*b.X + v.Y*b.Y + v.Z*b.Z)
}

func (v Vec3) Scale(t float32) Vec3 {
	return Vec3{
		X: v.X * t,
		Y: v.Y * t,
		Z: v.Z * t,
	}
}

type Sphere struct {
	Origin Vec3
	R      float32
}

type Ray struct {
	Origin Vec3
	D      Vec3
}

var sphere *Sphere = &Sphere{
	Origin: Vec3{0, 0, -30},
	R:      0.5,
}

func (r *Ray) intersectSphere(c *Sphere) (bool, float32) {
	/*
		   circumference:
		   x**2 + y**2 + z**2 = r**3

	       parametric (t) definition of a line:
		   x = Ox + Dx * t
		   y = Oy + Dy * t
	       z = Oz + Dz * t

		   eq:
		   (Ox + Dx * t)**2 + (Oy + Dy * t)**2 = r**2

		   (Dx^2 + Dy^2 + Dz^2)*t^2 + 2*(Ox*Dx+Oy*Dy*Oz*Dz)*t + (Ox^2+Oy^2+Oz^2) - r^3 = 0
		   ^^^^^^^^^^^^^^^^^^^^       ^^^^^^^^^^^^^^^^^^^^^^^     ^^^^^^^^^^^^^^^^^^^^
		   A                   B                   C

	*/

	// ray shifted so that the circle is at the zero origin
	or := r.Origin.Add(c.Origin.Neg())
	(*r).D = r.D.Unit()
	R := c.R

	// quadratic equation coefficients
	A := r.D.Dot(r.D)
	B := 2.0 * or.Dot(r.D)
	C := or.Dot(or) - R*R*R

	// discriminant
	D := B*B - 4.0*A*C

	// no intersection
	if D < 0 {
		return false, 0
	}
    Dsq := float32(math.Sqrt(float64(D)))

	t1 := (-B + Dsq) / (2.0 * A)
	t2 := (-B - Dsq) / (2.0 * A)

    t := t1
    if t2 < t1 && t2 >= 0 {
        t = t2
    }

	if t < 0 {
		// backward hit
		return false, 0
	}

	return true, t
}

func runShader(u, v, aspect float32) [4]byte {
	ray := Ray{
		Origin: Vec3{u * aspect, v, 1.0},
		D:      Vec3{0, 0, -1.0}.Unit(),
	}

	intersects, dist := ray.intersectSphere(sphere)
    bgr := [4]float32{23.0, 23.0, 23.0, 255.0}
    obj := [4]float32{200.0, 20.0, 220.0, 255.0}
	if intersects {
        hit := ray.Origin.Add( ray.D.Scale(dist) )
        normal := hit.Add(sphere.Origin.Neg()).Unit()

        cosalpha := normal.Z

        res := alerp(bgr[:], obj[:], cosalpha)
		return [4]byte{byte(res[0]), byte(res[1]), byte(res[2]), byte(res[3])}
	}

	return [4]byte{23, 23, 23, 255}
}

func main() {
	document := js.Global().Get("document")
	window := js.Global().Get("window")
	// get the dom element
	canvas := document.Call("getElementById", "tCanvas")
	tFrameCounter := document.Call("getElementById", "tFrameCounter")

	// initialize 2d context
	ctx := canvas.Call("getContext", "2d")

	var frameCounter int32 = 0
	var timePrev float32 = 0

	imageData := ctx.Call("createImageData", WIDTH, HEIGHT)
	pixelData := imageData.Get("data")

	var drawFrame js.Func
	drawFrame = js.FuncOf(func(this js.Value, args []js.Value) any {
		timeNow := float32(args[0].Float())
		deltaTime := timeNow - timePrev

		if deltaTime > 1000 {
			fps := float32(frameCounter*1000) / deltaTime
			perfStr := fmt.Sprintf("%0.2f last generation: %0.3fms", fps, deltaTime/float32(frameCounter))

			timePrev = timeNow
			tFrameCounter.Set("innerHTML", perfStr)
			frameCounter = 0
		}
		frameCounter++

		for row := 0; row < HEIGHT; row++ {
			for col := 0; col < WIDTH; col++ {
				u := (float32(col)/float32(WIDTH))*2.0 - 1
				v := (float32(row)/float32(HEIGHT))*2.0 - 1

				aspect := float32(WIDTH) / float32(HEIGHT)

				idx := 4 * (row*WIDTH + col)
				color := runShader(u, v, aspect)

				for off, v := range color {
					pixelData.SetIndex(idx+off, v)
				}
			}
		}
		ctx.Call("putImageData", imageData, 0, 0)

		window.Call("requestAnimationFrame", drawFrame)

		return nil
	})

	window.Call("requestAnimationFrame", drawFrame)
	done := make(chan bool)
	<-done
}
