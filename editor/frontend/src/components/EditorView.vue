<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import DaysPanel from './DaysPanel.vue'
import SceneList from './SceneList.vue'
import SceneCanvas from './SceneCanvas.vue'
import ChoicesEditor from './ChoicesEditor.vue'
import { useProjectStore } from '../stores/projectStore'
import { OpenProjectFolder, RunGame, BuildGame, BuildExeGame, HostGame, StopHost, GetHostStatus } from '../../wailsjs/go/main/App'
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
const buildStep = ref('')
const buildLog = ref('')

const isHosting = ref(false)
const hostUrl = ref('')
const isBuildingHost = ref(false)

const hostBuildStep = ref('')
const hostBuildLog = ref('')

async function toggleHost() {
  if (!store.projectPath) { alert('Brak ścieżki projektu'); return }
  if (isHosting.value) {
    try {
      await StopHost()
      isHosting.value = false
      hostUrl.value = ''
      statusMsg.value = '● PACZKA ZATRZYMANA'
      setTimeout(()=> statusMsg.value='', 2000)
    } catch(e:any) {
      statusMsg.value = '⚠ ' + ( (e as any)?.message || String(e) )
    }
    return
  }
  try {
    isBuildingHost.value = true
    hostBuildStep.value = '● ZAPISUJĘ PROJEKT...'
    hostBuildLog.value = ''
    await saveProject()
    
    hostBuildStep.value = '● DOTNET PUBLISH + PACZKA...'
    hostBuildLog.value += '[1/2] Buduję engine (10-15s)...\n'
    
    statusMsg.value = '● BUDUJĘ PACZKĘ...'
    const zipPath = await HostGame(store.projectPath, 8080)
    
    hostBuildLog.value += `[2/2] Gotowe: ${zipPath}\n✓ Paczka w Downloads/\n✓ BEZ LOCALHOSTA - otwieram folder\n`
    hostBuildStep.value = '✓ PACZKA GOTOWA! (bez localhosta)'
    hostUrl.value = zipPath
    statusMsg.value = '✓ PACZKA: ' + zipPath.split('\\').pop()

    setTimeout(()=> {
      isBuildingHost.value = false
      statusMsg.value = ''
    }, 3500)
  } catch (err:any) {
    hostBuildStep.value = '⚠ BŁĄD PACZKI'
    hostBuildLog.value += '\n⚠ BŁĄD: ' + ( (err as any)?.message || String(err) ) + '\n'
    statusMsg.value = '⚠ PACZKA: ' + ( (err as any)?.message || String(err) )
  }
}

function closeHostModal() {
  if (hostBuildStep.value.includes('BŁĄD') || hostBuildStep.value.includes('GOTOWA') || hostBuildStep.value.includes('GOTOWE') || hostBuildStep.value.includes('GOTOWA!')) {
    isBuildingHost.value = false
  }
}

onMounted(async () => {
  try {
    const [running, port, url] = await GetHostStatus()
    if (running) {
      isHosting.value = true
      hostUrl.value = url
    }
  } catch {}
})

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
      await RunGame(store.projectPath);
      statusMsg.value = '✓ PARKIET GOTOWY'
      window.open(`http://localhost:5020/?nocache=${Date.now()}`, '_blank')
      setTimeout(()=> statusMsg.value='', 2500)
    }
  } catch (err: any) { 
    const msg = (err as any)?.message || (typeof err === 'string' ? err : JSON.stringify(err)) || 'Nieznany błąd'
    console.error('[PARKIET] Blad:', err)
    statusMsg.value = '⚠ ' + msg
    alert('Błąd: '+msg) 
  } finally { 
    setTimeout(()=>isRunning.value=false, 1000) 
  }
}

