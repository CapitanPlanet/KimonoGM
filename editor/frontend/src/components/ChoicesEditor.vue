<script setup lang="ts">
import { computed, onMounted, ref, watch, onBeforeUnmount } from 'vue'
import { useProjectStore } from '../stores/projectStore'
import { ListAudioAssets, GetAudioBase64 } from '../../wailsjs/go/main/App'

const store = useProjectStore()
const scene = computed(() => store.currentScene)
const statDefs = computed(() => store.statsSystem?.stats || [])

const isEndDay = computed(() => {
  const id = scene.value?.Id || ''
  return !!(scene.value?.IsEndDay || (scene.value as any)?.Type === 'end_of_day' || id.startsWith('koniec_'))
})

const currentDay = computed(() => store.currentDay)
const allDays = computed(() => (store as any).daysList || Object.keys((store as any).days || {}))
const currentDayScenes = computed(() => store.currentDayScenes || [])

const currentDayNormal = computed(() => currentDayScenes.value.filter((s:any) => !s.IsEndDay && s.Type !== 'end_of_day' && !(s.Id||'').startsWith('koniec_')))
const currentDayEnds = computed(() => currentDayScenes.value.filter((s:any) => s.IsEndDay || s.Type === 'end_of_day' || (s.Id||'').startsWith('koniec_')))

const otherDaysScenes = computed(() => {
  const map: Record<string, any[]> = {}
  allDays.value.forEach((day:string) => {
    if (day === currentDay.value) return
    const scenes = (store as any).days?.[day] || []
    map[day] = scenes
  })
  return map
})

const endDayTargets = computed(() => {
  const targets: { value: string; label: string; day: string; isEnd: boolean }[] = []
  targets.push({ value: 'END', label: '🏁 KONIEC GRY', day: '', isEnd: false })
  allDays.value.forEach((day:string) => {
    if (day === currentDay.value) return
    const scenes = (store as any).days?.[day] || []
    const start = scenes.find((s:any) => !s.IsEndDay && s.Type !== 'end_of_day' && !(s.Id||'').startsWith('koniec_')) || scenes[0]
    if (start) {
      targets.push({ value: start.Id, label: `📅 ${day} → ${start.Id}`, day, isEnd: false })
    }
  })
  return targets
})

const backgrounds = computed(() => store.backgroundAssets)
const reactions = computed(() => store.reactionAssets)

const sfxAssets = ref<string[]>([])
const audioCache = ref<Record<string, string>>({})
const playingFile = ref<string | null>(null)
let currentAudio: HTMLAudioElement | null = null

function toSafeId(raw: string): string {
  let s = raw.trim().toLowerCase().replace(/\s+/g,'_').replace(/[^a-z0-9_\-]/g,'').replace(/_+/g,'_').replace(/^_+|_+$/g,'')
  return s.slice(0,40)
}

function ensureEndDayIntegrity() {
  const sc = scene.value as any
  if (!sc || !isEndDay.value) return
  if (!sc.Transfers) sc.Transfers = { keep_flags: [], keep_stats: store.statsSystem.stats.map(s=>s.name), summary_text: '' }
  if (!sc.Transfers.keep_stats?.length) sc.Transfers.keep_stats = store.statsSystem.stats.map(s=>s.name)
  if (!sc.Choices) sc.Choices = []
  if (sc.Choices.length === 0) {
    sc.Choices = [{ id: crypto.randomUUID(), Text:'Zakończ', Next: 'END', Stats:{}, ReactionText:'', ReactionImage:'', SoundFile:'' }]
  }
  if (sc.Choices.length > 1) sc.Choices = [sc.Choices[0]]
}

async function refreshSfx() {
  if (!store.projectPath) return
  try { sfxAssets.value = await ListAudioAssets(store.projectPath, 'sfx') as any } catch {}
}
async function ensureAudio(file: string) {
  if (!file) return ''
  if (audioCache.value[file]) return audioCache.value[file]
  if (!store.projectPath) return ''
  try {
    const b64 = await GetAudioBase64(store.projectPath, file)
    audioCache.value[file] = b64
    return b64
  } catch { return '' }
}
async function togglePlay(file: string) {
  if (!file) return
  if (playingFile.value === file) { stopPlay(); return }
  stopPlay()
  const url = await ensureAudio(file)
  if (!url) return
  currentAudio = new Audio(url)
  currentAudio.volume = 0.7
  currentAudio.onended = () => { playingFile.value = null }
  playingFile.value = file
  await currentAudio.play().catch(()=>{})
}
function stopPlay() {
  if (currentAudio) { currentAudio.pause(); currentAudio.src = ''; currentAudio = null }
  playingFile.value = null
}
function updateScene(field: string, value: any) {
  store.updateCurrentScene(field, value)
}
function ensureStatsForChoice(choice: any) {
  if (!choice.Stats) choice.Stats = {}
  statDefs.value.forEach((s:any) => { if (choice.Stats[s.id] === undefined) choice.Stats[s.id] = 0 })
}
function setStat(choice: any, statId: string, e: Event) {
  const val = Number((e.target as HTMLInputElement).value) || 0
  if (!choice.Stats) choice.Stats = {}
  choice.Stats[statId] = val
}

