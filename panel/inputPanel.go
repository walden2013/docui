package panel

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

var activeInput = 0

type Input struct {
	*Gui
	name string
	Position
	MaxLength int
	Items
	Data map[string]interface{}
}

type Items []Item

type Item struct {
	Label map[string]Position
	Input map[string]Position
}

func NewInput(gui *Gui, name string, x, y, w, h, maxLength int, items Items, data map[string]interface{}) Input {
	return Input{
		Gui:       gui,
		name:      name,
		Position:  Position{x, y, x + w, y + h},
		MaxLength: maxLength,
		Items:     items,
		Data:      data,
	}
}

func (i Input) Name() string {
	return i.name
}

func (i Input) SetView(g *gocui.Gui) (*gocui.View, error) {
	// create container panel
	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}

		v.Title = v.Name()
		v.Autoscroll = true
		v.Wrap = true
	}

	// create input panels
	for index, item := range i.Items {
		for name, p := range item.Label {
			if v, err := g.SetView(name, i.x*2+p.x, i.y+p.y, i.x*2+p.w, i.y+p.h); err != nil {
				if err != gocui.ErrUnknownView {
					return nil, err
				}
				v.Wrap = true
				v.Frame = false
				fmt.Fprint(v, name)
			}
		}

		for name, p := range item.Input {
			if v, err := g.SetView(name, i.x*2+p.x, i.y+p.y, i.x*2+p.w, i.y+p.h); err != nil {
				if err != gocui.ErrUnknownView {
					return nil, err
				}
				v.Wrap = true
				v.Editable = true
				v.Editor = i

				if index == 0 {
					SetCurrentPanel(g, name)
				}

				if name == "ImageInput" {
					fmt.Fprint(v, i.Data["Image"])
				}

				i.SetKeyBindWithItem(name)
			}
		}
	}

	i.SetKeybinds(i.Name())
	return v, nil
}

func (i Input) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > i.MaxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		v.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
}

func (i Input) Init(g *Gui) {
	_, err := i.SetView(g.Gui)

	if err != nil {
		panic(err)
	}
}

func (i Input) SetKeyBindWithItem(name string) {
	if err := i.SetKeybinding(name, gocui.KeyCtrlJ, gocui.ModNone, i.NextItem); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyCtrlK, gocui.ModNone, i.PreItem); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyCtrlC, gocui.ModNone, i.ClosePanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyCtrlQ, gocui.ModNone, i.quit); err != nil {
		panic(err)
	}
}

func (i Input) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	// パネルの位置とindexをリセットしないと、再度パネルを呼び出す時にinputの移動が変になる
	// 理由不明なので時間ある時調査
	for name, _ := range i.Items[0].Input {
		activeInput = 0
		SetCurrentPanel(g, name)
	}

	for _, item := range i.Items {
		for name, _ := range item.Label {
			i.DeleteView(name)
		}
		for name, _ := range item.Input {
			i.DeleteView(name)
			i.DeleteKeybindings(name)
		}
	}
	i.DeleteView(i.Name())
	SetCurrentPanel(g, ImageListPanel)

	return nil
}

func (i Input) NextItem(g *gocui.Gui, v *gocui.View) error {

	nextIndex := (activeInput + 1) % len(i.Items)
	item := i.Items[nextIndex]
	var name string
	for n, _ := range item.Input {
		name = n
	}

	if _, err := SetCurrentPanel(g, name); err != nil {
		return err
	}

	activeInput = nextIndex
	return nil
}

func (i Input) PreItem(g *gocui.Gui, v *gocui.View) error {
	nextIndex := activeInput - 1
	if nextIndex < 0 {
		nextIndex = len(i.Items) - 1
	} else {
		nextIndex = (activeInput - 1) % len(i.Items)
	}

	item := i.Items[nextIndex]

	var name string
	for n, _ := range item.Input {
		name = n
	}

	if _, err := SetCurrentPanel(g, name); err != nil {
		return err
	}

	activeInput = nextIndex
	return nil
}
