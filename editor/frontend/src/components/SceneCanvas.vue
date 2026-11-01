<script setup>
import { ref, computed, watch } from 'vue'
import { useProjectStore } from '../stores/projectStore'
import { GetImageBase64 } from '../../wailsjs/go/main/App'

const store = useProjectStore()
const bgCache = ref({})
const reactionCache = ref({})
const hoveredChoice = ref(null)

const bgUrl = computed(() => {
  const bg = store.currentScene?.Background
  if (!bg || !store.projectPath) return ''
  const key = bg.trim()
  return bgCache.value[key] || bgCache.value[bg] || ''
})

const choices = computed(() => {
  return store.currentScene?.Choices || []
})

const hoveredReactionUrl = computed(() => {
  const choice = hoveredChoice.value
  if (!choice?.ReactionImage) return ''
  const key = choice.ReactionImage.trim()
  return reactionCache.value[key] || reactionCache.value[choice.ReactionImage] || ''
})

function cleanPath(p) {
  if (!p) return ''
  return p.trim()
}

watch(() => store.currentScene?.Background, async (newBg) => {
  const clean = cleanPath(newBg)
  if (!clean || !store.projectPath) return
  if (bgCache.value[clean]) return
  try {
    const b64 = await GetImageBase64(clean)
    bgCache.value[clean] = b64
    bgCache.value[newBg] = b64
  } catch (e) {
    console.warn('[PREVIEW] Brak tła:', clean, e)
    bgCache.value[clean] = ''
    bgCache.value[newBg] = ''
  }
}, { immediate: true })

watch(choices, async (newChoices) => {
  if (!newChoices?.length || !store.projectPath) return
  for (const choice of newChoices) {
    const raw = cleanPath(choice.ReactionImage)
    if (!raw) continue
    if (reactionCache.value[raw]) continue
    try {
      const b64 = await GetImageBase64(raw)
      reactionCache.value[raw] = b64
      reactionCache.value[choice.ReactionImage] = b64
    } catch (e) {
      console.warn('[PREVIEW] Brak reakcji:', raw, e)
      reactionCache.value[raw] = ''
    }
  }
}, { immediate: true, deep: true })
</script>

<template>
  <div class="scene-canvas-wrapper" v-if="store.currentScene">
    <div class="preview-stage" :style="{ backgroundImage: bgUrl ? `url(${bgUrl})` : 'none' }">
      <div v-if="!bgUrl" class="no-bg">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
          <circle cx="8.5" cy="8.5" r="1.5"></circle>
          <polyline points="21 15 16 10 5 21"></polyline>
        </svg>
        <span>Brak tła - wrzuć bg_tutorial.webp</span>
      </div>

      <!-- TOP SAFE ZONE - tu będą avatary -->
      <div class="top-safe-zone">
        <div class="avatar-placeholder" v-if="!hoveredReactionUrl">
          <span class="avatar-label">AVATAR / REAKCJA</span>
          <div class="avatar-glow"></div>
        </div>
      </div>

      <!-- PAZUR: VHS grain + vignette -->
      <div class="vhs-overlay"></div>
      <div class="vignette"></div>

      <!-- REAKCJA HOVER - środek/prawo -->
      <transition name="reaction">
        <div v-if="hoveredReactionUrl" class="reaction-overlay">
          <img :src="hoveredReactionUrl" alt="reakcja" />
          <div class="reaction-label">{{ hoveredChoice?.ReactionImage?.replace('images/re_', '').replace('images/', '') }}</div>
        </div>
      </transition>

      <!-- DÓŁ: TEKST + WYBORY -->
      <div class="bottom-zone">
        <div class="preview-text">
          <h2>{{ store.currentScene.SceneTitle }}</h2>
          <p>{{ store.currentScene.Text }}</p>
        </div>

        <div v-if="choices.length" class="preview-choices">
          <div
            v-for="(choice, idx) in choices"
            :key="choice.id || idx"
            class="choice-btn"
            @mouseenter="hoveredChoice = choice"
            @mouseleave="hoveredChoice = null"
            :class="{ 'has-reaction': choice.ReactionImage, 'active': hoveredChoice === choice }"
          >
            <span class="choice-num">{{ idx + 1 }}</span>
            <span class="choice-text">{{ choice.Text || 'Pusty wybór' }}</span>
            <span v-if="choice.ReactionImage" class="choice-reaction-icon">🪩</span>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div v-else class="no-scene">
    Wybierz scenę z listy
  </div>
</template>

