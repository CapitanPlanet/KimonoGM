<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import DaysPanel from './DaysPanel.vue'
import SceneList from './SceneList.vue'
import SceneCanvas from './SceneCanvas.vue'
import ChoicesEditor from './ChoicesEditor.vue'
import TopBar from './editor/TopBar.vue'
import ObiBar from './editor/ObiBar.vue'
import BuildModal from './editor/BuildModal.vue'
import HostModal from './editor/HostModal.vue'
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
const absurdBadges = ['NO PAIN NO KIEŁBASA','RAN FOR JUR LAJF','FAJT KLUB MANIECKI','ZŁOTE KIMONO TEAM','PARKIET 02:17']
const currentAbsurd = computed(() => absurdBadges[absurdIndex.value])
function cycleAbsurd() { absurdIndex.value = (absurdIndex.value + 1) % absurdBadges.length }
const isLoaded = computed(() => !!store.meta && !!store.projectPath)

onMounted(() => { if (!(store as any).currentDayId && (store as any).days?.length) { (store as any).currentDayId = (store as any).days[0].id } })
async function saveProject() { try { await store.saveProject(); showSavedToast.value = true; setTimeout(() => showSavedToast.value = false, 2000) } catch (e) { alert('Błąd zapisu: ' + e) } }
async function openFolder() { if (!store.projectPath) return; try { await OpenProjectFolder(store.projectPath) } catch {} }

const statusMsg = ref('')
const isBuildingExe = ref(false); const buildStep = ref(''); const buildLog = ref('')
const isHosting = ref(false); const hostUrl = ref(''); const isBuildingHost = ref(false)
const hostBuildStep = ref(''); const hostBuildLog = ref('')

async function toggleHost() {
  if (!store.projectPath) return
  if (isHosting.value) { await StopHost(); isHosting.value=false; hostUrl.value=''; statusMsg.value='● PACZKA ZATRZYMANA'; setTimeout(()=>statusMsg.value='',2000); return }
  try {
    isBuildingHost.value=true; hostBuildStep.value='● ZAPISUJĘ PROJEKT...'; hostBuildLog.value=''; await saveProject()
    hostBuildStep.value='● DOTNET PUBLISH + PACZKA...'; hostBuildLog.value+='[1/2] Buduję engine (10-15s)...\n'
    statusMsg.value='● BUDUJĘ PACZKĘ...'; const zipPath = await HostGame(store.projectPath,8080)
    hostBuildLog.value+=`[2/2] Gotowe: ${zipPath}\n✓ Paczka w Downloads/\n`; hostBuildStep.value='✓ PACZKA GOTOWA!'; hostUrl.value=zipPath; statusMsg.value='✓ PACZKA: '+zipPath.split('\\').pop()
    setTimeout(()=>{ isBuildingHost.value=false; statusMsg.value='' },3500)
  } catch(err:any){ hostBuildStep.value='⚠ BŁĄD PACZKI'; hostBuildLog.value+='\n⚠ BŁĄD: '+(err?.message||String(err))+'\n'; statusMsg.value='⚠ PACZKA: '+(err?.message||String(err)) }
}
onMounted(async()=>{ try{ const [running,port,url]=await GetHostStatus(); if(running){ isHosting.value=true; hostUrl.value=url } }catch{} })

async function runParkiet(e: MouseEvent) {
  const isBuild = e.shiftKey; if (!store.projectPath) return
  try { isRunning.value=true; if(isBuild){ statusMsg.value='● BUILD RELEASE...'; await BuildGame(store.projectPath); statusMsg.value='✓ Zbudowano w dist/'; setTimeout(()=>statusMsg.value='',2500) }
  else { await saveProject(); statusMsg.value='● BUDUJĘ PARKIET...'; await RunGame(store.projectPath); statusMsg.value='✓ PARKIET GOTOWY'; window.open(`http://localhost:5020/?nocache=${Date.now()}`,'_blank'); setTimeout(()=>statusMsg.value='',2500) } }
  catch(err:any){ const msg=err?.message||String(err); statusMsg.value='⚠ '+msg; alert('Błąd: '+msg) } finally{ setTimeout(()=>isRunning.value=false,1000) }
}

