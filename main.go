//go:build windows

// Taklit — fare ve klavye işlemlerini kaydedip tekrar oynatan taşınabilir Windows aracı.
// Kurulum gerektirmez; ayarlar ve son kayıt exe'nin yanındaki dosyalarda tutulur.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	idRecord = 101 + iota
	idStop
	idPlay
	idRepeat
	idLoop
	idSpeed
	idPause
	idMoves
	idTopmost
	idOpen
	idSave
	idClear
)

const (
	hkRecord = 1
	hkStop   = 2
	hkPlay   = 3
	timerRec = 1
)

var speeds = []struct {
	label string
	v     float64
}{{"0.25x (yavaş)", 0.25}, {"0.5x", 0.5}, {"1x (gerçek)", 1}, {"2x", 2}, {"4x", 4}, {"10x", 10}}

var (
	mainHwnd uintptr
	ctl      = map[int]uintptr{}
	hStatus  uintptr
	hInfo    uintptr
	fontN    uintptr
	fontB    uintptr
	dpi      = uintptr(96)

	macro      *Macro
	macroName  string
	settings   Settings
	baseDir    string
	statusCol  uint32 // 0x00BBGGRR
	hotkeyWarn string
)

func sc(v int) int32 { return int32(uintptr(v) * dpi / 96) }

func settingsPath() string { return filepath.Join(baseDir, "taklit_ayar.json") }
func lastPath() string     { return filepath.Join(baseDir, "son_kayit.taklit") }

func main() {
	runtime.LockOSThread()
	if exe, err := os.Executable(); err == nil {
		baseDir = filepath.Dir(exe)
	}
	settings = LoadSettings(settingsPath())
	if m, err := LoadMacro(lastPath()); err == nil {
		macro, macroName = m, "son kayıt"
	}
	if pGetDpiForSystem.Find() == nil {
		dpi, _, _ = pGetDpiForSystem.Call()
	}
	fontN = mkFont(9, 400)
	fontB = mkFont(11, 700)

	hInst, _, _ := pGetModuleHandleW.Call(0)
	cls := u16("TaklitPencere")
	cursor, _, _ := pLoadCursorW.Call(0, 32512) // IDC_ARROW
	bg, _, _ := pGetSysColorBrush.Call(COLOR_BTNFACE)
	wc := WNDCLASSEXW{LpfnWndProc: syscall.NewCallback(wndProc), HInstance: hInst, HCursor: cursor, HbrBackground: bg, LpszClassName: cls, HIcon: appIcon(hInst)}
	wc.HIconSm = wc.HIcon
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	style := uintptr(WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_MINIMIZEBOX)
	r := RECT{0, 0, sc(436), sc(318)}
	pAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&r)), style, 0, 0)
	pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(u16("Taklit — Fare ve Klavye Kaydedici"))), style,
		CW_USEDEFAULT, CW_USEDEFAULT, uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top), 0, 0, hInst, 0)
	if mainHwnd == 0 {
		msgBox(0, "Pencere oluşturulamadı.", "Taklit", MB_ICONERROR)
		return
	}
	pShowWindow.Call(mainHwnd, SW_SHOW)
	pUpdateWindow.Call(mainHwnd)

	var msg MSG
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		if ok, _, _ := pIsDialogMessageW.Call(mainHwnd, uintptr(unsafe.Pointer(&msg))); ok != 0 {
			continue
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

// rsrc aracının gömdüğü simge; kimliği bilinmediği için ilk bulunan alınır
func appIcon(hInst uintptr) uintptr {
	for id := uintptr(1); id <= 16; id++ {
		if h, _, _ := pLoadIconW.Call(hInst, id); h != 0 {
			return h
		}
	}
	h, _, _ := pLoadIconW.Call(0, 32512) // IDI_APPLICATION
	return h
}

func mkFont(pt int, weight int) uintptr {
	h := -int32(uintptr(pt) * dpi / 72)
	f, _, _ := pCreateFontW.Call(uintptr(h), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16("Segoe UI"))))
	return f
}

