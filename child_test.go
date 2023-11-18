package furex

import (
	"testing"
	"time"

	"github.com/sedyh/furex/v2/geo"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		flex *View,
		h *mockHandler,
		frame geo.Rectangle,
	){
		// "button touch": testButtonTouch,
		// "mouse click":  testMouchClick,
		// "mouse move":   testMouseMove,
		"swipe": testSwipe,
	} {

		t.Run(scenario, func(t *testing.T) {

			flex := &View{
				Width:      300,
				Height:     500,
				Left:       100,
				Top:        50,
				Position:   PositionAbsolute,
				Direction:  Column,
				Justify:    JustifyCenter,
				AlignItems: AlignItemCenter,
			}

			flex2 := &View{
				Width:      100,
				Height:     200,
				Direction:  Column,
				Justify:    JustifyEnd,
				AlignItems: AlignItemEnd,
			}

			flex.AddChild(flex2)

			h := &mockHandler{}
			flex2.AddChild(&View{
				Width:   10,
				Height:  20,
				Handler: h,
			})

			// 	(0,0)
			// ┌───────────────────────────────────┐
			// │ view                              │
			// │      (100,50)                     │
			// │      ┌────────────────────────────┤
			// │      │flex(300x500)               │
			// │      │                            │
			// │      │                            │
			// │      │     (100,150)              │
			// │      │     ┌─────────────────┐    │
			// │      │     │flex2(100x200)   │    │
			// │      │     │                 │    │
			// │      │     │   ┌──────-──────┤    │
			// │      │     │   │button(10x20)│    │
			// │      │     │   │             │    │
			// │      │     │   │             │    │
			// │      │     │   │             │    │
			// │      │     │   │             │    │
			// │      │     │   │             │    │
			// │      │     └───┴──────────-──┘    │
			// │      │                  (300,400) │
			// │      │                            │
			// │      │                            │
			// └──────┴────────────────────────────┘
			//                                 (400,550)
			// expected button frame:
			// x = 300-10 = 290 to 300
			// y = 400-20 = 380 to 400

			flex.Update()
			flex.Draw(nil)

			frame := geo.Rect(290, 380, 300, 400)
			require.Equal(t, frame, h.Frame)

			fn(t, flex, h, frame)

		})
	}

}

func testButtonTouch(t *testing.T, flex *View, h *mockHandler, frame geo.Rectangle) {

	type result struct {
		IsPressed  bool
		IsReleased bool
		IsCanceled bool
	}

	var tests = []struct {
		Scenario string
		Start    geo.Point
		End      geo.Point
		Want     result
	}{
		{
			Scenario: "press inside and release inside",
			Start:    frame.Min,
			End:      frame.Min,
			Want:     result{true, true, false},
		},
		{
			Scenario: "press inside and release outside",
			Start:    frame.Min,
			End:      geo.Pt(frame.Min.X, frame.Min.Y-1),
			Want:     result{true, true, true},
		},
		{
			Scenario: "press inside and release inside (right-bottom)",
			Start:    frame.Max,
			End:      frame.Max,
			Want:     result{true, true, false},
		},
		{
			Scenario: "press inside and release outside (right-bottom)",
			Start:    frame.Max,
			End:      geo.Pt(frame.Max.X+1, frame.Max.Y),
			Want:     result{true, true, true},
		},
		{
			Scenario: "press outside and release inside",
			Start:    geo.Pt(frame.Min.X-1, frame.Min.Y),
			End:      geo.Pt(frame.Min.X+frame.Dx()/2, frame.Min.Y+frame.Dy()/2),
			Want:     result{false, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			h.Init()

			flex.HandleJustPressedTouchID(0, int(tt.Start.X), int(tt.Start.Y))
			flex.HandleJustReleasedTouchID(0, int(tt.End.X), int(tt.End.Y))

			assert.Equal(t, tt.Want, result{h.IsPressed, h.IsReleased, h.IsCancel})
		})
	}
}

func testMouchClick(t *testing.T, flex *View, h *mockHandler, frame geo.Rectangle) {

	type result struct {
		IsPressed  bool
		IsReleased bool
		IsCancel   bool
	}

	var tests = []struct {
		Scenario string
		Start    geo.Point
		End      geo.Point
		Want     result
	}{
		{
			Scenario: "press inside and release inside",
			Start:    frame.Min,
			End:      frame.Min,
			Want:     result{true, true, false},
		},
		{
			Scenario: "press inside left-top edge, release outside",
			Start:    frame.Min,
			End:      geo.Pt(frame.Min.X, frame.Min.Y-1),
			Want:     result{true, true, true},
		},
		{
			Scenario: "press inside righ-bottom edge, release inside",
			Start:    frame.Max,
			End:      frame.Max,
			Want:     result{true, true, false},
		},
		{
			Scenario: "press inside righ-bottom edge, release outside",
			Start:    frame.Max,
			End:      geo.Pt(frame.Max.X+1, frame.Max.Y),
			Want:     result{true, true, true},
		},
		{
			Scenario: "press outside, release inside",
			Start:    geo.Pt(frame.Min.X-1, frame.Min.Y),
			End:      geo.Pt(frame.Min.X+frame.Dx()/2, frame.Min.Y+frame.Dy()/2),
			Want:     result{false, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			h.Init()

			flex.handleMouseButtonLeftPressed(int(tt.Start.X), int(tt.Start.Y))
			flex.handleMouseButtonLeftReleased(int(tt.End.X), int(tt.End.Y))

			assert.Equal(t, tt.Want, result{h.IsPressed, h.IsReleased, h.IsCancel})
		})
	}
}

