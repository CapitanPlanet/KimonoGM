<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useProjectStore } from '../stores/projectStore'
import {
  SelectImageFile, ImportAsset, GetImageBase64, DeleteAsset,
  SelectAudioFile, ImportAudioAsset, ListAudioAssets, DeleteAudioAsset,
  SetProjectPath, DeleteFile
} from '../../wailsjs/go/main/App'

const store = useProjectStore()
const assetCache = ref<Record<string,string>>({})
const sectionsOpen = ref({ bg: true, re: true, av: true, sfx: true, vo: true, mu: true })

const backgroundAssets = computed(() => store.backgroundAssets?? [])
const reactionAssets = computed(() => store.reactionAssets?? [])
const avatarAssets = computed(() => store.avatarAssets?? [])
const avatarRulesCount = computed(() => store.avatarSystem?.rules?.length?? 0)
const avatarCount = computed(() => avatarAssets.value.length?? 0)
const totalFiles = computed(() =>
  (backgroundAssets.value.length + reactionAssets.value.length + avatarAssets.value.length + sfxAssets.value.length + voiceAssets.value.length + musicAssets.value.length)
)

const sfxAssets = ref<string[]>([])
const voiceAssets = ref<string[]>([])
const musicAssets = ref<string[]>([])

function norm(p: string) {
  return (p || '').replace(/\\/g, '/').trim()
}

function toSafeDayId(raw: string): string {
  // dowolna nazwa od użytkownika -> bezpieczny id pliku
  // np. "Wesele u Basi!!" -> "wesele_u_basi"
  let s = raw.trim().toLowerCase()
  s = s.replace(/\s+/g, '_')
  s = s.replace(/[^a-z0-9_\-]/g, '')
  s = s.replace(/_+/g, '_').replace(/-+/g, '-')
  s = s.replace(/^_+|_+$/g, '')
  return s.slice(0, 40)
}

async function refreshAssets() {
  if (!store.projectPath) return
  try {
    await SetProjectPath(store.projectPath)
    await store.loadAssets()
    for (const raw of (store.assets.all?? [])) {
      const key = norm(raw)
      if (!key) continue
      if (assetCache.value[key] || assetCache.value[raw]) continue
      try {
        const b64 = await GetImageBase64(key)
        assetCache.value[key] = b64
        assetCache.value[raw] = b64
        assetCache.value[raw.trim()] = b64
      } catch (e) {
        console.warn('[ASSETS] thumb fail', key, e)
      }
    }
    await refreshAudio()
  } catch(e) {
    console.warn('[ASSETS] refresh error', e)
  }
}

async function refreshAudio() {
  if (!store.projectPath) return
  try {
    await SetProjectPath(store.projectPath)
    const allAudio = (await ListAudioAssets() as any)?? [] as string[]
    const lower = (s:string) => s.toLowerCase()
    const normalized = allAudio.map((a:string)=>norm(a))
    sfxAssets.value = normalized.filter((a:string) => a.includes('sfx/') || lower(a).endsWith('.mp3') || lower(a).endsWith('.wav') || lower(a).endsWith('.ogg'))
    voiceAssets.value = normalized.filter((a:string) => a.includes('voice/'))
    musicAssets.value = normalized.filter((a:string) => a.includes('music/'))
  } catch (e) {
    console.warn('[AUDIO] list error', e)
    sfxAssets.value = []
    voiceAssets.value = []
    musicAssets.value = []
  }
}

const getAssetUrl = (p: string) => {
  if (!p) return ''
  const n = norm(p)
  return assetCache.value[p] || assetCache.value[n] || assetCache.value[p.trim()] || ''
}

async function importAsset(type: 'bg' | 're') {
  const filePath = await SelectImageFile()
  if (!filePath ||!store.projectPath) return
  await SetProjectPath(store.projectPath)
  let newRel = ""
  try { newRel = await ImportAsset(filePath, type) as string } catch(e) { return }
  const n = norm(newRel)
  for (let i = 0; i < 6; i++) {
    try {
      const b64 = await GetImageBase64(n)
      if (b64) { assetCache.value[n]=b64; assetCache.value[newRel]=b64; break }
    } catch { await new Promise(r=>setTimeout(r,150)) }
  }
  await new Promise(r => setTimeout(r, 200))
  await refreshAssets()
}

