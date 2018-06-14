// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"errors"

	"github.com/thermeon/termbox-go"
)

var (
	// ErrQuit is used to decide if the MainLoop finished successfully.
	ErrQuit = errors.New("quit")

	// ErrUnknownView allows to assert if a View must be initialized.
	ErrUnknownView = errors.New("unknown view")
)

// OutputMode represents the terminal's output mode (8 or 256 colors).
type OutputMode termbox.OutputMode

const (
	// OutputNormal provides 8-colors terminal mode.
	OutputNormal = OutputMode(termbox.OutputNormal)

	// Output256 provides 256-colors terminal mode.
	Output256 = OutputMode(termbox.Output256)
)

// Gui represents the whole User Interface, including the views, layouts
// and keybindings.
type Gui struct {
	tbEvents    chan termbox.Event
	userEvents  chan userEvent
	views       []Viewer
	currentView Viewer
	managers    []Manager
	keybindings []*keybinding
	maxX, maxY  int
	outputMode  OutputMode

	// BgColor and FgColor allow to configure the background and foreground
	// colors of the GUI.
	BgColor, FgColor Attribute

	// SelBgColor and SelFgColor allow to configure the background and
	// foreground colors of the frame of the current view.
	SelBgColor, SelFgColor Attribute

	// If Highlight is true, Sel{Bg,Fg}Colors will be used to draw the
	// frame of the current view.
	Highlight bool

	// If Cursor is true then the cursor is enabled.
	Cursor bool

	// If Mouse is true then mouse events will be enabled.
	Mouse bool

	// If InputEsc is true, when ESC sequence is in the buffer and it doesn't
	// match any known sequence, ESC means KeyEsc.
	InputEsc bool

	// If ASCII is true then use ASCII instead of unicode to draw the
	// interface. Using ASCII is more portable.
	ASCII bool
}

func (g *Gui) GetBgFgColor() (BgColor, FgColor Attribute) {
	return g.BgColor, g.FgColor
}

func (g *Gui) SetBgFgColor(BgColor, FgColor Attribute) {
	g.FgColor = FgColor
	g.BgColor = BgColor
}

func (g *Gui) GetSelBgFgColor() (BgColor, FgColor Attribute) {
	return g.SelBgColor, g.SelFgColor
}

func (g *Gui) SetSelBgFgColor(BgColor, FgColor Attribute) {
	g.SelFgColor = FgColor
	g.SelBgColor = BgColor
}

func (g *Gui) GetHighlight() bool {
	return g.Highlight
}

func (g *Gui) SetHighlight(h bool) {
	g.Highlight = h
}

func (g *Gui) GetCursor() bool {
	return g.Cursor
}

func (g *Gui) SetCursor(c bool) {
	g.Cursor = c
}

func (g *Gui) GetMouseEventsEnabled() bool {
	return g.Mouse
}

func (g *Gui) SetMouseEventsEnabled(e bool) {
	g.Mouse = e
}

func (g *Gui) GetInputEsc() bool {
	return g.InputEsc
}

func (g *Gui) SetInputEsc(e bool) {
	g.InputEsc = e
}

func (g *Gui) GetASCII() bool {
	return g.ASCII
}

func (g *Gui) SetASCII(a bool) {
	g.ASCII = a
}

// NewGui returns a new Gui object with a given output mode.
func NewGui(mode OutputMode) (Guier, error) {
	if err := termbox.Init(); err != nil {
		return nil, err
	}

	g := &Gui{}

	g.outputMode = mode
	termbox.SetOutputMode(termbox.OutputMode(mode))

	g.tbEvents = make(chan termbox.Event, 20)
	g.userEvents = make(chan userEvent, 20)

	g.maxX, g.maxY = termbox.Size()

	g.BgColor, g.FgColor = ColorDefault, ColorDefault
	g.SelBgColor, g.SelFgColor = ColorDefault, ColorDefault

	return g, nil
}

