<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')">
    <div class="tutorial-wrapper">
      <button class="close-btn" @click="$emit('close')">✕</button>

      <div class="main-container">
        <!-- Elementy dekoracyjne tła -->
        <div class="decor-blob blob-1"></div>
        <div class="decor-blob blob-2"></div>

        <!-- NAGŁÓWEK -->
        <header class="header">
          <div class="top-badges">
            <div class="badge-cyber font-display">
              <span class="dot"></span> Dysko, Judo... GAME MAKER!!!!
            </div>
            <div class="disco-ball-placeholder">🪩</div>
          </div>

          <div class="title-block">
            <div class="title-line-top font-display">JAK ZROBIĆ GRĘ W</div>
            <div class="title-line-middle">
              <span class="title-highlight font-display">KIMONO</span>
              <span class="subtitle font-display">W 7 RZUTACH</span>
              <span class="judo-badge font-display">夫に夕食を作れ KURWA</span>
            </div>
            <div class="no-code-tag">(bez kodowania, serio)</div>
          </div>

          <p class="description">
            Wystarczy inwencja twórcza i trochę
            <span class="belt pill-dark">CEBULI</span> – bo wszystko jest za darmo.
          </p>

          <!-- SEKCJA ZASOBÓW (Zawsze pionowo / wertykalnie) -->
          <div class="zasoby-dodatkowe-panel">
            <div class="zasoby-przyciski">
              <button @click="lightboxOpen = true" class="zasob-link ilustracja-link">
                NA SZYBKO - ZOBACZ ILUSTRACJĘ
              </button>
              <a href="https://www.youtube.com/" target="_blank" class="zasob-link video-link">
                Instrukcja 1 (YT)
              </a>
              <a href="https://www.youtube.com/" target="_blank" class="zasob-link video-link">
                Instrukcja 2 (YT)
              </a>
            </div>
          </div>
        </header>

        <!-- KROKI TUTORIALA -->
        <div class="steps-container">

          <!-- Krok 1 -->
          <section class="step-card">
            <aside class="step-aside">
              <div class="step-number-circle font-display" style="background: var(--gold); color: #000;">
                01
                <span class="step-kanji font-display">尻毛</span>
              </div>
              <div class="step-icon-wrap" style="color: var(--gold);">🔥</div>
            </aside>
            <div class="step-content">
              <div class="step-header-row">
                <div class="step-title-group">
                  <h3 class="font-display step-title"><span class="step-title-num">1.</span>ROZGRZEWKA — DODAJ DZIEŃ
                  </h3>
                  <div class="step-subtitle">"Najgęstszy las to ten, którego nie ma, a najmądrzejszy samuraj tnie
                    powietrze tam, gdzie wróg jeszcze nie pomyślał, żeby się urodzić" - Kien-shitsu</div>
                </div>
                <div class="step-tag font-display">Rzut na parkiet</div>
              </div>
              <p class="step-text">
                Każda gra to rundy. <b>Wesele, kac, poprawiny, dogrywka.</b> Klikasz <span class="pill">+ Dzień</span>.
                Nazwy bez spacji: <span class="code">wesele</span>, <span class="code">prolog</span>, <span
                  class="code">final_ippon</span>.
              </p>
            </div>
          </section>

          <!-- Krok 2 -->
          <section class="step-card">
            <aside class="step-aside">
              <div class="step-number-circle font-display" style="background: var(--purple); color: #fff;">
                02
                <span class="step-kanji font-display">鼻水</span>
              </div>
              <div class="step-icon-wrap" style="color: var(--purple);">🧱</div>
            </aside>
            <div class="step-content">
              <div class="step-header-row">
                <div class="step-title-group">
                  <h3 class="font-display step-title"><span class="step-title-num">2.</span>USTAW MATĘ — DODAJ SCENY
                  </h3>
                  <div class="step-subtitle">"Jeśli woda w dzbanie myśli, że jest winem, to ty jesteś suszoną śliwką." -
                    Jui Kuen</div>
                </div>
              </div>
              <p class="step-text">
                <span class="pill">+ Scena</span> to nowy chwyt. Ustaw <b>TŁO (BG)</b>, <b>TYTUŁ</b>, <b>TEKST</b>. Na
                końcu
                rundy zawsze <span
                  class="inline-flex bg-white text-black rounded-full px-2 py-0.5 font-bold text-[11px]">KONIEC DNIA →
                  END</span>. Inaczej zapętlisz się jak DJ na weselu u kuzyna.
              </p>
            </div>
          </section>

          <!-- Krok 3 -->
          <section class="step-card">
            <aside class="step-aside">
              <div class="step-number-circle font-display" style="background: var(--magenta); color: #fff;">
                03
                <span class="step-kanji font-display">脇毛</span>
              </div>
              <div class="step-icon-wrap" style="color: var(--magenta);">🤘</div>
            </aside>
            <div class="step-content">
              <div class="step-header-row">
                <div class="step-title-group">
                  <h3 class="font-display step-title"><span class="step-title-num">3.</span>TECHNIKA — WYBORY</h3>
                  <div class="step-subtitle">"Karma to bumerang, ale rzucony w ciemnym pokoju przez niewidomego
                    kucharza, który nie lubi kolendry" - Kim Ir Sen</div>
                </div>
                <div class="step-tag font-display">Ippon na WSTYD1</div>
              </div>
              <p class="step-text">
                W panelu <b>WYBORY</b> klikasz <span class="pill">+ Wybór</span>. <i>"Osu! Zamów oranżadę"</i> vs
                <i>"Seoi-nage na parkiet"</i>. Każdy wybór = nowy kierunek rzutu.
              </p>
            </div>
          </section>

          <!-- Krok 4 -->
          <section class="step-card">
            <aside class="step-aside">
              <div class="step-number-circle font-display" style="background: var(--cyan); color: #000;">
                04
                <span class="step-kanji font-display">天然</span>
              </div>
              <div class="step-icon-wrap" style="color: var(--cyan);">⚡️</div>
            </aside>
            <div class="step-content">
              <div class="step-header-row">
                <div class="step-title-group">
                  <h3 class="font-display step-title"><span class="step-title-num">4.</span>KONTRA — REAKCJE I STATS
                  </h3>
                  <div class="step-subtitle">"Człowiek, który nie ma nic do powiedzenia, powinien mówić bardzo głośno,
                    żeby zagłuszyć ciszę, która o nim plotkuje" - Gandhi</div>
                </div>
              </div>
              <p class="step-text">
                Każdy rzut ma <b>Kontrę</b>. Zmieniaj <b>Statystyki jako pasy (KONFIGURUJ GLOBALNIE)</b>: w okienku
                możesz wpisać nazwę statystyk i przypiąć pod avatara.
              </p>
            </div>
          </section>

          <!-- Krok 5 -->
          <section class="step-card">
            <aside class="step-aside">
              <div class="step-number-circle font-display" style="background: var(--gold); color: #000;">
                05
                <span class="step-kanji font-display">衣</span>
              </div>
              <div class="step-icon-wrap" style="color: var(--gold);">✨</div>
            </aside>
            <div class="step-content">
              <div class="step-header-row">
                <div class="step-title-group">
                  <h3 class="font-display step-title"><span class="step-title-num">5.</span>KIMONO Z AI — ASSETY</h3>
                  <div class="step-subtitle">"Kto patrzy w przepaść zbyt długo, ten w końcu zauważy, że przepaść też ma
                    dość patrzenia na jego krzywe zęby i pójdzie zrobić sobie herbaty" - Shang Tsung</div>
                </div>
              </div>
              <p class="step-text">
                Nie masz grafika? <b>TŁA (BG), AVATARY w judogach, REAKCJE</b> generuj w AI. Wrzuć <span
                  class="code">.webp</span> przez assety.
              </p>
            </div>
          </section>

          <!-- Krok 6 -->
          <section class="step-card">
            <aside class="step-aside">
              <div class="step-number-circle font-display" style="background: var(--magenta); color: #fff;">
                06
                <span class="step-kanji font-display">お尻</span>
              </div>
              <div class="step-icon-wrap" style="color: var(--magenta);">🎵</div>
            </aside>
            <div class="step-content">
              <div class="step-header-row">
                <div class="step-title-group">
                  <h3 class="font-display step-title"><span class="step-title-num">6.</span>MUZYKA I OKRZYK — AUDIO</h3>
                  <div class="step-subtitle">"Stare, Chińskie przysłowie mówi: jeśli nie wiesz co powiedzieć to powiedz stare Chińskie przysłowie" - Nieznany Chiński mędrzec</div>
                </div>
              </div>
              <p class="step-text">
                Wgraj <b>MUZYKĘ</b> (disco polo może być, ale uważaj na prawa autorskie), <b>SFX</b> rzut na matę,
                brzdęk cekinów, przekleństwa - obojętnie.
              </p>
            </div>
          </section>

          <!-- Krok 7 -->
          <section class="step-card">
            <aside class="step-aside">
              <div class="step-number-circle font-display"
                style="background: var(--white); color: #000; border: 2px solid var(--black);">
                07
                <span class="step-kanji font-display">オナ</span>
              </div>
              <div class="step-icon-wrap" style="color: var(--black);">👑</div>
            </aside>
            <div class="step-content">
              <div class="step-header-row">
                <div class="step-title-group">
                  <h3 class="font-display step-title"><span class="step-title-num">7.</span>IPPON! — PUBLIKUJ</h3>
                  <div class="step-subtitle">"Lepiej wypić sake z przyjacielem niż wodę z kałuży z wrogiem" - Generał ZZZXXZX (czy
                    jakoś tak)</div>
                </div>
              </div>
              <p class="step-text">
                Klikasz <span class="pill-dark pill">ZAPISZ</span>, potem <span class="pill">WEB-BUILD</span> lub <span
                  class="pill">EXE</span>. Zero kodu. Czysty ippon.
              </p>
            </div>
          </section>

        </div>

        <!-- STOPKA -->
        <footer class="footer-cta">
          <div class="master-belt">
            <div class="belt-white-side"></div>
            <div class="belt-red-side"></div>
          </div>
          <div class="cta-content">
            <h2 class="font-display cta-title">
              NIE TRZEBA KODOWAĆ.<br>
              WYSTARCZY INWENCJA<br>
              I DOBRY RZUT.<br>
              <span class="inline-block mt-2 bg-black text-[#F5D67B] rounded-full px-4 py-1 text-[26px] -rotate-1">NO
                PAIN, NO KIEŁBASA.</span>
            </h2>
            <div class="cta-details">
              <div class="tag-editor">KIMONO editor • DISCO DOJO</div>
              <div class="cta-text-small">
                <b>Disco Dojo 1986-2026</b> — Visual Novel z miłością do tatami, cekinów i promocji w Biedronce.
              </div>
            </div>
          </div>
          <div class="bottom-bar">
            <span>00:34:12 / DISCO DOJO — KIMONO_TUTORIAL_FINAL.mp4</span>
            <span style="color: var(--gold);">● HAJIME READY</span>
          </div>
        </footer>

      </div>
    </div>
  </div>

  <!-- LIGHTBOX jak na FB - duży obrazek na cały ekran -->
  <div v-if="lightboxOpen" ref="lightboxOverlay" class="lightbox-overlay" :class="{ 'lightbox-overlay--zoomed': lightboxZoomed, 'lightbox-overlay--dragging': isDragging }" @click.self="closeLightbox" @mousedown="startDrag" @mousemove="onDrag" @mouseup="endDrag" @mouseleave="endDrag">
    <button class="lightbox-close" @click="closeLightbox">✕</button>
    <img :src="tutorialImg" class="lightbox-img" :class="{ 'lightbox-img--zoomed': lightboxZoomed }" :style="{ transformOrigin: zoomOrigin }" alt="Tutorial - ilustracja" @click.stop="toggleZoom($event)" @dragstart.prevent :title="lightboxZoomed ? 'Kliknij żeby oddalić, przeciągnij żeby przesunąć' : 'Kliknij żeby powiększyć'" />
    <div class="lightbox-hint">Esc / klik poza obrazek = zamknij &nbsp;•&nbsp; Klik w obrazek = zoom &nbsp;•&nbsp; Przeciągnij = przesuwanie</div>
  </div>
