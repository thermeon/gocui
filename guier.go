// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

// Gui represents the whole User Interface, including the views, layouts
// and keybindings.
type Guier interface {
	Close()
	Size() (x, y int)
	SetRune(x, y int, ch rune, fgColor, bgColor Attribute) error
	Rune(x, y int) (rune, error)
	SetView(name string, x0, y0, x1, y1 int) (Viewer, error)
	SetViewOnTop(name string) (Viewer, error)
	SetViewOnBottom(name string) (Viewer, error)
	Views() []Viewer
	View(name string) (Viewer, error)
	ViewByPosition(x, y int) (Viewer, error)
	ViewPosition(name string) (x0, y0, x1, y1 int, err error)
	DeleteView(name string) error
	SetCurrentView(name string) (Viewer, error)
	CurrentView() Viewer
	SetKeybinding(viewname string, key interface{}, mod Modifier, handler func(Guier, Viewer) error) error
	DeleteKeybinding(viewname string, key interface{}, mod Modifier) error
	DeleteKeybindings(viewname string)
	Update(f func(Guier) error)
	SetManager(managers ...Manager)
	SetManagerFunc(manager func(Guier) error)
	MainLoop() error
	GetBgFgColor() (BgColor, FgColor Attribute)
	SetBgFgColor(BgColor, FgColor Attribute)
	GetSelBgFgColor() (BgColor, FgColor Attribute)
	SetSelBgFgColor(BgColor, FgColor Attribute)
	GetHighlight() bool
	SetHighlight(h bool)
	GetCursor() bool
	SetCursor(c bool)
	GetMouseEventsEnabled() bool
	SetMouseEventsEnabled(e bool)
	GetInputEsc() bool
	SetInputEsc(e bool)
	GetASCII() bool
	SetASCII(a bool)
}
