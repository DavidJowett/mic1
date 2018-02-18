package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type KeyBinding struct {
	View    string
	Key     interface{}
	Mod     gocui.Modifier
	Handler func(*gocui.Gui, *gocui.View) error
}

type Ui struct {
	Mic     *mic1
	Gui     *gocui.Gui
	MemAddr int
	MemHex  bool
	SymPos  int
	SymHex  bool
	MCPos   int
	VCycle  []*gocui.View
	CView   int
	MC      []string
}

func (u *Ui) Run() error {
	defer u.Gui.Close()
	if err := u.Gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func initGui(m *mic1) (*Ui, error) {
	var err error
	u := &Ui{Mic: m}
	u.MemAddr = 0x0000
	u.MemHex = true
	u.SymPos = 0
	u.SymHex = false
	u.MCPos = 0
	u.VCycle = make([]*gocui.View, 0, 4)
	u.CView = 1
	u.MC = make([]string, 256, 256)
	u.Gui, err = gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		return nil, err
	}

        /* Translate all the binary microcode instructions to a human readable format */
	for i, v := range u.Mic.MC {
		ins := Unpack(v)
		u.MC[i] = ins.ToString()
	}
	u.Gui.SetManagerFunc(u.Layout)

        /* Keybindings */
	var keys []KeyBinding = []KeyBinding{
		KeyBinding{"", gocui.KeyCtrlC, gocui.ModNone, quit},
		KeyBinding{"", 'q', gocui.ModNone, quit},
		KeyBinding{"", 's', gocui.ModNone, u.MicStep},
		KeyBinding{"", 'r', gocui.ModNone, u.MicRun},
		KeyBinding{"", 'h', gocui.ModNone, u.MicHalt},
		KeyBinding{"", 'c', gocui.ModNone, u.CycleView},
		KeyBinding{"symbols", 'j', gocui.ModNone, u.SymScrollDown},
		KeyBinding{"symbols", 'k', gocui.ModNone, u.SymScrollUp},
		KeyBinding{"symbols", 'g', gocui.ModNone, u.SymGoto},
		KeyBinding{"symbols", gocui.KeyEnter, gocui.ModNone, u.SymGoto},
		KeyBinding{"symbols", 'm', gocui.ModNone, u.SymModeToggle},
		KeyBinding{"memory", 'j', gocui.ModNone, u.MemScrollDown},
		KeyBinding{"memory", 'k', gocui.ModNone, u.MemScrollUp},
		KeyBinding{"memory", 'm', gocui.ModNone, u.MemModeToggle},
		KeyBinding{"microcode", 'j', gocui.ModNone, u.MicrocodeScrollDown},
		KeyBinding{"microcode", 'k', gocui.ModNone, u.MicrocodeScrollUp},
		KeyBinding{"microcode", 'b', gocui.ModNone, u.MicrocodeToggleBreakPoint},
	}

        /* Setup keybindngs */
	for _, k := range keys {
		err = u.Gui.SetKeybinding(k.View, k.Key, k.Mod, k.Handler)
		if err != nil {
			return nil, err
		}
	}

	u.Gui.Update(u.UpdateViews)

	return u, nil
}

func (u *Ui) UpdateViews(g *gocui.Gui) error {
	/* Registers View */
	err := u.UpdateRegistersView(g)
	if err != nil {
		return err
	}
	/* Symbols View */
	err = u.UpdateSymbolsView(g)
	if err != nil {
		return err
	}
	/* Microcode View */
	err = u.UpdateMicrocodeView(g)
	if err != nil {
		return err
	}
	/* Memory View */
	err = u.UpdateMemoryView(g)
	if err != nil {
		return err
	}
	return nil
}

func (u *Ui) UpdateRegistersView(g *gocui.Gui) error {
	u.Mic.RegistersLock.Lock()
	defer u.Mic.RegistersLock.Unlock()
	v, err := g.View("registers")
	if err != nil {
		return err
	}
	v.Clear()
	for i, r := range u.Mic.Registers {
		fmt.Fprintf(v, "%-7s: %#04x %-5d %016b\n", RegIdToNames[i], r, r, r)
	}
	fmt.Fprintf(v, "MAR    : %#04x %-5d %016b\n", u.Mic.MAR, u.Mic.MAR, u.Mic.MAR)
	fmt.Fprintf(v, "MBR    : %#04x %-5d %016b\n", u.Mic.MBR, u.Mic.MBR, u.Mic.MBR)
	if u.Mic.State == RUN {
		fmt.Fprintf(v, "Status : Running\n")
	} else {
		fmt.Fprintf(v, "Status : Halted\n")
	}
	fmt.Fprintf(v, "MPC    : %d\n", u.Mic.MPC)
	fmt.Fprintf(v, "Cycles : %d", u.Mic.Cycles)

	return nil
}

