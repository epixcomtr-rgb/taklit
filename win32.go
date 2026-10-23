//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// Win32 API köprüsü: yalnızca standart kütüphanedeki syscall kullanılır (cgo yok, ek bağımlılık yok).

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	winmm    = syscall.NewLazyDLL("winmm.dll")

	pRegisterClassExW    = user32.NewProc("RegisterClassExW")
	pCreateWindowExW     = user32.NewProc("CreateWindowExW")
	pDefWindowProcW      = user32.NewProc("DefWindowProcW")
	pGetMessageW         = user32.NewProc("GetMessageW")
	pIsDialogMessageW    = user32.NewProc("IsDialogMessageW")
	pTranslateMessage    = user32.NewProc("TranslateMessage")
	pDispatchMessageW    = user32.NewProc("DispatchMessageW")
	pPostMessageW        = user32.NewProc("PostMessageW")
	pSendMessageW        = user32.NewProc("SendMessageW")
	pPostQuitMessage     = user32.NewProc("PostQuitMessage")
	pDestroyWindow       = user32.NewProc("DestroyWindow")
	pShowWindow          = user32.NewProc("ShowWindow")
	pUpdateWindow        = user32.NewProc("UpdateWindow")
	pSetWindowTextW      = user32.NewProc("SetWindowTextW")
	pGetWindowTextW      = user32.NewProc("GetWindowTextW")
	pEnableWindow        = user32.NewProc("EnableWindow")
	pSetWindowPos        = user32.NewProc("SetWindowPos")
	pAdjustWindowRectEx  = user32.NewProc("AdjustWindowRectEx")
	pLoadCursorW         = user32.NewProc("LoadCursorW")
	pLoadIconW           = user32.NewProc("LoadIconW")
	pMessageBoxW         = user32.NewProc("MessageBoxW")
	pSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	pUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	pCallNextHookEx      = user32.NewProc("CallNextHookEx")
	pSendInput           = user32.NewProc("SendInput")
	pGetSystemMetrics    = user32.NewProc("GetSystemMetrics")
	pWindowFromPoint     = user32.NewProc("WindowFromPoint")
	pGetAncestor         = user32.NewProc("GetAncestor")
	pRegisterHotKey      = user32.NewProc("RegisterHotKey")
	pUnregisterHotKey    = user32.NewProc("UnregisterHotKey")
	pSetTimer            = user32.NewProc("SetTimer")
	pKillTimer           = user32.NewProc("KillTimer")
	pGetDpiForSystem     = user32.NewProc("GetDpiForSystem")
	pGetSysColorBrush    = user32.NewProc("GetSysColorBrush")
	pSetForegroundWindow = user32.NewProc("SetForegroundWindow")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")

	pCreateFontW  = gdi32.NewProc("CreateFontW")
	pSetTextColor = gdi32.NewProc("SetTextColor")
	pSetBkMode    = gdi32.NewProc("SetBkMode")

	pGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	pGetSaveFileNameW = comdlg32.NewProc("GetSaveFileNameW")

	pTimeBeginPeriod = winmm.NewProc("timeBeginPeriod")
	pTimeEndPeriod   = winmm.NewProc("timeEndPeriod")
)

const (
	WS_OVERLAPPED    = 0x00000000
	WS_CAPTION       = 0x00C00000
	WS_SYSMENU       = 0x00080000
	WS_MINIMIZEBOX   = 0x00020000
	WS_CHILD         = 0x40000000
	WS_VISIBLE       = 0x10000000
	WS_TABSTOP       = 0x00010000
	WS_VSCROLL       = 0x00200000
	WS_EX_TOPMOST    = 0x00000008
	WS_EX_CLIENTEDGE = 0x00000200

	BS_PUSHBUTTON    = 0x0
	BS_AUTOCHECKBOX  = 0x3
	ES_NUMBER        = 0x2000
	ES_AUTOHSCROLL   = 0x80
	ES_CENTER        = 0x1
	SS_LEFT          = 0x0
	CBS_DROPDOWNLIST = 0x3

	CW_USEDEFAULT = 0x80000000
	SW_SHOW       = 5

	WM_DESTROY        = 0x0002
	WM_CLOSE          = 0x0010
	WM_SETFONT        = 0x0030
	WM_COMMAND        = 0x0111
	WM_TIMER          = 0x0113
	WM_CTLCOLORSTATIC = 0x0138
	WM_HOTKEY         = 0x0312
	WM_APP            = 0x8000

	WM_KEYDOWN    = 0x0100
	WM_KEYUP      = 0x0101
	WM_SYSKEYDOWN = 0x0104
	WM_SYSKEYUP   = 0x0105

	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONDOWN = 0x0204
	WM_RBUTTONUP   = 0x0205
	WM_MBUTTONDOWN = 0x0207
	WM_MBUTTONUP   = 0x0208
	WM_MOUSEWHEEL  = 0x020A
	WM_XBUTTONDOWN = 0x020B
	WM_XBUTTONUP   = 0x020C
	WM_MOUSEHWHEEL = 0x020E

	BN_CLICKED   = 0
	BM_GETCHECK  = 0x00F0
	BM_SETCHECK  = 0x00F1
	CB_ADDSTRING = 0x0143
	CB_GETCURSEL = 0x0147
	CB_SETCURSEL = 0x014E

	WH_KEYBOARD_LL = 13
	WH_MOUSE_LL    = 14

	LLKHF_EXTENDED = 0x01

	INPUT_MOUSE    = 0
	INPUT_KEYBOARD = 1

	MOUSEEVENTF_MOVE        = 0x0001
	MOUSEEVENTF_LEFTDOWN    = 0x0002
	MOUSEEVENTF_LEFTUP      = 0x0004
	MOUSEEVENTF_RIGHTDOWN   = 0x0008
	MOUSEEVENTF_RIGHTUP     = 0x0010
	MOUSEEVENTF_MIDDLEDOWN  = 0x0020
	MOUSEEVENTF_MIDDLEUP    = 0x0040
	MOUSEEVENTF_XDOWN       = 0x0080
	MOUSEEVENTF_XUP         = 0x0100
	MOUSEEVENTF_WHEEL       = 0x0800
	MOUSEEVENTF_HWHEEL      = 0x1000
	MOUSEEVENTF_VIRTUALDESK = 0x4000
	MOUSEEVENTF_ABSOLUTE    = 0x8000

	KEYEVENTF_EXTENDEDKEY = 0x0001
	KEYEVENTF_KEYUP       = 0x0002

	SM_XVIRTUALSCREEN  = 76
	SM_YVIRTUALSCREEN  = 77
	SM_CXVIRTUALSCREEN = 78
	SM_CYVIRTUALSCREEN = 79

	GA_ROOT = 2

	MOD_NOREPEAT = 0x4000
	VK_ESCAPE    = 0x1B
	VK_F9        = 0x78
	VK_F10       = 0x79
	VK_F11       = 0x7A

	HWND_TOPMOST   = ^uintptr(0) // -1
	HWND_NOTOPMOST = ^uintptr(1) // -2
	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002

	MB_OK        = 0x0
	MB_ICONERROR = 0x10
	MB_ICONINFO  = 0x40

	OFN_OVERWRITEPROMPT = 0x00000002
	OFN_PATHMUSTEXIST   = 0x00000800
	OFN_FILEMUSTEXIST   = 0x00001000
	OFN_NOCHANGEDIR     = 0x00000008

	COLOR_BTNFACE = 15
	TRANSPARENT   = 1
)