// JEDEN SELECT dla końca dnia
function onEndDayTargetChange(e: Event) {
  const nextId = (e.target as HTMLSelectElement).value
  const sc = scene.value as any
  if (!sc) return
  if (!sc.Choices) sc.Choices = [{ id: crypto.randomUUID(), Text:'Zakończ', Next: 'END', Stats:{} }]
  sc.Choices[0].Next = nextId
  if (nextId === 'END') {
    sc.NextDayId = undefined
    sc.NextDay = undefined
  } else {
    for (const d of allDays.value) {
      const scenes = (store as any).days?.[d] || []
      if (scenes.some((s:any) => s.Id === nextId)) {
        sc.NextDayId = d
        sc.NextDay = d
        break
      }
    }
  }
  store.saveProject()
}

const idError = ref('')
function renameSceneId(e: Event){
  const input = e.target as HTMLInputElement
  let newId = toSafeId(input.value)
  if(!newId){ idError.value='ID nie może być puste'; input.value = store.currentScene?.Id || ''; return }
  if(newId.length < 2){ idError.value='Min 2 znaki'; return }
  const oldId = store.currentScene?.Id
  if(!oldId || oldId === newId) { idError.value=''; return }
  if(store.sceneIdsInCurrentDay.includes(newId)){ idError.value=`ID "${newId}" już istnieje w ${store.currentDay}`; input.value = oldId; return }
  store.currentScene!.Id = newId
  store.currentSceneId = newId
  Object.values(store.days as any).forEach((scenes:any) => {
    scenes.forEach((s:any) => { s.Choices?.forEach((c:any) => { if(c.Next === oldId){ c.Next = newId } }) })
  })
  if((store.meta as any)?.startScene === oldId && store.currentDay === (store as any).meta.startDay){ (store.meta as any).startScene = newId }
  store.saveProject()
  idError.value = ''
}

function toggleKeepStat(statId: string, e: Event) {
  if (!scene.value) return
  if (!(scene.value as any).Transfers) (scene.value as any).Transfers = { keep_flags: [], keep_stats: [], summary_text: "" }
  if (!(scene.value as any).Transfers.keep_stats) (scene.value as any).Transfers.keep_stats = []
  const checked = (e.target as HTMLInputElement).checked
  const arr = (scene.value as any).Transfers.keep_stats as string[]
  if (checked) { if (!arr.includes(statId)) arr.push(statId) }
  else { (scene.value as any).Transfers.keep_stats = arr.filter((x:string)=>x!==statId) }
  store.saveProject()
}

onMounted(() => {
  ensureEndDayIntegrity()
  scene.value?.Choices?.forEach((c: any) => { if (!c.id) c.id = crypto.randomUUID(); ensureStatsForChoice(c) })
  refreshSfx()
})
onBeforeUnmount(() => stopPlay())
watch(() => store.projectPath, refreshSfx)
watch(() => scene.value?.Id, () => { stopPlay(); refreshSfx(); ensureEndDayIntegrity(); scene.value?.Choices?.forEach((c: any) => ensureStatsForChoice(c)) })
watch(statDefs, () => { scene.value?.Choices?.forEach((c: any) => ensureStatsForChoice(c)) }, { deep: true })

function addChoice() {
  if (!scene.value || isEndDay.value) return
  if (!scene.value.Choices) scene.value.Choices = []
  const initStats: Record<string, number> = {}
  statDefs.value.forEach((s:any) => initStats[s.id] = 0)
  const firstEnd = currentDayEnds.value[0]?.Id
  const firstNormal = currentDayNormal.value.find((s:any) => s.Id !== scene.value!.Id)?.Id
  const target = firstEnd || firstNormal || scene.value.Id
  scene.value.Choices.push({ id: crypto.randomUUID(), Text: 'Nowy wybór', Next: target, Stats: initStats, ReactionText: '', ReactionImage: '', SoundFile: '' } as any)
  store.saveProject()
}
function deleteChoice(index: number) { scene.value?.Choices?.splice(index, 1); store.saveProject() }
</script>