func add(class, text string, id int, style uintptr, x, y, w, h int, exStyle uintptr) uintptr {
	hInst, _, _ := pGetModuleHandleW.Call(0)
	hw, _, _ := pCreateWindowExW.Call(exStyle, uintptr(unsafe.Pointer(u16(class))), uintptr(unsafe.Pointer(u16(text))),
		WS_CHILD|WS_VISIBLE|style, uintptr(sc(x)), uintptr(sc(y)), uintptr(sc(w)), uintptr(sc(h)), mainHwnd, uintptr(id), hInst, 0)
	sendMsg(hw, WM_SETFONT, fontN, 1)
	if id > 0 {
		ctl[id] = hw
	}
	return hw
}

func buildUI() {
	hStatus = add("STATIC", "", 0, SS_LEFT, 14, 10, 410, 26, 0)
	sendMsg(hStatus, WM_SETFONT, fontB, 1)

	add("BUTTON", "● Kayıt (F9)", idRecord, WS_TABSTOP|BS_PUSHBUTTON, 12, 42, 134, 38, 0)
	add("BUTTON", "■ Durdur (F10)", idStop, WS_TABSTOP|BS_PUSHBUTTON, 151, 42, 134, 38, 0)
	add("BUTTON", "▶ Oynat (F11)", idPlay, WS_TABSTOP|BS_PUSHBUTTON, 290, 42, 134, 38, 0)

	add("STATIC", "Tekrar sayısı:", 0, SS_LEFT, 14, 99, 92, 20, 0)
	add("EDIT", strconv.Itoa(settings.Repeat), idRepeat, WS_TABSTOP|ES_NUMBER|ES_CENTER|ES_AUTOHSCROLL, 108, 96, 56, 24, WS_EX_CLIENTEDGE)
	add("BUTTON", "Durdurana kadar sürekli tekrarla", idLoop, WS_TABSTOP|BS_AUTOCHECKBOX, 180, 97, 246, 22, 0)

	add("STATIC", "Hız:", 0, SS_LEFT, 14, 133, 40, 20, 0)
	cb := add("COMBOBOX", "", idSpeed, WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST, 108, 129, 120, 200, 0)
	for _, s := range speeds {
		sendMsg(cb, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(s.label))))
	}
	add("STATIC", "Tekrar arası (sn):", 0, SS_LEFT, 246, 133, 118, 20, 0)
	add("EDIT", strconv.Itoa(settings.PauseSec), idPause, WS_TABSTOP|ES_NUMBER|ES_CENTER, 368, 130, 56, 24, WS_EX_CLIENTEDGE)

	add("BUTTON", "Fare hareketlerini kaydet", idMoves, WS_TABSTOP|BS_AUTOCHECKBOX, 14, 164, 200, 22, 0)
	add("BUTTON", "Pencere hep üstte", idTopmost, WS_TABSTOP|BS_AUTOCHECKBOX, 246, 164, 180, 22, 0)

	add("BUTTON", "Aç…", idOpen, WS_TABSTOP|BS_PUSHBUTTON, 12, 198, 134, 30, 0)
	add("BUTTON", "Farklı kaydet…", idSave, WS_TABSTOP|BS_PUSHBUTTON, 151, 198, 134, 30, 0)
	add("BUTTON", "Kaydı temizle", idClear, WS_TABSTOP|BS_PUSHBUTTON, 290, 198, 134, 30, 0)

	hInfo = add("STATIC", "", 0, SS_LEFT, 14, 240, 410, 72, 0)

	if settings.SpeedIdx < 0 || settings.SpeedIdx >= len(speeds) {
		settings.SpeedIdx = 2
	}
	sendMsg(cb, CB_SETCURSEL, uintptr(settings.SpeedIdx), 0)
	setChecked(ctl[idLoop], settings.Loop)
	setChecked(ctl[idMoves], settings.RecMoves)
	setChecked(ctl[idTopmost], settings.Topmost)
	applyTopmost()

	var bad []string
	for _, k := range []struct {
		id   uintptr
		vk   uintptr
		name string
	}{{hkRecord, VK_F9, "F9"}, {hkStop, VK_F10, "F10"}, {hkPlay, VK_F11, "F11"}} {
		if r, _, _ := pRegisterHotKey.Call(mainHwnd, k.id, MOD_NOREPEAT, k.vk); r == 0 {
			bad = append(bad, k.name)
		}
	}
	if len(bad) > 0 {
		hotkeyWarn = "Uyarı: " + strings.Join(bad, ", ") + " kısayolu başka bir programda kullanılıyor; düğmeleri kullanın."
	}
	setIdle("Hazır")
}

