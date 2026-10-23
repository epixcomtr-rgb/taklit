//go:build windows

package main

import (
	"syscall"
	"time"
	"unsafe"
)

// Kayıt: düşük seviyeli fare ve klavye kancaları (WH_MOUSE_LL / WH_KEYBOARD_LL).
// Kanca geri çağrıları onları kuran iş parçacığında (GUI iş parçacığı) mesaj döngüsü içinde çalışır,
// bu yüzden aşağıdaki durum yalnızca o iş parçacığından değiştirilir.

const (
	injectMagic   = 0x54414B4C // "TAKL": kendi gönderdiğimiz girdileri tanımak için dwExtraInfo
	llmhfInjected = 0x01
	llkhfInjected = 0x10
)

var rec struct {
	on      bool
	moves   bool
	t0      time.Time
	events  []Event
	lastX   int32
	lastY   int32
	skipBtn [6]bool // kendi penceremizde başlayan tıklamalar (Durdur düğmesi vb.) kayda girmez
}

var (
	mouseHook, kbdHook uintptr
	mouseCb            = syscall.NewCallback(mouseProc)
	kbdCb              = syscall.NewCallback(kbdProc)
)

func installHooks() bool {
	hInst, _, _ := pGetModuleHandleW.Call(0)
	if mouseHook == 0 {
		mouseHook, _, _ = pSetWindowsHookExW.Call(WH_MOUSE_LL, mouseCb, hInst, 0)
	}
	if kbdHook == 0 {
		kbdHook, _, _ = pSetWindowsHookExW.Call(WH_KEYBOARD_LL, kbdCb, hInst, 0)
	}
	return mouseHook != 0 && kbdHook != 0
}

func removeHooks() {
	if mouseHook != 0 {
		pUnhookWindowsHookEx.Call(mouseHook)
		mouseHook = 0
	}
	if kbdHook != 0 {
		pUnhookWindowsHookEx.Call(kbdHook)
		kbdHook = 0
	}
}

func startRecording(moves bool) bool {
	rec.on, rec.moves, rec.t0, rec.events = true, moves, time.Now(), nil
	rec.lastX, rec.lastY = -1<<30, -1<<30
	rec.skipBtn = [6]bool{}
	if !installHooks() {
		rec.on = false
		removeHooks()
		return false
	}
	return true
}

func stopRecording() []Event {
	rec.on = false
	if !player.running.Load() {
		removeHooks()
	}
	ev := rec.events
	rec.events = nil
	// baştaki boşluğu at: ilk olay 0. ms'ye çekilir
	if len(ev) > 0 {
		off := ev[0].T
		for i := range ev {
			ev[i].T -= off
		}
	}
	return ev
}

func elapsedMs() int64 { return time.Since(rec.t0).Milliseconds() }

func onOwnWindow(pt POINT) bool {
	h, _, _ := pWindowFromPoint.Call(uintptr(uint32(pt.X)) | uintptr(uint32(pt.Y))<<32)
	if h == 0 {
		return false
	}
	root, _, _ := pGetAncestor.Call(h, GA_ROOT)
	return root == mainHwnd
}

func mouseProc(nCode, wParam, lParam uintptr) uintptr {
	if int32(nCode) >= 0 && rec.on {
		m := (*MSLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		if m.ExtraInfo != injectMagic {
			recordMouse(uint32(wParam), m)
		}
	}
	r, _, _ := pCallNextHookEx.Call(0, nCode, wParam, lParam)
	return r
}

func recordMouse(msg uint32, m *MSLLHOOKSTRUCT) {
	x, y := m.Pt.X, m.Pt.Y
	btn, down := 0, false
	switch msg {
	case WM_MOUSEMOVE:
		if !rec.moves || (x == rec.lastX && y == rec.lastY) {
			return
		}
		rec.lastX, rec.lastY = x, y
		rec.events = append(rec.events, Event{T: elapsedMs(), Type: EvMove, X: x, Y: y})
		return
	case WM_MOUSEWHEEL, WM_MOUSEHWHEEL:
		if onOwnWindow(m.Pt) {
			return
		}
		t := EvWheel
		if msg == WM_MOUSEHWHEEL {
			t = EvHWheel
		}
		rec.events = append(rec.events, Event{T: elapsedMs(), Type: t, X: x, Y: y, Delta: int32(int16(m.MouseData >> 16))})
		return
	case WM_LBUTTONDOWN:
		btn, down = BtnLeft, true
	case WM_LBUTTONUP:
		btn = BtnLeft
	case WM_RBUTTONDOWN:
		btn, down = BtnRight, true
	case WM_RBUTTONUP:
		btn = BtnRight
	case WM_MBUTTONDOWN:
		btn, down = BtnMiddle, true
	case WM_MBUTTONUP:
		btn = BtnMiddle
	case WM_XBUTTONDOWN, WM_XBUTTONUP:
		btn, down = BtnX1, msg == WM_XBUTTONDOWN
		if m.MouseData>>16 == 2 {
			btn = BtnX2
		}
	default:
		return
	}
	if down {
		if onOwnWindow(m.Pt) {
			rec.skipBtn[btn] = true
			return
		}
	} else if rec.skipBtn[btn] {
		rec.skipBtn[btn] = false
		return
	}
	t := EvUp
	if down {
		t = EvDown
	}
	rec.events = append(rec.events, Event{T: elapsedMs(), Type: t, X: x, Y: y, Btn: btn})
}

func kbdProc(nCode, wParam, lParam uintptr) uintptr {
	if int32(nCode) >= 0 {
		k := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		msg := uint32(wParam)
		isDown := msg == WM_KEYDOWN || msg == WM_SYSKEYDOWN
		physical := k.Flags&llkhfInjected == 0
		// oynatma sırasında elle basılan Esc oynatmayı durdurur
		if player.running.Load() && physical && isDown && k.VkCode == VK_ESCAPE {
			player.stop.Store(true)
		}
		if rec.on && k.ExtraInfo != injectMagic {
			switch k.VkCode {
			case VK_F9, VK_F10, VK_F11: // kısayol tuşları kayda girmez
			default:
				t := EvKeyUp
				if isDown {
					t = EvKeyDown
				}
				rec.events = append(rec.events, Event{T: elapsedMs(), Type: t, Vk: uint16(k.VkCode), Scan: uint16(k.ScanCode), Ext: k.Flags&LLKHF_EXTENDED != 0})
			}
		}
	}
	r, _, _ := pCallNextHookEx.Call(0, nCode, wParam, lParam)
	return r
}