</template>

<script>
import tutorialImg from '../assets/tutorial_basic.jpg'

export default {
  props: {
    isOpen: {
      type: Boolean,
      required: true
    }
  },
  data() {
    return {
      tutorialImg,
      lightboxOpen: false,
      lightboxZoomed: false,
      zoomOrigin: '50% 50%',
      isDragging: false,
      dragStartX: 0,
      dragStartY: 0,
      dragScrollLeft: 0,
      dragScrollTop: 0,
      dragged: false
    }
  },
  watch: {
    lightboxOpen(open) {
      if (open) {
        window.addEventListener('keydown', this.onLightboxKey)
      } else {
        window.removeEventListener('keydown', this.onLightboxKey)
      }
    }
  },
  methods: {
    closeLightbox() {
      this.lightboxOpen = false
      this.lightboxZoomed = false
      this.zoomOrigin = '50% 50%'
    },
    toggleZoom(e) {
      if (this.dragged) { this.dragged = false; return }
      if (!this.lightboxZoomed) {
        // zoom DOKLADNIE w miejsce klikniecia
        const rect = e.target.getBoundingClientRect()
        const x = ((e.clientX - rect.left) / rect.width) * 100
        const y = ((e.clientY - rect.top) / rect.height) * 100
        this.zoomOrigin = x.toFixed(1) + '% ' + y.toFixed(1) + '%'
        this.lightboxZoomed = true
      } else {
        this.lightboxZoomed = false
      }
    },
    startDrag(e) {
      if (!this.lightboxZoomed) return
      const el = this.$refs.lightboxOverlay
      if (!el) return
      this.isDragging = true
      this.dragged = false
      this.dragStartX = e.clientX
      this.dragStartY = e.clientY
      this.dragScrollLeft = el.scrollLeft
      this.dragScrollTop = el.scrollTop
    },
    onDrag(e) {
      if (!this.isDragging) return
      const el = this.$refs.lightboxOverlay
      if (!el) return
      const dx = e.clientX - this.dragStartX
      const dy = e.clientY - this.dragStartY
      if (Math.abs(dx) > 4 || Math.abs(dy) > 4) this.dragged = true
      el.scrollLeft = this.dragScrollLeft - dx
      el.scrollTop = this.dragScrollTop - dy
    },
    endDrag() {
      this.isDragging = false
    },
    onLightboxKey(e) {
      if (e.key === 'Escape') this.closeLightbox()
    }
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.onLightboxKey)
  }
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;700&family=Syne:wght@700;800&display=swap');