// Close finalizes the library. It should be called after a successful
// initialization and when gocui is not needed anymore.
func (g *Gui) Close() {
	termbox.Close()
}

// Size returns the terminal's size.
func (g *Gui) Size() (x, y int) {
	return g.maxX, g.maxY
}

// SetRune writes a rune at the given point, relative to the top-left
// corner of the terminal. It checks if the position is valid and applies
// the given colors.
func (g *Gui) SetRune(x, y int, ch rune, fgColor, bgColor Attribute) error {
	if x < 0 || y < 0 || x >= g.maxX || y >= g.maxY {
		return errors.New("invalid point")
	}
	termbox.SetCell(x, y, ch, termbox.Attribute(fgColor), termbox.Attribute(bgColor))
	return nil
}

// Rune returns the rune contained in the cell at the given position.
// It checks if the position is valid.
func (g *Gui) Rune(x, y int) (rune, error) {
	if x < 0 || y < 0 || x >= g.maxX || y >= g.maxY {
		return ' ', errors.New("invalid point")
	}
	c := termbox.CellBuffer()[y*g.maxX+x]
	return c.Ch, nil
}

// SetView creates a new view with its top-left corner at (x0, y0)
// and the bottom-right one at (x1, y1). If a view with the same name
// already exists, its dimensions are updated; otherwise, the error
// ErrUnknownView is returned, which allows to assert if the View must
// be initialized. It checks if the position is valid.
func (g *Gui) SetView(name string, x0, y0, x1, y1 int) (Viewer, error) {
	if x0 >= x1 || y0 >= y1 {
		return nil, errors.New("invalid dimensions")
	}
	if name == "" {
		return nil, errors.New("invalid name")
	}

	if v, err := g.View(name); err == nil {
		v.SetBounds(x0, y0, x1, y1)
		v.Invalidate()
		return v, nil
	}

	v := newView(name, x0, y0, x1, y1, g.outputMode)
	v.SetBgFgColor(g.BgColor, g.FgColor)
	v.SetSelBgFgColor(g.SelBgColor, g.SelFgColor)
	g.views = append(g.views, v)
	return v, ErrUnknownView
}