type POINT struct{ X, Y int32 }

type RECT struct{ Left, Top, Right, Bottom int32 }

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
	_       uint32
}

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type MSLLHOOKSTRUCT struct {
	Pt        POINT
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type KBDLLHOOKSTRUCT struct {
	VkCode    uint32
	ScanCode  uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

// INPUT (amd64'te 40 bayt): tip + hizalama + birleşim. Fare ve klavye için ayrı ayrı tanımlanır,
// ikisi de aynı boyuttadır, SendInput'a tek tek gönderilir.
type MOUSEINPUT struct {
	Type      uint32
	_         uint32
	Dx, Dy    int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type KEYBDINPUT struct {
	Type      uint32
	_         uint32
	Vk, Scan  uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
	_         [8]byte
}

type OPENFILENAMEW struct {
	LStructSize       uint32
	HwndOwner         uintptr
	HInstance         uintptr
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
	PvReserved        uintptr
	DwReserved        uint32
	FlagsEx           uint32
}

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func sendMsg(h uintptr, msg uint32, w, l uintptr) uintptr {
	r, _, _ := pSendMessageW.Call(h, uintptr(msg), w, l)
	return r
}

func setText(h uintptr, s string) { pSetWindowTextW.Call(h, uintptr(unsafe.Pointer(u16(s)))) }

func getText(h uintptr) string {
	buf := make([]uint16, 256)
	pGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func enable(h uintptr, on bool) {
	v := uintptr(0)
	if on {
		v = 1
	}
	pEnableWindow.Call(h, v)
}

func checked(h uintptr) bool { return sendMsg(h, BM_GETCHECK, 0, 0) == 1 }

func setChecked(h uintptr, on bool) {
	v := uintptr(0)
	if on {
		v = 1
	}
	sendMsg(h, BM_SETCHECK, v, 0)
}

func msgBox(h uintptr, text, title string, flags uintptr) {
	pMessageBoxW.Call(h, uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(title))), flags)
}

func sysMetric(i int) int32 {
	r, _, _ := pGetSystemMetrics.Call(uintptr(i))
	return int32(r)
}

// dosya iletişim kutusu filtresi: "Ad\0*.uzantı\0...\0\0"
func filterStr(parts ...string) *uint16 {
	var out []uint16
	for _, p := range parts {
		out = append(out, syscall.StringToUTF16(p)...)
	}
	out = append(out, 0)
	return &out[0]
}

func fileDialog(owner uintptr, save bool, title, initial string) (string, bool) {
	buf := make([]uint16, 1024)
	copy(buf, syscall.StringToUTF16(initial))
	ofn := OPENFILENAMEW{
		HwndOwner:   owner,
		LpstrFilter: filterStr("Taklit kaydı (*.taklit)", "*.taklit", "Tüm dosyalar", "*.*"),
		LpstrFile:   &buf[0],
		NMaxFile:    uint32(len(buf)),
		LpstrTitle:  u16(title),
		LpstrDefExt: u16("taklit"),
		Flags:       OFN_PATHMUSTEXIST | OFN_NOCHANGEDIR,
	}
	ofn.LStructSize = uint32(unsafe.Sizeof(ofn))
	proc := pGetOpenFileNameW
	if save {
		ofn.Flags |= OFN_OVERWRITEPROMPT
		proc = pGetSaveFileNameW
	} else {
		ofn.Flags |= OFN_FILEMUSTEXIST
	}
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return "", false
	}
	return syscall.UTF16ToString(buf), true
}
