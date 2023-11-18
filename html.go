package furex

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/vanng822/go-premailer/premailer"
	"golang.org/x/net/html"
)

// The Component can be either a handler instance (e.g., DrawHandler), a factory function
// func() furex.Handler, or a function component func() *furex.View.
// This allows flexibility in usage:
// If you want to reuse the same handler instance for multiple HTML tags, pass the instance;
// otherwise, pass the factory function to create separate handler instances for each tag.
type Component interface{}

// ComponentsMap is a type alias for a dictionary that associates
// custom HTML tags with their respective components.
// This enables a convenient way to manage and reference components
// based on their corresponding HTML tags.
type ComponentsMap map[string]Component

// ParseOptions represents options for parsing HTML.
type ParseOptions struct {
	// Components is a map of component name and handler.
	// For example, if you have a component named "start-button", you can define a handler
	// for it like this:
	// 	opts.Components := map[string]Handler{
	// 		"start-button": <your handler>,
	//  }
	// The component name must be in kebab-case.
	// You can use the component in your HTML like this:
	// 	<start-button></start-button>
	// Note: self closing tag is not supported.
	Components ComponentsMap
	// Width and Height is the size of the root view.
	// This is useful when you want to specify the width and height
	// outside of the HTML.
	Width  float64
	Height float64

	// Handler is the handler for the root view.
	Handler Handler
}

func Parse(input string, opts *ParseOptions) *View {
	if opts == nil {
		opts = &ParseOptions{}
	}

	inlinedHTML := inlineCSS(input)
	z := html.NewTokenizer(strings.NewReader(inlinedHTML))
	dummy := &View{}
	stack := &stack{stack: []*View{dummy}}
	depth := 0
	inBody := false
	cms := []ComponentsMap{opts.Components, registerdComponents}
Loop:
	for {
		tt := z.Next()
		tn, _ := z.TagName()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				break Loop
			}
			panic(z.Err())
		case html.StartTagToken:
			if string(tn) == "body" {
				inBody = true
				continue
			}
			if !inBody {
				continue
			}
			view := processTag(z, string(tn), opts, depth, cms)
			if view == nil {
				continue
			}
			stack.peek().AddChild(view)
			stack.push(view)

			depth++
		case html.SelfClosingTagToken:
			view := processTag(z, string(tn), opts, depth, cms)
			if view == nil {
				continue
			}
			stack.peek().AddChild(view)
		case html.TextToken:
			if stack.len() > 0 {
				stack.peek().Text = strings.TrimSpace(string(z.Text()))
			}
		case html.EndTagToken:
			if string(tn) == "body" {
				inBody = false
				continue
			}
			if !inBody {
				continue
			}
			stack.pop()
			depth--
		}
	}
	if len(dummy.children) != 1 {
		panic(fmt.Sprintf("invalid html: %s", input))
	}
	view := dummy.PopChild()
	// the root view should be dirty for the first time
	// even if the view does not have any children
	view.isDirty = true
	if opts.Handler != nil {
		view.Handler = opts.Handler
	}
	return view
}

func inlineCSS(doc string) string {
	prem, err := premailer.NewPremailerFromString(doc, &premailer.Options{})
	if err != nil {
		println(fmt.Errorf("invalid css: %s", err))
		return doc
	}
	html, err := prem.Transform()
	if err != nil {
		println(fmt.Errorf("error transform html: %s", err))
		return doc
	}
	return html
}

type stack struct {
	stack []*View
}

func (s *stack) push(v *View) {
	s.stack = append(s.stack, v)
}

func (s *stack) len() int {
	return len(s.stack)
}

func (s *stack) peek() *View {
	return s.stack[len(s.stack)-1]
}

func (s *stack) pop() *View {
	v := s.peek()
	s.stack = s.stack[:len(s.stack)-1]
	return v
}

var (
	defaultComponents   = ComponentsMap{"div": nil, "view": nil}
	registerdComponents = defaultComponents
)

func RegisterComponents(cs ComponentsMap) {
	for k, v := range cs {
		register(k, v)
	}
}

