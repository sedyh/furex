package graphic

import (
	"image/color"
	"sync"

	"github.com/sedyh/furex/v2/geo"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	g    graphic
	once sync.Once
)

type graphic struct {
	imgOfAPixel *ebiten.Image
}

func (g *graphic) setup() {
	once.Do(func() {
		g.imgOfAPixel = ebiten.NewImage(1, 1)
	})
}

type FillRectOpts struct {
	Rect  geo.Rectangle
	Color color.Color
}

func FillRect(target *ebiten.Image, opts *FillRectOpts) {
	g.setup()
	r, c := &opts.Rect, &opts.Color
	g.imgOfAPixel.Fill(*c)
	op := &ebiten.DrawImageOptions{}
	w, h := r.Size().X, r.Size().Y
	op.GeoM.Translate(r.Min.X*(1/float64(w)), r.Min.Y*(1/float64(h)))
	op.GeoM.Scale(w, h)
	target.DrawImage(g.imgOfAPixel, op)
}

type DrawRectOpts struct {
	Rect        geo.Rectangle
	Color       color.Color
	StrokeWidth int
}

func DrawRect(target *ebiten.Image, opts *DrawRectOpts) {
	g.setup()
	r, c, sw := &opts.Rect, &opts.Color, float64(opts.StrokeWidth)
	FillRect(target, &FillRectOpts{
		Rect: geo.Rect(r.Min.X, r.Min.Y, r.Min.X+sw, r.Max.Y), Color: *c,
	})
	FillRect(target, &FillRectOpts{
		Rect: geo.Rect(r.Min.X, r.Min.Y, r.Min.X+sw, r.Max.Y), Color: *c,
	})
	FillRect(target, &FillRectOpts{
		Rect: geo.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+sw), Color: *c,
	})
	FillRect(target, &FillRectOpts{
		Rect: geo.Rect(r.Min.X, r.Max.Y-sw, r.Max.X, r.Max.Y), Color: *c,
	})
}