async function buildExe() {
  if (!store.projectPath) { alert('Brak ścieżki projektu'); return }
  try {
    isBuildingExe.value = true
    buildStep.value = '● ZAPISUJĘ PROJEKT...'
    buildLog.value = ''
    await saveProject()
    
    buildStep.value = '● TASKKILL + DOTNET BUILD-SERVER SHUTDOWN...'
    buildLog.value += '[1/3] Czyszczę stare procesy...\n'
    await new Promise(r => setTimeout(r, 300))
    
    buildStep.value = '● DOTNET PUBLISH - BUDUJĘ ENGINE...'
    buildLog.value += '[2/3] dotnet publish - to potrwa 10-15s...\n'
    
    const log = await BuildExeGame(store.projectPath)
    buildLog.value += log + '\n'
    
    buildStep.value = '✓ EXE GOTOWE!'
    buildLog.value += '\n✓ GOTOWE: Downloads/ - Maniecki Parkiet czeka!\n'
    statusMsg.value = '✓ EXE GOTOWE - Downloads/'
    
    // auto zamknij po 2.5s
    setTimeout(()=> {
      isBuildingExe.value = false
      statusMsg.value = ''
    }, 2500)
    
  } catch (err: any) {
    buildStep.value = '⚠ BŁĄD BUDOWANIA'
    buildLog.value += '\n⚠ BŁĄD: ' + ((err as any)?.message || String(err)) + '\n'
    statusMsg.value = '⚠ EXE: ' + ( (err as any)?.message || String(err) )
    // nie zamykaj automatycznie przy błędzie - niech user przeczyta
  } finally {
    // isBuildingExe zamykamy w setTimeout powyżej przy sukcesie, przy błędzie user zamyka X
    setTimeout(()=>{ if (!isBuildingExe.value) isBuildingExe.value = false }, 1000)
  }
}

function closeBuildModal() {
  if (buildStep.value.includes('BŁĄD')) {
    isBuildingExe.value = false
  }
}
</script>

<template>
  <div class="editor">
    <header class="top-bar-neon">
      <div class="neon-tube"></div>
      <div class="top-bar-inner">
        <div class="top-bar-left">
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
        
        <button @click="toggleHost" :disabled="isBuildingHost" class="btn-footer host cyan" :class="{ running: isHosting }" :title="isHosting ? hostUrl : 'Hostuj paczkę na :8080 (to samo co PUBLIC)'">
          <span v-if="!isBuildingHost && !isHosting">🌐 WEB-BUILD</span>
          <span v-else-if="isBuildingHost">● BUDUJĘ...</span>
          <span v-else>● {{ hostUrl.replace('http://','') }} ✕</span>
        </button>

        <button @click="buildExe" :disabled="isBuildingExe" class="btn-footer exe-build" :class="{ running: isBuildingExe }" title="Zbuduj EXE (Downloads/)">
          <span v-if="!isBuildingExe">💿 EXE</span>
          <span v-else>● BUDUJĘ...</span>
        </button>

        <button @click="runParkiet" :disabled="isRunning" class="btn-footer parkiet olive" :class="{ running: isRunning }" title="Klik = Odpal grę na :5020 | Shift+Klik = Build">
          <span v-if="!isRunning">▶ WPUŚĆ NA PARKIET</span>
          <span v-else>● ODPALAM...</span>
        </button>



        <div class="obi-maker-tag">
          <div class="tag-inner"><span class="kanji">先生、玉舐めて</span><span class="maker">KIMONO</span></div>
        </div>
      </div>
    </div>

    <!-- ========== BUILD EXE MODAL - VHS ESTETYKA - NIE RUSZA LAYOUTU ========== -->
    <transition name="build-modal">
      <div v-if="isBuildingExe" class="build-overlay" @click.self="closeBuildModal">
        <div class="build-modal-vhs">
          <div class="vhs-scanlines"></div>
          <div class="vhs-noise"></div>
          
          <div class="build-header">
            <div class="build-title-wrap">
              <span class="build-icon">💿</span>
              <h2 class="build-title">BUDUJĘ EXE</h2>
              <span class="build-vhs">VHS • REC • 02:17</span>
            </div>
            <button v-if="buildStep.includes('BŁĄD') || buildStep.includes('GOTOWE')" @click="isBuildingExe = false" class="build-close">✕</button>
            <div v-else class="build-rec">
              <span class="rec-dot"></span>REC
            </div>
          </div>

          <div class="build-body">
            <div class="build-step" :class="{ error: buildStep.includes('BŁĄD'), done: buildStep.includes('GOTOWE') }">
              {{ buildStep }}
            </div>
            
            <div class="build-progress-bar">
              <div class="build-progress-fill" :class="{ done: buildStep.includes('GOTOWE'), error: buildStep.includes('BŁĄD') }"></div>
            </div>

            <div class="build-log">
              <pre>{{ buildLog || 'Uruchamiam builder v41...' }}</pre>
            </div>

            <div class="build-footer">
              <div class="build-tip">
                <span class="tip-label">TIP:</span> Argus + Kustosz + 6 rano = elegancki build ✨
              </div>
              <div v-if="!buildStep.includes('BŁĄD') && !buildStep.includes('GOTOWE')" class="build-wait">
                Nie zamykaj edytora, to potrwa ~15s... dotnet publish mieli...
              </div>
              <div v-if="buildStep.includes('GOTOWE')" class="build-success">
                🎉 Maniecki Parkiet gotowy w Downloads! 🎉
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>

    <!-- ========== BUILD HOST MODAL - CYAN VHS - zamiast szarego guzika ========== -->
    <transition name="build-modal">
      <div v-if="isBuildingHost" class="build-overlay" @click.self="closeHostModal">
        <div class="build-modal-vhs host-modal">
          <div class="vhs-scanlines"></div>
          <div class="vhs-noise"></div>
          
          <div class="build-header host-header">
            <div class="build-title-wrap">
              <span class="build-icon">📦</span>
              <h2 class="build-title host-title">BUDUJĘ PACZKĘ</h2>
              <span class="build-vhs host-vhs">ZIP • DOWNLOADS • NO-LOCALHOST</span>
            </div>
            <button v-if="hostBuildStep.includes('BŁĄD') || hostBuildStep.includes('GOTOWA') || hostBuildStep.includes('GOTOWE')" @click="isBuildingHost = false" class="build-close host-close">✕</button>
            <div v-else class="build-rec host-rec">
              <span class="rec-dot host-dot"></span>REC
            </div>
          </div>

          <div class="build-body">
            <div class="build-step host-step" :class="{ error: hostBuildStep.includes('BŁĄD'), done: hostBuildStep.includes('GOTOWA') || hostBuildStep.includes('GOTOWE') }">
              {{ hostBuildStep }}
            </div>
            
            <div class="build-progress-bar host-bar">
              <div class="build-progress-fill host-fill" :class="{ done: hostBuildStep.includes('GOTOWA') || hostBuildStep.includes('GOTOWE'), error: hostBuildStep.includes('BŁĄD') }"></div>
            </div>

            <div class="build-log host-log">
              <pre>{{ hostBuildLog || 'Przygotowuję piękną paczkę...' }}</pre>
            </div>

            <div class="build-footer">
              <div class="build-tip">
                <span class="tip-label host-label">TIP:</span> Paczka leci do Pobranych - BEZ localhosta, BEZ cache ✨
              </div>
              <div v-if="!hostBuildStep.includes('BŁĄD') && !hostBuildStep.includes('GOTOWA') && !hostBuildStep.includes('GOTOWE')" class="build-wait host-wait">
                Nie zamykaj edytora, to potrwa ~10s... dotnet publish + zip...
              </div>
              <div v-if="hostBuildStep.includes('GOTOWA') || hostBuildStep.includes('GOTOWE')" class="build-success host-success">
                🎉 Paczka gotowa w Downloads! Zero localhosta! 🎉
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>

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

