<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import DaysPanel from './DaysPanel.vue'
import SceneList from './SceneList.vue'
import SceneCanvas from './SceneCanvas.vue'
import ChoicesEditor from './ChoicesEditor.vue'
import JanuszModal from './JanuszModal.vue'
import { useProjectStore } from '../stores/projectStore'
import { OpenProjectFolder } from '../../wailsjs/go/main/App'
import tutorialBasic from '../assets/tutorial_basic.jpg'

const store = useProjectStore()
const emit = defineEmits(['go-to-menu'])
const showSavedToast = ref(false)
const showTutorial = ref(false)

const isLoaded = computed(() => !!store.meta && !!store.projectPath)
const rulesCount = computed(() => store.avatarSystem?.rules?.length ?? 0)

const absurdMottos = [
  "RAN FOR JUR LAJF",
  "FAJT KLUB MANIECKI",
  "KIMONO Z BAZARU • 50 ZŁ",
  "SENSEJ JANUSZ • 3 DAN",
  "DOJO ZA REMIZĄ",
  "KARATE KID Z WOLI",
  "ZŁOTE KIMONO • CZARNY PAS",
  "NO PAIN NO KIEŁBASA"
]
const motto = ref(absurdMottos[0])
function losujMotto(){ motto.value = absurdMottos[Math.floor(Math.random()*absurdMottos.length)] }

onMounted(() => {
  if (!(store as any).currentDayId && (store as any).days?.length) {
    ;(store as any).currentDayId = (store as any).days[0].id
  }
  losujMotto()
})

async function saveProject() {
  try {
    await store.saveProject()
    showSavedToast.value = true
    setTimeout(() => { showSavedToast.value = false }, 2000)
  } catch (e) { alert('Błąd zapisu: ' + e) }
}
async function openFolder() {
  if (!store.projectPath) return
  try { await OpenProjectFolder(store.projectPath) } catch (e) { console.warn(e) }
}
</script>

<template>
  <div class="editor">
    <header class="top-bar-neon" @click="losujMotto">
      <div class="neon-tube"></div>
      
      <div class="top-left">
        <!-- ABSURD BADGE -->
        <div class="absurd-badge" :title="'Kliknij żeby zmienić motto'">
          <div class="absurd-top">SEKCJA</div>
          <div class="absurd-main">{{ motto }}</div>
          <div class="absurd-bottom"><span class="belt-line"></span></div>
        </div>
        <div class="brand">
          <div class="kimono-title">KIMONO<span class="ver">editor 1.0</span><span class="vhs">VHS • 1994</span></div>
          <div class="project-line">
            <span class="proj-name">{{ store.meta?.gameName || 'test' }}</span>
            <span class="dot">•</span>
            <span class="path clickable" @click.stop="openFolder">{{ store.projectPath || '...' }}</span>
          </div>
        </div>
      </div>

      <div class="top-center">
        <div class="fight-badge"><span class="neon-dot"></span> MANIECKI • 02:17 • PARKIET</div>
      </div>

      <div class="top-right">
        <button @click.stop="store.ui.showJanuszModal = true" class="btn-janusz-neon" :class="{ empty: rulesCount===0 }">🥋 JANUSZ [{{ rulesCount }}]</button>
      </div>
    </header>

    <div v-if="isLoaded" class="main-grid">
      <DaysPanel class="panel" />
      <SceneList class="panel" />
      <SceneCanvas class="panel panel-preview" />
      <ChoicesEditor class="panel" />
    </div>
    <div v-else class="loading">Ładowanie projektu... {{ store.projectPath || '' }}</div>

    <div class="bottom-bar kohaku-obi">
      <div class="obi-vignette"></div>
      <div class="obi-h5">
        <span class="h-line"></span><span class="h-line"></span><span class="h-line mid"></span><span class="h-line"></span><span class="h-line"></span>
      </div>
      <div class="obi-3szy">
        <div class="szy"><i></i></div><div class="szy"><i></i></div><div class="szy"><i></i></div>
      </div>
      <div class="obi-content">
        <button @click="saveProject" class="btn-footer save"><span>💾</span> Zapisz projekt</button>
        <button @click="emit('go-to-menu')" class="btn-footer menu">📁 Menu</button>
        <button @click="showTutorial = true" class="btn-footer tutorial"><span>📖</span> Pokaż Tutorial</button>
      </div>
    </div>

    <transition name="toast"><div v-if="showSavedToast" class="toast-saved">✓ Zapisano</div></transition>

    <div v-if="showTutorial" class="modal-overlay" @click.self="showTutorial = false">
      <div class="modal-content tutorial-content">
        <button class="close-btn" @click="showTutorial = false">✕</button>
        <img :src="tutorialBasic" alt="Instrukcja" class="tutorial-img" />
      </div>
    </div>

    <JanuszModal />
  </div>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;700&display=swap');