.main-container {
  --bg-color: #050408;
  --card-bg: #0b0a12;
  --text-main: #f5f1ea;
  --text-muted: rgba(245, 241, 234, 0.65);
  --magenta: #d4238e;
  --purple: #6222cc;
  --gold: #d4b563;
  --cyan: #22cccc;
  --white: #ffffff;
  --black: #000000;
  --brand-gradient: linear-gradient(135deg, var(--gold), var(--magenta));
  --belt-gradient: linear-gradient(90deg, var(--white) 0%, var(--gold) 50%, var(--white) 100%);

  position: relative;
  width: 100%;
  max-width: 880px;
  background-color: var(--card-bg);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 40px;
  overflow: hidden;
  box-shadow: 0 40px 120px -20px rgba(0, 0, 0, 0.7);
  font-family: 'Space Grotesk', sans-serif;
  color: var(--text-main);
  text-align: left;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: #040306;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  z-index: 9999;
  overflow-y: auto;
  padding: 40px 20px;
}

.tutorial-img {
  width: 100%;
  max-width: 100%;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  margin-top: 15px;
}

.tutorial-wrapper {
  position: relative;
  max-width: 880px;
  width: 100%;
}

.close-btn {
  position: absolute;
  top: 15px;
  right: 20px;
  background: #1a1724;
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 50%;
  width: 36px;
  height: 36px;
  font-size: 18px;
  font-weight: bold;
  cursor: pointer;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.5);
}