func readSettings() {
	settings.Repeat, _ = strconv.Atoi(getText(ctl[idRepeat]))
	if settings.Repeat < 1 {
		settings.Repeat = 1
		setText(ctl[idRepeat], "1")
	}
	settings.PauseSec, _ = strconv.Atoi(getText(ctl[idPause]))
	settings.Loop = checked(ctl[idLoop])
	settings.RecMoves = checked(ctl[idMoves])
	settings.Topmost = checked(ctl[idTopmost])
	settings.SpeedIdx = int(sendMsg(ctl[idSpeed], CB_GETCURSEL, 0, 0))
	if settings.SpeedIdx < 0 || settings.SpeedIdx >= len(speeds) {
		settings.SpeedIdx = 2
	}
}

func applyTopmost() {
	after := HWND_NOTOPMOST
	if checked(ctl[idTopmost]) {
		after = HWND_TOPMOST
	}
	pSetWindowPos.Call(mainHwnd, after, 0, 0, 0, 0, SWP_NOMOVE|SWP_NOSIZE)
}

func setStatus(text string, color uint32) {
	statusCol = color
	setText(hStatus, text)
}

func refreshInfo() {
	src := ""
	if macro != nil && macroName != "" {
		src = "Kayıt: " + macroName + "\r\n"
	}
	help := "Esc (elle) ya da F10 oynatmayı durdurur."
	if hotkeyWarn != "" {
		help = hotkeyWarn
	}
	setText(hInfo, src+macro.Summary()+"\r\n"+help)
}

// düğmelerin etkinliği duruma göre
func setButtons(recording, playing bool) {
	idle := !recording && !playing
	has := macro != nil && len(macro.Events) > 0
	enable(ctl[idRecord], idle)
	enable(ctl[idStop], !idle)
	enable(ctl[idPlay], idle && has)
	for _, id := range []int{idRepeat, idLoop, idSpeed, idPause, idMoves, idOpen} {
		enable(ctl[id], idle)
	}
	enable(ctl[idSave], idle && has)
	enable(ctl[idClear], idle && has)
}

func setIdle(text string) {
	setStatus(text, 0x00404040)
	setButtons(false, false)
	refreshInfo()
}

func doRecord() {
	if rec.on || player.running.Load() {
		return
	}
	readSettings()
	if !startRecording(settings.RecMoves) {
		msgBox(mainHwnd, "Kayıt başlatılamadı (Windows fare/klavye kancasına izin vermedi).", "Taklit", MB_ICONERROR)
		return
	}
	setStatus("● Kaydediliyor… (F10 ile bitir)", 0x002020D0)
	setButtons(true, false)
	pSetTimer.Call(mainHwnd, timerRec, 250, 0)
}

func doStop() {
	if rec.on {
		pKillTimer.Call(mainHwnd, timerRec)
		ev := stopRecording()
		if len(ev) == 0 {
			setIdle("Kayıt boş — hiçbir işlem yakalanmadı")
			return
		}
		scr := virtualScreen()
		macro = &Macro{Version: 1, Created: time.Now(), Screen: [4]int32{scr.x, scr.y, scr.w, scr.h}, Events: ev}
		macroName = "son kayıt (" + time.Now().Format("15:04") + ")"
		if err := SaveMacro(lastPath(), macro); err != nil {
			hotkeyWarn = "Son kayıt diske yazılamadı: " + err.Error()
		}
		setIdle("✓ Kayıt tamam — Oynat (F11) ile tekrarlayın")
		return
	}
	if player.running.Load() {
		player.stop.Store(true)
	}
}