<style scoped>
/* ==========================================================================
   ODCHUDZONY I ZREFRAKTOROWANY CSS DLA SCENECANVAS (DESKTOP DRIVEN)
   ========================================================================== */
.scene-canvas-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background: #08080a;
  padding: 12px;
}

.preview-stage {
  position: relative;
  width: 100%;
  height: 100%; /* Usuwamy sztywne aspect-ratio, scena bierze 100% wysokości kontenera */
  background-size: cover;
  background-repeat: no-repeat;
  background-position: center 30%;
  background-color: #000;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 
    0 0 0 1px rgba(255,45,149,0.15),
    0 15px 45px rgba(0,0,0,0.8); /* Odchudzona poświata, która nie marnuje przestrzeni */
}

.no-bg {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #5a5a6a;
  gap: 12px;
  background: radial-gradient(ellipse at center, #1a1a23 0%, #08080a 100%);
}

/* TOP SAFE ZONE - 18% (Zmniejszone, by zyskać przestrzeń na środku) */
.top-safe-zone {
  position: relative;
  z-index: 3;
  height: 18%;
  background: linear-gradient(180deg, rgba(0,0,0,0.8) 0%, transparent 100%);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 12px;
  pointer-events: none;
}

.avatar-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.avatar-label {
  font-size: 9px;
  letter-spacing: 0.2em;
  color: rgba(255,255,255,0.25);
  font-weight: 700;
}

.avatar-glow {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  border: 1px dashed rgba(255,45,149,0.3);
  background: radial-gradient(circle, rgba(255,45,149,0.08) 0%, transparent 70%);
}

/* KOSMETYKA VHS */
.vhs-overlay {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
  opacity: 0.04;
  background-image: repeating-linear-gradient(0deg, transparent, transparent 2px, white 2px, white 3px);
}

.vignette {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
  background: radial-gradient(ellipse at center, transparent 55%, rgba(0,0,0,0.5) 100%);
}

/* DOLNA SEKCJA - TEKST + WYBORY */
.bottom-zone {
  position: relative;
  z-index: 4;
  margin-top: auto;
  display: flex;
  flex-direction: column;
}

.preview-text {
  background: linear-gradient(0deg, rgba(0,0,0,0.95) 0%, rgba(0,0,0,0.7) 70%, transparent 100%);
  padding: 16px 20px 12px 20px;
  border-top: 1px solid rgba(255,45,149,0.15);
}

.preview-text h2 {
  color: #ff2d95;
  font-size: 11px;
  margin: 0 0 4px 0;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  text-shadow: 0 0 12px rgba(255,45,149,0.6);
  font-weight: 800;
}

.preview-text p {
  color: #fff;
  font-size: 13px;
  line-height: 1.4;
  margin: 0;
  white-space: pre-wrap;
  font-weight: 500;
}

/* INTERAKTYWNE WYBORY */
.preview-choices {
  background: rgba(8,8,10,0.95);
  border-top: 1px solid rgba(255,255,255,0.08);
  padding: 10px;
  display: grid;
  grid-template-columns: 1fr 1fr; /* Zawsze dwie kolumny na desktopie */
  gap: 8px;
  max-height: 160px;
  overflow-y: auto;
}

.choice-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: linear-gradient(180deg, #1f1f2a 0%, #15151e 100%);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 6px;
  color: #b8b8c8;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
  text-align: left;
}

.choice-btn:hover {
  background: linear-gradient(180deg, #2a2a3a 0%, #1f1f2a 100%);
  border-color: rgba(255,45,149,0.5);
  color: #fff;
  box-shadow: 0 0 12px rgba(255,45,149,0.2);
}

.choice-num {
  width: 20px;
  height: 20px;
  background: rgba(0,0,0,0.6);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-size: 10px;
  font-weight: 800;
  flex-shrink: 0;
  color: #ff2d95;
}

.choice-btn:hover .choice-num {
  background: #ff2d95;
  color: #000;
  border-color: #ff2d95;
}

.choice-text {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* REAKCJA I WARUNKI PUSTE */
.reaction-overlay {
  position: absolute;
  top: 45%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 10;
  background: rgba(12,12,18,0.92);
  border: 1px solid rgba(255,45,149,0.4);
  border-radius: 8px;
  padding: 8px;
  max-width: 60%;
}
.reaction-overlay img { width: 100%; aspect-ratio: 1; object-fit: cover; border-radius: 4px; display: block; }
.reaction-label { font-size: 9px; color: #8a8a9a; margin-top: 6px; text-align: center; }

.no-scene { display: flex; align-items: center; justify-content: center; height: 100%; color: #5a5a6a; font-size: 12px; }

</style>