.font-display {
  font-family: 'Syne', sans-serif;
  font-weight: 800;
}

.pill {
  display: inline-flex;
  align-items: center;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 2px 10px;
  font-size: 11px;
  font-weight: 700;
  color: var(--gold);
  border-radius: 9999px;
  margin: 0 2px;
}

.pill-dark {
  background: var(--black);
  color: var(--gold);
}

.code {
  display: inline-block;
  background: var(--black);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 1px 6px;
  font-size: 11px;
  color: var(--cyan);
  border-radius: 9999px;
  font-family: monospace;
}

.belt {
  display: inline-block;
  background: var(--gold);
  color: var(--black);
  padding: 0 8px;
  border-radius: 9999px;
  font-weight: 700;
  font-size: 11px;
}

.decor-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px);
  pointer-events: none;
  opacity: 0.15;
}

.blob-1 {
  top: -10%;
  left: -10%;
  width: 60%;
  height: 60%;
  background: var(--purple);
}

.blob-2 {
  top: 10%;
  right: -10%;
  width: 50%;
  height: 50%;
  background: var(--gold);
}

.header {
  position: relative;
  padding: 50px 50px 30px;
  z-index: 1;
}

.top-badges {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.disco-ball-placeholder {
  width: 70px;
  height: 70px;
  background: #15131d;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  box-shadow: 0 0 15px rgba(0, 0, 0, 0.5);
}

.badge-cyber {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  padding: 4px 12px;
  border-radius: 9999px;
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  color: rgba(255, 255, 255, 0.5);
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--gold);
}

