package furex

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sedyh/furex/v2/geo"

	"github.com/hajimehoshi/ebiten/v2"
)

// View represents a UI element.
// You can set flex options, size, position and so on.
// Handlers can be set to create custom component such as button or list.
type View struct {
	// TODO: Remove these fields in the future.
	Left         float64
	Right        *float64
	Top          float64
	Bottom       *float64
	Width        float64
	WidthInPct   float64
	Height       float64
	HeightInPct  float64
	MarginLeft   float64
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	Position     Position
	Direction    Direction
	Wrap         FlexWrap
	Justify      Justify
	AlignItems   AlignItem
	AlignContent AlignContent
	Grow         float64
	Shrink       float64
	Display      Display

	ID      string
	Raw     string
	TagName string
	Text    string
	Attrs   map[string]string
	Hidden  bool

	Handler Handler

	containerEmbed
	flexEmbed
	lock      sync.Mutex
	hasParent bool
	parent    *View
}

// Update updates the view
func (v *View) Update() {
	if v.isDirty {
		v.startLayout()
	}
	v.processHandler()
	for _, v := range v.children {
		v.item.Update()
		v.item.processHandler()
	}
	if !v.hasParent {
		v.processEvent()
	}
}

func (v *View) processHandler() {
	if u, ok := v.Handler.(UpdateHandler); ok {
		u.HandleUpdate()
		return
	}
	if u, ok := v.Handler.(Updater); ok {
		u.Update(v.frame, v)
		return
	}
}

func (v *View) startLayout() {
	v.lock.Lock()
	defer v.lock.Unlock()
	if !v.hasParent {
		v.frame = geo.Rect(v.Left, v.Top, v.Left+v.Width, v.Top+v.Height)
	}
	v.flexEmbed.View = v

	for _, child := range v.children {
		if child.item.Position == PositionStatic {
			child.item.startLayout()
		}
	}

	v.layout(v.frame.Dx(), v.frame.Dy(), &v.containerEmbed)
	v.isDirty = false
}

// UpdateWithSize the view with modified height and width
func (v *View) UpdateWithSize(width, height float64) {
	if !v.hasParent && (v.Width != width || v.Height != height) {
		v.Height = height
		v.Width = width
		v.isDirty = true
	}
	v.Update()
}

// Layout marks the view as dirty
func (v *View) Layout() {
	v.isDirty = true
	if v.hasParent {
		v.parent.isDirty = true
	}
}

// Draw draws the view
func (v *View) Draw(screen *ebiten.Image) {
	if v.isDirty {
		v.startLayout()
	}
	if !v.hasParent {
		v.handleDrawRoot(screen, v.frame)
	}
	if !v.Hidden && v.Display != DisplayNone {
		v.containerEmbed.Draw(screen)
	}
	if Debug && !v.hasParent && v.Display != DisplayNone {
		debugBorders(screen, v.containerEmbed)
	}
}

// AddTo add itself to a parent view
func (v *View) AddTo(parent *View) *View {
	if v.hasParent {
		panic("this view has been already added to a parent")
	}
	parent.AddChild(v)
	return v
}

// AddChild adds one or multiple child views
func (v *View) AddChild(views ...*View) *View {
	for _, vv := range views {
		v.addChild(vv)
	}
	return v
}

// RemoveChild removes a specified view
func (v *View) RemoveChild(cv *View) bool {
	for i, child := range v.children {
		if child.item == cv {
			v.children = append(v.children[:i], v.children[i+1:]...)
			v.isDirty = true
			cv.hasParent = false
			cv.parent = nil
			return true
		}
	}
	return false
}

// RemoveAll removes all children view
func (v *View) RemoveAll() {
	v.isDirty = true
	for _, child := range v.children {
		child.item.hasParent = false
		child.item.parent = nil
	}
	v.children = []*child{}
}