<template>
  <div v-if="scene" class="choices-panel">
    <div class="panel-content">
      <div class="id-edit-block" :class="{ 'is-end': isEndDay }">
        <div class="field">
          <label>ID SCENY * <span class="hint-inline">(bez spacji)</span> <span v-if="isEndDay" class="end-tag">🌙 KONIEC DNIA</span></label>
          <input class="input-id" :value="scene.Id" @blur="renameSceneId" @keydown.enter="(e: any) => (e.target as HTMLInputElement).blur()" />
          <div v-if="idError" class="field-error">{{ idError }}</div>
        </div>
      </div>

      <h3>Edycja: {{ scene.Id }}</h3>
      <div class="field">
        <label>Tytuł sceny</label>
        <input :value="scene.SceneTitle" @input="updateScene('SceneTitle', ($event.target as HTMLInputElement).value)" />
      </div>
      <div class="field">
        <label>Tło</label>
        <select :value="scene.Background" @change="updateScene('Background', ($event.target as HTMLSelectElement).value)">
          <option value="">— Brak —</option>
          <option v-for="bg in backgrounds" :key="bg" :value="bg">{{ bg.replace('images/','') }}</option>
        </select>
      </div>
      <div class="field">
        <label>Tekst sceny</label>
        <textarea :value="scene.Text" @input="updateScene('Text', ($event.target as HTMLTextAreaElement).value)" rows="5"></textarea>
      </div>

      <!-- KONIEC DNIA - JEDEN SELECT -->
      <template v-if="isEndDay">
        <div class="end-day-box">
          <h4>🌙 TRANSPORT MIĘDZY DNIAMI</h4>
          <div class="field">
            <label>Co się stanie po tej scenie?</label>
            <select :value="(scene as any).Choices?.[0]?.Next || 'END'" @change="onEndDayTargetChange" class="big-select highlight">
              <option v-for="t in endDayTargets" :key="t.value" :value="t.value">{{ t.label }}</option>
            </select>
          </div>
          <div class="explain-box" :class="{ end: (scene as any).Choices?.[0]?.Next === 'END' }">
            <template v-if="(scene as any).Choices?.[0]?.Next === 'END'">
              <b>🏁 Koniec gry</b><br/>Po tej scenie gra się zakończy i pokaże ekran końcowy.
            </template>
            <template v-else>
              <b>➡️ Przejście do {{ (scene as any).NextDayId }}</b><br/>
              Po zakończeniu dnia gracz trafi do: <code>{{ (scene as any).Choices?.[0]?.Next }}</code><br/>
              <span class="small">To pierwsza scena dnia {{ (scene as any).NextDayId }}. Jeśli chcesz iść do innej sceny w tym dniu, zmień to w silniku, ale nie zalecamy.</span>
            </template>
          </div>

          <div class="field" style="margin-top:12px">
            <label>Co zachować do następnego dnia?</label>
            <div class="keep-grid">
              <label v-for="st in statDefs" :key="st.id" class="keep-check">
                <input type="checkbox" :checked="(scene as any).Transfers?.keep_stats?.includes(st.id)" @change="toggleKeepStat(st.id, $event)" />
                {{ st.name }}
              </label>
            </div>
          </div>
          <div class="field">
            <label>Podsumowanie dnia (pokazuje się przy przejściu)</label>
            <textarea :value="(scene as any).Transfers?.summary_text || ''" @input="(scene as any).Transfers.summary_text = ($event.target as HTMLTextAreaElement).value; store.saveProject()" rows="2" placeholder="np. Janusz przeżył dzień 1..."></textarea>
          </div>
        </div>
      </template>

      <!-- ZWYKŁA SCENA -->
      <template v-else>
        <div class="choices-header">
          <h4>WYBORY [{{ scene.Choices?.length || 0 }}]</h4>
          <button @click="addChoice" class="btn-add">+ Wybór</button>
        </div>
        <div class="choices-list">
          <div v-for="(choice, idx) in scene.Choices || []" :key="choice.id" class="choice-card">
            <div class="choice-top">
              <input v-model="choice.Text" placeholder="Tekst wyboru" class="choice-input" @change="store.saveProject()" />
              <button @click="deleteChoice(idx)" class="btn-delete">✕</button>
            </div>
            <div class="field">
              <label>Przejdź do →</label>
              <select v-model="choice.Next" @change="store.saveProject()" class="big-select">
                <optgroup :label="`📍 ${currentDay} - sceny`">
                  <option v-for="s in currentDayNormal" :key="s.Id" :value="s.Id">{{ s.Id }}</option>
                </optgroup>
                <optgroup :label="`🌙 ${currentDay} - koniec dnia`">
                  <option v-for="s in currentDayEnds" :key="s.Id" :value="s.Id">🌙 {{ s.Id }} → {{ (s as any).NextDayId || 'END' }}</option>
                </optgroup>
                <optgroup v-for="(scenes, day) in otherDaysScenes" :key="day" :label="`📅 ${day}`">
                  <option v-for="s in scenes" :key="s.Id" :value="s.Id">{{ s.Id }}</option>
                </optgroup>
              </select>
            </div>
            <div class="row-2">
              <div class="inline-field">
                <span>Reakcja:</span>
                <select v-model="choice.ReactionImage" @change="store.saveProject()">
                  <option value="">— Brak —</option>
                  <option v-for="re in reactions" :key="re" :value="re">{{ re.replace('images/','') }}</option>
                </select>
              </div>
              <div class="inline-field sfx-field">
                <span>🔊 SFX:</span>
                <select v-model="choice.SoundFile" @change="stopPlay(); store.saveProject()">
                  <option value="">— Brak —</option>
                  <option v-for="sfx in sfxAssets" :key="sfx" :value="sfx">{{ sfx.replace('sounds/sfx/','') }}</option>
                </select>
                <button v-if="choice.SoundFile" @click="togglePlay(choice.SoundFile)" class="btn-play" :class="{ playing: playingFile === choice.SoundFile }">{{ playingFile === choice.SoundFile? '■' : '▶' }}</button>
                <button v-if="choice.SoundFile" @click="choice.SoundFile=''; stopPlay(); store.saveProject()" class="btn-clear">✕</button>
              </div>
            </div>
            <div class="stats-title">Statystyki ({{ statDefs.length }})</div>
            <div class="stats-grid" :style="{ gridTemplateColumns: `repeat(${Math.min(Math.max(statDefs.length,2),4)}, 1fr)` }">
              <div v-for="st in statDefs" :key="st.id" class="stat">
                <label :title="st.id">{{ st.name }}</label>
                <input type="number" :value="choice.Stats?.[st.id]?? 0" @input="setStat(choice, st.id, $event); store.saveProject()" />
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
  <div v-else class="empty">Wybierz scenę</div>