async function importAudio(type: 'sfx' | 'voice' | 'music') {
  const filePath = await SelectAudioFile()
  if (!filePath ||!store.projectPath) return
  await SetProjectPath(store.projectPath)
  await ImportAudioAsset(filePath, type)
  await new Promise(r => setTimeout(r, 150))
  await refreshAudio()
}

async function deleteAsset(asset: string) {
  if (!confirm(`Usunąć ${norm(asset).replace('images/', '')}?`)) return
  await SetProjectPath(store.projectPath!)
  await DeleteAsset(norm(asset))
  const n = norm(asset)
  delete assetCache.value[asset]
  delete assetCache.value[n]
  delete assetCache.value[n.trim()]
  await refreshAssets()
}

async function deleteAudioAsset(asset: string) {
  if (!confirm(`Usunąć ${norm(asset)}?`)) return
  await SetProjectPath(store.projectPath!)
  await DeleteAudioAsset(norm(asset))
  await refreshAudio()
}

function toggleSection(t: 'bg'|'re'|'av'|'sfx'|'vo'|'mu') { (sectionsOpen.value as any)[t] =!(sectionsOpen.value as any)[t] }

function selectDay(day: string) {
  store.currentDay = day
  store.currentSceneId = store.days[day]?.[0]?.Id || null
}

// === DOWOLNE NAZWY DNI - FIX ===
function handleAddDay() {
  const raw = prompt('Nazwa nowego dnia? np. wesele_u_basi, prolog, final', `dzien_${Object.keys(store.days).length+1}`)
  if (!raw) return
  const id = toSafeDayId(raw)
  if (!id) { alert('Nieprawidłowa nazwa'); return }
  if (id.length < 2) { alert('Min 2 znaki'); return }
  if ((store as any).days[id]) { alert(`Dzień "${id}" już istnieje!`); return }
  store.addDay(id)
  store.saveProject()
}

async function handleDeleteDay(e: Event, day: string) {
  e.stopPropagation()
  if ((store.dayFileList?.length?? 0) <= 1) { alert('Musisz zostawić minimum 1 dzień!'); return }
  if (!confirm(`Na pewno usunąć dzień "${day}"? Plik Data/${day}.json zostanie skasowany.`)) return
  await store.deleteDay(day)
  await store.saveProject()
}

async function handleRenameDay(e: Event, oldId: string) {
  e.stopPropagation()
  const raw = prompt(`Zmień nazwę dnia "${oldId}" na:`, oldId)
  if (!raw) return
  const newId = toSafeDayId(raw)
  if (!newId) { alert('Nieprawidłowa nazwa'); return }
  if (newId === oldId) return
  if ((store as any).days[newId]) { alert(`Dzień "${newId}" już istnieje!`); return }

  if (!confirm(`Przemianować "${oldId}" -> "${newId}"?\nZaktualizuję wszystkie NextDayId.`)) return

  // 1. Skopiuj dane
  const scenes = (store as any).days[oldId]
  if (!scenes) return
  // zaktualizuj NextDayId w scenach które wskazywały na stary dzień
  Object.values((store as any).days).forEach((scList: any) => {
    (scList as any[]).forEach((s: any) => {
      if (s.NextDayId === oldId) s.NextDayId = newId
      if (s.NextDay === oldId) s.NextDay = newId
      if (s.Transfers?.nextDay === oldId) s.Transfers.nextDay = newId
    })
  })
  // zaktualizuj startDay w meta
  if ((store as any).meta?.startDay === oldId) (store as any).meta.startDay = newId

  (store as any).days[newId] = scenes
  delete (store as any).days[oldId]

  if (store.currentDay === oldId) store.currentDay = newId

  // 2. Skasuj stary plik
  if (store.projectPath) {
    try { await DeleteFile(store.projectPath, `Data/${oldId}.json`) } catch {}
  }
  await store.saveProject()
}

