<template>
  <div class="kimono-launcher">
    <img :src="bgImage" class="bg-image" alt="" />
    <div class="top-right-beer">
      <button class="kimono-btn beer-btn" @click="donate">ZA DARMO, ALE MOŻESZ PISTAWIĆ PIWO SENSEI 🍺</button>
    </div>
    <div class="projects-panel">
      <div class="projects-header"><span class="dot" /> PROJECTS // ŁÓDŹ DRIVE</div>
      <div class="projects-list">
        <div v-for="p in displayProjects" :key="p.path" class="project-item" :class="{ active: selectedPath === p.path }" @click="openRecent(p.path)">
          <div class="proj-name">{{ p.name }}</div>
          <div class="proj-meta"><span>{{ p.date }}</span><span>{{ p.size }}</span></div>
        </div>
      </div>
      <div class="disco-hint">DISCO BALL: ACTIVE<br/>REFLECTIVE FLOOR: ON<br/>JAWLINE: CHISELED ✓</div>
    </div>
    <div class="bottom-bar">
      <button class="kimono-btn" @click="handleNew">[ NEW DOJO ]</button>
      <button class="kimono-btn" @click="handleOpen">[ OPEN PROJECT ]</button>
      <button class="kimono-btn ghost" disabled>[ BUILD ]</button>
    </div>
    <div class="version">v2.86 - 1986 EDITION | ŁÓDŹ PL</div>
    <div v-if="err" class="error-box">{{ err }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import bgImage from '../assets/bg.jpg'
import { useProjectStore } from '../stores/projectStore'
const projectStore = useProjectStore()
const displayProjects = ref<any[]>([])
const selectedPath = ref('')
const err = ref('')
function getApp(): any { return (window as any).go?.main?.App }
function mapProjects(raw:any[]){
  return raw.map((r:any)=>{
    if(typeof r === 'string'){ const n=r.split(/[\\/]/).pop()!.toUpperCase(); return { path:r, name:n.slice(0,28), date:'1996-04-12', size:'12.4 MB' } }
    const path=r.path||r.Path||''; const name=(r.name||path.split(/[\\/]/).pop()||path).toString().replace('.janproj','').toUpperCase()
    return { path, name:name.slice(0,28), date:(r.lastOpened||r.date||'1996-04-12').toString().slice(0,10), size:'12.4 MB' }
  })
}
async function handleNew(){
  const App = getApp()
  err.value=''
  try{
    // NOWY UX: 1 prompt, bez Explorera
    let defaultPath = 'C:\\Users\\MD-Core\\KimonoGM\\games'
    try{
      if(App?.GetDefaultProjectPath){
        defaultPath = await App.GetDefaultProjectPath()
      }
    }catch{}
    // 1. Nazwa projektu
    const projectName = prompt(`Nazwa nowego dojo:\n(Zostanie utworzone w ${defaultPath})`, 'Moja Gra')
    if(!projectName || projectName.trim()==='') return

    const cleanName = projectName.trim().replace(/[\\\/:*?"<>|]/g, '_')
    const fullPath = defaultPath.replace(/\\/g,'/').replace(/\/$/,'') + '/' + cleanName
    
    // 2. Twórz od razu, bez wyboru folderu
    console.log('Creating project at:', fullPath)
    await projectStore.createProjectAtPath(fullPath, cleanName)
    console.log('Created:', projectStore.projectPath)
    
  }catch(e:any){
    err.value='NEW ERR: '+e.message
    console.error(e)
    // Fallback: jeśli zawiedzie (np. brak uprawnień), dopiero wtedy pokaż Explorer jako opcję wyboru innego miejsca
    const App2 = getApp()
    if(confirm('Nie udało się utworzyć w domyślnym folderze. Chcesz wybrać inną lokalizację?')){
      try{
        const parent = await App2.SelectFolder()
        if(!parent) return
        const name = prompt('Nazwa projektu:', 'Moja gra')
        if(!name) return
        const fp = parent.replace(/\\/g,'/') + '/' + name.trim()
        await projectStore.createProjectAtPath(fp, name.trim())
      }catch(e2:any){ err.value='NEW ERR2: '+e2.message }
    }
  }
}
async function handleOpen(){
  const App=getApp(); try{ const f=App?.SelectFolder?await App.SelectFolder():null; if(!f) return; await projectStore.loadProjectFromPath(f.replace(/\\/g,'/')); if(App?.AddRecentProject) await App.AddRecentProject(f) }catch(e:any){ err.value='OPEN ERR: '+e.message }
}
async function openRecent(p:string){ selectedPath.value=p; try{ await projectStore.loadProjectFromPath(p.replace(/\\/g,'/')) }catch(e:any){ err.value='RECENT ERR: '+e.message } }
function donate(){ const rt=(window as any).runtime; if(rt?.BrowserOpenURL) rt.BrowserOpenURL('https://buycoffee.to/'); else window.open('https://buycoffee.to/','_blank') }
onMounted(async()=>{ const App=getApp(); let t=0; const it=setInterval(async()=>{ const a=getApp(); if(a||t>20){clearInterval(it); if(a?.GetRecentProjects){try{const r=await a.GetRecentProjects(); if(r?.length>0) displayProjects.value=mapProjects(r)}catch{}}} t++},300) })
</script>

<style scoped>
.kimono-launcher{position:relative;width:100vw;height:100vh;overflow:hidden;background:black}
.bg-image{position:absolute;inset:0;width:100%;height:100%;object-fit:cover;object-position:left center}
.top-right-beer{position:absolute;top:18px;right:24px;z-index:5;max-width:520px}
.beer-btn{font-size:13px !important;padding:12px 16px !important;white-space:normal !important;text-align:left;max-width:520px;opacity:0.9}
.projects-panel{position:absolute;right:28px;top:50%;transform:translateY(-50%);width:360px;z-index:2;display:flex;flex-direction:column;max-height:72vh}
.projects-header{color:#777;font-size:11px;letter-spacing:2.5px;margin-bottom:14px;display:flex;align-items:center;gap:8px;font-family:'Courier New',monospace;font-weight:700}
.dot{width:8px;height:8px;background:#FF0000;border-radius:50%;box-shadow:0 0 6px red}
.projects-list{flex:1;overflow-y:auto;padding-right:6px;display:flex;flex-direction:column;gap:8px;max-height:52vh}
.projects-list::-webkit-scrollbar{width:4px}
.projects-list::-webkit-scrollbar-thumb{background:rgba(255,235,59,0.3)}
.project-item{background:rgba(10,10,10,0.85);border:1px solid rgba(255,235,59,0.35);border-left:3px solid #FFEB3B;padding:14px 16px;cursor:pointer;transition:all 0.15s;flex-shrink:0}
.project-item:hover{background:rgba(30,30,15,0.9);border-color:#FFEB3B;transform:translateX(-2px)}
.project-item.active{background:#FFEB3B;color:black;border:1px solid #FF2D7B;border-left:3px solid #FF2D7B;box-shadow:4px 4px 0 #00E5FF}
.proj-name{font-weight:800;font-size:13px;font-family:'Courier New',monospace}
.project-item:not(.active) .proj-name{color:#FFEB3B}
.proj-meta{font-size:10px;opacity:0.6;display:flex;justify-content:space-between;margin-top:6px;font-family:'Courier New',monospace}
.disco-hint{margin-top:14px;color:#555;font-size:9px;line-height:1.6;font-family:monospace;opacity:0.7;flex-shrink:0}
.bottom-bar{position:absolute;bottom:32px;left:50%;transform:translateX(-50%);display:flex;gap:16px;z-index:3}
.kimono-btn{background:#FFEB3B;color:black;font-family:'Courier New',monospace;font-weight:900;font-size:14px;padding:14px 24px;border:2px solid #FF2D7B;box-shadow:4px 4px 0 #00E5FF;cursor:pointer}
.kimono-btn:hover{transform:translate(-2px,-2px);box-shadow:6px 6px 0 #00E5FF;background:#FFF176}
.kimono-btn.ghost{opacity:0.35;background:rgba(0,0,0,0.6);color:#777;border-color:#333;box-shadow:none}
.version{position:absolute;bottom:12px;right:18px;color:rgba(255,255,255,0.35);font-size:10px;font-family:monospace}
.error-box{position:absolute;bottom:90px;left:20px;background:rgba(0,0,0,0.9);color:#FFEB3B;padding:10px 14px;font-family:monospace;font-size:11px;max-width:500px;border:1px solid #FFEB3B;z-index:20}
</style>
