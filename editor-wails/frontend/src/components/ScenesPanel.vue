<script setup lang="ts">
import { computed, ref } from 'vue'
import { useProjectStore } from '../stores/projectStore'

const store = useProjectStore()

const scenes = computed(() => store.currentDayScenes)
const currentDay = computed(() => store.currentDay)

function toSafeId(raw: string): string {
  let s = raw.trim().toLowerCase().replace(/\s+/g,'_').replace(/[^a-z0-9_\-]/g,'').replace(/_+/g,'_').replace(/^_+|_+$/g,'')
  return s.slice(0,40)
}

const newSceneName = ref('')

function isEnd(s: any): boolean {
  return !!s.IsEndDay || s.Type === 'end_of_day' || (s.Id||'').startsWith('koniec_')
}

function addNormalScene() {
  let raw = newSceneName.value.trim()
  if (!raw) raw = `scena_${Date.now().toString(36)}`
  const id = toSafeId(raw)
  if (!id || id.length < 2) { alert('Min 2 znaki'); return }
  if (scenes.value.some(x=>x.Id===id)) { alert(`ID "${id}" już istnieje w ${currentDay.value}`); return }

  store.addSceneToCurrentDay(id, {
    SceneTitle: raw,
    Background: 'images/bg_tutorial.webp',
    Text: '',
    Type: 'normal',
    IsEndDay: false
  })
  newSceneName.value = ''
  store.saveProject()
}

function addEndDayScene() {
  const base = currentDay.value
  const existing = scenes.value.filter(s=>isEnd(s)).map(s=>s.Id)
  let idx = 1
  let endId = `koniec_${base}`
  if (existing.includes(endId)) {
    idx = 2
    while (existing.includes(`koniec_${base}_${idx}`)) idx++
    endId = `koniec_${base}_${idx}`
  }
  if (scenes.value.some(x=>x.Id===endId)) { alert(`ID "${endId}" już istnieje`); return }

  store.addSceneToCurrentDay(endId, {
    SceneTitle: `KONIEC ${base.toUpperCase()} ${idx>1?idx:''}`.trim(),
    Background: 'images/bg_tutorial.webp',
    Text: `Koniec ${base}.`,
    IsEndDay: true,
    Type: 'end_of_day',
    NextDayId: undefined,
    NextDay: undefined,
    Transfers: { keep_flags: [], keep_stats: store.statsSystem.stats.map(s=>s.name), summary_text: '' },
    Choices: [{ Text:'Zakończ', Next:'END', Stats:{} } as any]
  })
  store.saveProject()
}

function move(idx: number, dir: -1|1) {
  store.moveSceneInCurrentDay(idx, dir)
  store.saveProject()
}
function dup(id: string) {
  store.duplicateScene(id)
  store.saveProject()
}
function del(id: string) {
  store.deleteScene(id)
  store.saveProject()
}
</script>

<template>
  <div class="scenes-panel">
    <div class="header">
      <div class="title-row">
        <span class="title">SCENY: {{ currentDay }}</span>
        <span class="count">[{{ scenes.length }}]</span>
      </div>
      <div class="actions">
        <input v-model="newSceneName" placeholder="id nowej sceny np. wesele" @keydown.enter="addNormalScene" class="input-new" />
        <div class="btns">
          <button @click="addNormalScene" class="btn">+ Scena</button>
          <button @click="addEndDayScene" class="btn end">+ Koniec dnia</button>
        </div>
      </div>
    </div>

    <div class="list">
      <div v-for="(s, i) in scenes" :key="s.Id" :class="['scene-item', { active: store.currentSceneId===s.Id, end: isEnd(s) }]" @click="store.selectScene(s.Id)" :title="s.Id">
        <div class="top-row">
          <span class="idx">{{ i }}</span>
          <span class="id-full" :title="s.Id">{{ s.Id }}</span>
        </div>
        <div class="bottom-row">
          <span v-if="isEnd(s)" class="tag">🌙 KONIEC</span>
          <span v-if="isEnd(s)" class="next-tag" :class="{ 'is-end': !(s as any).NextDayId }">-> {{ (s as any).NextDayId || 'END' }}</span>
          <span v-else class="next-tag normal">normalna</span>
          <div class="right">
            <button @click.stop="move(i,-1)" class="mini" title="w górę">▲</button>
            <button @click.stop="move(i,1)" class="mini" title="w dół">▼</button>
            <button @click.stop="dup(s.Id)" class="mini" title="duplikuj">⎘</button>
            <button @click.stop="del(s.Id)" class="mini del" title="usuń">✕</button>
          </div>
        </div>
      </div>
      <div v-if="scenes.length===0" class="empty-list">Brak scen - dodaj pierwszą</div>
    </div>

    <div class="hint">
      <div>• Kliknij scenę żeby edytować</div>
      <div>• 🌙 = koniec dnia, zawsze 1 choice -> END lub start kolejnego dnia</div>
      <div>• ID pełne widać po najechaniu myszką</div>
    </div>
  </div>