func (u *Ui) UpdateSymbolsView(g *gocui.Gui) error {
	v, err := g.View("symbols")
	if err != nil {
		return err
	}
	v.Clear()
	_, maxY := v.Size()
	if u.SymHex {
		for i := 0; i < maxY && (i+u.SymPos) < len(u.Mic.MemSymbols); i++ {
			fmt.Fprintf(v, "%-24s : %#04x\n", u.Mic.MemSymbols[i+u.SymPos].Name, u.Mic.MemSymbols[i+u.SymPos].Val)
		}
	} else {
		for i := 0; i < maxY && (i+u.SymPos) < len(u.Mic.MemSymbols); i++ {
			fmt.Fprintf(v, "%-24s : %-6d\n", u.Mic.MemSymbols[i+u.SymPos].Name, u.Mic.MemSymbols[i+u.SymPos].Val)
		}
	}

	return nil
}

func (u *Ui) UpdateMicrocodeView(g *gocui.Gui) error {
	v, err := g.View("microcode")
	if err != nil {
		return err
	}
	u.Mic.RegistersLock.Lock()
	mpc := u.Mic.MPC
	u.Mic.RegistersLock.Unlock()
	v.Clear()
	_, maxY := v.Size()
        var br rune
        var cur rune
	for i := 0; i < maxY && (i+u.MCPos) < 256; i++ {
                if i+u.MCPos == int(mpc) {
                        cur = '>'
                } else {
                        cur = ' '
                }
                if u.Mic.MCC[i+u.MCPos].BR {
                        br = '*'
                } else {
                        br = ' '
                }
		fmt.Fprintf(v, "%c%c%3d: %s\n", cur, br, i+u.MCPos, u.MC[i+u.MCPos])
	}
	return nil
}

func (u *Ui) UpdateMemoryView(g *gocui.Gui) error {
	u.Mic.RegistersLock.Lock()
	defer u.Mic.RegistersLock.Unlock()
	v, err := g.View("memory")
	if err != nil {
		return err
	}
	v.Clear()
	_, maxY := v.Size()
	if u.MemHex {
		for i := 0; i < maxY && (i*8+int(u.MemAddr)) < 4096; i++ {
			fmt.Fprintf(v, "%#04x: ", int(u.MemAddr)+(i*8))
			for j := 0; j < 8; j++ {
				fmt.Fprintf(v, "%#04x ", u.Mic.Memory[int(u.MemAddr)+(i*8)+j])
			}
			fmt.Fprint(v, "\n")
		}
	} else {
		for i := 0; i < maxY && (i*8+int(u.MemAddr)) < 4096; i++ {
			fmt.Fprintf(v, "%6d: ", int(u.MemAddr)+(i*8))
			for j := 0; j < 8; j++ {
				fmt.Fprintf(v, "%6d ", u.Mic.Memory[int(u.MemAddr)+(i*8)+j])
			}
			fmt.Fprint(v, "\n")
		}

	}

	return nil
}

func (u *Ui) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	maxX--
	maxY--
	col1x := (maxX - 4) * 5 / 12
	if col1x > 44 {
		col1x = 44
	}
	cell1y := (maxY - 4) * 7 / 8
	if cell1y > 22 {
		cell1y = 22
	}
	if v, err := g.SetView("registers", 0, 0, col1x, cell1y); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = "registers"
	}
	if v, err := g.SetView("symbols", 0, cell1y+1, col1x, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Highlight = true
		v.Title = "symbols"
		v.SetCursor(0, 0)
		DefocusView(g, v)

		u.VCycle = append(u.VCycle, v)
	}
	if v, err := g.SetView("microcode", col1x+1, 0, maxX, (maxY-4)/2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Highlight = true
		v.Title = "microcode"

		v.SetCursor(0, 0)
		FocusView(g, v)

		u.VCycle = append(u.VCycle, v)
	}
	if v, err := g.SetView("memory", col1x+1, (maxY-4)/2+1, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Highlight = true
		v.Title = "memory"

		v.SetCursor(0, 0)
		DefocusView(g, v)

		u.VCycle = append(u.VCycle, v)
	}
	u.UpdateViews(g)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (u *Ui) MicStep(g *gocui.Gui, v *gocui.View) error {
	u.Mic.Step()
	g.Update(u.UpdateViews)
	return nil
}