async function buildExe() {
  if(!store.projectPath) return
  try{ isBuildingExe.value=true; buildStep.value='● ZAPISUJĘ PROJEKT...'; buildLog.value=''; await saveProject()
    buildStep.value='● TASKKILL + DOTNET BUILD-SERVER SHUTDOWN...'; buildLog.value+='[1/3] Czyszczę stare procesy...\n'; await new Promise(r=>setTimeout(r,300))
    buildStep.value='● DOTNET PUBLISH - BUDUJĘ ENGINE...'; buildLog.value+='[2/3] dotnet publish - to potrwa 10-15s...\n'
    const log = await BuildExeGame(store.projectPath); buildLog.value+=log+'\n'; buildStep.value='✓ EXE GOTOWE!'; buildLog.value+='\n✓ GOTOWE: Downloads/\n'; statusMsg.value='✓ EXE GOTOWE - Downloads/'
    setTimeout(()=>{ isBuildingExe.value=false; statusMsg.value='' },2500)
  }catch(err:any){ buildStep.value='⚠ BŁĄD BUDOWANIA'; buildLog.value+='\n⚠ BŁĄD: '+(err?.message||String(err))+'\n'; statusMsg.value='⚠ EXE: '+(err?.message||String(err)) }
}
</script>

<template>
  <div class="editor">
    <TopBar :project-path="store.projectPath" :status-msg="statusMsg" :absurd="currentAbsurd" @open-folder="openFolder" @cycle-absurd="cycleAbsurd" />

    <div v-if="isLoaded" class="main-grid">
      <DaysPanel class="panel" /><SceneList class="panel" /><SceneCanvas class="panel panel-preview" /><ChoicesEditor class="panel" />
    </div>
    <div v-else class="loading">Ładowanie projektu... {{ store.projectPath || '' }}</div>

    <ObiBar :is-hosting="isHosting" :host-url="hostUrl" :is-building-host="isBuildingHost" :is-building-exe="isBuildingExe" :is-running="isRunning"
      @save="saveProject" @menu="emit('go-to-menu')" @tutorial="showTutorial=true" @host="toggleHost" @build-exe="buildExe" @parkiet="runParkiet" />

    <BuildModal v-if="isBuildingExe" :step="buildStep" :log="buildLog" @close="isBuildingExe=false" />
    <HostModal v-if="isBuildingHost" :step="hostBuildStep" :log="hostBuildLog" @close="isBuildingHost=false" />

    <transition name="toast"><div v-if="showSavedToast" class="toast-saved">✓ Zapisano</div></transition>
    <div v-if="showTutorial" class="modal-overlay" @click.self="showTutorial=false"><div class="modal-content"><button class="close-btn" @click="showTutorial=false">✕</button><img :src="tutorialBasic" class="tutorial-img"/></div></div>
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
.main-grid { display:grid; grid-template-columns:280px 280px minmax(400px,1fr) 380px; gap:12px; padding:12px; flex:1; overflow:hidden; min-height:0; padding-bottom:0; min-width:1376px; }
.loading { flex:1; display:flex; align-items:center; justify-content:center; color:var(--text-dim); font-family:'JetBrains Mono',monospace; }
.toast-saved { position:fixed; bottom:80px; left:50%; transform:translateX(-50%); background:#1a2e1a; border:1.5px solid #2a5a2a; color:#4ade80; padding:10px 20px; border-radius:6px; font-family:'JetBrains Mono',monospace; font-size:12px; font-weight:700; box-shadow:0 4px 12px rgba(0,0,0,0.4),2px 2px 0 #000; z-index:1000; }
.toast-enter-active,.toast-leave-active{transition:all 0.3s ease} .toast-enter-from,.toast-leave-to{opacity:0;transform:translateX(-50%) translateY(10px)}
.modal-overlay{position:fixed;inset:0;background:rgba(5,3,10,0.85);backdrop-filter:blur(8px);display:flex;align-items:center;justify-content:center;z-index:9998}
.modal-content{background:var(--panel);border:1.5px solid var(--border);border-radius:8px;padding:20px;max-width:90vw;max-height:90vh;overflow:auto;position:relative;box-shadow:0 8px 32px rgba(0,0,0,0.6)}
.close-btn{position:absolute;top:10px;right:10px;background:#1a1214;border:1.5px solid #2a1a1e;color:var(--text-dim);width:28px;height:28px;border-radius:4px;cursor:pointer}
.tutorial-img{max-width:100%;height:auto;display:block}
</style>
