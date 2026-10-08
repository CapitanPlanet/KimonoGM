# KIMONO GM — janusz-maker / editor 1.0

> **NO PAIN NO KIEŁBASA — RAN FOR JUR LAJF — FAJT KLUB MANIECKI**
> `VHS • 1994 • MANIECKI • 02:17 • PARKIET`

**Non-profitowy edytor do robienia absurdalnych visual noveli dla znajomych.** 
Twój Janusz w złotym kimono, kula dysko, kolejka jak za komuny przy barze. Ty to piszesz, silnik to odtwarza.

[Kimono Editor Screenshot](docs/screenshots/editor-1.3-obi-textured.png)

---

### O co chodzi?

**KimonoGM** to monorepo: edytor + silnik + przykładowe gry. Robisz grę w edytorze Wails/Vue, klikasz BUILD i dostajesz EXE lub wersję web na GitHub Pages.

Inspiracja: polskie wesela, lata 90., Janusze, disco, oranżada przy barze i decyzje typu `Zamów oranżadę przy barze` vs `Wbij na parkiet`.

```
Wchodzisz do Manieczek. Kula dysko kręci się nad głową, 
a przy barze kolejka jak za komuny. Janusz w złotym kimono już tańczy.
```

To jest gra którą robisz w 15 minut. Bez kodu.

### Co potrafi edytor 1.3

- **DNI** — struktura gry po dniach (`day1`, `wesele`, `prolog`, `final_boss`). Każdy dzień to lista scen.
- **SCENY** — id, tytuł, tło (webp), tekst narratora, wybory. Drag & drop, strzałki, duplikacja.
- **WYBORY [n]** — tekst + `PRZEJDŹ DO →` scena, opcjonalnie reakcja, SFX, statystyki.
- **STATYSTYKI** — `CEBULA / WSTYD / PORTFEL / REPUTACJA` — customizowalne w projekcie.
- **ASSETY** — BG / RE (reakcje) / SFX / VOICE / MUSIC / TŁA / AVATARY / NARRATOR / MUZYKA — wszystko w jednym miejscu z podglądem.
- **PODGLĄD LIVE** — środkowy canvas to dokładnie to co zobaczy gracz.
- **ZAPIS** — projekt to folder z JSON + assety. Możesz go otworzyć w eksploratorze jednym klikiem.

UI: neon tube na górze, stonowany pas obi na dole z fakturą płótna, VHS vibe. Na przypale albo wcale.

### Stack

| Warstwa | Tech | Po co |
|---|---|---|
| **editor** | Go + Wails + Vue 3 + TypeScript | natywna apka desktop, szybka, bez Electrona |
| **engine** | C# / JanuszSimulator (własny) | odtwarza gry, logika scen |
| **runtime** | Go / WebAssembly | eksport WEB |
| **hosting** | GitHub Pages `/docs` | demo wersji web |

```
KimonoGM/
├── editor/               # edytor Wails (to co widzisz na screenie)
│   ├── frontend/src/components/
│   │   ├── EditorView.vue        # główny layout + OBI 2.2 textured
│   │   ├── DaysPanel.vue
│   │   ├── SceneList.vue
│   │   ├── SceneCanvas.vue       # podgląd gry
│   │   └── ChoicesEditor.vue
│   └── main.go
├── engine/               # silnik gry (C#)
├── games/                # przykładowe projekty (każdy folder = gra)
│   └── test/
├── docs/                 # GitHub Pages + screenshoty
└── README.md
```

### Jak odpalić

**Wymagania:** Go 1.22+, Node 20+, Wails `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

```powershell
# 1. Sklonuj
git clone https://github.com/CapitanPlanet/KimonoGM.git
cd KimonoGM

# 2. Odpal edytor w trybie dev
cd editor
wails dev
# otworzy się okno KIMONO editor 1.0

# 3. Zrób grę
# - DODAJ DZIEŃ -> day1
# - + Scena -> start / bar / parkiet
# - ustaw tło, tekst, wybory
# - ZAPISZ PROJEKT

# 4. Build produkcyjny
wails build
# exe w build/bin/
```

Tworzenie projektu:
- Projekt = folder w `games/` (np. `games/manieczki/`)
- W środku `project.json`, `days/*.json`, `assets/`
- Możesz skopiować `games/test` jako template

### Model danych

```json
{
  "days": ["day1"],
  "scenes": {
    "start": {
      "title": "Manieczki - Wejście",
      "bg": "bg_tutorial.webp",
      "text": "Wchodzisz do Manieczek...",
      "choices": [
        { "text": "Zamów oranżadę przy barze", "next": "bar", "stats": {"CEBULA": 1} },
        { "text": "Wbij na parkiet", "next": "parkiet" }
      ]
    }
  }
}
```

### Roadmap

- [x] **ETAP 1** — Monorepo + Wails + Vue, podstawowy edytor
- [x] **ETAP 2** — Gry jako projekty w `/games`, DNI / SCENY / WYBORY, asset manager
- [x] **ETAP 2.2 OBI** — textured stonowany pas, faktura płótna, przyciski left row, neon tube — **DONE 1.3 RAN FOR JUR LAJF**
- [ ] **ETAP 3** — Global settings zamiast KONFIGURUJ JANUSZA, statystyki custom, theme editor
- [ ] **ETAP 4** — Przycisk BUILD -> EXE / WEB, eksport na GitHub Pages
- [ ] **ETAP 5** — Steam Workshop dla Januszy, multiplayer na parkiecie

### Styl projektu

Nie jesteśmy korpo. Jesteśmy fajtem z Manieczek.

- Fonty: `Bebas Neue` / `JetBrains Mono` / `Inter`
- Kolory: `#0D0B14` bg, `#D4A574` mustard, `#E5E7EB` shiro, `#a89c85` obi beige, `#6e1f28` bordo
- Zasady: VHS, ziarnistość, cieniowanie, na przypale albo wcale, no pain no kiełbasa.

Badge rotacyjny w headerze:
`NO PAIN NO KIEŁBASA / RAN FOR JUR LAJF / FAJT KLUB MANIECKI / ZŁOTE KIMONO TEAM / PARKIET 02:17` — kliknij żeby zmienić.

### Contributing

Projekt non-profit, robiony po godzinach przy piwie. PRy mile widziane, ale trzymamy klimat.

1. Fork
2. Branch `feat/nazwa` lub `fix/nazwa`
3. `git commit -m "feat: ..."`
4. PR na `main` — opis po polsku, może być z humorem

Issues: pisz jak Janusz — co nie działa, gdzie, jaki asset.

### Licencja

Non-profit / MIT-ish — rób gry dla znajomych, nie sprzedawaj Janusza. Jeśli chcesz komercyjnie — dogadamy się przy oranżadzie.

---

**CapitanPlanet / KIMONO TEAM 1994-2026**

> *Wbij na parkiet. Zapisz projekt. Fajrant.*