async function handleDuplicateDay(e: Event, srcId: string) {
  e.stopPropagation()
  const base = `${srcId}_kopia`
  let newId = base
  let c = 2
  while ((store as any).days[newId]) newId = `${base}_${c++}`

  const raw = prompt(`Duplikuj "${srcId}" jako:`, newId)
  if (!raw) return
  const id = toSafeDayId(raw)
  if (!id) return
  if ((store as any).days[id]) { alert(`Dzień "${id}" już istnieje!`); return }

  const cloned = JSON.parse(JSON.stringify((store as any).days[srcId]))
  // nowe Day number
  const dayNum = Object.keys((store as any).days).length + 1
  cloned.forEach((s: any) => { s.Day = dayNum })

  ;(store as any).days[id] = cloned
  store.currentDay = id
  store.currentSceneId = cloned[0]?.Id || null
  await store.saveProject()
}

onMounted(async () => {
  if (store.projectPath) {
    await SetProjectPath(store.projectPath)
  }
  await refreshAssets()
})

watch(() => store.projectPath, async (newPath) => {
  if (newPath) {
    await SetProjectPath(newPath)
  }
  await refreshAssets()
})

watch(() => store.ui.showAvatarEditor, async (open) => {
  if (!open) {
    await refreshAssets()
  }
})
</script>

