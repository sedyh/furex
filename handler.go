package furex

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sedyh/furex/v2/geo"
)

// Handler represents a component that can be added to a container.
type Handler interface{}

// Drawer represents a component that can be added to a container.
type Drawer interface {
	// Draw function draws the content of the component inside the frame.
	Draw(screen *ebiten.Image, frame geo.Rectangle, v *View)
}

// Updater represents a component that updates by one tick.
type Updater interface {
	// Update updates the state of the component by one tick.
	Update(frame geo.Rectangle, v *View)
}

// DrawHandler represents a component that can be added to a container.
// Deprectead: use Drawer instead
type DrawHandler interface {
	// HandleDraw function draws the content of the component inside the frame.
	// The frame parameter represents the location (x,y) and size (width,height) relative to the window (0,0).
	HandleDraw(screen *ebiten.Image, frame geo.Rectangle)
}

// UpdateHandler represents a component that updates by one tick.
// Deprectead: use Updater instead
type UpdateHandler interface {
	// Updater updates the state of the component by one tick.
	HandleUpdate()
}

// ButtonHandler represents a button component.
type ButtonHandler interface {
	// HandlePress handle the event when user just started pressing the button
	// The parameter (x, y) is the location relative to the window (0,0).
	// touchID is the unique ID of the touch.
	// If the button is pressed by a mouse, touchID is -1.
	HandlePress(x, y int, t ebiten.TouchID)

	// HandleRelease handle the event when user just released the button.
	// The parameter (x, y) is the location relative to the window (0,0).
	// The parameter isCancel is true when the touch/left click is released outside of the button.
	HandleRelease(x, y int, isCancel bool)
}

// NotButton represents a component that is not a button.
// TODO: update HandlePress to return bool in the next major version.
type NotButton interface {
	// IsButton returns true if the handler is a button.
	IsButton() bool
}

// TouchHandler represents a component that handle touches.
type TouchHandler interface {
	// HandleJustPressedTouchID handles the touchID just pressed and returns true if it handles the TouchID
	HandleJustPressedTouchID(touch ebiten.TouchID, x, y int) bool
	// HandleJustReleasedTouchID handles the touchID just released
	// Should be called only when it handled the TouchID when pressed
	HandleJustReleasedTouchID(touch ebiten.TouchID, x, y int)
}

// MouseHandler represents a component that handle mouse move.
type MouseHandler interface {
	// HandleMouse handles the much move and returns true if it handles the mouse move.
	// The parameter (x, y) is the location relative to the window (0,0).
	HandleMouse(x, y int) bool
}

// MouseLeftButtonHandler represents a component that handle mouse button left click.
type MouseLeftButtonHandler interface {
	// HandleJustPressedMouseButtonLeft handle left mouse button click just pressed.
	// The parameter (x, y) is the location relative to the window (0,0).
	// It returns true if it handles the mouse move.
	HandleJustPressedMouseButtonLeft(x, y int) bool
	// HandleJustReleasedMouseButtonLeft handles the touchID just released.
	// The parameter (x, y) is the location relative to the window (0,0).
	HandleJustReleasedMouseButtonLeft(x, y int)
}

// MouseEnterLeaveHandler represets a component that handle mouse enter.
type MouseEnterLeaveHandler interface {
	// HandleMouseEnter handles the mouse enter.
	HandleMouseEnter(x, y int) bool
	// HandleMouseLeave handles the mouse leave.
	HandleMouseLeave()
}

// SwipeDirection represents different swipe directions.
type SwipeDirection int

const (
	SwipeDirectionLeft SwipeDirection = iota
	SwipeDirectionRight
	SwipeDirectionUp
	SwipeDirectionDown
)

// SwipeHandler represents a component that handle swipe.
type SwipeHandler interface {
	// HandleSwipe handles swipes.
	HandleSwipe(dir SwipeDirection)
}

type handler struct {
	opts HandlerOpts
}

// HandlerOpts represents the options for a handler.
type HandlerOpts struct {
	Update        func(v *View)
	Draw          func(screen *ebiten.Image, frame geo.Rectangle, v *View)
	HandlePress   func(x, y int, t ebiten.TouchID)
	HandleRelease func(x, y int, isCancel bool)
}

// NewHandler creates a new handler.
func NewHandler(opts HandlerOpts) Handler {
	return &handler{opts: opts}
}

func (h *handler) Update(v *View) {
	if h.opts.Update != nil {
		h.opts.Update(v)
	}
}

func (h *handler) Draw(screen *ebiten.Image, frame geo.Rectangle, v *View) {
	if h.opts.Draw != nil {
		h.opts.Draw(screen, frame, v)
	}
}

func (h *handler) HandlePress(x, y int, t ebiten.TouchID) {
	if h.opts.HandlePress != nil {
		h.opts.HandlePress(x, y, t)
	}
}

func (h *handler) HandleRelease(x, y int, isCancel bool) {
	if h.opts.HandleRelease != nil {
		h.opts.HandleRelease(x, y, isCancel)
	}
}
