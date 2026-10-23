# Taklit

**Fare ve klavye işlemlerini kaydedip tekrar oynatan, kurulum gerektirmeyen Windows aracı.**

Tek bir `Taklit.exe` dosyası. Kurulum yok, .NET/çalışma zamanı gerekmez. Tekrarlayan tıklama ve klavye işlerini bir kez kaydedin, istediğiniz kadar tekrarlatın.

> **İndir:** Depodaki [`Taklit.exe`](Taklit.exe) dosyası (ya da [Releases](../../releases/latest) sayfası)

## Özellikler

- **Kayıt:** Sol / sağ / orta / yan (X1, X2) fare düğmeleri, dikey ve yatay tekerlek, fare hareketleri ve tüm klavye tuşları (Ctrl, Alt, Shift kombinasyonları dahil) milisaniye zamanlamasıyla kaydedilir.
- **Oynatma:** Kayıt aynı zamanlamayla `SendInput` üzerinden yeniden üretilir.
  - Tekrar sayısı ya da durdurana kadar sürekli tekrar
  - Hız: 0.25x, 0.5x, 1x, 2x, 4x, 10x
  - Tekrarlar arası bekleme süresi
- **Kısayollar:** `F9` kayıt · `F10` durdur · `F11` oynat · oynatma sırasında elle basılan `Esc` durdurur.
- **Güvenli durdurma:** Oynatma yarıda kesilirse basılı kalan tuş ve fare düğmeleri otomatik bırakılır.
- **Kendini kaydetmez:** Taklit penceresine yapılan tıklamalar ve F9/F10/F11 kayda girmez.
- **Dosyaya kaydetme:** Kayıtlar `.taklit` (JSON) dosyası olarak saklanıp açılabilir. Son kayıt ve ayarlar exe'nin yanında tutulur.
- **Ekran ölçekleme ve çoklu monitör:** Per-Monitor DPI farkında. %125 / %150 ölçeklemede ve birden çok monitörde tıklamalar doğru piksele gider.
- **Taşınabilir:** Kayıt defterine, AppData'ya hiçbir şey yazmaz.

## Kullanım

1. `Taklit.exe`'yi yazılabilir bir klasöre koyup çalıştırın.
2. **● Kayıt (F9)** → yapmak istediğiniz işlemleri yapın → **■ Durdur (F10)**.
3. Tekrar sayısı, hız gibi ayarları seçin → **▶ Oynat (F11)**.
4. Kaydı saklamak için **Farklı kaydet…**, sonra tekrar kullanmak için **Aç…**.

Ayrıntılı Türkçe kılavuz: [KULLANIM.txt](KULLANIM.txt)

### Bilinmesi gerekenler

- Taklit ekrana bakmaz, **kaydedilen koordinatlara** tıklar. Pencerelerin yeri ve ekran çözünürlüğü kayıt anındakiyle aynı olmalı. Çözünürlük farklıysa program uyarır.
- **Yönetici olarak çalışan** bir programa girdi göndermek için Taklit'i de yönetici olarak çalıştırın (Windows UIPI kısıtı).
- Exe imzasızdır. İlk açılışta SmartScreen uyarısı çıkabilir: *Ek bilgi → Yine de çalıştır*. Düşük seviyeli klavye kancası kullandığı için bazı antivirüsler sezgisel olarak işaretleyebilir. Kaynak kod açıktır, kendiniz derleyebilirsiniz.
- UAC onay ekranı, Ctrl+Alt+Del gibi güvenli masaüstüne girdi gönderilemez (Windows kısıtı).

## Derleme

Go 1.22+ gerekir. cgo kullanılmaz, bu yüzden Linux, macOS ya da Windows'ta çapraz derlenebilir.

```bash
./derle.sh          # → Taklit.exe (windows/amd64)
```

Betik sırasıyla:
1. [`rsrc`](https://github.com/akavel/rsrc) ile manifest ve simgeyi gömer (`rsrc_windows_amd64.syso`).
2. `go vet` çalıştırır.
3. `-H windowsgui` ile derler.

Windows'ta elle derlemek için:

```powershell
go run github.com/akavel/rsrc@v0.10.2 -manifest taklit.manifest -ico taklit.ico -arch amd64 -o rsrc_windows_amd64.syso
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -trimpath -ldflags "-H windowsgui -s -w" -o Taklit.exe .
```

## Kod yapısı

| Dosya | İçerik |
|---|---|
| `main.go` | Win32 pencere, düğmeler, kısayol tuşları, durum yönetimi |
| `record.go` | `WH_MOUSE_LL` / `WH_KEYBOARD_LL` kancalarıyla kayıt |
| `play.go` | `SendInput` ile zamanlamalı oynatma, durdurma, takılı tuşları bırakma |
| `macro.go` | `.taklit` dosya biçimi (JSON) ve ayarlar |
| `win32.go` | Windows API tanımları (yalnızca standart `syscall`, ek bağımlılık yok) |
| `taklit.manifest` | Common Controls v6 görünümü, PerMonitorV2 DPI, `asInvoker` |

### `.taklit` dosya biçimi

```json
{
  "version": 1,
  "created": "2026-10-23T20:15:00+03:00",
  "screen": [0, 0, 1920, 1080],
  "events": [
    {"t": 0,   "e": "move",  "x": 812, "y": 440},
    {"t": 120, "e": "down",  "x": 812, "y": 440, "b": 1},
    {"t": 190, "e": "up",    "x": 812, "y": 440, "b": 1},
    {"t": 900, "e": "kdown", "vk": 65, "sc": 30},
    {"t": 960, "e": "kup",   "vk": 65, "sc": 30}
  ]
}
```

- `t`: kayıt başından itibaren milisaniye.
- `e`: olay türü. `move`, `down`, `up`, `wheel`, `hwheel`, `kdown`, `kup` olabilir.
- `b`: fare düğmesi. 1 sol, 2 sağ, 3 orta, 4 X1, 5 X2.
- `vk` / `sc` / `ext`: sanal tuş kodu / tarama kodu / genişletilmiş tuş.

## Lisans

[MIT](LICENSE)