<template>
  <aside v-if="store.meta" class="side-panel">
    <div class="side-scroll">
      <div class="panel-section highlight">
        <div class="label">PROJEKT JANUSZA</div>
        <div class="project-name" :title="store.meta?.gameName?? ''">{{ store.meta?.gameName?? '...' }}</div>
        <button @click="store.ui.showAvatarEditor = true" class="btn-janusz">
          <span class="icon">🧠</span>
          <span class="text">
            <b>KONFIGURUJ JANUSZA</b>
            <small>{{ avatarRulesCount }} rules • {{ avatarCount }} avatars</small>
          </span>
        </button>
      </div>

      <div class="panel-section days-section">
        <div class="section-header">
          <h4>DNI [{{ store.dayFileList?.length?? 0 }}]</h4>
          <button @click="handleAddDay" class="btn-add-day">+ Dzień</button>
        </div>
        <div class="days-list">
          <div v-for="day in (store as any).daysList" :key="day" :class="['day-item', { active: store.currentDay === day }]" @click="selectDay(day)">
            <div class="day-left">
              <span class="day-name">{{ day }}</span>
              <span class="day-count">{{ (store as any).days[day]?.length?? 0 }}</span>
            </div>
            <div class="day-right">
              <button class="btn-mini" title="Duplikuj" @click="handleDuplicateDay($event, day)">⎘</button>
              <button class="btn-mini" title="Zmień nazwę" @click="handleRenameDay($event, day)">✎</button>
              <button class="btn-mini del" title="Usuń" @click="handleDeleteDay($event, day)">✕</button>
            </div>
          </div>
          <div v-if="!(store as any).daysList?.length" class="empty">Brak dni</div>
        </div>
        <div class="hint">Nazwy dowolne: wesele, prolog, final_boss. Bez spacji - zamienią się na _</div>
      </div>

      <div class="panel-section">
        <div class="label">ASSETY [{{ totalFiles }}]</div>
        <div class="asset-buttons">
          <button @click="importAsset('bg')" class="btn-asset">+ BG</button>
          <button @click="importAsset('re')" class="btn-asset">+ RE</button>
          <button @click="refreshAssets" class="btn-asset">↻ Odśwież</button>
        </div>
        <div class="asset-buttons" style="margin-top:6px">
          <button @click="importAudio('sfx')" class="btn-asset">+ SFX</button>
          <button @click="importAudio('voice')" class="btn-asset">+ VOICE</button>
          <button @click="importAudio('music')" class="btn-asset">+ MUSIC</button>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('bg')"><span class="arrow">{{ sectionsOpen.bg? '▼':'▶' }}</span><span>TŁA</span><span class="count">[{{ backgroundAssets.length }}]</span></div>
          <div v-if="sectionsOpen.bg" class="asset-list">
            <div v-for="a in backgroundAssets" :key="a" class="asset-item"><img v-if="getAssetUrl(a)" :src="getAssetUrl(a)" /><div v-else class="thumb-placeholder"></div><span class="name">{{ norm(a).replace('images/','') }}</span><button @click.stop="deleteAsset(a)" class="btn-del">✕</button></div>
            <div v-if="!backgroundAssets.length" class="empty">Brak teł</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('re')"><span class="arrow">{{ sectionsOpen.re? '▼':'▶' }}</span><span>REAKCJE</span><span class="count">[{{ reactionAssets.length }}]</span></div>
          <div v-if="sectionsOpen.re" class="asset-list">
            <div v-for="a in reactionAssets" :key="a" class="asset-item"><img v-if="getAssetUrl(a)" :src="getAssetUrl(a)" /><div v-else class="thumb-placeholder"></div><span class="name">{{ norm(a).replace('images/','') }}</span><button @click.stop="deleteAsset(a)" class="btn-del">✕</button></div>
            <div v-if="!reactionAssets.length" class="empty">Brak reakcji</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('av')"><span class="arrow">{{ sectionsOpen.av? '▼':'▶' }}</span><span>AVATARY</span><span class="count">[{{ avatarAssets.length }}]</span></div>
          <div v-if="sectionsOpen.av" class="asset-list av-list">
            <div v-for="a in avatarAssets" :key="a" class="asset-item"><img v-if="getAssetUrl(a)" :src="getAssetUrl(a)" /><div v-else class="thumb-placeholder"></div><span class="name">{{ norm(a).replace('images/av_','').replace('images/','') }}</span><button @click.stop="deleteAsset(a)" class="btn-del">✕</button></div>
            <div v-if="!avatarAssets.length" class="empty">Brak avatarów</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('sfx')"><span class="arrow">{{ sectionsOpen.sfx? '▼':'▶' }}</span><span>DŹWIĘKI</span><span class="count">[{{ sfxAssets.length }}]</span></div>
          <div v-if="sectionsOpen.sfx" class="asset-list">
            <div v-for="a in sfxAssets" :key="a" class="asset-item audio-item">
              <span class="audio-icon">🔊</span>
              <span class="name">{{ norm(a).replace('sounds/sfx/','').replace('sounds/','') }}</span>
              <button @click.stop="deleteAudioAsset(a)" class="btn-del">✕</button>
            </div>
            <div v-if="!sfxAssets.length" class="empty">Brak dźwięków</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('vo')"><span class="arrow">{{ sectionsOpen.vo? '▼':'▶' }}</span><span>NARRATOR</span><span class="count">[{{ voiceAssets.length }}]</span></div>
          <div v-if="sectionsOpen.vo" class="asset-list">
            <div v-for="a in voiceAssets" :key="a" class="asset-item audio-item">
              <span class="audio-icon">🎙</span>
              <span class="name">{{ norm(a).replace('sounds/voice/','').replace('sounds/','') }}</span>
              <button @click.stop="deleteAudioAsset(a)" class="btn-del">✕</button>
            </div>
            <div v-if="!voiceAssets.length" class="empty">Brak nagrań</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('mu')"><span class="arrow">{{ sectionsOpen.mu? '▼':'▶' }}</span><span>MUZYKA</span><span class="count">[{{ musicAssets.length }}]</span></div>
          <div v-if="sectionsOpen.mu" class="asset-list">
            <div v-for="a in musicAssets" :key="a" class="asset-item audio-item">
              <span class="audio-icon">🎵</span>
              <span class="name">{{ norm(a).replace('sounds/music/','').replace('sounds/','') }}</span>
              <button @click.stop="deleteAudioAsset(a)" class="btn-del">✕</button>
            </div>
            <div v-if="!musicAssets.length" class="empty">Brak muzyki</div>
          </div>
        </div>

      </div>
    </div>
  </aside>
  <aside v-else class="side-panel">
    <div class="side-scroll"><div class="empty">Ładowanie projektu...</div></div>
  </aside>
</template>