func register(name string, c Component) { registerdComponents[name] = c }
func resetComponents()                  { registerdComponents = defaultComponents }

type cms []ComponentsMap

func processTag(z *html.Tokenizer, tagName string, opts *ParseOptions, depth int, cms cms) *View {
	view := createView(tagName, cms)

	if depth == 0 {
		processRootView(view, opts)
	}

	view.TagName = tagName
	view.Raw = string(z.Raw())

	setStyleProps(view, readAttrs(z))

	return view
}

func setStyleProps(view *View, attrs attrs) {
	parseStyle(view, attrs.style)

	view.ID = attrs.id
	view.Attrs = attrs.miscs
	view.Hidden = attrs.hidden
}

func processRootView(view *View, opts *ParseOptions) {
	if opts.Width != 0 {
		view.Width = opts.Width
	}
	if opts.Height != 0 {
		view.Height = opts.Height
	}
}

func createView(name string, cms cms) *View {
	view := &View{}
	for _, cm := range cms {
		if ok := component(name, cm, view); ok {
			return view
		}
	}
	panic(fmt.Sprintf("unknown component: %s", name))
}

func component(name string, m ComponentsMap, v *View) bool {
	c, ok := m[name]
	if c == nil {
		return ok
	}
	if c, ok := c.(func() Handler); ok {
		v.Handler = c()
		return true
	}
	if c, ok := c.(func() *View); ok {
		*v = *c()
		return true
	}
	v.Handler = c
	return true
}

func parseStyle(view *View, style string) {
	pairs := strings.Split(style, ";")
	errs := &ErrorList{}
	for _, pair := range pairs {
		kv := strings.Split(pair, ":")
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])

		mapper, ok := styleMapper[k]
		if !ok {
			errs.Add(fmt.Errorf("unknown style: %s", k))
			continue
		}
		parsed, err := mapper.parseFunc(v)
		if err != nil {
			errs.Add(err)
			continue
		}
		mapper.setFunc(view, parsed)
	}
	if errs.HasErrors() {
		println(fmt.Sprintf("parse style errors: %v", errs))
	}
}

func Int(i int) *int           { return &i }
func Float(f float64) *float64 { return &f }

