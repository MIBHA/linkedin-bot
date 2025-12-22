package stealth

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Point struct {
	X float64
	Y float64
}

type BezierCurve struct {
	Start    Point
	End      Point
	Control1 Point
	Control2 Point
}

func NewBezierCurve(start, end Point) *BezierCurve {
	dx := end.X - start.X
	dy := end.Y - start.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	offsetMagnitude := distance * (0.2 + rand.Float64()*0.3)
	angle1 := rand.Float64() * math.Pi * 2
	angle2 := rand.Float64() * math.Pi * 2
	
	control1 := Point{
		X: start.X + dx*0.25 + math.Cos(angle1)*offsetMagnitude,
		Y: start.Y + dy*0.25 + math.Sin(angle1)*offsetMagnitude,
	}
	
	control2 := Point{
		X: start.X + dx*0.75 + math.Cos(angle2)*offsetMagnitude,
		Y: start.Y + dy*0.75 + math.Sin(angle2)*offsetMagnitude,
	}
	
	return &BezierCurve{Start: start, End: end, Control1: control1, Control2: control2}
}

func (bc *BezierCurve) CalculatePoint(t float64) Point {
	oneMinusT := 1 - t
	x := math.Pow(oneMinusT, 3)*bc.Start.X + 3*math.Pow(oneMinusT, 2)*t*bc.Control1.X + 3*oneMinusT*math.Pow(t, 2)*bc.Control2.X + math.Pow(t, 3)*bc.End.X
	y := math.Pow(oneMinusT, 3)*bc.Start.Y + 3*math.Pow(oneMinusT, 2)*t*bc.Control1.Y + 3*oneMinusT*math.Pow(t, 2)*bc.Control2.Y + math.Pow(t, 3)*bc.End.Y
	return Point{X: x, Y: y}
}

func (bc *BezierCurve) GeneratePath(numPoints int) []Point {
	points := make([]Point, numPoints)
	for i := 0; i < numPoints; i++ {
		t := float64(i) / float64(numPoints-1)
		points[i] = bc.CalculatePoint(t)
	}
	return points
}

type MoveMouseOptions struct {
	Duration  int
	Overshoot bool
	Steps     int
}

func DefaultMoveMouseOptions() *MoveMouseOptions {
	return &MoveMouseOptions{Duration: 300 + rand.Intn(500), Overshoot: rand.Float64() < 0.15, Steps: 20 + rand.Intn(30)}
}

func MoveMouseBezier(page *rod.Page, targetX, targetY float64, opts *MoveMouseOptions) error {
	if opts == nil {
		opts = DefaultMoveMouseOptions()
	}
	currentX, currentY := 500.0, 500.0
	start := Point{X: currentX, Y: currentY}
	end := Point{X: targetX, Y: targetY}
	curve := NewBezierCurve(start, end)
	path := curve.GeneratePath(opts.Steps)
	stepDelay := time.Duration(opts.Duration/opts.Steps) * time.Millisecond
	
	for i, point := range path {
		delay := stepDelay
		if i < len(path)/4 || i > len(path)*3/4 {
			delay = delay * 3 / 2
		}
		if err := page.Mouse.MoveTo(proto.Point{X: point.X, Y: point.Y}); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	
	if opts.Overshoot {
		overshootDistance := 5.0 + rand.Float64()*15.0
		angle := rand.Float64() * math.Pi * 2
		overshootX := targetX + math.Cos(angle)*overshootDistance
		overshootY := targetY + math.Sin(angle)*overshootDistance
		_ = page.Mouse.MoveTo(proto.Point{X: overshootX, Y: overshootY})
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		_ = page.Mouse.MoveTo(proto.Point{X: targetX, Y: targetY})
	}
	return nil
}

func ClickWithHumanDelay(page *rod.Page) error {
	time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)
	err := page.Mouse.Click("left", 1)
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
	return nil
}

func MoveAndClick(page *rod.Page, x, y float64, opts *MoveMouseOptions) error {
	if err := MoveMouseBezier(page, x, y, opts); err != nil {
		return err
	}
	return ClickWithHumanDelay(page)
}

func ScrollWithVariation(page *rod.Page, deltaY float64) error {
	chunks := 3 + rand.Intn(5)
	chunkSize := deltaY / float64(chunks)
	for i := 0; i < chunks; i++ {
		variation := chunkSize * (0.8 + rand.Float64()*0.4)
		err := page.Mouse.Scroll(0, variation, chunks)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)
	}
	if rand.Float64() < 0.2 {
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
		_ = page.Mouse.Scroll(0, -deltaY*0.1, 1)
	}
	return nil
}