/* HOST BUTTON - cyan */
.btn-footer.host.cyan {
  background: linear-gradient(180deg, #0e2a2f 0%, #0a1e22 100%);
  border-color: #00d9ff;
  color: #7defff;
  box-shadow: 0 0 0 1px #001a1e, 2px 2px 0 #000, 0 0 12px rgba(0,217,255,0.25);
}
.btn-footer.host.cyan:hover {
  background: linear-gradient(180deg, #12343a 0%, #0e2a2f 100%);
  color: #aaffff;
  box-shadow: 0 0 0 1px #00d9ff, 3px 3px 0 #000, 0 0 20px rgba(0,217,255,0.4);
}
.btn-footer.host.cyan.running {
  background: #00d9ff;
  color: #001a1e;
  border-color: #7defff;
  box-shadow: 0 0 0 1px #7defff, 0 0 20px rgba(0,217,255,0.6);
  animation: host-pulse 1.2s infinite;
}
@keyframes host-pulse { 0%,100% { box-shadow: 0 0 0 1px #7defff, 0 0 12px rgba(0,217,255,0.4); } 50% { box-shadow: 0 0 0 1px #7defff, 0 0 24px rgba(0,217,255,0.8); } }

/* HOST MODAL - CYAN VARIANT - zamiast szarego guzika */
.host-modal { border-color: #00d9ff !important; box-shadow: 0 0 0 1px #001a1e, 0 8px 32px rgba(0,0,0,0.6), 0 0 24px rgba(0,217,255,0.15) !important; }
.host-header { background: linear-gradient(90deg, #0a1e22 0%, #0e2a2f 100%) !important; border-bottom-color: #00d9ff !important; }
.host-title { color: #7defff !important; }
.host-vhs { background: #001a1e !important; color: #00d9ff !important; border: 1px solid rgba(0,217,255,0.3); }
.host-rec { color: #00d9ff !important; }
.host-dot { background: #00d9ff !important; box-shadow: 0 0 8px #00d9ff !important; }
.host-close { border-color: #00d9ff !important; color: #7defff !important; }
.host-close:hover { background: #0e2a2f !important; color: #aaffff !important; }
.host-step { color: #7defff !important; background: rgba(0,217,255,0.08) !important; border-color: rgba(0,217,255,0.2) !important; }
.host-step.done { color: #4ade80 !important; }
.host-bar { border-color: #0e2a2f !important; }
.host-fill { background: linear-gradient(90deg, #0a1e22 0%, #00d9ff 50%, #7defff 100%) !important; box-shadow: 0 0 8px rgba(0,217,255,0.4) !important; }
.host-fill.done { background: #4ade80 !important; }
.host-log { border-color: #0e2a2f !important; }
.host-label { color: #00d9ff !important; }
.host-wait { color: #7defff !important; background: rgba(0,217,255,0.08) !important; border-color: rgba(0,217,255,0.2) !important; }
.host-success { color: #7defff !important; }
</style>

<style scoped>
.editor { height:100vh; display:flex; flex-direction:column; background:var(--bg); }
.main-grid { display:grid; grid-template-columns:280px 280px minmax(400px,1fr) 380px; gap:12px; padding:12px; flex:1; overflow:hidden; min-height:0; padding-bottom:68px; min-width:1376px; }
.top-bar-neon { position:relative; background:#0F0D18; border-bottom:1px solid #1E1B2E; flex-shrink:0; overflow:hidden; }
.neon-tube { position:absolute; top:0; left:0; right:0; height:2px; background:linear-gradient(90deg, #ff0080, #D4A574, #7C3AED, #ff0080); background-size:200% 100%; animation: tube-flow 3s linear infinite; box-shadow: 0 0 8px rgba(212,165,116,.6); }
@keyframes tube-flow { 0% { background-position: 0% 0%; } 100% { background-position: 200% 0%; } }
.top-bar-inner { padding:10px 16px; display:flex; align-items:center; justify-content:space-between; position:relative; }
.top-bar-left { display:flex; align-items:center; gap:18px; }
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
  max-width: 360px;
  min-width: 0;
  flex-shrink: 1;
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
.logo-wrap { display:flex; align-items:center; gap:12px; flex-shrink:0; }
.absurd-badge { background:#16131F; border:1px solid #2A2440; padding:4px 8px; border-radius:4px; cursor:pointer; user-select:none; }
.absurd-label { font-size:7px; color:var(--text-dim); font-family:'JetBrains Mono',monospace; }
.absurd-text { font-size:10px; color:var(--mustard); font-weight:700; font-family:'JetBrains Mono',monospace; }
.logo-main { display:flex; align-items:baseline; gap:8px; }
.logo-main h1 { font-family:'Bebas Neue',sans-serif; font-size:22px; margin:0; letter-spacing:0.06em; }
.ver { font-size:10px; color:var(--mustard); font-family:'JetBrains Mono',monospace; }
.vhs { font-size:9px; color:var(--text-dim); font-family:'JetBrains Mono',monospace; }
.top-bar-center { position:relative; display:flex; align-items:center; gap:8px; }
.fight-badge { font-size:10px; background:#1a1626; border:1px solid #2A2440; padding:4px 10px; border-radius:12px; display:flex; align-items:center; gap:6px; font-family:'JetBrains Mono',monospace; color:var(--text-dim); }
.status-msg {
  position:absolute;
  left:50%;
  top:calc(100% + 6px);
  transform:translateX(-50%);
  z-index:50;
  font-family:'JetBrains Mono',monospace;
  font-size:11px;
  color:#E7C36A;
  background:#1a1510;
  border:1.5px solid #3d2e15;
  padding:6px 12px;
  border-radius:4px;
  white-space:nowrap;
  box-shadow: 0 4px 12px rgba(0,0,0,0.5), 2px 2px 0 #000;
  pointer-events:none;
  animation: status-pop 0.2s ease;
}
@keyframes status-pop { from { opacity:0; transform:translateX(-50%) translateY(-4px); } to { opacity:1; transform:translateX(-50%) translateY(0); } }
.fight-badge .dot { width:6px; height:6px; background:#ff2e63; border-radius:50%; display:inline-block; box-shadow:0 0 6px #ff2e63; animation: blink 1.2s infinite; }
@keyframes blink { 0%,100% { opacity:1; } 50% { opacity:0.4; } }
.loading { flex:1; display:flex; align-items:center; justify-content:center; color:var(--text-dim); font-family:'JetBrains Mono',monospace; }

/* OBI - LEKKO ZMIENIONY - niższy, bardziej bawełniany, nie rozciąga metki */
.bottom-bar.kohaku-obi-2.kohaku-striped.dark {
  height: 52px;
  background: #1e1816;
  border-top: 1.5px solid #000;
  border-bottom: 1.5px solid #000;
  position: relative;
  display: flex;
  align-items: center;
  padding: 0 10px;
  flex-shrink: 0;
  overflow: hidden;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.06), inset 0 -1px 0 rgba(0,0,0,0.4);
}
.dark .obi-stripes { position: absolute; inset: 0; background: repeating-linear-gradient(90deg, #ddd5c0 0px, #ddd5c0 86px, #8a1a2e 86px, #8a1a2e 172px); filter: brightness(0.92) contrast(1.05) saturate(0.9); z-index: 0; }
.dark .obi-stripes::after { content:''; position:absolute; inset:0; background: repeating-linear-gradient(0deg, transparent 0px, transparent 1px, rgba(0,0,0,0.04) 1px, rgba(0,0,0,0.04) 2px); pointer-events:none; }
.dark .obi-fabric { position:absolute; inset:0; background-image: repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.12) 2px, rgba(0,0,0,0.12) 3px), repeating-linear-gradient(90deg, transparent, transparent 2px, rgba(0,0,0,0.08) 2px, rgba(0,0,0,0.08) 3px), url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E"); opacity:0.35; mix-blend-mode: multiply; z-index: 1; }
.dark .obi-dirt { position:absolute; inset:0; background: radial-gradient(ellipse at 20% 50%, rgba(0,0,0,0.15) 0%, transparent 50%), radial-gradient(ellipse at 80% 30%, rgba(0,0,0,0.12) 0%, transparent 40%), radial-gradient(ellipse at 50% 80%, rgba(0,0,0,0.08) 0%, transparent 60%); z-index: 2; }
.dark .obi-h5 { position:absolute; inset:0; display:flex; flex-direction:column; justify-content:center; gap:3px; padding:8px 0; z-index:3; opacity:0.18; }
.dark .obi-h5 .h-line { height:1px; background: #000; width:100%; }
.dark .obi-h5 .h-line.mid { height:2px; background: #000; opacity:0.8; }
.obi-content { position:relative; z-index:10; display:flex; align-items:center; gap:8px; width:100%; flex-wrap:nowrap; }
.obi-spacer { flex:1; }
.obi-maker-tag {
  position: relative;
  width: 68px;
  height: 44px;
  flex-shrink:0;
  flex-grow:0;
  background: #c19a5b;
  background-image: 
    repeating-linear-gradient(90deg, transparent, transparent 1px, rgba(0,0,0,0.05) 1px, rgba(0,0,0,0.05) 2px),
    radial-gradient(ellipse at 20% 30%, rgba(0,0,0,0.18) 0%, transparent 45%),
    radial-gradient(ellipse at 80% 20%, rgba(60,30,5,0.20) 0%, transparent 35%),
    radial-gradient(ellipse at 50% 80%, rgba(0,0,0,0.14) 0%, transparent 50%),
    radial-gradient(ellipse at 30% 70%, rgba(122,21,37,0.10) 0%, transparent 30%);
  border: 1.5px solid #0a0706;
  border-left: 2px solid #1a1512;
  display: flex;
  align-items: center;
  justify-content: center;
  transform: rotate(-1.2deg) translateY(1px);
  box-shadow: 
    0 0 0 1px #000,
    1px 2px 0 #000,
    2px 4px 8px rgba(0,0,0,0.6),
    inset 0 1px 0 rgba(255,255,255,0.35),
    inset 0 -2px 0 rgba(0,0,0,0.25);
  margin-left: 14px;
  margin-right: 6px;
  border-radius: 1px;
  overflow: visible;
  pointer-events: none;
  user-select: none;
  z-index: 5;
}
.obi-maker-tag::before {
  content: '';
  position: absolute;
  top: 3px;
  left: 50%;
  transform: translateX(-50%);
  width: 7px;
  height: 7px;
  background: #1a1512;
  border-radius: 50%;
  border: 1px solid #2a2015;
  box-shadow: inset 0 1px 2px rgba(0,0,0,1), 0 0 0 1px #000, 0 1px 0 rgba(255,255,255,0.15);
  z-index: 3;
}
.obi-maker-tag::after {
  content: '';
  position: absolute;
  inset: 3px 3px 3px 3px;
  border: 1px dashed rgba(0,0,0,0.28);
  pointer-events: none;
  border-radius: 0px;
}
.obi-maker-tag .tag-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
  margin-top: 6px;
  line-height: 1;
}
.obi-maker-tag .kanji {
  font-size: 12px;
  font-weight: 900;
  color: #1a0f08;
  font-family: 'Noto Sans JP', serif;
  text-shadow: 0 0.5px 0 rgba(255,255,255,0.4);
  line-height: 1;
  letter-spacing: 0.04em;
}
.obi-maker-tag .maker {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 9.5px;
  letter-spacing: 0.16em;
  color: #2a1510;
  font-weight: 400;
  line-height: 1;
  margin-top: 2px;
  border-top: 1px solid rgba(0,0,0,0.20);
  padding-top: 2px;
  text-shadow: 0 0.5px 0 rgba(255,255,255,0.25);
}


.btn-footer { font-family:'JetBrains Mono',monospace; font-size:11px; font-weight:700; padding:7px 12px; border-radius:4px; border:1.5px solid #000; cursor:pointer; background:#1C1929; color:var(--text); box-shadow: 2px 2px 0 #000; transition: all 0.12s ease; white-space:nowrap; flex-shrink:0; display:inline-flex; align-items:center; justify-content:center; min-height:32px; }
.btn-footer:hover { transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #000; }
.btn-footer:active { transform: translate(1px,1px); box-shadow: 1px 1px 0 #000; }
.btn-footer:disabled { opacity:0.6; cursor:not-allowed; transform:none; }
.btn-footer.save { background:#1a1f2e; border-color:#2A3441; color:#7DD3FC; min-width:132px; }
.btn-footer.menu { background:#16131F; border-color:#2A2440; color:var(--text-dim); }
.btn-footer.tutorial { background:#1a1610; border-color:#2a2015; color:#D4A574; }
.btn-footer.exe-build { background: linear-gradient(180deg, #2a1f0f 0%, #1a150a 100%); border-color:#3d2e15; color:#E7C36A; min-width:86px; width:86px; }
.btn-footer.exe-build.running { background:#3d2e15; color:#E7C36A; animation: pulse-gold 1s infinite; }
.btn-footer.parkiet { background: linear-gradient(180deg, #4a5a2a 0%, #2a3515 100%); border-color:#5a6a35; color:#a3e635; }
.btn-footer.parkiet.olive { background: linear-gradient(180deg, #4a5a2a 0%, #2a3515 100%); }
.btn-footer.parkiet.running { animation: pulse-green 1s infinite; }
@keyframes pulse-gold { 0%,100% { box-shadow: 2px 2px 0 #000, 0 0 8px rgba(231,195,106,0.4); } 50% { box-shadow: 2px 2px 0 #000, 0 0 16px rgba(231,195,106,0.7); } }
@keyframes pulse-green { 0%,100% { box-shadow: 2px 2px 0 #000, 0 0 8px rgba(163,230,53,0.4); } 50% { box-shadow: 2px 2px 0 #000, 0 0 16px rgba(163,230,53,0.7); } }

.menu-frame { background:#0a0a0a; border:1.5px solid #000; padding:2px; box-shadow: 2px 2px 0 #000; transform: rotate(0.5deg); }
.menu-frame .dark-framed { box-shadow:none; border:none; transform:none; }

/* TOAST */
.toast-saved { position:fixed; bottom:80px; left:50%; transform:translateX(-50%); background:#1a2e1a; border:1.5px solid #2a5a2a; color:#4ade80; padding:10px 20px; border-radius:6px; font-family:'JetBrains Mono',monospace; font-size:12px; font-weight:700; box-shadow: 0 4px 12px rgba(0,0,0,0.4), 2px 2px 0 #000; z-index:1000; }
.toast-enter-active, .toast-leave-active { transition: all 0.3s ease; }
.toast-enter-from, .toast-leave-to { opacity:0; transform:translateX(-50%) translateY(10px); }

/* ========== BUILD EXE MODAL - VHS ESTETYKA ========== */
.build-overlay {
  position: fixed;
  inset: 0;
  background: rgba(5, 3, 10, 0.92);
  backdrop-filter: blur(12px) saturate(0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

.build-modal-vhs {
  position: relative;
  width: 100%;
  max-width: 560px;
  background: linear-gradient(180deg, #1a1214 0%, #121010 100%);
  border: 2px solid #000;
  border-radius: 8px;
  box-shadow: 0 0 0 1px #2a2020, 0 8px 32px rgba(0,0,0,0.8), 0 0 80px rgba(122,21,37,0.15);
  overflow: hidden;
}

.vhs-scanlines {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.15) 2px, rgba(0,0,0,0.15) 3px);
  pointer-events: none;
  z-index: 5;
}

.vhs-noise {
  position: absolute;
  inset: 0;
  opacity: 0.03;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
  pointer-events: none;
  z-index: 6;
  animation: noise-shift 0.1s infinite;
}
@keyframes noise-shift { 0% { transform: translate(0,0); } 50% { transform: translate(-1px,1px); } 100% { transform: translate(1px,0); } }

.build-header {
  position: relative;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  background: linear-gradient(90deg, #1a1214 0%, #2a1a1e 100%);
  border-bottom: 2px solid #000;
}

.build-title-wrap { display:flex; align-items:center; gap:10px; }
.build-icon { font-size:18px; filter: drop-shadow(0 0 6px rgba(231,195,106,0.6)); }
.build-title { font-family:'Bebas Neue',sans-serif; font-size:20px; margin:0; letter-spacing:0.08em; color:#E7C36A; }
.build-vhs { font-family:'JetBrains Mono',monospace; font-size:9px; color:#7a1525; background:#000; padding:2px 6px; border-radius:3px; letter-spacing:0.05em; }

.build-rec { display:flex; align-items:center; gap:6px; font-family:'JetBrains Mono',monospace; font-size:11px; color:#ff2e63; font-weight:700; }
.rec-dot { width:8px; height:8px; background:#ff2e63; border-radius:50%; box-shadow:0 0 8px #ff2e63; animation: rec-blink 0.8s infinite; }
@keyframes rec-blink { 0%,100% { opacity:1; } 50% { opacity:0.3; } }

.build-close {
  background:#1a1214; border:1.5px solid #2a1a1e; color:#D4A574; width:28px; height:28px; border-radius:4px;
  font-size:14px; cursor:pointer; display:flex; align-items:center; justify-content:center;
  box-shadow: 2px 2px 0 #000; transition: all 0.12s;
}
.build-close:hover { background:#2a1a1e; color:#E7C36A; transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #000; }

.build-body { position:relative; z-index:10; padding:18px; }

.build-step {
  font-family:'JetBrains Mono',monospace; font-size:13px; font-weight:700; color:#E7C36A;
  background: rgba(231,195,106,0.08); border:1px solid rgba(231,195,106,0.2); padding:10px 12px; border-radius:4px;
  margin-bottom:14px; letter-spacing:0.02em;
}
.build-step.done { color:#4ade80; background: rgba(74,222,128,0.08); border-color: rgba(74,222,128,0.25); }
.build-step.error { color:#ff6b6b; background: rgba(255,107,107,0.08); border-color: rgba(255,107,107,0.25); }

.build-progress-bar { height:6px; background:#0a0808; border:1px solid #1a1214; border-radius:3px; overflow:hidden; margin-bottom:16px; position:relative; }
.build-progress-fill {
  height:100%; width:0%; background: linear-gradient(90deg, #7a1525 0%, #D4A574 50%, #E7C36A 100%);
  border-radius:2px; animation: progress-move 2s ease-in-out infinite; box-shadow: 0 0 8px rgba(231,195,106,0.4);
}
.build-progress-fill.done { width:100% !important; animation:none; background:#4ade80; box-shadow: 0 0 12px rgba(74,222,128,0.6); }
.build-progress-fill.error { width:100% !important; animation:none; background:#ff6b6b; }
@keyframes progress-move {
  0% { width:10%; transform: translateX(0); }
  50% { width:70%; transform: translateX(20%); }
  100% { width:10%; transform: translateX(90%); }
}

.build-log {
  background:#050304; border:1.5px solid #1a1214; border-radius:4px; padding:12px; max-height:220px; overflow-y:auto;
  font-family:'JetBrains Mono',monospace; font-size:10px; line-height:1.5; color:#9CA3AF;
  box-shadow: inset 0 2px 8px rgba(0,0,0,0.6);
}
.build-log pre { margin:0; white-space:pre-wrap; word-break:break-all; }
.build-log::-webkit-scrollbar { width:6px; }
.build-log::-webkit-scrollbar-track { background:#0a0808; }
.build-log::-webkit-scrollbar-thumb { background:#2a1a1e; border-radius:3px; }

.build-footer { margin-top:14px; display:flex; flex-direction:column; gap:8px; }
.build-tip { font-family:'JetBrains Mono',monospace; font-size:10px; color:#6B7280; }
.tip-label { color:#D4A574; font-weight:700; }
.build-wait { font-family:'JetBrains Mono',monospace; font-size:11px; color:#D4A574; background: rgba(212,165,116,0.08); border:1px dashed rgba(212,165,116,0.2); padding:8px 10px; border-radius:4px; animation: wait-pulse 1.5s infinite; }
@keyframes wait-pulse { 0%,100% { opacity:0.8; } 50% { opacity:1; } }
.build-success { font-family:'Bebas Neue',sans-serif; font-size:16px; color:#4ade80; text-align:center; letter-spacing:0.05em; animation: success-glow 1s infinite; }
@keyframes success-glow { 0%,100% { text-shadow: 0 0 8px rgba(74,222,128,0.4); } 50% { text-shadow: 0 0 16px rgba(74,222,128,0.8); } }

.build-modal-enter-active, .build-modal-leave-active { transition: all 0.25s cubic-bezier(0.4,0,0.2,1); }
.build-modal-enter-from, .build-modal-leave-to { opacity:0; transform: scale(0.96) translateY(8px); }

/* MODAL OVERLAY OGOLNY */
.modal-overlay { position:fixed; inset:0; background: rgba(5,3,10,0.85); backdrop-filter: blur(8px); display:flex; align-items:center; justify-content:center; z-index:9998; }
.modal-content { background:var(--panel); border:1.5px solid var(--border); border-radius:8px; padding:20px; max-width:90vw; max-height:90vh; overflow:auto; position:relative; box-shadow: 0 8px 32px rgba(0,0,0,0.6); }
.close-btn { position:absolute; top:10px; right:10px; background:#1a1214; border:1.5px solid #2a1a1e; color:var(--text-dim); width:28px; height:28px; border-radius:4px; cursor:pointer; }
.tutorial-img { max-width:100%; height:auto; display:block; }

/* HOST BUTTON - cyan */
.btn-footer.host.cyan {
  background: linear-gradient(180deg, #0e2a2f 0%, #0a1e22 100%);
  border-color: #00d9ff;
  color: #7defff;
  box-shadow: 0 0 0 1px #001a1e, 2px 2px 0 #000, 0 0 12px rgba(0,217,255,0.25);
}
.btn-footer.host.cyan:hover {
  background: linear-gradient(180deg, #12343a 0%, #0e2a2f 100%);
  color: #aaffff;
  box-shadow: 0 0 0 1px #00d9ff, 3px 3px 0 #000, 0 0 20px rgba(0,217,255,0.4);
}
.btn-footer.host.cyan.running {
  background: #00d9ff;
  color: #001a1e;
  border-color: #7defff;
  box-shadow: 0 0 0 1px #7defff, 0 0 20px rgba(0,217,255,0.6);
  animation: host-pulse 1.2s infinite;
}
@keyframes host-pulse { 0%,100% { box-shadow: 0 0 0 1px #7defff, 0 0 12px rgba(0,217,255,0.4); } 50% { box-shadow: 0 0 0 1px #7defff, 0 0 24px rgba(0,217,255,0.8); } }

/* HOST MODAL - CYAN VARIANT - zamiast szarego guzika */
.host-modal { border-color: #00d9ff !important; box-shadow: 0 0 0 1px #001a1e, 0 8px 32px rgba(0,0,0,0.6), 0 0 24px rgba(0,217,255,0.15) !important; }
.host-header { background: linear-gradient(90deg, #0a1e22 0%, #0e2a2f 100%) !important; border-bottom-color: #00d9ff !important; }
.host-title { color: #7defff !important; }
.host-vhs { background: #001a1e !important; color: #00d9ff !important; border: 1px solid rgba(0,217,255,0.3); }
.host-rec { color: #00d9ff !important; }
.host-dot { background: #00d9ff !important; box-shadow: 0 0 8px #00d9ff !important; }
.host-close { border-color: #00d9ff !important; color: #7defff !important; }
.host-close:hover { background: #0e2a2f !important; color: #aaffff !important; }
.host-step { color: #7defff !important; background: rgba(0,217,255,0.08) !important; border-color: rgba(0,217,255,0.2) !important; }
.host-step.done { color: #4ade80 !important; }
.host-bar { border-color: #0e2a2f !important; }
.host-fill { background: linear-gradient(90deg, #0a1e22 0%, #00d9ff 50%, #7defff 100%) !important; box-shadow: 0 0 8px rgba(0,217,255,0.4) !important; }
.host-fill.done { background: #4ade80 !important; }
.host-log { border-color: #0e2a2f !important; }
.host-label { color: #00d9ff !important; }
.host-wait { color: #7defff !important; background: rgba(0,217,255,0.08) !important; border-color: rgba(0,217,255,0.2) !important; }
.host-success { color: #7defff !important; }
</style>