.title-block {
  margin-bottom: 15px;
}

.title-line-top {
  font-size: 38px;
  line-height: 1;
  letter-spacing: -0.02em;
  color: #ffffff;
  margin-bottom: 5px;
}

.title-line-middle {
  display: flex;
  align-items: baseline;
  gap: 15px;
  flex-wrap: wrap;
}

.title-highlight {
  background: var(--brand-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-size: 56px;
  font-weight: 800;
  line-height: 1;
}

.subtitle {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.14em;
  color: rgba(255, 255, 255, 0.5);
  text-transform: uppercase;
}

.judo-badge {
  background: #1a1724;
  color: var(--gold);
  border: 1px solid rgba(212, 181, 99, 0.2);
  font-size: 11px;
  font-weight: 800;
  padding: 2px 8px;
  border-radius: 9999px;
  transform: rotate(3deg);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.3);
}

.no-code-tag {
  font-size: 14px;
  font-family: sans-serif;
  color: rgba(255, 255, 255, 0.35);
  font-weight: 400;
  margin-top: 6px;
}

.description {
  font-size: 15px;
  color: var(--text-muted);
  max-width: 600px;
  margin-top: 15px;
}

/* SEKCJA ZASOBÓW - Sztywno wertykalnie (jeden pod drugim) */
.zasoby-dodatkowe-panel {
  background-color: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  padding: 15px;
  margin-top: 20px;
  max-width: 380px;
}

.zasoby-przyciski {
  display: flex;
  flex-direction: column;
  /* Wertykalnie */
  gap: 8px;
}

.zasob-link {
  display: flex;
  align-items: center;
  text-decoration: none;
  background-color: rgba(255, 255, 255, 0.04);
  color: #fff;
  padding: 8px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  transition: background-color 0.2s;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.zasob-link:hover {
  background-color: rgba(255, 255, 255, 0.08);
}

.video-link:hover {
  border-color: #d32f2f;
}

.ilustracja-link:hover {
  border-color: #1976d2;
}

.steps-container {
  padding: 0 30px 40px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
  z-index: 1;
}

.step-card {
  background: rgba(255, 255, 255, 0.015);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 20px;
  display: flex;
  overflow: hidden;
  transition: all 0.2s ease;
}

.step-card:hover {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.1);
}

.step-aside {
  width: 75px;
  flex-shrink: 0;
  background: rgba(0, 0, 0, 0.2);
  border-right: 1px solid rgba(255, 255, 255, 0.03);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 0;
}

.step-number-circle {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 800;
  position: relative;
}

