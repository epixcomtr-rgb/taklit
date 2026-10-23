package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Olay türleri
const (
	EvMove    = "move"
	EvDown    = "down"  // fare düğmesi bas
	EvUp      = "up"    // fare düğmesi bırak
	EvWheel   = "wheel" // dikey tekerlek
	EvHWheel  = "hwheel"
	EvKeyDown = "kdown"
	EvKeyUp   = "kup"
)

// Fare düğmeleri
const (
	BtnLeft   = 1
	BtnRight  = 2
	BtnMiddle = 3
	BtnX1     = 4
	BtnX2     = 5
)

// Event: kayıt başlangıcından itibaren T milisaniyede olan tek bir fare/klavye olayı.
// Koordinatlar fiziksel ekran pikselidir (program DPI farkında çalışır).
type Event struct {
	T     int64  `json:"t"`
	Type  string `json:"e"`
	X     int32  `json:"x,omitempty"`
	Y     int32  `json:"y,omitempty"`
	Btn   int    `json:"b,omitempty"`
	Delta int32  `json:"d,omitempty"`
	Vk    uint16 `json:"vk,omitempty"`
	Scan  uint16 `json:"sc,omitempty"`
	Ext   bool   `json:"ext,omitempty"`
}

type Macro struct {
	Version int       `json:"version"`
	Created time.Time `json:"created"`
	Screen  [4]int32  `json:"screen"` // kayıt anındaki sanal ekran: x, y, genişlik, yükseklik
	Events  []Event   `json:"events"`
}

func (m *Macro) Duration() time.Duration {
	if m == nil || len(m.Events) == 0 {
		return 0
	}
	return time.Duration(m.Events[len(m.Events)-1].T) * time.Millisecond
}

func (m *Macro) Summary() string {
	if m == nil || len(m.Events) == 0 {
		return "Kayıt yok"
	}
	var clicks, keys, moves int
	for _, e := range m.Events {
		switch e.Type {
		case EvDown:
			clicks++
		case EvKeyDown:
			keys++
		case EvMove:
			moves++
		}
	}
	return fmt.Sprintf("%d olay · %d tıklama · %d tuş · %d hareket · süre %s", len(m.Events), clicks, keys, moves, fmtDur(m.Duration()))
}

func fmtDur(d time.Duration) string {
	s := int(d.Round(100*time.Millisecond) / (100 * time.Millisecond))
	if s >= 600 {
		return fmt.Sprintf("%d dk %d sn", s/600, (s%600)/10)
	}
	return fmt.Sprintf("%d.%d sn", s/10, s%10)
}

func SaveMacro(path string, m *Macro) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadMacro(path string) (*Macro, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Macro
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("dosya okunamadı: %v", err)
	}
	if m.Version != 1 {
		return nil, fmt.Errorf("desteklenmeyen kayıt sürümü: %d", m.Version)
	}
	return &m, nil
}

// Settings: exe'nin yanındaki taklit_ayar.json (taşınabilir).
type Settings struct {
	Repeat   int  `json:"repeat"`
	Loop     bool `json:"loop"`
	SpeedIdx int  `json:"speed"`
	PauseSec int  `json:"pause"`
	RecMoves bool `json:"moves"`
	Topmost  bool `json:"topmost"`
}

func DefaultSettings() Settings {
	return Settings{Repeat: 1, SpeedIdx: 2, RecMoves: true, Topmost: true}
}

func LoadSettings(path string) Settings {
	s := DefaultSettings()
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	return s
}

func SaveSettings(path string, s Settings) {
	if b, err := json.MarshalIndent(s, "", "  "); err == nil {
		_ = os.WriteFile(path, b, 0644)
	}
}
