package controllers

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type MainViewController struct {
	baseController
	c *ControllerCommon

	context      *context.MainContext
	otherContext *context.MainContext
}

var _ types.IController = &MainViewController{}

func NewMainViewController(
	c *ControllerCommon,
	context *context.MainContext,
	otherContext *context.MainContext,
) *MainViewController {
	return &MainViewController{
		baseController: baseController{},
		c:              c,
		context:        context,
		otherContext:   otherContext,
	}
}

func (self *MainViewController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Key:             opts.GetKey(opts.Config.Universal.TogglePanel),
			Handler:         self.togglePanel,
			Description:     self.c.Tr.ToggleStagingView,
			Tooltip:         self.c.Tr.ToggleStagingViewTooltip,
			DisplayOnScreen: true,
		},
		{
			Key:         opts.GetKey(opts.Config.Universal.Return),
			Handler:     self.escape,
			Description: self.c.Tr.ExitFocusedMainView,
		},
		{
			// overriding this because we want to read all of the task's output before we start searching
			Key:         opts.GetKey(opts.Config.Universal.StartSearch),
			Handler:     self.openSearch,
			Description: self.c.Tr.StartSearch,
			Tag:         "navigation",
		},
	}
}

func (self *MainViewController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: self.context.GetViewName(),
			Key:      gocui.MouseLeft,
			Handler: func(opts gocui.ViewMouseBindingOpts) error {
				if self.isFocused() {
					return self.onClick(opts)
				}

				self.context.SetParentContext(self.otherContext.GetParentContext())
				self.c.Context().Push(self.context, types.OnFocusOpts{
					ClickedWindowName:  self.context.GetWindowName(),
					ClickedViewLineIdx: opts.Y,
				})

				return nil
			},
		},
	}
}

func (self *MainViewController) Context() types.Context {
	return self.context
}

func (self *MainViewController) togglePanel() error {
	if self.otherContext.GetView().Visible {
		self.otherContext.SetParentContext(self.context.GetParentContext())
		self.c.Context().Push(self.otherContext, types.OnFocusOpts{})
	}

	return nil
}

func (self *MainViewController) escape() error {
	self.c.Context().Pop()
	return nil
}

func (self *MainViewController) onClick(opts gocui.ViewMouseBindingOpts) error {
	parentCtx := self.context.GetParentContext()
	if parentCtx.GetOnClickFocusedMainView() != nil {
		return parentCtx.GetOnClickFocusedMainView()(self.context.GetViewName(), opts.Y)
	}
	return nil
}

func (self *MainViewController) openSearch() error {
	if manager := self.c.GetViewBufferManagerForView(self.context.GetView()); manager != nil {
		manager.ReadToEnd(func() {
			self.c.OnUIThread(func() error {
				return self.c.Helpers().Search.OpenSearchPrompt(self.context)
			})
		})
	}

	return nil
}

func (self *MainViewController) isFocused() bool {
	return self.c.Context().Current().GetKey() == self.context.GetKey()
}
