<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import DaysPanel from './DaysPanel.vue'
import SceneList from './SceneList.vue'
import SceneCanvas from './SceneCanvas.vue'
import ChoicesEditor from './ChoicesEditor.vue'
import { useProjectStore } from '../stores/projectStore'
import { OpenProjectFolder, RunGame, BuildGame, BuildExeGame } from '../../wailsjs/go/main/App'
import tutorialBasic from '../assets/tutorial_basic.jpg'
import JanuszModal from './JanuszModal.vue'

const store = useProjectStore()
const emit = defineEmits(['go-to-menu'])
const showSavedToast = ref(false)
const showTutorial = ref(false)
const isRunning = ref(false)
const absurdIndex = ref(0)
const absurdBadges = ['NO PAIN NO KIEŁBASA','RAN FOR JUR LAJF','FAJT KLUB MANIECKI','ZŁOTE KIMONO TEAM','PARKIET 02:17',]
const currentAbsurd = computed(() => absurdBadges[absurdIndex.value])
function cycleAbsurd() { absurdIndex.value = (absurdIndex.value + 1) % absurdBadges.length }
const isLoaded = computed(() => !!store.meta && !!store.projectPath)
onMounted(() => { if (!(store as any).currentDayId && (store as any).days?.length) { ;(store as any).currentDayId = (store as any).days[0].id } })
async function saveProject() { try { await store.saveProject(); showSavedToast.value = true; setTimeout(() => { showSavedToast.value = false }, 2000) } catch (e) { alert('Błąd zapisu: ' + e) } }
async function openFolder() { if (!store.projectPath) return; try { await OpenProjectFolder(store.projectPath) } catch (e) { console.warn(e) } }

const statusMsg = ref('')
const isBuildingExe = ref(false)


async function runParkiet(e: MouseEvent) {
  const isBuild = e.shiftKey
  if (!store.projectPath) { alert('Brak ścieżki projektu'); return }
  try {
    isRunning.value = true
    if (isBuild) { 
      statusMsg.value = '● BUILD RELEASE...'
      await BuildGame(store.projectPath); 
      statusMsg.value = '✓ Zbudowano w dist/'
      setTimeout(()=> statusMsg.value='', 2500)
    } else { 
      await saveProject(); 
      statusMsg.value = '● BUDUJĘ PARKIET... (Go czeka na :5020)'
      await RunGame(store.projectPath); // Go czeka 45s na port, bez CORS
      statusMsg.value = '✓ PARKIET GOTOWY'
      window.open(`http://localhost:5020/?nocache=${Date.now()}`, '_blank')
      setTimeout(()=> statusMsg.value='', 2500)
    }
  } catch (err: any) { 
    statusMsg.value = '⚠ ' + err.message
    alert('Błąd: '+err.message) 
  } finally { 
    setTimeout(()=>isRunning.value=false, 1000) 
  }
}

async function buildExe() {
  if (!store.projectPath) { alert('Brak ścieżki projektu'); return }
  try {
    isBuildingExe.value = true
    await saveProject()
    statusMsg.value = '● BUDUJĘ EXE... (taskkill + dotnet publish)'
    const log = await BuildExeGame(store.projectPath)
    console.log('BUILD EXE LOG:', log)
    statusMsg.value = '✓ EXE GOTOWE - Downloads/'
    setTimeout(()=> statusMsg.value='', 3500)
  } catch (err: any) {
    statusMsg.value = '⚠ EXE: ' + err.message
    alert('Błąd budowania EXE: '+err.message)
  } finally {
    setTimeout(()=>isBuildingExe.value=false, 1000)
  }
}
</script>

