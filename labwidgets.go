package labwidgets

import (
	"math"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var (
	TitleFont = Font{
		Family:    "Tahoma",
		PointSize: 12,
		Bold:      true,
	}
	ContentFont = Font{
		Family:    "Tahoma",
		PointSize: 10,
		Bold:      false,
	}
	ButtonFont = Font{
		Family:    "Tahoma",
		PointSize: 10,
		Bold:      false,
	}
	LabelFont = Font{
		Family:    "Tahoma",
		PointSize: 10,
		Bold:      false,
	}

	defaultMargins = Margins{
		Top:    3,
		Left:   6,
		Right:  6,
		Bottom: 3,
	}
)

func StaticLabel(label string) Label {
	labelSize := fontToSize(LabelFont, len(label))
	return Label{
		Alignment: AlignHFarVCenter,
		Text:      label,
		Font:      LabelFont,
		MinSize:   labelSize,
		MaxSize:   labelSize,
	}
}

func _LineEdit(assignTo **walk.LineEdit, initialValue string, length int, readonly bool) LineEdit {
	editSize := fontToSize(ContentFont, length)
	return LineEdit{
		AssignTo: assignTo,
		Text:     initialValue,
		Font:     ContentFont,
		MinSize:  editSize,
		MaxSize:  editSize,
		Enabled:  readonly,
	}
}

func _NumberEdit(assignTo **walk.NumberEdit, digits, decimals int, unit string, readonly bool) NumberEdit {
	numberSize := fontToSize(ContentFont, digits+decimals+len(unit)+1+1)
	return NumberEdit{
		AssignTo: assignTo,
		Suffix:   " " + unit,
		Decimals: decimals,
		MinSize:  numberSize,
		MaxSize:  numberSize,
		Font:     ContentFont,
		Enabled:  readonly,
	}
}

func MomentaryButton(enabled chan bool, btnText string, action func()) PushButton {
	s := &struct {
		self *walk.PushButton
	}{}

	if enabled == nil {
		enabled = make(chan bool, 1)
	}

	go func() {
		for enabled := range enabled {
			s.self.Synchronize(func() {
				s.self.SetEnabled(enabled)
			})
		}
	}()

	btnSize := fontToSize(ButtonFont, len(btnText)+4)
	return PushButton{
		AssignTo: &s.self,
		Text:     btnText,
		Font:     ButtonFont,
		MaxSize:  btnSize,
		MinSize:  btnSize,

		OnClicked: func() {
			if action == nil {
				return
			}
			enabled <- false
			go func() {
				action()
				enabled <- true
			}()
		},
	}
}

func ToggleButton(enabled chan bool, btnText string, state *bool, action func(bool) bool) PushButton {
	s := &struct {
		self   *walk.PushButton
		state  *bool
		_state bool
	}{}
	if enabled == nil {
		enabled = make(chan bool, 1)
	}
	if state != nil {
		s.state = state
	} else {
		s.state = &s._state
	}
	go func() {
		for enabled := range enabled {
			s.self.Synchronize(func() {
				s.self.SetEnabled(enabled)
			})
		}
	}()

	btnSize := fontToSize(ButtonFont, len(btnText)+4)
	return PushButton{
		AssignTo: &s.self,
		Text:     btnText,
		Font:     ButtonFont,
		MaxSize:  btnSize,
		MinSize:  btnSize,

		OnClicked: func() {
			if action == nil {
				return
			}
			enabled <- false
			go func() {
				result := action(!*s.state)
				if result {
					s.self.Synchronize(func() {
						*s.state = !*s.state
					})

				}
				enabled <- true
			}()
		},
	}
}

// LabeledToggleButton returns a ToggleButton with a Label next to it indicating the current state of the controlled property.
// The button automatically disables itself unti the user action function returns.
// If the user function call returns the result of the execution, it should be passed back to the button
// so that it can update its internal state. If no such feedback mechanism exists, the user function should always return true for the toggle to happen.
func LabeledToggleButton(enable chan bool, btnText string, stateStrings map[bool]string, initialState bool, action func(bool) bool) Composite {
	s := &struct {
		label        *walk.Label
		State        bool
		StateString  string
		StateStrings map[bool]string
	}{
		State:        initialState,
		StateString:  stateStrings[initialState],
		StateStrings: stateStrings,
	}
	labelSize := fontToSize(LabelFont, int(math.Max(float64(len(stateStrings[true])), float64(len(stateStrings[false])))))
	return Composite{
		Layout: HBox{
			Margins:     Margins{},
			MarginsZero: true,
		},
		Children: []Widget{
			ToggleButton(enable, btnText, &s.State, func(b bool) bool {
				result := action(b)
				if result {
					s.label.Synchronize(func() {
						s.label.SetText(s.StateStrings[b])
					})
				}
				return result
			}),
			Label{
				AssignTo: &s.label,
				Text:     stateStrings[initialState],
				Font:     LabelFont,
				MaxSize:  labelSize,
				MinSize:  labelSize,
			},
			HSpacer{},
		},
	}
}

func StringSetter(enabled chan bool, btnText string, initialvalue string, length int, action func(string)) Composite {
	s := &struct {
		self   *walk.Composite
		edit   *walk.LineEdit
		button *walk.PushButton
	}{}
	go func() {
		for enabled := range enabled {
			s.self.Synchronize(func() {
				s.self.SetEnabled(enabled)
			})
		}
	}()

	return Composite{
		AssignTo: &s.self,
		Layout: HBox{
			Margins:     Margins{},
			MarginsZero: true,
		},
		Children: []Widget{
			_LineEdit(&s.edit, initialvalue, length, true),
			MomentaryButton(nil, btnText, func() {
				if action != nil {
					action(s.edit.Text())
				}
			}),
			HSpacer{},
		},
	}
}

func StringGetter(enabled chan bool, btnText string, length int, action func() string) Composite {
	s := &struct {
		self   *walk.Composite
		edit   *walk.LineEdit
		button *walk.PushButton
	}{}

	if enabled == nil {
		enabled = make(chan bool, 1)
	}
	go func() {
		for enabled := range enabled {
			s.self.Synchronize(func() {
				s.self.SetEnabled(enabled)
			})
		}
	}()

	return Composite{
		AssignTo: &s.self,
		Layout: HBox{
			Margins:     Margins{},
			MarginsZero: true,
		},
		Children: []Widget{
			MomentaryButton(nil, btnText, func() {
				if action == nil {
					return
				}
				str := action()
				s.self.Synchronize(func() {
					s.edit.SetText(str)
				})

			}),
			_LineEdit(&s.edit, "", length, false),
			HSpacer{},
		},
	}
}

func StringReadout(input <-chan string, label string, length int) Composite {
	s := &struct {
		self *walk.LineEdit
	}{}

	go func() {
		for v := range input {
			s.self.Synchronize(func() {
				s.self.SetText(v)
			})
		}
	}()

	return Composite{

		Layout: HBox{
			Margins:     defaultMargins,
			MarginsZero: true,
			Alignment:   AlignHNearVCenter,
		},

		Children: []Widget{
			StaticLabel(label),
			_LineEdit(&s.self, "", length, false),
			HSpacer{},
		},
	}
}

func FloatSetter(enabled chan bool, btnText string, unit string, initialvalue, min, max float64, digits, decimals int, action func(float64)) Composite {
	s := &struct {
		self   *walk.Composite
		db     *walk.DataBinder
		button *walk.PushButton
		edit   *walk.NumberEdit
		Value  float64
	}{
		Value: initialvalue,
	}

	go func() {
		for enabled := range enabled {
			s.self.Synchronize(func() {
				s.self.SetEnabled(enabled)
			})
		}
	}()

	buttonEnable := make(chan bool, 1)
	numEdit := _NumberEdit(&s.edit, digits, decimals, unit, true)
	numEdit.Value = Bind("Value", Range{Min: min, Max: max})
	numEdit.OnValueChanged = func() {
		s.db.Submit()
	}

	return Composite{
		DataBinder: DataBinder{
			AssignTo:       &s.db,
			Name:           "FloatInput",
			DataSource:     s,
			ErrorPresenter: ToolTipErrorPresenter{},
			OnCanSubmitChanged: func() {
				if s.db.CanSubmit() {
					buttonEnable <- true
				} else {
					buttonEnable <- false
				}
			},
		},
		AssignTo: &s.self,
		Layout: HBox{
			Margins:     defaultMargins,
			MarginsZero: true,
		},
		Children: []Widget{
			numEdit,
			MomentaryButton(buttonEnable, btnText, func() {
				if action == nil {
					return
				}
				action(s.Value)
			}),
			HSpacer{},
		},
	}
}

func FloatGetter(enabled chan bool, btnText string, unit string, digits, decimals int, action func() float64) Composite {
	s := &struct {
		self   *walk.Composite
		db     *walk.DataBinder
		button *walk.PushButton
		edit   *walk.NumberEdit
	}{}

	go func() {
		for enabled := range enabled {
			s.self.Synchronize(func() {
				s.self.SetEnabled(enabled)
			})
		}
	}()

	return Composite{
		DataBinder: DataBinder{
			AssignTo:       &s.db,
			Name:           "FloatInput",
			DataSource:     &s,
			ErrorPresenter: ToolTipErrorPresenter{},
			OnCanSubmitChanged: func() {
				if !s.db.CanSubmit() {
					enabled <- false
				}
			},
		},
		AssignTo: &s.self,
		Layout: HBox{
			Margins:     defaultMargins,
			MarginsZero: true,
		},
		Children: []Widget{
			MomentaryButton(nil, btnText, func() {
				if action == nil {
					return
				}
				val := action()
				s.self.Synchronize(func() {
					s.edit.SetValue(val)
				})
			}),
			_NumberEdit(&s.edit, digits, decimals, unit, false),
			HSpacer{},
		},
	}
}

func FloatReadout(input <-chan float64, label string, digits, decimals int, unit string) Composite {
	s := &struct {
		self *walk.NumberEdit
	}{}
	go func() {
		for v := range input {
			s.self.Synchronize(func() {
				s.self.SetValue(v)
			})
		}
	}()

	return Composite{

		Layout: HBox{
			Margins:     defaultMargins,
			MarginsZero: true,
		},
		Children: []Widget{
			StaticLabel(label),
			_NumberEdit(&s.self, digits, decimals, unit, false),
			HSpacer{},
		},
	}

}

func fontToSize(font Font, length int) Size {
	if length == 0 {
		return Size{}
	}
	width := int(float64(font.PointSize*length) * 0.8)
	height := font.PointSize * 3
	return Size{Width: width, Height: height}
}

type ControlDisablers struct {
	chns map[string]chan bool
}

// Create a collection of control signals for enabling/disabling a group of widgets
func NewDisablers(btn ...string) *ControlDisablers {
	c := ControlDisablers{
		chns: make(map[string]chan bool, len(btn)),
	}

	for name, _ := range c.chns {
		c.chns[name] = make(chan bool, 1)
	}
	return &c
}

func (c *ControlDisablers) EnableAll() {
	for _, chn := range c.chns {
		chn <- true
	}
}

func (c *ControlDisablers) DisableAll() {
	for _, chn := range c.chns {
		chn <- false
	}
}

func (c *ControlDisablers) Enable(name string) {
	if chn, ok := c.chns[name]; ok {
		chn <- true
	}
}

func (c *ControlDisablers) Disable(name string) {
	if chn, ok := c.chns[name]; ok {
		chn <- false
	}
}
