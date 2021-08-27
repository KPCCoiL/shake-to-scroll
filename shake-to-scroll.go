package main

import (
	"io/ioutil"
	"log"
	"math"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type Vector struct {
	x float64
	y float64
}

type PastPosition struct {
	pos  Vector
	time time.Time
}

func logPosition(win *gtk.Window) PastPosition {
	x, y := win.GetPosition()
	return PastPosition{
		pos:  Vector{x: float64(x), y: float64(y)},
		time: time.Now(),
	}
}

func velocity(win *gtk.Window, prev PastPosition) Vector {
	x, y := win.GetPosition()
	dx := float64(x) - prev.pos.x
	dy := float64(y) - prev.pos.y
	dt := float64(time.Since(prev.time).Microseconds()) / 1e6
	return Vector{x: dx / dt, y: dy / dt}
}

func addSampleText(buf *gtk.TextBuffer) {
	// generated thanks to https://lipsum.com/
	content, err := ioutil.ReadFile("lorem-ipsum.txt")
	if err != nil {
		log.Fatal("Could not read terms of use: ", err)
	}

	buf.SetText("Terms of Use\n\n" + string(content))
	tag := buf.CreateTag("headline", map[string]interface{}{
		"size-points": 24.0,
	})
	buf.ApplyTag(tag, buf.GetStartIter(), buf.GetIterAtLine(1))
}

func main() {
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Terms of Use")
	win.SetDefaultSize(600, 400)
	win.SetResizable(false)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	scroll, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create scrolled window:", err)
	}
	win.Add(scroll)
	scroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_EXTERNAL)

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	if err != nil {
		log.Fatal("Unable to create box:", err)
	}
	scroll.Add(box)

	textView, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal("Unable to create TextView:", err)
	}
	box.Add(textView)
	textView.SetWrapMode(gtk.WRAP_WORD)
	textView.SetEditable(false)

	buffer, err := textView.GetBuffer()
	if err != nil {
		log.Fatal("Unable to get the buffer of TextView:", err)
	}

	confirmation, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Unable to create box: ", err)
	}
	box.Add(confirmation)

	checkButton, err := gtk.CheckButtonNewWithLabel("I have read and accept the Terms of Use")
	if err != nil {
		log.Fatal("Unable to create check button: ", err)
	}
	confirmation.Add(checkButton)

	button, err := gtk.ButtonNew()
	if err != nil {
		log.Fatal("Unable to create button: ", err)
	}
	confirmation.Add(button)
	button.SetLabel("Continue")
	button.SetSensitive(false)

	checkButton.Connect("toggled", func() {
		button.SetSensitive(checkButton.GetActive())
	})
	button.Connect("clicked", func() {
		button.SetLabel("Yay!")
	})

	first_time := true
	record := logPosition(win)
	adjustment := scroll.GetVAdjustment()
	acceleration := 0.0
	speed := 0.0
	ratio := 0.0
	glib.TimeoutAdd(10, func() bool {
		v := velocity(win, record)
		record = logPosition(win)

		if first_time {
			first_time = false
			return true
		}

		const (
			lambda          = 1e-3
			spring_constant = 1
		)
		drag_speed := math.Hypot(v.x, v.y)
		acceleration = lambda*drag_speed - spring_constant*ratio
		log.Println(ratio, speed, acceleration, drag_speed)

		return true
	})
	lastTime := time.Now()
	glib.TimeoutAdd(10, func() bool {
		now := time.Now()
		dt := float64(now.Sub(lastTime)) / 1e9
		lastTime = now

		ratio += speed * dt
		speed += acceleration * dt

		if ratio < 0 {
			ratio = 0
			speed = 0
		}

		if ratio > 1 {
			ratio = 1
			speed = 0
		}

		minimum := adjustment.GetLower()
		maximum := adjustment.GetUpper() - adjustment.GetPageSize()

		adjustment.SetValue(minimum + (maximum-minimum)*ratio)

		return true
	})

	win.ShowAll()

	addSampleText(buffer)
	gtk.Main()

}