func testMouseMove(t *testing.T, flex *View, h *mockHandler, frame geo.Rectangle) {
	type result struct {
		IsMouseMoved bool
		MousePoint   geo.Point
	}
	var tests = []struct {
		Scenario string
		Point    geo.Point
		Want     result
	}{
		{
			Scenario: "move mouse left-top inside",
			Point:    geo.Point{frame.Min.X, frame.Min.Y},
			Want:     result{IsMouseMoved: true, MousePoint: geo.Point{frame.Min.X, frame.Min.Y}},
		},
		{
			Scenario: "move mouse right-bottom inside",
			Point:    geo.Point{frame.Max.X, frame.Max.Y},
			Want:     result{IsMouseMoved: true, MousePoint: geo.Point{frame.Max.X, frame.Max.Y}},
		},
		{
			Scenario: "move mouse left outside",
			Point:    geo.Point{frame.Min.X - 1, frame.Min.Y},
			Want:     result{IsMouseMoved: false, MousePoint: geo.Point{-1, -1}},
		},
		{
			Scenario: "move mouse right outside",
			Point:    geo.Point{frame.Max.X + 1, frame.Min.Y},
			Want:     result{IsMouseMoved: false, MousePoint: geo.Point{-1, -1}},
		},
		{
			Scenario: "move mouse top outside",
			Point:    geo.Point{frame.Min.X, frame.Min.Y - 1},
			Want:     result{IsMouseMoved: false, MousePoint: geo.Point{-1, -1}},
		},
		{
			Scenario: "move mouse bottom outside",
			Point:    geo.Point{frame.Min.X, frame.Max.Y + 1},
			Want:     result{IsMouseMoved: false, MousePoint: geo.Point{-1, -1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			h.Init()

			flex.handleMouse(int(tt.Point.X), int(tt.Point.Y))

			assert.Equal(t, tt.Want, result{h.IsMouseMoved, h.MousePoint})
		})
	}
}

func testSwipe(t *testing.T, flex *View, h *mockHandler, frame geo.Rectangle) {
	type result struct {
		IsSwiped bool
		SwipeDir SwipeDirection
	}
	var tests = []struct {
		Scenario string
		From     geo.Point
		To       geo.Point
		Time     time.Duration
		Want     result
	}{
		{
			Scenario: "swipe left",
			From:     geo.Point{frame.Min.X, frame.Min.Y},
			To:       geo.Point{frame.Min.X - 50, frame.Min.Y},
			Time:     time.Duration(0),
			Want:     result{IsSwiped: true, SwipeDir: SwipeDirectionLeft},
		},
		{
			Scenario: "swipe right",
			From:     geo.Point{frame.Min.X, frame.Min.Y},
			To:       geo.Point{frame.Min.X + 50, frame.Min.Y},
			Time:     time.Millisecond * 50,
			Want:     result{IsSwiped: true, SwipeDir: SwipeDirectionRight},
		},
		{
			Scenario: "swipe down",
			From:     geo.Point{frame.Min.X, frame.Min.Y},
			To:       geo.Point{frame.Min.X, frame.Min.Y + 50},
			Time:     time.Millisecond * 50,
			Want:     result{IsSwiped: true, SwipeDir: SwipeDirectionDown},
		},
		{
			Scenario: "swipe slow",
			From:     geo.Point{frame.Min.X, frame.Min.Y},
			To:       geo.Point{frame.Min.X, frame.Min.Y + 50},
			Time:     time.Millisecond * 301,
			Want:     result{IsSwiped: false},
		},
		{
			Scenario: "swipe short",
			From:     geo.Point{frame.Min.X, frame.Min.Y},
			To:       geo.Point{frame.Min.X, frame.Min.Y + 49},
			Time:     time.Millisecond * 50,
			Want:     result{IsSwiped: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			h.Init()

			flex.HandleJustPressedTouchID(0, int(tt.From.X), int(tt.From.Y))
			<-time.After(tt.Time)
			flex.HandleJustReleasedTouchID(0, int(tt.To.X), int(tt.To.Y))
			if tt.Want.IsSwiped {
				assert.Equal(t, tt.Want, result{h.IsSwiped, h.SwipeDir})
			} else {
				assert.False(t, h.IsSwiped)
			}
		})
	}
}

type mockHandler struct {
	mockFlags
	Frame      geo.Rectangle
	MousePoint geo.Point
	SwipeDir   SwipeDirection
}

type mockFlags struct {
	IsPressed    bool
	IsReleased   bool
	IsCancel     bool
	IsUpdated    bool
	IsDrawn      bool
	IsMouseMoved bool
	IsSwiped     bool
}

var _ DrawHandler = (*mockHandler)(nil)
var _ UpdateHandler = (*mockHandler)(nil)
var _ ButtonHandler = (*mockHandler)(nil)
var _ MouseHandler = (*mockHandler)(nil)
var _ SwipeHandler = (*mockHandler)(nil)

func (h *mockHandler) Init() {
	h.mockFlags = mockFlags{}
	h.MousePoint = geo.Pt(-1, -1)
}

func (h *mockHandler) HandleUpdate() {
	h.IsUpdated = true
}

func (h *mockHandler) HandleDraw(screen *ebiten.Image, frame geo.Rectangle) {
	h.Frame = frame
	h.IsDrawn = true
}

func (h *mockHandler) HandlePress(x, y int, t ebiten.TouchID) {
	h.IsPressed = true
}

func (h *mockHandler) HandleRelease(x, y int, isCancel bool) {
	h.IsReleased = true
	h.IsCancel = isCancel
}

func (h *mockHandler) HandleMouse(x, y int) bool {
	h.IsMouseMoved = true
	h.MousePoint = geo.Pt(float64(x), float64(y))
	return true
}

func (h *mockHandler) HandleSwipe(dir SwipeDirection) {
	h.IsSwiped = true
	h.SwipeDir = dir
}