// SetViewOnTop sets the given view on top of the existing ones.
func (g *Gui) SetViewOnTop(name string) (Viewer, error) {
	for i, v := range g.views {
		if v.Name() == name {
			s := append(g.views[:i], g.views[i+1:]...)
			g.views = append(s, v)
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// SetViewOnBottom sets the given view on bottom of the existing ones.
func (g *Gui) SetViewOnBottom(name string) (Viewer, error) {
	for i, v := range g.views {
		if v.Name() == name {
			s := append(g.views[:i], g.views[i+1:]...)
			g.views = append([]Viewer{v}, s...)
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// Views returns all the views in the GUI.
func (g *Gui) Views() []Viewer {
	return g.views
}

// View returns a pointer to the view with the given name, or error
// ErrUnknownView if a view with that name does not exist.
func (g *Gui) View(name string) (Viewer, error) {
	for _, v := range g.views {
		if v.Name() == name {
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// ViewByPosition returns a pointer to a view matching the given position, or
// error ErrUnknownView if a view in that position does not exist.
func (g *Gui) ViewByPosition(x, y int) (Viewer, error) {
	// traverse views in reverse order checking top views first
	for i := len(g.views); i > 0; i-- {
		v := g.views[i-1]
		x0, y0, x1, y1 := v.GetBounds()
		if x > x0 && x < x1 && y > y0 && y < y1 {
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// ViewPosition returns the coordinates of the view with the given name, or
// error ErrUnknownView if a view with that name does not exist.
func (g *Gui) ViewPosition(name string) (x0, y0, x1, y1 int, err error) {
	for _, v := range g.views {
		if v.Name() == name {
			x0, y0, x1, y1 := v.GetBounds()

			return x0, y0, x1, y1, nil
		}
	}
	return 0, 0, 0, 0, ErrUnknownView
}

// DeleteView deletes a view by name.
func (g *Gui) DeleteView(name string) error {
	for i, v := range g.views {
		if v.Name() == name {
			g.views = append(g.views[:i], g.views[i+1:]...)
			return nil
		}
	}
	return ErrUnknownView
}

// SetCurrentView gives the focus to a given view.
func (g *Gui) SetCurrentView(name string) (Viewer, error) {
	for _, v := range g.views {
		if v.Name() == name {
			g.currentView = v
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// CurrentView returns the currently focused view, or nil if no view
// owns the focus.
func (g *Gui) CurrentView() Viewer {
	return g.currentView
}

// SetKeybinding creates a new keybinding. If viewname equals to ""
// (empty string) then the keybinding will apply to all views. key must
// be a rune or a Key.
func (g *Gui) SetKeybinding(viewname string, key interface{}, mod Modifier, handler func(Guier, Viewer) error) error {
	var kb *keybinding

	k, ch, err := getKey(key)
	if err != nil {
		return err
	}
	kb = newKeybinding(viewname, k, ch, mod, handler)
	g.keybindings = append(g.keybindings, kb)
	return nil
}

// DeleteKeybinding deletes a keybinding.
func (g *Gui) DeleteKeybinding(viewname string, key interface{}, mod Modifier) error {
	k, ch, err := getKey(key)
	if err != nil {
		return err
	}

	for i, kb := range g.keybindings {
		if kb.viewName == viewname && kb.ch == ch && kb.key == k && kb.mod == mod {
			g.keybindings = append(g.keybindings[:i], g.keybindings[i+1:]...)
			return nil
		}
	}
	return errors.New("keybinding not found")
}

// DeleteKeybindings deletes all keybindings of view.
func (g *Gui) DeleteKeybindings(viewname string) {
	var s []*keybinding
	for _, kb := range g.keybindings {
		if kb.viewName != viewname {
			s = append(s, kb)
		}
	}
	g.keybindings = s
}

// getKey takes an empty interface with a key and returns the corresponding
// typed Key or rune.
func getKey(key interface{}) (Key, rune, error) {
	switch t := key.(type) {
	case Key:
		return t, 0, nil
	case rune:
		return 0, t, nil
	default:
		return 0, 0, errors.New("unknown type")
	}
}

// userEvent represents an event triggered by the user.
type userEvent struct {
	f func(Guier) error
}

// Update executes the passed function. This method can be called safely from a
// goroutine in order to update the GUI. It is important to note that the
// passed function won't be executed immediately, instead it will be added to
// the user events queue. Given that Update spawns a goroutine, the order in
// which the user events will be handled is not guaranteed.
func (g *Gui) Update(f func(Guier) error) {
	go func() { g.userEvents <- userEvent{f: f} }()
}

// A Manager is in charge of GUI's layout and can be used to build widgets.
type Manager interface {
	// Layout is called every time the GUI is redrawn, it must contain the
	// base views and its initializations.
	Layout(Guier) error
}

// The ManagerFunc type is an adapter to allow the use of ordinary functions as
// Managers. If f is a function with the appropriate signature, ManagerFunc(f)
// is an Manager object that calls f.
type ManagerFunc func(Guier) error

// Layout calls f(g)
func (f ManagerFunc) Layout(g Guier) error {
	return f(g)
}

// SetManager sets the given GUI managers. It deletes all views and
// keybindings.
func (g *Gui) SetManager(managers ...Manager) {
	g.managers = managers
	g.currentView = nil
	g.views = nil
	g.keybindings = nil

	go func() { g.tbEvents <- termbox.Event{Type: termbox.EventResize} }()
}

// SetManagerFunc sets the given manager function. It deletes all views and
// keybindings.
func (g *Gui) SetManagerFunc(manager func(Guier) error) {
	g.SetManager(ManagerFunc(manager))
}

// MainLoop runs the main loop until an error is returned. A successful
// finish should return ErrQuit.
func (g *Gui) MainLoop() error {
	go func() {
		for {
			g.tbEvents <- termbox.PollEvent()
		}
	}()

	inputMode := termbox.InputAlt
	if g.InputEsc {
		inputMode = termbox.InputEsc
	}
	if g.Mouse {
		inputMode |= termbox.InputMouse
	}
	termbox.SetInputMode(inputMode)

	if err := g.flush(); err != nil {
		return err
	}
	for {
		select {
		case ev := <-g.tbEvents:
			if err := g.handleEvent(&ev); err != nil {
				return err
			}
		case ev := <-g.userEvents:
			if err := ev.f(g); err != nil {
				return err
			}
		}
		if err := g.consumeevents(); err != nil {
			return err
		}
		if err := g.flush(); err != nil {
			return err
		}
	}
}

// consumeevents handles the remaining events in the events pool.
func (g *Gui) consumeevents() error {
	for {
		select {
		case ev := <-g.tbEvents:
			if err := g.handleEvent(&ev); err != nil {
				return err
			}
		case ev := <-g.userEvents:
			if err := ev.f(g); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

// handleEvent handles an event, based on its type (key-press, error,
// etc.)
func (g *Gui) handleEvent(ev *termbox.Event) error {
	switch ev.Type {
	case termbox.EventKey, termbox.EventMouse:
		return g.onKey(ev)
	case termbox.EventError:
		return ev.Err
	default:
		return nil
	}
}

// flush updates the gui, re-drawing frames and buffers.
func (g *Gui) flush() error {
	termbox.Clear(termbox.Attribute(g.FgColor), termbox.Attribute(g.BgColor))

	maxX, maxY := termbox.Size()
	// if GUI's size has changed, we need to redraw all views
	if maxX != g.maxX || maxY != g.maxY {
		for _, v := range g.views {
			v.Invalidate()
		}
	}
	g.maxX, g.maxY = maxX, maxY

	for _, m := range g.managers {
		if err := m.Layout(g); err != nil {
			return err
		}
	}
	for _, v := range g.views {
		if v.HasFrame() {
			var fgColor, bgColor Attribute
			if g.Highlight && v == g.currentView {
				fgColor = g.SelFgColor
				bgColor = g.SelBgColor
			} else {
				fgColor = g.FgColor
				bgColor = g.BgColor
			}

			if err := g.drawFrameEdges(v, fgColor, bgColor); err != nil {
				return err
			}
			if err := g.drawFrameCorners(v, fgColor, bgColor); err != nil {
				return err
			}
			if v.GetTitle() != "" {
				if err := g.drawTitle(v, fgColor, bgColor); err != nil {
					return err
				}
			}
		}
		if err := g.draw(v); err != nil {
			return err
		}
	}
	termbox.Flush()
	return nil
}

// drawFrameEdges draws the horizontal and vertical edges of a view.
func (g *Gui) drawFrameEdges(v Viewer, fgColor, bgColor Attribute) error {
	runeH, runeV := '─', '│'
	if g.ASCII {
		runeH, runeV = '-', '|'
	}

	x0, y0, x1, y1 := v.GetBounds()

	for x := x0 + 1; x < x1 && x < g.maxX; x++ {
		if x < 0 {
			continue
		}
		if y0 > -1 && y0 < g.maxY {
			if err := g.SetRune(x, y0, runeH, fgColor, bgColor); err != nil {
				return err
			}
		}
		if y1 > -1 && y1 < g.maxY {
			if err := g.SetRune(x, y1, runeH, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	for y := y0 + 1; y < y1 && y < g.maxY; y++ {
		if y < 0 {
			continue
		}
		if x0 > -1 && x0 < g.maxX {
			if err := g.SetRune(x0, y, runeV, fgColor, bgColor); err != nil {
				return err
			}
		}
		if x1 > -1 && x1 < g.maxX {
			if err := g.SetRune(x1, y, runeV, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	return nil
}

// drawFrameCorners draws the corners of the view.
func (g *Gui) drawFrameCorners(v Viewer, fgColor, bgColor Attribute) error {
	runeTL, runeTR, runeBL, runeBR := '┌', '┐', '└', '┘'
	if g.ASCII {
		runeTL, runeTR, runeBL, runeBR = '+', '+', '+', '+'
	}

	x0, y0, x1, y1 := v.GetBounds()

	corners := []struct {
		x, y int
		ch   rune
	}{{x0, y0, runeTL}, {x1, y0, runeTR}, {x0, y1, runeBL}, {x1, y1, runeBR}}

	for _, c := range corners {
		if c.x >= 0 && c.y >= 0 && c.x < g.maxX && c.y < g.maxY {
			if err := g.SetRune(c.x, c.y, c.ch, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	return nil
}

// drawTitle draws the title of the view.
func (g *Gui) drawTitle(v Viewer, fgColor, bgColor Attribute) error {

	x0, y0, x1, _ := v.GetBounds()

	if y0 < 0 || y0 >= g.maxY {
		return nil
	}

	for i, ch := range v.GetTitle() {
		x := x0 + i + 2
		if x < 0 {
			continue
		} else if x > x1-2 || x >= g.maxX {
			break
		}
		if err := g.SetRune(x, y0, ch, fgColor, bgColor); err != nil {
			return err
		}
	}
	return nil
}

// draw manages the cursor and calls the draw function of a view.
func (g *Gui) draw(v Viewer) error {
	if g.Cursor {
		if curview := g.currentView; curview != nil {
			vMaxX, vMaxY := curview.Size()
			cx, cy := curview.Cursor()
			x0, y0, _, _ := curview.GetBounds()
			if cx < 0 {
				cx = 0
			} else if cx >= vMaxX {
				cx = vMaxX - 1
			}
			if cy < 0 {
				cy = 0
			} else if cy >= vMaxY {
				cy = vMaxY - 1
			}

			curview.SetCursor(cx, cy)

			gMaxX, gMaxY := g.Size()
			cx, cy = x0+cx+1, y0+cy+1
			if cx >= 0 && cx < gMaxX && cy >= 0 && cy < gMaxY {
				termbox.SetCursor(cx, cy)
			} else {
				termbox.HideCursor()
			}
		}
	} else {
		termbox.HideCursor()
	}

	v.ClearRunes()
	if err := v.Draw(); err != nil {
		return err
	}
	return nil
}

// onKey manages key-press events. A keybinding handler is called when
// a key-press or mouse event satisfies a configured keybinding. Furthermore,
// currentView's internal buffer is modified if currentView.Editable is true.
func (g *Gui) onKey(ev *termbox.Event) error {
	switch ev.Type {
	case termbox.EventKey:
		matched, err := g.execKeybindings(g.currentView, ev)
		if err != nil {
			return err
		}
		if matched {
			break
		}
		if g.currentView != nil && g.currentView.IsEditable() && g.currentView.GetEditor() != nil {
			g.currentView.GetEditor().Edit(g.currentView, Key(ev.Key), ev.Ch, Modifier(ev.Mod))
		}
	case termbox.EventMouse:
		mx, my := ev.MouseX, ev.MouseY
		v, err := g.ViewByPosition(mx, my)
		if err != nil {
			break
		}
		x0, _, y0, _ := v.GetBounds()
		if err := v.SetCursor(mx-x0-1, my-y0-1); err != nil {
			return err
		}
		if _, err := g.execKeybindings(v, ev); err != nil {
			return err
		}
	}

	return nil
}

// execKeybindings executes the keybinding handlers that match the passed view
// and event. The value of matched is true if there is a match and no errors.
func (g *Gui) execKeybindings(v Viewer, ev *termbox.Event) (matched bool, err error) {
	matched = false
	for _, kb := range g.keybindings {
		if kb.handler == nil {
			continue
		}
		if kb.matchKeypress(Key(ev.Key), ev.Ch, Modifier(ev.Mod)) && kb.matchView(v) {
			if err := kb.handler(g, v); err != nil {
				return false, err
			}
			matched = true
		}
	}
	return matched, nil
}