func (u *Ui) MicUpdate(g *gocui.Gui, v *gocui.View) error {
	g.Update(u.UpdateViews)
	return nil
}

func (u *Ui) MicRun(g *gocui.Gui, v *gocui.View) error {
	u.Mic.DesiredState = RUN
	u.Gui.Update(u.UpdateViews)
	go u.Mic.Run()
	go u.MicWatcher()
	return nil
}

func (u *Ui) MicHalt(g *gocui.Gui, v *gocui.View) error {
	u.Mic.DesiredState = HALT
	return nil
}

func (u *Ui) MicWatcher() {
	for {
		u.Mic.StateChange.L.Lock()
		u.Mic.StateChange.Wait()
		u.Gui.Update(u.UpdateViews)
		u.Mic.StateChange.L.Unlock()
	}
}

func (u *Ui) CycleView(g *gocui.Gui, v *gocui.View) error {
	DefocusView(g, u.VCycle[u.CView])
	u.CView = (u.CView + 1) % len(u.VCycle)
	FocusView(g, u.VCycle[u.CView])

	return nil
}

func (u *Ui) SymScrollDown(g *gocui.Gui, v *gocui.View) error {
	u.SymPos += 1
	if u.SymPos >= len(u.Mic.MemSymbols) {
		u.SymPos = len(u.Mic.MemSymbols) - 1
	}
	u.Gui.Update(u.UpdateSymbolsView)
	return nil
}

func (u *Ui) SymScrollUp(g *gocui.Gui, v *gocui.View) error {
	u.SymPos -= 1
	if u.SymPos < 0 {
		u.SymPos = 0
	}
	u.Gui.Update(u.UpdateSymbolsView)
	return nil
}

func (u *Ui) MemScrollDown(g *gocui.Gui, v *gocui.View) error {
	u.MemAddr += 8
	if u.MemAddr >= 4096 {
		u.MemAddr = 4088
	}
	u.Gui.Update(u.UpdateMemoryView)
	return nil
}

func (u *Ui) MemScrollUp(g *gocui.Gui, v *gocui.View) error {
	u.MemAddr -= 8
	if u.MemAddr < 0 {
		u.MemAddr = 0
	}
	u.Gui.Update(u.UpdateMemoryView)
	return nil
}

func (u *Ui) MicrocodeScrollDown(g *gocui.Gui, v *gocui.View) error {
	u.MCPos += 1
	if u.MCPos > 255 {
		u.MCPos = 255
	}
	u.Gui.Update(u.UpdateMicrocodeView)
	return nil
}

func (u *Ui) MicrocodeScrollUp(g *gocui.Gui, v *gocui.View) error {
	u.MCPos -= 1
	if u.MCPos < 0 {
		u.MCPos = 0
	}
	u.Gui.Update(u.UpdateMicrocodeView)
	return nil
}

func (u *Ui) SymGoto(g *gocui.Gui, v *gocui.View) error {
	v, err := g.View("symbols")
	if err != nil {
		return err
	}
	_, symi := v.Cursor()
	symi += u.SymPos
	sym := u.Mic.MemSymbols[symi].Val
	u.MemAddr = int(sym - (sym % 8))
	return nil
}

func (u *Ui) MemModeToggle(g *gocui.Gui, v *gocui.View) error {
	u.MemHex = !u.MemHex
	return nil
}

func (u *Ui) SymModeToggle(g *gocui.Gui, v *gocui.View) error {
	u.SymHex = !u.SymHex
	return nil
}


func (u *Ui) MicrocodeToggleBreakPoint(g *gocui.Gui, v *gocui.View) error {
	v, err := g.View("microcode")
	if err != nil {
		return err
	}
        u.Mic.RegistersLock.Lock()
        defer u.Mic.RegistersLock.Unlock()
        _, mci := v.Cursor()
        mci += u.MCPos
        u.Mic.MCC[mci].BR = !u.Mic.MCC[mci].BR
        return nil
}

/* Util functions */
func FocusView(g *gocui.Gui, v *gocui.View) {
	v.SelBgColor = gocui.ColorDefault
	v.SelFgColor = gocui.ColorGreen
	g.SetCurrentView(v.Name())
}

func DefocusView(g *gocui.Gui, v *gocui.View) {
	v.SelBgColor = gocui.ColorDefault
	v.SelFgColor = gocui.ColorRed
	g.SetCurrentView("")
}
