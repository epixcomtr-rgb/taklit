//go:build windows

package main

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

// Oynatma: kayıttaki olaylar aynı zamanlamayla (hız çarpanıyla) SendInput ile yeniden üretilir.
// Ayrı bir goroutine'de çalışır; ilerleme GUI'ye PostMessage ile bildirilir.

const (
	msgPlayProgress = WM_APP + 1 // wParam: tekrar no, lParam: olay sırası
	msgPlayDone     = WM_APP + 2 // wParam: 1 = kullanıcı durdurdu
)

var player struct {
	running atomic.Bool
	stop    atomic.Bool
}

type playOpts struct {
	Repeat int
	Loop   bool
	Speed  float64
	Pause  time.Duration
}

type screenBox struct{ x, y, w, h int32 }

func virtualScreen() screenBox {
	return screenBox{sysMetric(SM_XVIRTUALSCREEN), sysMetric(SM_YVIRTUALSCREEN), sysMetric(SM_CXVIRTUALSCREEN), sysMetric(SM_CYVIRTUALSCREEN)}
}

// fiziksel piksel → SendInput'un 0..65535 mutlak koordinatı (tüm monitörler)
func (s screenBox) abs(x, y int32) (int32, int32) {
	dx := (float64(x-s.x) + 0.5) * 65536 / float64(s.w)
	dy := (float64(y-s.y) + 0.5) * 65536 / float64(s.h)
	return int32(dx), int32(dy)
}

func sendMouse(flags uint32, dx, dy int32, data uint32) {
	in := MOUSEINPUT{Type: INPUT_MOUSE, Dx: dx, Dy: dy, MouseData: data, Flags: flags, ExtraInfo: injectMagic}
	pSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
}

func sendKey(vk, scan uint16, ext, up bool) {
	var f uint32
	if ext {
		f |= KEYEVENTF_EXTENDEDKEY
	}
	if up {
		f |= KEYEVENTF_KEYUP
	}
	in := KEYBDINPUT{Type: INPUT_KEYBOARD, Vk: vk, Scan: scan, Flags: f, ExtraInfo: injectMagic}
	pSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
}

func btnFlags(btn int, down bool) (uint32, uint32) {
	switch btn {
	case BtnLeft:
		if down {
			return MOUSEEVENTF_LEFTDOWN, 0
		}
		return MOUSEEVENTF_LEFTUP, 0
	case BtnRight:
		if down {
			return MOUSEEVENTF_RIGHTDOWN, 0
		}
		return MOUSEEVENTF_RIGHTUP, 0
	case BtnMiddle:
		if down {
			return MOUSEEVENTF_MIDDLEDOWN, 0
		}
		return MOUSEEVENTF_MIDDLEUP, 0
	case BtnX1, BtnX2:
		data := uint32(1)
		if btn == BtnX2 {
			data = 2
		}
		if down {
			return MOUSEEVENTF_XDOWN, data
		}
		return MOUSEEVENTF_XUP, data
	}
	return 0, 0
}

// basılı kalan tuş/düğmeler: durdurulunca bırakılır ki klavye/fare takılı kalmasın
type held struct {
	keys map[uint16]Event
	btns map[int]Event
}

func (h *held) send(scr screenBox, e Event) {
	const absMove = MOUSEEVENTF_MOVE | MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_VIRTUALDESK
	switch e.Type {
	case EvMove:
		dx, dy := scr.abs(e.X, e.Y)
		sendMouse(absMove, dx, dy, 0)
	case EvDown, EvUp:
		f, data := btnFlags(e.Btn, e.Type == EvDown)
		if f == 0 {
			return
		}
		dx, dy := scr.abs(e.X, e.Y)
		sendMouse(absMove|f, dx, dy, data)
		if e.Type == EvDown {
			h.btns[e.Btn] = e
		} else {
			delete(h.btns, e.Btn)
		}
	case EvWheel, EvHWheel:
		f := uint32(MOUSEEVENTF_WHEEL)
		if e.Type == EvHWheel {
			f = MOUSEEVENTF_HWHEEL
		}
		dx, dy := scr.abs(e.X, e.Y)
		sendMouse(absMove|f, dx, dy, uint32(e.Delta))
	case EvKeyDown:
		sendKey(e.Vk, e.Scan, e.Ext, false)
		h.keys[e.Vk] = e
	case EvKeyUp:
		sendKey(e.Vk, e.Scan, e.Ext, true)
		delete(h.keys, e.Vk)
	}
}

func (h *held) releaseAll(scr screenBox) {
	for _, e := range h.keys {
		sendKey(e.Vk, e.Scan, e.Ext, true)
	}
	for _, e := range h.btns {
		f, data := btnFlags(e.Btn, false)
		dx, dy := scr.abs(e.X, e.Y)
		sendMouse(MOUSEEVENTF_MOVE|MOUSEEVENTF_ABSOLUTE|MOUSEEVENTF_VIRTUALDESK|f, dx, dy, data)
	}
	h.keys, h.btns = map[uint16]Event{}, map[int]Event{}
}

// durdurulana kadar d süre bekler; durdurulduysa false
func sleepUntil(deadline time.Time) bool {
	for {
		if player.stop.Load() {
			return false
		}
		left := time.Until(deadline)
		if left <= 0 {
			return true
		}
		if left > 15*time.Millisecond {
			left = 15 * time.Millisecond
		}
		time.Sleep(left)
	}
}

func startPlayback(hwnd uintptr, events []Event, o playOpts) {
	player.stop.Store(false)
	player.running.Store(true)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		pTimeBeginPeriod.Call(1)
		defer pTimeEndPeriod.Call(1)

		scr := virtualScreen()
		h := &held{keys: map[uint16]Event{}, btns: map[int]Event{}}
		stopped := false
		// Oynat düğmesine/F11'e basan el çekilsin
		if !sleepUntil(time.Now().Add(400 * time.Millisecond)) {
			stopped = true
		}
		for rep := 1; !stopped && (o.Loop || rep <= o.Repeat); rep++ {
			pPostMessageW.Call(hwnd, msgPlayProgress, uintptr(rep), 0)
			start := time.Now()
			for i, e := range events {
				at := start.Add(time.Duration(float64(e.T) / o.Speed * float64(time.Millisecond)))
				if !sleepUntil(at) {
					stopped = true
					break
				}
				h.send(scr, e)
				if i%25 == 0 {
					pPostMessageW.Call(hwnd, msgPlayProgress, uintptr(rep), uintptr(i+1))
				}
			}
			h.releaseAll(scr)
			if !stopped && (o.Loop || rep < o.Repeat) && o.Pause > 0 {
				if !sleepUntil(time.Now().Add(o.Pause)) {
					stopped = true
				}
			}
		}
		h.releaseAll(scr)
		player.running.Store(false)
		w := uintptr(0)
		if stopped {
			w = 1
		}
		pPostMessageW.Call(hwnd, msgPlayDone, w, 0)
	}()
}