<template>
  <div class="editor">
    <header class="top-bar-neon">
      <div class="neon-tube"></div>
      <div class="top-bar-inner">
        <div class="top-bar-left">
          <!-- PEŁNA ŚCIEŻKA NA POCZĄTKU - ZIELONY LINK -->
          <div class="project-path clickable green-link" @click="openFolder" :title="store.projectPath || 'Kliknij aby otworzyć folder'">
            <span class="folder-icon">📁</span>
            <span class="path-full">{{ store.projectPath || 'test / test' }}</span>
          </div>

          <div class="logo-wrap">
            <div class="absurd-badge" @click="cycleAbsurd"><div class="absurd-label">SEKCJA</div><div class="absurd-text">{{ currentAbsurd }}</div></div>
            <div class="logo-main"><h1>KIMONO</h1><span class="ver">editor 1.0</span><span class="vhs">VHS • 1994</span></div>
          </div>
        </div>
        <div class="top-bar-center"><div class="fight-badge"><span class="dot"></span>MANIECKI • 02:17 • PARKIET</div><div v-if="statusMsg" class="status-msg">{{ statusMsg }}</div></div>
        <div class="top-bar-right"></div>
      </div>
    </header>

    <div v-if="isLoaded" class="main-grid">
      <DaysPanel class="panel" /><SceneList class="panel" /><SceneCanvas class="panel panel-preview" /><ChoicesEditor class="panel" />
    </div>
    <div v-else class="loading">Ładowanie projektu... {{ store.projectPath || '' }}</div>

    <!-- BOTTOM OBI 2.7 - DARK MENU W RAMCE -->
    <div class="bottom-bar kohaku-obi-2 kohaku-striped dark">
      <div class="obi-stripes"></div>
      <div class="obi-fabric"></div>
      <div class="obi-dirt"></div>
      <div class="obi-h5">
        <span class="h-line"></span>
        <span class="h-line"></span>
        <span class="h-line mid"></span>
        <span class="h-line"></span>
        <span class="h-line"></span>
      </div>

      <div class="obi-content">
        <button @click="saveProject" class="btn-footer save">💾 ZAPISZ PROJEKT</button>
        
        <div class="menu-frame">
          <button @click="emit('go-to-menu')" class="btn-footer menu dark-framed">📁 MENU</button>
        </div>

        <button @click="showTutorial = true" class="btn-footer tutorial">📖 POKAŻ TUTORIAL</button>
        
        <div class="obi-spacer"></div>

        <button @click="buildExe" :disabled="isBuildingExe" class="btn-footer exe-build" :class="{ running: isBuildingExe }" title="Zbuduj EXE (Downloads/)">
          <span v-if="!isBuildingExe">💿 EXE</span>
          <span v-else>● BUDUJĘ...</span>
        </button>

        <button @click="runParkiet" :disabled="isRunning" class="btn-footer parkiet olive" :class="{ running: isRunning }" title="Klik = Odpal grę na :5020 | Shift+Klik = Build">
          <span v-if="!isRunning">▶ WPUŚĆ NA PARKIET</span>
          <span v-else>● ODPALAM...</span>
        </button>

        <div class="obi-maker-tag">
          <div class="tag-inner"><span class="kanji">武道</span><span class="maker">KIMONO</span></div>
        </div>
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
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;700&family=Bebas+Neue&display=swap');
:root { --bg:#0D0B14; --panel:#16131F; --card:#1C1929; --input:#151221; --border:#2A2440; --mustard:#D4A574; --mustard-bright:#E7C36A; --shiro:#E5E7EB; --text:#E6EDF3; --text-dim:#9CA3AF; --pl-red:#CE0E2D; }
* { box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:var(--bg); margin:0; color:var(--text); overflow-x:hidden; }
</style>

<style scoped>
.editor { height:100vh; display:flex; flex-direction:column; background:var(--bg); }
.main-grid { display:grid; grid-template-columns:280px 280px minmax(400px,1fr) 380px; gap:12px; padding:12px; flex:1; overflow:hidden; min-height:0; padding-bottom:68px; min-width:1376px; }
.top-bar-neon { position:relative; background:#0F0D18; border-bottom:1px solid #1E1B2E; flex-shrink:0; overflow:hidden; }
.neon-tube { position:absolute; top:0; left:0; right:0; height:2px; background:linear-gradient(90deg, #ff0080, #D4A574, #7C3AED, #ff0080); background-size:200% 100%; animation: tube-flow 3s linear infinite; box-shadow: 0 0 8px rgba(212,165,116,.6); }
@keyframes tube-flow { 0% { background-position: 0% 0%; } 100% { background-position: 200% 0%; } }
.top-bar-inner { padding:10px 16px; display:flex; align-items:center; justify-content:space-between; }
.top-bar-left { display:flex; align-items:center; gap:18px; }

/* ZIELONY LINK - NA POCZĄTKU - PEŁNA ŚCIEŻKA */
.project-path.green-link {
  font-size: 11px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 7px;
  background: rgba(34, 197, 94, 0.08);
  border: 1.5px solid rgba(34, 197, 94, 0.3);
  padding: 6px 12px;
  border-radius: 4px;
  transition: all 0.15s ease;
  max-width: 480px;
  overflow: hidden;
  font-family: 'JetBrains Mono', monospace;
}
.project-path.green-link:hover {
  background: rgba(34, 197, 94, 0.15);
  border-color: #22c55e;
  box-shadow: 0 0 10px rgba(34, 197, 94, 0.25), 0 2px 0 #000;
  transform: translateY(-1px);
}
.project-path.green-link:active { transform: translateY(0); }
.folder-icon { font-size: 12px; flex-shrink: 0; }
.path-full {
  color: #22c55e;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-decoration: underline;
  text-decoration-color: rgba(34, 197, 94, 0.35);
  text-underline-offset: 3px;
  letter-spacing: 0.01em;
}
.project-path.green-link:hover .path-full {
  color: #4ade80;
  text-decoration-color: #22c55e;
}

.logo-wrap { display:flex; align-items:center; gap:12px; }
.absurd-badge { background:#16131F; border:1px solid #2A2440; padding:4px 8px; border-radius:4px; cursor:pointer; user-select:none; }
.absurd-label { font-size:7px; color:var(--text-dim); font-family:'JetBrains Mono',monospace; }
.absurd-text { font-size:10px; color:var(--mustard); font-weight:700; font-family:'JetBrains Mono',monospace; }
.logo-main { display:flex; align-items:baseline; gap:8px; }
.logo-main h1 { font-family:'Bebas Neue',sans-serif; font-size:22px; margin:0; letter-spacing:0.06em; }
.ver { font-size:10px; color:var(--mustard); font-family:'JetBrains Mono',monospace; }
.vhs { font-size:9px; color:var(--text-dim); font-family:'JetBrains Mono',monospace; }
.fight-badge { font-size:10px; background:#1a1626; border:1px solid #2A2440; padding:4px 10px; border-radius:12px; display:flex; align-items:center; gap:6px; font-family:'JetBrains Mono',monospace; color:var(--text-dim); }
.status-msg { margin-left:16px; font-family:'JetBrains Mono',monospace; font-size:11px; color:#D4A574; background:#1a1510; border:1px solid #2a2015; padding:4px 8px; border-radius:4px; white-space:nowrap; }
.fight-badge .dot { width:6px; height:6px; background:#ff2e63; border-radius:50%; display:inline-block; box-shadow:0 0 6px #ff2e63; animation: blink 1.2s infinite; }
@keyframes blink { 0%,100% { opacity:1; } 50% { opacity:0.4; } }
.loading { flex:1; display:flex; align-items:center; justify-content:center; color:var(--text-dim); font-family:'JetBrains Mono',monospace; }

/* ========== OBI 2.7 ========== */
.bottom-bar.kohaku-obi-2.kohaku-striped.dark {
  height: 60px;
  background: #1a1214;
  border-top: 2px solid #000;
  border-bottom: 2px solid #000;
  position: relative;
  display: flex;
  align-items: center;
  padding: 0 14px;
  flex-shrink: 0;
  overflow: hidden;
}
.dark .obi-stripes {
  position: absolute; inset: 0;
  background: repeating-linear-gradient(90deg, #d5cdb8 0px, #d5cdb8 88px, #7a1525 88px, #7a1525 176px);
  filter: brightness(0.88) contrast(1.1);
  z-index: 0;
}
.dark .obi-fabric {
  position:absolute; inset:0;
  background-image: 
    repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.12) 2px, rgba(0,0,0,0.12) 3px),
    repeating-linear-gradient(90deg, transparent, transparent 2px, rgba(0,0,0,0.08) 2px, rgba(0,0,0,0.08) 3px),
    url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)' opacity='0.18'/%3E%3C/svg%3E");
  opacity:0.85; mix-blend-mode: multiply; pointer-events:none; z-index:1;
}
.dark .obi-dirt {
  position:absolute; inset:0;
  background:
    radial-gradient(ellipse 280px 90px at 18% 75%, rgba(60,35,15,0.28) 0%, transparent 70%),
    radial-gradient(ellipse 200px 70px at 52% 40%, rgba(80,20,20,0.22) 0%, transparent 65%),
    radial-gradient(ellipse 320px 110px at 85% 80%, rgba(30,15,5,0.32) 0%, transparent 70%),
    radial-gradient(ellipse 12px 12px at 34% 55%, rgba(0,0,0,0.45) 0%, transparent 70%),
    radial-gradient(ellipse 18px 18px at 67% 32%, rgba(40,10,5,0.5) 0%, transparent 70%),
    linear-gradient(90deg, rgba(0,0,0,0.18) 0%, transparent 12%, transparent 88%, rgba(0,0,0,0.25) 100%);
  mix-blend-mode: multiply;
  pointer-events:none; z-index:2;
}
.dark .obi-h5 { position:absolute; inset:0; display:flex; flex-direction:column; justify-content:space-between; padding:6px 0; pointer-events:none; z-index:3; }
.dark .obi-h5 .h-line { height:0; border-top:1px solid rgba(0,0,0,0.55); width:100%; box-shadow: 0 1px 0 rgba(255,255,255,0.12); }
.dark .obi-h5 .h-line.mid { border-top:1.5px solid rgba(0,0,0,0.7); box-shadow: 0 1px 0 rgba(255,255,255,0.18); }
.obi-content { position:relative; z-index:5; display:flex; width:100%; justify-content:flex-start; align-items:center; gap:12px; }
.obi-spacer { flex: 1; }
.btn-footer {
  height:36px; padding:0 18px; font-size:11px; font-weight:800; border-radius:2px;
  display:flex; gap:8px; align-items:center; letter-spacing:0.06em; text-transform:uppercase;
  font-family:'JetBrains Mono',monospace; cursor:pointer; transition: all 0.15s ease;
  white-space: nowrap;
}
.btn-footer.save {
  background:#0a0a0a; border:1.5px solid #D4A574; color:#e7dcc0;
  box-shadow: 0 2px 0 #000, 0 4px 12px rgba(0,0,0,0.6), inset 0 1px 0 rgba(255,255,255,0.1);
}
.btn-footer.save:hover { background:#151310; border-color:#E7C36A; transform: translateY(-1px); box-shadow: 0 3px 0 #000, 0 6px 16px rgba(0,0,0,0.7); }
.menu-frame {
  position: relative; padding: 3px; background: #f0ece1; border: 2px solid #000; border-radius: 3px;
  box-shadow: 0 2px 0 #000, 0 4px 10px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.8); transform: rotate(-0.5deg);
}
.menu-frame::before { content: ''; position: absolute; inset: 2px; border: 1px dashed rgba(0,0,0,0.25); pointer-events: none; border-radius: 1px; }
.btn-footer.menu.dark-framed {
  background: #0e0e0e; border: 1.5px solid #1a1a1a; color: #f0ece1; height: 30px; min-width: 90px; justify-content: center;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.12), inset 0 -1px 0 rgba(0,0,0,0.8); margin: 0; position: relative;
}
.btn-footer.menu.dark-framed:hover { background: #1a1a1a; color: #fff; border-color: #333; transform: translateY(-1px); }
.btn-footer.tutorial {
  background:#1a0a0e; border:1.5px solid #a81c32; color:#ffdde0;
  box-shadow: 0 2px 0 #000, 0 4px 10px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.08);
}
.btn-footer.tutorial:hover { background:#2a0e16; border-color:#ce223c; transform: translateY(-1px); }
.btn-footer.parkiet {
  height: 36px; background: #22281a; border: 1.5px solid #8a9a5b; color: #d5d8a0;
  box-shadow: 0 2px 0 #000, 0 4px 12px rgba(0,0,0,0.6), inset 0 1px 0 rgba(255,255,255,0.08);
}
.btn-footer.parkiet:hover { background: #2f3a1f; border-color: #a8b87a; color: #e8ebc0; transform: translateY(-1px); box-shadow: 0 3px 0 #000, 0 6px 16px rgba(0,0,0,0.7); }
.btn-footer.parkiet.running { background: #000; border-color: #383f26; color: #383f26; animation: blink 0.7s infinite; pointer-events: none; }
.btn-footer.exe-build {
  height: 36px; background: #0e1018; border: 1.5px solid #E7C36A; color: #E7C36A;
  box-shadow: 0 2px 0 #000, 0 4px 12px rgba(0,0,0,0.6), inset 0 1px 0 rgba(255,255,255,0.08);
}
.btn-footer.exe-build:hover { background: #1a1c0e; border-color: #FFD700; color: #FFD700; transform: translateY(-1px); box-shadow: 0 3px 0 #000, 0 6px 16px rgba(231,195,106,0.25); }
.btn-footer.exe-build.running { background: #000; border-color: #444; color: #666; animation: blink 0.7s infinite; pointer-events: none; }
.obi-maker-tag {
  margin-left: 12px; background:#0a0a0a; border:1.5px solid #222; padding:4px 10px 4px 8px;
  display:flex; align-items:center; transform: rotate(-1.5deg); box-shadow: 0 2px 0 #000, 0 3px 8px rgba(0,0,0,0.6); position:relative;
}
.obi-maker-tag::before { content:'320'; position:absolute; top:-8px; right:-6px; background:#f0ece1; color:#111; font-size:7px; font-weight:900; padding:1px 4px; border:1px solid #000; font-family:'JetBrains Mono',monospace; transform: rotate(3deg); }
.tag-inner { display:flex; flex-direction:column; align-items:center; line-height:1; }
.kanji { font-size:16px; color:#c9a84c; font-weight:900; letter-spacing:0.1em; text-shadow: 0 1px 0 #000; }
.maker { font-size:7px; color:#e8e0d0; letter-spacing:0.15em; font-family:'JetBrains Mono',monospace; margin-top:2px; }
.toast-saved { position:fixed; bottom:74px; right:20px; background:var(--card); border:1px solid var(--mustard); color:var(--mustard); padding:10px 16px; border-radius:6px; z-index:200; font-family:'JetBrains Mono',monospace; }
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,0.9); display:flex; align-items:center; justify-content:center; z-index:300; }
.modal-content { background:#0d1117; border-radius:8px; border:1px solid var(--border); max-width:95vw; max-height:95vh; overflow:auto; position:relative; }
.modal-content.tutorial-content { background:#0D0B14; border:1px solid var(--border); padding:0; line-height:0; }
.close-btn { position:absolute; top:16px; right:16px; background:#dc2626; color:white; border:none; width:40px; height:40px; border-radius:50%; cursor:pointer; z-index:10; }
.tutorial-img { width:100%; max-width:1100px; height:auto; display:block; }
</style>