var styleMapper = map[string]mapper[View]{
	"left": {
		parseFunc: parseNumber,
		setFunc:   setFunc(func(v *View, val float64) { v.Left = val }),
	},
	"right": {
		parseFunc: parseNumber,
		setFunc:   setFunc(func(v *View, val float64) { v.Right = Float(val) }),
	},
	"top": {
		parseFunc: parseNumber,
		setFunc:   setFunc(func(v *View, val float64) { v.Top = val }),
	},
	"bottom": {
		parseFunc: parseNumber,
		setFunc:   setFunc(func(v *View, val float64) { v.Bottom = Float(val) }),
	},
	"width": {
		parseFunc: parseLength,
		setFunc: setFunc(func(v *View, val cssLength) {
			switch val.unit {
			case cssUnitPx:
				v.Width = val.val
			case cssUnitPct:
				v.WidthInPct = val.val
			}
		}),
	},
	"height": {
		parseFunc: parseLength,
		setFunc: setFunc(func(v *View, val cssLength) {
			switch val.unit {
			case cssUnitPx:
				v.Height = val.val
			case cssUnitPct:
				v.HeightInPct = val.val
			}
		}),
	},
	"margin-left": {
		parseFunc: parseNumber,
		setFunc:   setFunc(func(v *View, val float64) { v.MarginLeft = val }),
	},
	"margin-top": {
		parseFunc: parseNumber,
		setFunc:   setFunc(func(v *View, val float64) { v.MarginTop = val }),
	},
	"margin-right": {
		parseFunc: parseNumber,
		setFunc:   setFunc(func(v *View, val float64) { v.MarginRight = val }),
	},
	"margin-bottom": {
		parseFunc: parseNumber,
		setFunc:   setFunc(func(v *View, val float64) { v.MarginBottom = val }),
	},
	"position": {
		parseFunc: parsePosition,
		setFunc:   setFunc(func(v *View, val Position) { v.Position = val }),
	},
	"direction": {
		parseFunc: parseDirection,
		setFunc:   setFunc(func(v *View, val Direction) { v.Direction = val }),
	},
	"flex-direction": {
		parseFunc: parseDirection,
		setFunc:   setFunc(func(v *View, val Direction) { v.Direction = val }),
	},
	"flex-wrap": {
		parseFunc: parseWrap,
		setFunc:   setFunc(func(v *View, val FlexWrap) { v.Wrap = val }),
	},
	"wrap": {
		parseFunc: parseWrap,
		setFunc:   setFunc(func(v *View, val FlexWrap) { v.Wrap = val }),
	},
	"justify": {
		parseFunc: parseJustify,
		setFunc:   setFunc(func(v *View, val Justify) { v.Justify = val }),
	},
	"justify-content": {
		parseFunc: parseJustify,
		setFunc:   setFunc(func(v *View, val Justify) { v.Justify = val }),
	},
	"align-items": {
		parseFunc: parseAlignItem,
		setFunc:   setFunc(func(v *View, val AlignItem) { v.AlignItems = val }),
	},
	"align-content": {
		parseFunc: parseAlignContent,
		setFunc:   setFunc(func(v *View, val AlignContent) { v.AlignContent = val }),
	},
	"flex-grow": {
		parseFunc: parseFloat,
		setFunc:   setFunc(func(v *View, val float64) { v.Grow = val }),
	},
	"grow": {
		parseFunc: parseFloat,
		setFunc:   setFunc(func(v *View, val float64) { v.Grow = val }),
	},
	"flex-shrink": {
		parseFunc: parseFloat,
		setFunc:   setFunc(func(v *View, val float64) { v.Shrink = val }),
	},
	"shrink": {
		parseFunc: parseFloat,
		setFunc:   setFunc(func(v *View, val float64) { v.Shrink = val }),
	},
	"display": {
		parseFunc: parseDisplay,
		setFunc:   setFunc(func(v *View, val Display) { v.Display = val }),
	},
}

// setFunc creates a function that takes an entity and a value as an interface{}.
// The created function type asserts the value to the correct type U and then calls
// the given function f with the entity and the value of type U.
// If the type U is a pointer, the function will handle it accordingly:
// - If the input value is nil, it will pass a nil value of type U to the given function f.
// - If the input value is non-nil, it will create a pointer to the value, type-assert it to U, and pass it to f.
func setFunc[T, U any](f func(entity T, value U)) func(T, any) {
	return func(e T, v any) {
		argType := reflect.TypeOf(f).In(1)

		// If the value v is nil
		if v == nil {
			if argType.Kind() == reflect.Ptr {
				// pass nil value if the type is pointer
				nilValue := reflect.Zero(argType).Interface().(U)
				f(e, nilValue)
			} else {
				// pass deafult value if the type is not pointer
				var u U
				defaultValue := reflect.Zero(reflect.TypeOf(u))
				f(e, defaultValue.Interface().(U))
			}
			return
		}

		// If the second input parameter type of f is a pointer
		if argType.Kind() == reflect.Ptr {
			// Create a pointer to the value and set its value to v
			valuePtr := reflect.New(reflect.TypeOf(v)).Elem()
			valuePtr.Set(reflect.ValueOf(v))

			// If the pointer created can be type asserted to U, call f with e and the pointer
			if ptr, ok := valuePtr.Addr().Interface().(U); ok {
				f(e, ptr)
				return
			}
		}

		// If the value v is of type U, call f directly with e and v
		if v, ok := v.(U); ok {
			f(e, v)
			return
		}

		// If the type of the value is incorrect, panic with an error message
		var u U
		panic(fmt.Sprintf("type of the value is incorrect: %v, %T vs %T", v, v, u))
	}
}

type mapper[T any] struct {
	parseFunc func(string) (any, error)
	setFunc   func(*T, any)
}