<style scoped>
.side-panel{width:280px;min-width:280px;height:100%;background:#0D1117;border-right:1px solid #21262D;display:flex;flex-direction:column;overflow:hidden;box-sizing:border-box}
.side-scroll{flex:1;overflow-y:auto;overflow-x:hidden;display:flex;flex-direction:column;gap:16px;padding:12px;box-sizing:border-box}
.side-scroll::-webkit-scrollbar{width:6px}
.side-scroll::-webkit-scrollbar-thumb{background:#30363D;border-radius:3px}
.panel-section{flex-shrink:0;display:flex;flex-direction:column;gap:8px}
.panel-section.highlight{background:rgba(212,165,116,0.08);border:1px solid rgba(212,165,116,0.25);border-radius:8px;padding:10px}
.label{font-size:10px;font-weight:700;letter-spacing:1px;color:#7D8590}
.project-name{font-size:12px;font-family:monospace;color:#E6EDF3;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.btn-janusz{margin-top:6px;width:100%;display:flex;gap:10px;align-items:center;background:#161B22;border:1px solid #D4A574;border-radius:6px;padding:10px 12px;color:#E6EDF3;cursor:pointer;text-align:left;box-sizing:border-box}
.btn-janusz:hover{background:#211E12}
.days-section{padding-bottom:8px;border-bottom:1px solid #21262D}
.section-header{display:flex;justify-content:space-between;align-items:center}
.section-header h4{margin:0;font-size:11px;color:#D4A574;letter-spacing:1px}
.btn-add-day{background:#21262D;border:1px solid #30363D;color:#E6EDF3;padding:4px 8px;border-radius:4px;font-size:10px;cursor:pointer;font-weight:700}
.btn-add-day:hover{border-color:#D4A574;color:#D4A574}
.days-list{display:flex;flex-direction:column;gap:4px;margin-top:6px}
.day-item{display:flex;justify-content:space-between;align-items:center;padding:6px 8px;background:#161B22;border-radius:4px;font-size:12px;cursor:pointer;border-left:2px solid transparent;gap:8px;flex-shrink:0}
.day-item:hover{border-left-color:#D4A574}.day-item.active{background:#211E12;border-left-color:#D4A574}
.day-left{display:flex;align-items:center;gap:6px;min-width:0}
.day-name{font-family:monospace;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.day-count{background:#000;padding:1px 6px;border-radius:10px;font-size:10px;color:#7D8590;flex:0 0 auto}
.day-right{display:flex;align-items:center;gap:2px;flex:0 0 auto}
.btn-mini{width:20px;height:20px;background:#21262D;border:1px solid #30363D;color:#7D8590;border-radius:3px;cursor:pointer;font-size:10px;display:flex;align-items:center;justify-content:center}
.btn-mini:hover{background:#30363D;color:#E6EDF3}
.btn-mini.del:hover{background:#3a1212;border-color:#7f1d1d;color:#fca5a5}
.hint{font-size:9px;color:#484F58;margin-top:4px;font-style:italic}
.asset-buttons{display:grid;grid-template-columns:1fr 1fr 1fr;gap:6px}
.btn-asset{height:34px;background:#21262D;border:1px solid #30363D;color:#cbd5e1;border-radius:4px;font-size:11px;font-weight:700;cursor:pointer;display:flex;align-items:center;justify-content:center}
.btn-asset:hover{border-color:#8B949E;color:white;background:#30363D}
.asset-count{font-size:10px;color:#484F58;text-align:center}
.asset-category{margin-top:4px;flex-shrink:0}
.category-header{display:flex;gap:6px;align-items:center;padding:6px 8px;background:#161B22;border:1px solid #21262D;border-radius:4px;cursor:pointer;font-size:11px;font-weight:700;color:#E6EDF3}
.count{margin-left:auto;color:#7D8590;font-weight:400}
.asset-list{display:flex;flex-direction:column;gap:4px;margin-top:6px;max-height:160px;overflow-y:auto;flex-shrink:0}
.asset-list.av-list{max-height:200px}
.asset-item{display:flex;align-items:center;gap:6px;background:#161B22;border-radius:4px;padding:4px;flex-shrink:0}
.asset-item img{width:28px;height:28px;object-fit:cover;border-radius:3px;flex-shrink:0}
.thumb-placeholder{width:28px;height:28px;background:#21262D;border-radius:3px;flex-shrink:0}
.audio-item .audio-icon{width:28px;height:28px;display:flex;align-items:center;justify-content:center;background:#21262D;border-radius:3px;font-size:12px;flex-shrink:0}
.name{flex:1;font-size:10px;font-family:monospace;color:#94a3b8;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.btn-del{width:18px;height:18px;flex:0 0 18px;background:#dc2626;color:white;border:0;border-radius:3px;cursor:pointer;font-size:10px;display:flex;align-items:center;justify-content:center;line-height:1}
.btn-del:hover{background:#ef4444}
.empty{font-size:11px;color:#484F58;font-style:italic;padding:8px;text-align:center}
</style>
