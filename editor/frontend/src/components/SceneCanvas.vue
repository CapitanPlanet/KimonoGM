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
/* WRAPPER - mobile first */
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
  max-width: 480px; /* max jak telefon */
  aspect-ratio: 9/16; /* DEFAULT NA TELEFON */
  background-size: cover; /* FIX NR 1 - było contain i robiło pasy */
  background-repeat: no-repeat;
  background-position: center 30%; /* lekko do góry żeby kula dysko była widoczna */
  background-color: #000;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 
    0 0 0 1px rgba(255,45,149,0.2),
    0 20px 60px rgba(0,0,0,0.8),
    0 0 40px rgba(255,45,149,0.15);
  border: 1px solid rgba(255,255,255,0.08);
}

/* DESKTOP - szerszy */
@media (min-width: 1024px) {
  .preview-stage {
    aspect-ratio: 16/10;
    max-width: 100%;
    border-radius: 12px;
  }
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

/* TOP SAFE ZONE - 22% na avatary */
.top-safe-zone {
  position: relative;
  z-index: 3;
  height: 22%;
  background: linear-gradient(180deg, rgba(0,0,0,0.7) 0%, transparent 100%);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 16px;
  pointer-events: none;
}

.avatar-placeholder {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.avatar-label {
  font-size: 9px;
  letter-spacing: 0.2em;
  color: rgba(255,255,255,0.25);
  font-weight: 700;
}

.avatar-glow {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  border: 1px dashed rgba(255,45,149,0.3);
  background: radial-gradient(circle, rgba(255,45,149,0.08) 0%, transparent 70%);
}

/* VHS / PAZUR */
.vhs-overlay {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
  opacity: 0.04;
  background-image: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    white 2px,
    white 3px
  );
}

.vignette {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
  background: radial-gradient(ellipse at center, transparent 55%, rgba(0,0,0,0.5) 100%);
}

/* BOTTOM */
.bottom-zone {
  position: relative;
  z-index: 4;
  margin-top: auto;
}

.preview-text {
  background: linear-gradient(0deg, rgba(0,0,0,0.95) 0%, rgba(0,0,0,0.7) 60%, transparent 100%);
  padding: 20px 20px 14px 20px;
  border-top: 1px solid rgba(255,45,149,0.15);
}

.preview-text h2 {
  color: #ff2d95;
  font-size: 11px;
  margin: 0 0 6px 0;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  text-shadow: 0 0 12px rgba(255,45,149,0.6);
  font-weight: 800;
}

.preview-text p {
  color: #fff;
  font-size: 14px;
  line-height: 1.5;
  margin: 0;
  white-space: pre-wrap;
  text-shadow: 0 1px 4px rgba(0,0,0,1);
  font-weight: 500;
}

/* REAKCJA */
.reaction-overlay {
  position: absolute;
  top: 46%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 10;
  background: rgba(12,12,18,0.92);
  border: 1px solid rgba(255,45,149,0.4);
  border-radius: 12px;
  padding: 10px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.9), 0 0 20px rgba(255,45,149,0.2);
  max-width: 72%;
  backdrop-filter: blur(8px);
}

.reaction-overlay img {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
  border-radius: 8px;
  display: block;
}

.reaction-label {
  font-size: 10px;
  color: #8a8a9a;
  font-family: monospace;
  margin-top: 8px;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reaction-enter-active,
.reaction-leave-active {
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.reaction-enter-from,
.reaction-leave-to {
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.85);
}

/* WYBORY - neon disco */
.preview-choices {
  background: rgba(8,8,10,0.92);
  border-top: 1px solid rgba(255,255,255,0.08);
  padding: 10px;
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  max-height: 34%;
  overflow-y: auto;
  backdrop-filter: blur(12px);
}

@media (min-width: 768px) {
  .preview-choices {
    grid-template-columns: 1fr 1fr;
  }
  .choice-btn:last-child:nth-child(odd) {
    grid-column: 1 / -1;
  }
}

.preview-choices::-webkit-scrollbar { width: 4px; }
.preview-choices::-webkit-scrollbar-thumb { background: #ff2d95; border-radius: 4px; }

.choice-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 12px;
  background: linear-gradient(180deg, #1f1f2a 0%, #15151e 100%);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  color: #b8b8c8;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.18s;
  text-align: left;
  position: relative;
  overflow: hidden;
}

.choice-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(255,45,149,0.08), transparent);
  transform: translateX(-100%);
  transition: transform 0.6s;
}

.choice-btn:hover::before {
  transform: translateX(100%);
}

.choice-btn:hover {
  background: linear-gradient(180deg, #2a2a3a 0%, #1f1f2a 100%);
  border-color: rgba(255,45,149,0.5);
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.5), 0 0 12px rgba(255,45,149,0.25);
}

.choice-btn.active {
  border-color: #ff2d95;
  background: linear-gradient(180deg, #2a1f2e 0%, #1f151e 100%);
}

.choice-btn.has-reaction {
  border-left: 2px solid #ff2d95;
}

.choice-num {
  width: 22px;
  height: 22px;
  background: rgba(0,0,0,0.6);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-size: 11px;
  font-weight: 800;
  flex-shrink: 0;
  color: #ff2d95;
}

.choice-btn:hover .choice-num {
  background: #ff2d95;
  color: #000;
  border-color: #ff2d95;
  box-shadow: 0 0 8px rgba(255,45,149,0.5);
}

.choice-text {
  flex: 1;
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.choice-reaction-icon {
  font-size: 14px;
  opacity: 0.7;
  flex-shrink: 0;
}

.no-scene {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #5a5a6a;
  font-size: 13px;
  letter-spacing: 0.05em;
}
</style>