func parseNumber(val string) (any, error) {
	val = strings.TrimSuffix(val, "px")
	return strconv.Atoi(val)
}

func parseFloat(val string) (any, error) {
	return strconv.ParseFloat(val, 64)
}

func parsePosition(val string) (any, error) {
	switch val {
	case "absolute":
		return PositionAbsolute, nil
	case "static", "relative":
		return PositionStatic, nil
	}
	return PositionStatic, fmt.Errorf("unknown position: %s", val)
}

func parseDirection(val string) (any, error) {
	switch val {
	case "row":
		return Row, nil
	case "column":
		return Column, nil
	}
	return Column, fmt.Errorf("unknown direction: %s", val)
}

func parseWrap(val string) (any, error) {
	switch val {
	case "wrap":
		return Wrap, nil
	case "nowrap":
		return NoWrap, nil
	}
	return NoWrap, fmt.Errorf("unknown wrap: %s", val)
}

func parseJustify(val string) (any, error) {
	switch val {
	case "flex-start", "start":
		return JustifyStart, nil
	case "flex-end", "end":
		return JustifyEnd, nil
	case "space-between":
		return JustifySpaceBetween, nil
	case "space-around":
		return JustifySpaceAround, nil
	case "center":
		return JustifyCenter, nil
	}
	return JustifyStart, fmt.Errorf("unknown justify: %s", val)
}

func parseAlignItem(val string) (any, error) {
	switch val {
	case "flex-start", "start":
		return AlignItemStart, nil
	case "flex-end", "end":
		return AlignItemEnd, nil
	case "center":
		return AlignItemCenter, nil
	case "stretch":
		return AlignItemStretch, nil
	}
	return AlignItemStretch, fmt.Errorf("unknown align-items: %s", val)
}

func parseAlignContent(val string) (any, error) {
	switch val {
	case "flex-start", "start":
		return AlignContentStart, nil
	case "flex-end", "end":
		return AlignContentEnd, nil
	case "center":
		return AlignContentCenter, nil
	case "stretch":
		return AlignContentStretch, nil
	case "space-between":
		return AlignContentSpaceBetween, nil
	case "space-around":
		return AlignContentSpaceAround, nil
	}
	return AlignContentStart, fmt.Errorf("unknown align-content: %s", val)
}

func parseDisplay(val string) (any, error) {
	switch val {
	case "none":
		return DisplayNone, nil
	case "", "flex":
		return DisplayFlex, nil
	}
	return DisplayFlex, fmt.Errorf("unknown display: %s", val)
}

type cssLength struct {
	unit cssUnit
	val  float64
}

func parseLength(val string) (any, error) {
	switch {
	case strings.HasSuffix(val, "%"):
		val = strings.TrimSuffix(val, "%")
		v, err := parseFloat(val)
		if err != nil || v.(float64) <= 0 {
			return cssLength{}, nil
		}
		return cssLength{unit: cssUnitPct, val: v.(float64)}, nil
	default:
		val = strings.TrimSuffix(val, "px")
		v, err := parseFloat(val)
		if err != nil {
			return cssLength{}, nil
		}
		return cssLength{unit: cssUnitPx, val: v.(float64)}, nil
	}
}

type attrs struct {
	id     string
	style  string
	hidden bool
	miscs  map[string]string
}

func readAttrs(z *html.Tokenizer) attrs {
	attr := attrs{
		miscs: make(map[string]string),
	}
	for {
		key, val, more := z.TagAttr()
		attr.miscs[string(key)] = string(val)
		switch string(key) {
		case "id":
			attr.id = string(val)
		case "style":
			attr.style = string(val)
		case "hidden":
			v := string(val)
			if v == "" {
				attr.hidden = true
			} else {
				attr.hidden = parseBool(v)
			}
		}
		if !more {
			break
		}
	}
	return attr
}

func parseBool(val string) bool {
	return val == "true"
}

type cssUnit int

const (
	cssUnitPx cssUnit = iota
	cssUnitPct
)