// PopChild remove the last child view add to this view
func (v *View) PopChild() *View {
	if len(v.children) == 0 {
		return nil
	}
	c := v.children[len(v.children)-1]
	v.children = v.children[:len(v.children)-1]
	v.isDirty = true
	c.item.hasParent = false
	c.item.parent = nil
	return c.item
}

func (v *View) addChild(cv *View) *View {
	child := &child{item: cv, handledTouchID: -1}
	v.children = append(v.children, child)
	v.isDirty = true
	cv.hasParent = true
	cv.parent = v
	return v
}

func (v *View) isWidthFixed() bool {
	return v.Width != 0 || v.WidthInPct != 0
}

func (v *View) width() float64 {
	if v.Width == 0 {
		return v.calculatedWidth
	}
	return v.Width
}

func (v *View) isHeightFixed() bool {
	return v.Height != 0 || v.HeightInPct != 0
}

func (v *View) height() float64 {
	if v.Height == 0 {
		return v.calculatedHeight
	}
	return v.Height
}

func (v *View) getChildren() []*View {
	if v == nil || v.children == nil {
		return nil
	}
	ret := make([]*View, len(v.children))
	for i, child := range v.children {
		ret[i] = child.item
	}
	return ret
}

// GetByID returns the view with the specified id.
// It returns nil if not found.
func (v *View) GetByID(id string) (*View, bool) {
	if v.ID == id {
		return v, true
	}
	for _, child := range v.children {
		if v, ok := child.item.GetByID(id); ok {
			return v, true
		}
	}
	return nil, false
}

// MustGetByID returns the view with the specified id.
// It panics if not found.
func (v *View) MustGetByID(id string) *View {
	vv, ok := v.GetByID(id)
	if !ok {
		panic("view not found")
	}
	return vv
}

// SetLeft sets the left position of the view.
func (v *View) SetLeft(left float64) {
	v.Left = left
	v.Layout()
}

// SetRight sets the right position of the view.
func (v *View) SetRight(right float64) {
	v.Right = Float(right)
	v.Layout()
}

// SetTop sets the top position of the view.
func (v *View) SetTop(top float64) {
	v.Top = top
	v.Layout()
}

// SetBottom sets the bottom position of the view.
func (v *View) SetBottom(bottom float64) {
	v.Bottom = Float(bottom)
	v.Layout()
}

// SetWidth sets the width of the view.
func (v *View) SetWidth(width float64) {
	v.Width = width
	v.Layout()
}

// SetHeight sets the height of the view.
func (v *View) SetHeight(height float64) {
	v.Height = height
	v.Layout()
}

// SetMarginLeft sets the left margin of the view.
func (v *View) SetMarginLeft(marginLeft float64) {
	v.MarginLeft = marginLeft
	v.Layout()
}

// SetMarginTop sets the top margin of the view.
func (v *View) SetMarginTop(marginTop float64) {
	v.MarginTop = marginTop
	v.Layout()
}

// SetMarginRight sets the right margin of the view.
func (v *View) SetMarginRight(marginRight float64) {
	v.MarginRight = marginRight
	v.Layout()
}

// SetMarginBottom sets the bottom margin of the view.
func (v *View) SetMarginBottom(marginBottom float64) {
	v.MarginBottom = marginBottom
	v.Layout()
}

// SetPosition sets the position of the view.
func (v *View) SetPosition(position Position) {
	v.Position = position
	v.Layout()
}

// SetDirection sets the direction of the view.
func (v *View) SetDirection(direction Direction) {
	v.Direction = direction
	v.Layout()
}

// SetWrap sets the wrap property of the view.
func (v *View) SetWrap(wrap FlexWrap) {
	v.Wrap = wrap
	v.Layout()
}

// SetJustify sets the justify property of the view.
func (v *View) SetJustify(justify Justify) {
	v.Justify = justify
	v.Layout()
}

// SetAlignItems sets the align items property of the view.
func (v *View) SetAlignItems(alignItems AlignItem) {
	v.AlignItems = alignItems
	v.Layout()
}