.step-kanji {
  position: absolute;
  bottom: -4px;
  right: -4px;
  background: #0b0a12;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 50%;
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 7px;
  color: rgba(255, 255, 255, 0.5);
}

.step-icon-wrap {
  margin-top: 15px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.03);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.step-content {
  flex: 1;
  padding: 20px 25px;
}

.step-header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 4px;
}

.step-title-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

/* Podpis blisko tytułu */
.step-title {
  font-size: 16px;
  font-weight: 800;
  color: var(--white);
  margin-bottom: 0px;
}

.step-title-num {
  opacity: 0.3;
  margin-right: 5px;
  font-weight: 600;
}

.step-subtitle {
  font-size: 10px;
  letter-spacing: 0.15em;
  color: rgba(255, 255, 255, 0.4);
  text-transform: uppercase;
  margin-bottom: 10px;
}

.step-tag {
  background: rgba(255, 255, 255, 0.06);
  color: var(--white);
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 9999px;
  white-space: nowrap;
}

.step-text {
  font-size: 13px;
  color: rgba(245, 241, 234, 0.65);
  margin-top: 8px;
  line-height: 1.5;
}

.footer-cta {
  position: relative;
  margin-top: 20px;
}

.master-belt {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 14px;
  display: flex;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  z-index: 2;
}

.belt-white-side {
  flex: 1;
  background: var(--belt-gradient);
}

.belt-red-side {
  flex: 1;
  background: linear-gradient(90deg, #b71c1c, #8e0000);
}

.cta-content {
  background: linear-gradient(135deg, #12101a, #0b0a12);
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  color: var(--text-main);
  padding: 40px 40px 30px;
  border-radius: 0 0 40px 40px;
  position: relative;
  z-index: 1;
  margin-top: 7px;
}

.cta-title {
  font-size: 26px;
  line-height: 1.1;
  letter-spacing: -0.01em;
  color: var(--white);
}

.cta-details {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-top: 25px;
  flex-wrap: wrap;
}

.tag-editor {
  background: var(--black);
  color: var(--gold);
  border: 1px solid rgba(212, 181, 99, 0.2);
  padding: 4px 12px;
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 700;
}

.cta-text-small {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.4);
  max-width: 45ch;
}

.bottom-bar {
  background: var(--black);
  color: rgba(255, 255, 255, 0.25);
  font-size: 9px;
  padding: 10px 30px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  display: flex;
  justify-content: space-between;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

/* LIGHTBOX */
.ilustracja-link {
  cursor: pointer;
  width: 100%;
  text-align: left;
  font-family: inherit;
}

.lightbox-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.92);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  padding: 40px;
  cursor: zoom-out;
  animation: lightboxFade 0.2s ease;
}

@keyframes lightboxFade {
  from { opacity: 0; }
  to { opacity: 1; }
}

.lightbox-overlay--zoomed {
  overflow: auto;
  align-items: flex-start;
  justify-content: flex-start;
  cursor: grab;
  user-select: none;
  -webkit-user-select: none;
}

.lightbox-overlay--dragging {
  cursor: grabbing;
}

.lightbox-overlay--zoomed .lightbox-img {
  margin: auto;
  /* zeby dało sie zescrollowac do kazdej krawedzi */
  max-width: none;
}

.lightbox-img {
  transition: transform 0.2s ease;
  cursor: zoom-in;
  max-width: 92vw;
  max-height: 88vh;
  width: auto;
  height: auto;
  object-fit: contain;
  border-radius: 12px;
  box-shadow: 0 20px 80px rgba(0, 0, 0, 0.8);
  cursor: default;
}

.lightbox-img--zoomed {
  max-width: none;
  max-height: none;
  width: auto;
  height: auto;
  transform: scale(2.2);
  cursor: zoom-out;
}

.lightbox-close {
  position: absolute;
  top: 20px;
  right: 20px;
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  width: 44px;
  height: 44px;
  font-size: 20px;
  cursor: pointer;
  z-index: 1;
}

.lightbox-close:hover {
  background: rgba(255, 255, 255, 0.25);
}

.lightbox-hint {
  position: absolute;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
  white-space: nowrap;
}
</style>