func doPlay() {
	if rec.on || player.running.Load() || macro == nil || len(macro.Events) == 0 {
		return
	}
	readSettings()
	SaveSettings(settingsPath(), settings)
	scr := virtualScreen()
	if macro.Screen[2] != 0 && (macro.Screen[2] != scr.w || macro.Screen[3] != scr.h) {
		hotkeyWarn = fmt.Sprintf("Dikkat: kayıt %dx%d ekranda yapılmış, şu an %dx%d — tıklamalar kayabilir.", macro.Screen[2], macro.Screen[3], scr.w, scr.h)
	}
	installHooks() // elle basılan Esc'i yakalamak için
	setButtons(false, true)
	setStatus("▶ Oynatılıyor…", 0x00208020)
	refreshInfo()
	startPlayback(mainHwnd, macro.Events, playOpts{
		Repeat: settings.Repeat, Loop: settings.Loop, Speed: speeds[settings.SpeedIdx].v,
		Pause: time.Duration(settings.PauseSec) * time.Second,
	})
}

func doOpen() {
	p, ok := fileDialog(mainHwnd, false, "Kayıt aç", "")
	if !ok {
		return
	}
	m, err := LoadMacro(p)
	if err != nil {
		msgBox(mainHwnd, err.Error(), "Kayıt açılamadı", MB_ICONERROR)
		return
	}
	macro, macroName = m, filepath.Base(p)
	setIdle("Kayıt yüklendi")
}

func doSave() {
	if macro == nil {
		return
	}
	p, ok := fileDialog(mainHwnd, true, "Kaydı farklı kaydet", "kayit_"+time.Now().Format("2006-01-02_15-04")+".taklit")
	if !ok {
		return
	}
	if err := SaveMacro(p, macro); err != nil {
		msgBox(mainHwnd, err.Error(), "Kaydedilemedi", MB_ICONERROR)
		return
	}
	macroName = filepath.Base(p)
	setIdle("✓ Kaydedildi: " + macroName)
}

func doClear() {
	macro, macroName = nil, ""
	_ = os.Remove(lastPath())
	setIdle("Kayıt temizlendi")
}

func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch uint32(msg) {
	case 0x0001: // WM_CREATE
		mainHwnd = hwnd
		buildUI()
		return 0
	case WM_COMMAND:
		id, code := int(wParam&0xFFFF), uint32(wParam>>16)
		if code == BN_CLICKED {
			switch id {
			case idRecord:
				doRecord()
			case idStop:
				doStop()
			case idPlay:
				doPlay()
			case idOpen:
				doOpen()
			case idSave:
				doSave()
			case idClear:
				doClear()
			case idTopmost:
				applyTopmost()
			}
		}
		return 0
	case WM_HOTKEY:
		switch wParam {
		case hkRecord:
			doRecord()
		case hkStop:
			doStop()
		case hkPlay:
			doPlay()
		}
		return 0
	case WM_TIMER:
		if wParam == timerRec && rec.on {
			setStatus(fmt.Sprintf("● Kaydediliyor… %d olay · %s (F10 ile bitir)", len(rec.events), fmtDur(time.Since(rec.t0))), 0x002020D0)
		}
		return 0
	case msgPlayProgress:
		total := len(macro.Events)
		reps := strconv.Itoa(settings.Repeat)
		if settings.Loop {
			reps = "∞"
		}
		setStatus(fmt.Sprintf("▶ Oynatılıyor — tekrar %d/%s · olay %d/%d", wParam, reps, lParam, total), 0x00208020)
		return 0
	case msgPlayDone:
		if !rec.on {
			removeHooks()
		}
		if wParam == 1 {
			setIdle("■ Oynatma durduruldu")
		} else {
			setIdle("✓ Oynatma bitti")
		}
		return 0
	case WM_CTLCOLORSTATIC:
		if lParam == hStatus {
			pSetTextColor.Call(wParam, uintptr(statusCol))
			pSetBkMode.Call(wParam, TRANSPARENT)
			bg, _, _ := pGetSysColorBrush.Call(COLOR_BTNFACE)
			return bg
		}
	case WM_CLOSE:
		player.stop.Store(true)
		if rec.on {
			stopRecording()
		}
		readSettings()
		SaveSettings(settingsPath(), settings)
		pDestroyWindow.Call(hwnd)
		return 0
	case WM_DESTROY:
		removeHooks()
		for _, id := range []uintptr{hkRecord, hkStop, hkPlay} {
			pUnregisterHotKey.Call(hwnd, id)
		}
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return r
}