// SetAlignContent sets the align content property of the view.
func (v *View) SetAlignContent(alignContent AlignContent) {
	v.AlignContent = alignContent
	v.Layout()
}

// SetGrow sets the grow property of the view.
func (v *View) SetGrow(grow float64) {
	v.Grow = grow
	v.Layout()
}

// SetShrink sets the shrink property of the view.
func (v *View) SetShrink(shrink float64) {
	v.Shrink = shrink
	v.Layout()
}

// SetDisplay sets the display property of the view.
func (v *View) SetDisplay(display Display) {
	v.Display = display
	v.Layout()
}

// SetHidden sets the hidden property of the view.
func (v *View) SetHidden(hidden bool) {
	v.Hidden = hidden
	v.Layout()
}

func (v *View) Config() ViewConfig {
	cfg := ViewConfig{
		TagName:      v.TagName,
		ID:           v.ID,
		Left:         v.Left,
		Right:        v.Right,
		Top:          v.Top,
		Bottom:       v.Bottom,
		Width:        v.Width,
		Height:       v.Height,
		MarginLeft:   v.MarginLeft,
		MarginTop:    v.MarginTop,
		MarginRight:  v.MarginRight,
		MarginBottom: v.MarginBottom,
		Position:     v.Position,
		Direction:    v.Direction,
		Wrap:         v.Wrap,
		Justify:      v.Justify,
		AlignItems:   v.AlignItems,
		AlignContent: v.AlignContent,
		Grow:         v.Grow,
		Shrink:       v.Shrink,
		children:     []ViewConfig{},
	}
	for _, child := range v.getChildren() {
		cfg.children = append(cfg.children, child.Config())
	}
	return cfg
}

func (v *View) handleDrawRoot(screen *ebiten.Image, b geo.Rectangle) {
	if h, ok := v.Handler.(DrawHandler); ok {
		h.HandleDraw(screen, b)
		return
	}
	if h, ok := v.Handler.(Drawer); ok {
		h.Draw(screen, b, v)
	}
}

// This is for debugging and testing.
type ViewConfig struct {
	TagName      string
	ID           string
	Left         float64
	Right        *float64
	Top          float64
	Bottom       *float64
	Width        float64
	Height       float64
	MarginLeft   float64
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	Position     Position
	Direction    Direction
	Wrap         FlexWrap
	Justify      Justify
	AlignItems   AlignItem
	AlignContent AlignContent
	Grow         float64
	Shrink       float64
	children     []ViewConfig
}

func (cfg ViewConfig) Tree() string {
	return cfg.tree("")
}

// TODO: This is a bit of a mess. Clean it up.
func (cfg ViewConfig) tree(indent string) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s<%s ", indent, cfg.TagName))
	if cfg.ID != "" {
		sb.WriteString(fmt.Sprintf("id=\"%s\" ", cfg.ID))
	}
	sb.WriteString("style=\"")
	sb.WriteString(
		fmt.Sprintf(
			"left: %f, right: %f, top: %f, bottom: %f, width: %f, height: %f, marginLeft: %f, marginTop: %f, marginRight: %f, marginBottom: %f, position: %s, direction: %s, wrap: %s, justify: %s, alignItems: %s, alignContent: %s, grow: %f, shrink: %f",
			cfg.Left, *cfg.Right, cfg.Top, *cfg.Bottom, cfg.Width, cfg.Height, cfg.MarginLeft, cfg.MarginTop, cfg.MarginRight, cfg.MarginBottom, cfg.Position, cfg.Direction, cfg.Wrap, cfg.Justify, cfg.AlignItems, cfg.AlignContent, cfg.Grow, cfg.Shrink))
	sb.WriteString("\">\n")
	for _, child := range cfg.children {
		sb.WriteString(child.tree(indent + "  "))
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("%s</%s>", indent, cfg.TagName))
	sb.WriteString("\n")
	return sb.String()
}