</template>

<style scoped>
.scenes-panel{
  background:#0D1117;
  height:100%;
  display:flex;
  flex-direction:column;
  min-height:0;
  width:100%;
  min-width:280px;
  max-width:420px;
  box-sizing:border-box;
  padding:12px;
  gap:10px;
  border-right:1px solid #21262D;
}
.header{display:flex;flex-direction:column;gap:10px}
.title-row{display:flex;align-items:center;gap:6px}
.title{color:#00FF94;font-size:12px;font-family:monospace;letter-spacing:1px;font-weight:800}
.count{color:#7D8590;font-size:11px;font-family:monospace}
.actions{display:flex;flex-direction:column;gap:6px}
.btns{display:flex;gap:6px}
.input-new{width:100%;height:32px;padding:0 10px;background:#161B22;border:1px solid #21262D;color:#E6EDF3;border-radius:6px;font-size:12px;font-family:monospace;box-sizing:border-box}
.input-new:focus{outline:none;border-color:#00FF94}
.input-new::placeholder{color:#484F58}
.btn{flex:1;height:32px;padding:0 10px;background:#21262D;border:1px solid #30363D;color:#E6EDF3;border-radius:6px;font-size:11px;font-weight:700;cursor:pointer;white-space:nowrap}
.btn:hover{border-color:#00FF94;background:#2A313C}
.btn.end{background:rgba(251,191,36,0.15);border-color:#fbbf24;color:#fbbf24}
.btn.end:hover{background:rgba(251,191,36,0.25)}
.list{flex:1;overflow-y:auto;overflow-x:hidden;display:flex;flex-direction:column;gap:6px;min-height:0;padding-right:2px}
.scene-item{
  display:flex;
  flex-direction:column;
  gap:4px;
  padding:8px 8px;
  background:#161B22;
  border-radius:8px;
  border:1px solid #21262D;
  border-left:3px solid transparent;
  cursor:pointer;
  transition:all 0.15s;
}
.scene-item:hover{border-color:#30363D;background:#1C2129}
.scene-item.active{background:#14261E;border-color:#00FF94;border-left-color:#00FF94}
.scene-item.end{border-left-color:#fbbf24;background:#211E0E}
.scene-item.end.active{background:#2A2310;border-color:#fbbf24}
.top-row{display:flex;align-items:flex-start;gap:6px;width:100%;min-width:0}
.idx{font-size:10px;color:#484F58;flex:0 0 16px;padding-top:2px;font-family:monospace}
.id-full{
  font-family:monospace;
  font-size:12px;
  color:#E6EDF3;
  word-break:break-all;
  white-space:normal;
  line-height:1.3;
  flex:1;
  min-width:0;
  font-weight:600;
}
.scene-item.active .id-full{color:#00FF94}
.scene-item.end.active .id-full{color:#fbbf24}
.bottom-row{display:flex;align-items:center;gap:6px;width:100%;flex-wrap:wrap;padding-left:22px;box-sizing:border-box}
.tag{background:#fbbf24;color:#000;font-size:8px;padding:2px 5px;border-radius:4px;font-weight:800;letter-spacing:0.5px;flex:0 0 auto}
.next-tag{font-size:10px;color:#7D8590;font-family:monospace;background:#0D1117;padding:2px 6px;border-radius:4px;flex:0 0 auto;border:1px solid #21262D}
.next-tag.is-end{color:#f85149;border-color:#3A1A20;background:#1A0F0F}
.next-tag.normal{color:#484F58}
.right{display:flex;gap:3px;margin-left:auto;flex:0 0 auto}
.mini{width:22px;height:22px;background:#21262D;border:1px solid #30363D;color:#7D8590;border-radius:4px;cursor:pointer;font-size:10px;display:flex;align-items:center;justify-content:center;transition:all 0.1s}
.mini:hover{background:#30363D;color:#fff;border-color:#484F58}
.mini.del:hover{background:#3A1212;border-color:#7f1d1d;color:#fca5a5}
.empty-list{padding:20px;text-align:center;color:#484F58;font-size:11px;border:1px dashed #21262D;border-radius:8px}
.hint{font-size:10px;color:#484F58;line-height:1.5;border-top:1px solid #21262D;padding-top:10px;display:flex;flex-direction:column;gap:2px}
</style>