</template>

<style scoped>
.choices-panel{background:#0D1117;height:100%;display:flex;flex-direction:column;min-height:0;width:100%;box-sizing:border-box}
.panel-content{padding:16px;overflow-y:auto;overflow-x:hidden;flex:1;min-height:0;display:flex;flex-direction:column;gap:14px;box-sizing:border-box}
.choices-panel h3{margin:0;color:#D4A574;font-size:11px;font-family:monospace;text-transform:uppercase;letter-spacing:1px;opacity:.9}
.id-edit-block{background:#161B22;border:1px solid #D4A574;border-radius:8px;padding:10px;margin-bottom:4px}
.id-edit-block.is-end{border-color:#fbbf24}
.input-id{font-family:monospace;font-weight:700;letter-spacing:0.5px}
.hint-inline{font-weight:400;color:#484F58;text-transform:none;letter-spacing:0}
.end-tag{background:#fbbf24;color:#000;font-size:9px;padding:2px 6px;border-radius:4px;margin-left:8px;font-weight:800}
.field-error{font-size:11px;color:#F85149;margin-top:4px}
.field{display:flex;flex-direction:column;gap:6px}
.field label{font-size:10px;color:#7D8590;text-transform:uppercase;letter-spacing:.6px;font-weight:700}
.field input,.field select,.field textarea{width:100%;box-sizing:border-box;padding:8px 10px;background:#161B22;border:1px solid #21262D;color:#E6EDF3;font-size:12px;border-radius:6px;color-scheme:dark}
.field input:focus,.field select:focus,.field textarea:focus{outline:none;border-color:#D4A574}
.field textarea{resize:vertical;line-height:1.5;min-height:80px}
.big-select{height:40px!important;font-size:12px!important;font-family:monospace!important;font-weight:600!important}
.big-select.highlight{border-color:#fbbf24!important;background:#211E0E!important;color:#fbbf24!important}
.end-day-box{background:rgba(251,191,36,0.08);border:1px dashed #fbbf24;border-radius:10px;padding:14px;display:flex;flex-direction:column;gap:12px}
.end-day-box h4{margin:0;color:#fbbf24;font-size:11px;letter-spacing:1px}
.explain-box{background:#0D1117;border:1px solid #21262D;border-radius:8px;padding:10px 12px;font-size:11px;color:#E6EDF3;line-height:1.5}
.explain-box.end{border-color:#fbbf24;background:#211E0E}
.explain-box b{color:#D4A574}
.explain-box.end b{color:#fbbf24}
.explain-box code{background:#161B22;padding:2px 6px;border-radius:4px;font-family:monospace;color:#D4A574;border:1px solid #21262D}
.explain-box .small{font-size:10px;color:#7D8590;margin-top:4px;display:block}
.keep-grid{display:grid;grid-template-columns:1fr 1fr;gap:6px;margin-top:4px}
.keep-check{font-size:11px;color:#E6EDF3;display:flex;align-items:center;gap:6px;background:#0D1117;padding:6px 8px;border-radius:4px;border:1px solid #21262D}
.choices-header{width:100%;box-sizing:border-box;display:flex;flex-direction:row;justify-content:space-between;align-items:center;gap:12px;min-height:38px;padding:0;padding-top:14px;margin-top:10px;border-top:1px solid #21262D}
.choices-header h4{margin:0;color:#E6EDF3;font-size:11px;letter-spacing:1px;line-height:1;white-space:nowrap}
.btn-add{flex:0 0 auto;height:30px;padding:0 14px;background:#D4A574;border:0;color:#000;font-size:12px;font-weight:800;border-radius:6px;cursor:pointer;white-space:nowrap}
.btn-add:hover{filter:brightness(1.1)}
.choices-list{display:flex;flex-direction:column;gap:10px;width:100%;box-sizing:border-box}
.choice-card{background:#161B22;border:1px solid #21262D;border-radius:8px;padding:10px;display:flex;flex-direction:column;gap:10px;box-sizing:border-box;width:100%}
.choice-top{display:flex;gap:8px;align-items:center;width:100%}
.choice-input{flex:1;min-width:0;height:34px;padding:0 10px;background:#0D1117;border:1px solid #2A313C;color:#fff;font-size:12px;border-radius:6px;box-sizing:border-box}
.choice-input:focus{outline:none;border-color:#D4A574}
.btn-delete{width:34px;height:34px;flex:0 0 34px;background:#2A1215;border:1px solid #3A1A20;color:#FF8A9B;border-radius:6px;cursor:pointer;font-size:14px;font-weight:700}
.btn-delete:hover{background:#3A1A20;color:#fff}
.row-2{display:grid;grid-template-columns:1fr 1fr;gap:8px;width:100%}
.inline-field{display:flex;align-items:center;gap:8px;background:#0D1117;border:1px solid #21262D;border-radius:6px;padding:0 8px;height:34px;box-sizing:border-box;min-width:0}
.inline-field span{font-size:10px;color:#7D8590;white-space:nowrap;flex:0 0 auto}
.inline-field select{flex:1;min-width:0;background:#0D1117;border:0;color:#E6EDF3;font-size:11px;font-family:monospace;outline:none;padding:0;color-scheme:dark}
.btn-clear{width:18px;height:18px;flex:0 0 18px;background:#21262D;border:0;border-radius:3px;color:#7D8590;cursor:pointer;font-size:10px;display:flex;align-items:center;justify-content:center}
.btn-clear:hover{background:#30363D;color:#fff}
.btn-play{width:26px;height:26px;flex:0 0 26px;background:#21262D;border:1px solid #30363D;border-radius:6px;color:#D4A574;cursor:pointer;font-size:11px;display:flex;align-items:center;justify-content:center}
.btn-play:hover{background:#30363D;border-color:#D4A574}
.btn-play.playing{background:#D4A574;color:#000;border-color:#D4A574;animation:pulse 1s infinite}
@keyframes pulse{0%{box-shadow:0 0 0 0 rgba(212,165,116,.4)}70%{box-shadow:0 0 0 6px rgba(212,165,116,0)}100%{box-shadow:0 0 0 0 rgba(212,165,116,0)}}
.stats-title{font-size:9px;color:#7D8590;text-transform:uppercase;letter-spacing:1px;margin-top:2px}
.stats-grid{display:grid;gap:8px;width:100%}
.stat{display:flex;flex-direction:column;gap:4px;min-width:0}
.stat label{font-size:10px;color:#7D8590;text-align:center;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;font-weight:700}
.stat input{width:100%;box-sizing:border-box;height:32px;background:#0D1117;border:1px solid #21262D;color:#E6EDF3;font-size:12px;border-radius:6px;text-align:center}
.stat input:focus{outline:none;border-color:#D4A574}
.empty{padding:60px 20px;text-align:center;color:#484F58;font-size:12px}
@media(max-width:600px){.row-2{grid-template-columns:1fr}.stats-grid{grid-template-columns:1fr 1fr!important}}
</style>