:root {
  --bg:#0D0B14; --panel:#16131F; --card:#1C1929; --input:#151221; --border:#2A2440;
  --disco:#7C3AED; --mustard:#D4A574; --mustard-bright:#E7C36A;
  --neon-pink:#FF2E93; --shiro:#E5E7EB; --text:#E6EDF3; --text-dim:#9CA3AF;
  --accent: var(--mustard);
}
* { box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:var(--bg); margin:0; color:var(--text); overflow-x:hidden; }
::-webkit-scrollbar{ width:8px; height:8px; }
::-webkit-scrollbar-track{ background:#0D0B14; }
::-webkit-scrollbar-thumb{ background: linear-gradient(180deg, #D4A574, #8e1b26); border-radius:4px; }
</style>

<style scoped>
.editor { height:100vh; display:flex; flex-direction:column; background:var(--bg); }
.main-grid { display:grid; grid-template-columns:280px 280px minmax(400px,1fr) 380px; gap:12px; padding:12px; flex:1; overflow:hidden; min-height:0; padding-bottom:72px; min-width:1376px; }
.loading { flex:1; display:flex; align-items:center; justify-content:center; color:var(--text-dim); font-family:'JetBrains Mono',monospace; }

.top-bar-neon{
  position:relative; height:62px; display:flex; align-items:center; justify-content:space-between;
  padding:0 18px; background:#16131F; border-bottom:1px solid #2A2440; flex-shrink:0; overflow:hidden;
  cursor:pointer; user-select:none;
}
.neon-tube{
  position:absolute; top:0; left:0; right:0; height:2px;
  background: linear-gradient(90deg, #D4A574 0%, #FF2E93 40%, #7C3AED 75%, #D4A574 100%);
  box-shadow: 0 0 8px rgba(255,46,147,0.6);
}

.top-left{ display:flex; align-items:center; gap:14px; z-index:2; }

/* ABSURD BADGE - RAN FOR JUR LAJF */
.absurd-badge{
  min-width:132px; height:42px; background:#0a0a0a; border:1.5px solid #000; border-radius:4px;
  display:flex; flex-direction:column; align-items:center; justify-content:center;
  box-shadow: 0 2px 0 #000, 0 0 0 1px #2A2440, inset 0 1px 0 rgba(255,255,255,0.08);
  padding:2px 8px; gap:1px; transform: rotate(-1deg);
  transition: transform 0.15s;
}
.absurd-badge:hover{ transform: rotate(1deg) scale(1.03); }
.absurd-top{ font-size:7px; letter-spacing:2px; color:#7D8590; font-family:'JetBrains Mono',monospace; font-weight:700; }
.absurd-main{
  font-family:'JetBrains Mono',monospace; font-weight:900; font-size:10px; letter-spacing:0.5px; color:#E7C36A;
  text-shadow: 0 0 6px rgba(231,195,106,0.4); white-space:nowrap;
}
.absurd-bottom{ width:100%; display:flex; justify-content:center; margin-top:2px; }
.belt-line{ display:block; width:80%; height:4px; background:#111; border-top:1px solid #D4A574; border-bottom:1px solid #D4A574; box-shadow: 0 0 6px rgba(212,165,116,0.3); }

.kimono-title{
  font-family:'JetBrains Mono',monospace; font-weight:800; font-size:15px; letter-spacing:1.5px; color:#E5E7EB;
  text-shadow: 0 0 6px rgba(255,46,147,0.45), 0 0 12px rgba(124,58,237,0.3);
  display:flex; align-items:center; gap:8px;
}
.kimono-title .ver{ font-weight:400; font-size:10px; color:var(--mustard); opacity:0.8; }
.kimono-title .vhs{ font-weight:700; font-size:8px; color:#FF2E93; background:rgba(255,46,147,0.12); border:1px solid rgba(255,46,147,0.25); padding:1px 5px; border-radius:3px; }
.project-line{ display:flex; align-items:center; gap:8px; font-family:'JetBrains Mono',monospace; font-size:11px; margin-top:2px; }
.proj-name{ color:var(--mustard-bright); font-weight:700; }
.dot{ color:#3A344E; }
.path{ color:var(--text-dim); cursor:pointer; font-size:10px; max-width:260px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.path:hover{ color:var(--mustard); }

.top-center{ position:absolute; left:50%; transform:translateX(-50%); z-index:2; pointer-events:none; }
.fight-badge{
  display:flex; align-items:center; gap:8px; background: rgba(0,0,0,0.35); border:1px solid #2A2440; 
  padding:5px 12px; border-radius:20px; font-family:'JetBrains Mono',monospace; font-size:10px; letter-spacing:1px; color:#9CA3AF;
}
.neon-dot{ width:7px; height:7px; background:#FF2E93; border-radius:50%; box-shadow: 0 0 6px #FF2E93; animation: pulse 1.2s infinite; }
@keyframes pulse{ 0%,100%{ opacity:1; transform:scale(1)} 50%{ opacity:0.6; transform:scale(0.85)} }

.top-right{ display:flex; align-items:center; z-index:2; }
.btn-janusz-neon{
  background:#16131F; border:1px solid #D4A574; color:#D4A574; 
  padding:8px 14px; border-radius:6px; font-size:11px; font-weight:800; font-family:'JetBrains Mono',monospace;
  cursor:pointer; transition:0.2s;
}
.btn-janusz-neon:hover{ background:#D4A574; color:#000; box-shadow:0 0 14px rgba(212,165,116,0.35); }
.btn-janusz-neon.empty{ border-color:#FF2E93; color:#FF2E93; animation: bleed 2s infinite; }
@keyframes bleed{ 0%,100%{ box-shadow:0 0 10px rgba(255,46,147,0.2);} 50%{ box-shadow:0 0 18px rgba(255,46,147,0.5);} }

.panel { background:var(--panel); border:1px solid var(--border); border-radius:8px; overflow:hidden; display:flex; flex-direction:column; min-height:0; }

.bottom-bar.kohaku-obi {
  position:fixed; bottom:0; left:0; right:0; height:62px;
  background: repeating-linear-gradient(90deg, #c9bfa3 0px, #c9bfa3 108px, #7a1620 108px, #7a1620 110px, #8e1b26 110px, #8e1b26 218px, #7a1620 218px, #7a1620 220px);
  border-top:2px solid #0a0a0a; border-bottom:2px solid #0a0a0a;
  display:flex; align-items:center; padding:0 14px; z-index:100; overflow:hidden;
}
.bottom-bar.kohaku-obi::before { content:''; position:absolute; inset:0; background:linear-gradient(to bottom, rgba(0,0,0,0.7) 0%, transparent 50%, rgba(0,0,0,0.7) 100%); pointer-events:none; z-index:1; }
.obi-vignette { position:absolute; inset:0; background: radial-gradient(ellipse at center, transparent 60%, rgba(0,0,0,0.28) 100%); pointer-events:none; z-index:1; }
.obi-h5 { position:absolute; inset:0; display:flex; flex-direction:column; justify-content:space-between; padding:4px 0; pointer-events:none; z-index:2; }
.obi-h5 .h-line { height:0; border-top:1.5px dashed rgba(10,10,10,0.85); width:100%; }
.obi-h5 .h-line.mid { border-top:1.8px dashed #000; }
.obi-3szy { position:absolute; inset:0; display:flex; justify-content:center; gap:220px; pointer-events:none; z-index:2; }
.obi-3szy .szy { width:0; height:100%; border-left:1px dashed rgba(0,0,0,0.6); position:relative; }
.obi-3szy .szy i { position:absolute; top:50%; left:50%; transform:translate(-50%,-50%); width:10px; height:10px; background:#b8aa8a; border:1px solid #000; border-radius:50%; }
.obi-3szy .szy i::after { content:'×'; position:absolute; inset:0; display:flex; align-items:center; justify-content:center; font-size:7px; font-weight:900; color:#000; }
.obi-content { position:relative; z-index:3; display:flex; width:100%; justify-content:space-between; align-items:center; }
.btn-footer { height:38px; padding:0 18px; background:#101014; border:1.5px solid #000; color:#f5f0e8; cursor:pointer; font-size:11px; font-weight:800; border-radius:3px; display:flex; gap:8px; align-items:center; text-transform:uppercase; }
.btn-footer:hover { border-color:#d4b86a; color:#d4b86a; }
.btn-footer.menu { border-color:var(--mustard); color:var(--mustard); }
.toast-saved { position:fixed; bottom:80px; right:20px; background:var(--card); border:1px solid var(--mustard); color:var(--mustard); padding:10px 16px; border-radius:6px; z-index:200; }
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,0.9); display:flex; align-items:center; justify-content:center; z-index:300; }
.modal-content { background:#0d1117; border-radius:8px; border:1px solid var(--border); max-width:95vw; max-height:95vh; overflow:auto; position:relative; }
.modal-content.tutorial-content { background:#0D0B14; padding:0; }
.close-btn { position:absolute; top:16px; right:16px; background:#dc2626; color:white; border:none; width:40px; height:40px; border-radius:50%; cursor:pointer; z-index:10; }
.tutorial-img { width:100%; max-width:1100px; height:auto; display:block; }
</style>
