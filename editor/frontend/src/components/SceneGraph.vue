<script setup lang="ts">
import { computed, ref } from 'vue'
import { useProjectStore } from '../stores/projectStore'

const store = useProjectStore()
const isCollapsed = ref(false)
const zoom = ref(1)
const hideShortcuts = ref(true)
const wiggly = ref(true)

const scenes = computed(() => store.currentDayScenes || [])
const currentId = computed(() => store.currentSceneId)
const startId = computed(() => (store as any).meta?.startScene || 'start')
const isEndNode = (id:string, s?:any) => id.startsWith('koniec_') || (s && (s.IsEndDay || s.Type==='end_of_day'))

const rawEdges = computed(() => {
  const list:any[] = []
  const ids = new Set(scenes.value.map((s:any)=>s.Id))
  for (const s of scenes.value as any[]) {
    if(isEndNode(s.Id, s)) continue
    for (let i=0;i<(s.Choices||[]).length;i++){
      const c = (s.Choices as any[])[i]
      const next = (c.Next || '').toString().trim()
      if (!next || next === 'END' || !ids.has(next)) continue
      list.push({ from: s.Id, to: next, label: (c.Text||'').trim() || `Wybór ${i+1}` })
    }
  }
  return list
})
const filteredEdges = computed(() => {
  if(!hideShortcuts.value) return rawEdges.value
  const hasIntermediate = rawEdges.value.some(e => !isEndNode(e.from) && isEndNode(e.to))
  if(!hasIntermediate) return rawEdges.value
  return rawEdges.value.filter(e => {
    if(e.from===startId.value && isEndNode(e.to)){
      const others = rawEdges.value.filter(x=>x.from===startId.value && !isEndNode(x.to))
      if(others.length>0) return false
    }
    return true
  })
})
const groupedEdges = computed(() => {
  const map = new Map<string, {from:string, to:string, count:number, labels:string[]}>()
  for(const e of filteredEdges.value){
    const key = e.from + '::' + e.to
    if(!map.has(key)) map.set(key, { from: e.from, to: e.to, count: 0, labels: [] })
    const g = map.get(key)!; g.count++; g.labels.push(e.label)
  }
  return Array.from(map.values())
})

const layout = computed(() => {
  const map = new Map<string, {x:number, y:number, level:number}>()
  const visited = new Set<string>()
  const levels: Map<number, string[]> = new Map()
  const queue: {id:string, level:number}[] = [{ id: startId.value, level: 0 }]
  visited.add(startId.value)
  while(queue.length){
    const {id, level} = queue.shift()!
    if(isEndNode(id)) continue
    if(!levels.has(level)) levels.set(level, [])
    levels.get(level)!.push(id)
    const outs = [...new Set(groupedEdges.value.filter(e=>e.from===id).map(e=>e.to).filter(to=>!isEndNode(to)))].sort()
    for(const nid of outs){ if(!visited.has(nid)){ visited.add(nid); queue.push({id:nid, level: level+1}) } }
  }
  for(const s of scenes.value as any[]){
    if(isEndNode(s.Id, s)) continue
    if(!visited.has(s.Id)){
      const maxLevel = Math.max(...Array.from(levels.keys()), 0)
      if(!levels.has(maxLevel+1)) levels.set(maxLevel+1, [])
      levels.get(maxLevel+1)!.push(s.Id); visited.add(s.Id)
    }
  }
  const endNodes = (scenes.value as any[]).filter(s=>isEndNode(s.Id, s)).map(s=>s.Id).sort()
  if(endNodes.length>0){
    const maxLevel = Math.max(...Array.from(levels.keys()), 0)
    levels.set(maxLevel+1, endNodes)
  }
  levels.forEach((ids, lvl) => {
    ids.sort()
    ids.forEach((id, idx) => {
      const total = ids.length
      const spacingY = 70
      const startY = 14 + ((3 - total) * spacingY / 2)
      map.set(id, { x: 20 + lvl * 155, y: Math.max(8, startY + idx * spacingY), level: lvl })
    })
  })
  return map
})

function getPos(id:string){ return layout.value.get(id) || {x:20, y:8, level:0} }

function getOffsets(fromId:string, toId:string){
  const outs = groupedEdges.value.filter(e=>e.from===fromId).sort((a,b)=>a.to.localeCompare(b.to))
  const ins = groupedEdges.value.filter(e=>e.to===toId).sort((a,b)=>a.from.localeCompare(b.from))
  const outIdx = outs.findIndex(e=>e.to===toId)
  const inIdx = ins.findIndex(e=>e.from===fromId)
  // MAX rozproszenie - każda linia z innej wysokości
  const outOff = outs.length>1 ? (outIdx - (outs.length-1)/2) * 16 : 0
  const inOff = ins.length>1 ? (inIdx - (ins.length-1)/2) * 16 : 0
  return { outOff, inOff, outIdx }
}

function getPath(fromId:string, toId:string){
  const a = getPos(fromId)
  const b = getPos(toId)
  const { outOff, inOff, outIdx } = getOffsets(fromId, toId)
  // Wyjście z różnych wysokości noda - to naprawia Twoje zielone strzałki które się zlewały
  const x1 = a.x + 105
  const y1 = a.y + 19 + outOff
  const x2 = b.x
  const y2 = b.y + 19 + inOff
  if(!wiggly.value){
    const midX = (x1 + x2)/2
    return `M ${x1} ${y1} C ${midX} ${y1}, ${midX} ${y2}, ${x2} ${y2}`
  }
  const dx = x2 - x1
  const dy = y2 - y1
  const hash = (fromId.charCodeAt(0) + toId.charCodeAt(0) + outIdx) % 3
  if(Math.abs(dy) < 15){
    const wave = (hash - 1) * 28
    return `M ${x1} ${y1} C ${x1 + dx*0.35} ${y1+wave}, ${x2 - dx*0.35} ${y2-wave}, ${x2} ${y2}`
  } else {
    return `M ${x1} ${y1} C ${x1 + dx*0.45} ${y1 + (hash-1)*6}, ${x2 - dx*0.45} ${y2 - (hash-1)*6}, ${x2} ${y2}`
  }
}
function getMid(fromId:string, toId:string){
  const a = getPos(fromId); const b = getPos(toId)
  return { x: (a.x + b.x)/2 + 52, y: (a.y + b.y)/2 + 19 }
}
function getEdgeColor(fromId:string){
  if(fromId===startId.value) return '#22c55e'
  if(fromId==='bar') return '#38bdf8'
  if(fromId==='parkiet') return '#a78bfa'
  if(isEndNode(fromId)) return '#fbbf24'
  let h=0; for(let i=0;i<fromId.length;i++) h+=fromId.charCodeAt(i)
  return ['#34d399','#fb7185','#fbbf24','#a78bfa','#38bdf8','#f472b6'][h%6]
}
function getBounds(){
  let maxX = 580, maxY = 240
  layout.value.forEach(v => { if(v.x+115 > maxX) maxX = v.x+115; if(v.y+42 > maxY) maxY = v.y+42 })
  return { w: maxX + 30, h: maxY + 30 }
}
function select(id:string){ store.selectScene(id) }
</script>

<template>
  <div class="graph-wrap" :class="{ collapsed: isCollapsed }">
    <div class="graph-header" @click="isCollapsed = !isCollapsed">
      <span class="g-title">GRAF: {{ store.currentDay }} [{{ scenes.length }}] • MAX COMPACT</span>
      <span class="g-path">{{ filteredEdges.length }} tras • koniec na końcu</span>
      <div class="g-tools" @click.stop>
        <label class="chk"><input type="checkbox" v-model="hideShortcuts" /> skróty</label>
        <label class="chk"><input type="checkbox" v-model="wiggly" /> wygibasy</label>
        <button @click="zoom = Math.max(0.4, zoom-0.1)" class="zbtn">-</button>
        <span class="zval">{{ Math.round(zoom*100) }}%</span>
        <button @click="zoom = Math.min(1.6, zoom+0.1)" class="zbtn">+</button>
      </div>
      <span class="g-toggle">{{ isCollapsed ? 'ZWIN' : 'ROZWIN' }}</span>
    </div>
    <div v-if="!isCollapsed" class="graph-canvas">
      <div class="graph-inner" :style="{ width: getBounds().w + 'px', height: getBounds().h + 'px', transform: 'scale(' + zoom + ')', transformOrigin: '0 0' }">
        <div class="grid-bg"></div>
        <svg class="edges-layer" :width="getBounds().w" :height="getBounds().h">
          <defs>
            <marker id="arrow-green" viewBox="0 0 10 10" refX="8" refY="4" markerWidth="8" markerHeight="8" orient="auto-start-reverse"><path fill="#22c55e" d="M 0 0 L 10 5 L 0 10 z"/></marker>
            <marker id="arrow-blue" viewBox="0 0 10 10" refX="8" refY="4" markerWidth="8" markerHeight="8" orient="auto-start-reverse"><path fill="#38bdf8" d="M 0 0 L 10 5 L 0 10 z"/></marker>
            <marker id="arrow-purple" viewBox="0 0 10 10" refX="8" refY="4" markerWidth="8" markerHeight="8" orient="auto-start-reverse"><path fill="#a78bfa" d="M 0 0 L 10 5 L 0 10 z"/></marker>
            <marker id="arrow-yellow" viewBox="0 0 10 10" refX="8" refY="4" markerWidth="8" markerHeight="8" orient="auto-start-reverse"><path fill="#fbbf24" d="M 0 0 L 10 5 L 0 10 z"/></marker>
            <marker id="arrow-gray" viewBox="0 0 10 10" refX="8" refY="4" markerWidth="8" markerHeight="8" orient="auto-start-reverse"><path fill="#6b7280" d="M 0 0 L 10 5 L 0 10 z"/></marker>
          </defs>
          <g v-for="(e,i) in groupedEdges" :key="i">
            <path :d="getPath(e.from, e.to)" class="ink-path" :class="{ active: currentId===e.from || currentId===e.to }" :style="{ stroke: getEdgeColor(e.from) }" :marker-end="'url(#arrow-' + (e.from===startId ? 'green' : e.from==='bar' ? 'blue' : e.from==='parkiet' ? 'purple' : isEndNode(e.from) ? 'yellow' : 'gray') + ')'" />
            <g :transform="'translate(' + getMid(e.from, e.to).x + ',' + getMid(e.from, e.to).y + ')'">
              <rect x="-10" y="-7" width="20" height="14" rx="7" class="badge-bg" :style="{ stroke: getEdgeColor(e.from) }" />
              <text text-anchor="middle" y="2.5" class="badge-text">{{ e.count }}</text>
            </g>
          </g>
        </svg>
        <div v-for="s in scenes" :key="s.Id" class="node" :class="{ active: currentId===s.Id, end: isEndNode(s.Id, s), start: s.Id===startId }" :style="{ left: getPos(s.Id).x + 'px', top: getPos(s.Id).y + 'px', borderLeftColor: getEdgeColor(s.Id) }" @click="select(s.Id)">
          <div class="n-top"><span class="n-id">{{ s.Id }}</span><span class="dot" :style="{ background: getEdgeColor(s.Id) }"></span></div>
          <span class="n-count">{{ filteredEdges.filter(e=>e.from===s.Id).length }} • lvl {{ getPos(s.Id).level }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.graph-wrap{ background:#0D1117; border:1px solid #21262D; border-bottom:none; border-radius:8px 8px 0 0; display:flex; flex-direction:column; flex-shrink:0; }
.graph-wrap.collapsed{ border-radius:8px; border-bottom:1px solid #21262D; }
.graph-header{ height:24px; display:flex; align-items:center; gap:6px; padding:0 8px; background:#161B22; border-bottom:1px solid #21262D; cursor:pointer; user-select:none; }
.g-title{ font-family:'JetBrains Mono',monospace; font-size:9px; font-weight:800; color:#D4A574; }
.g-path{ font-family:'JetBrains Mono',monospace; font-size:7px; color:#7D8590; margin-left:auto; }
.g-tools{ display:flex; align-items:center; gap:5px; margin-left:6px; }
.chk{ font-family:'JetBrains Mono',monospace; font-size:7px; color:#9CA3AF; display:flex; align-items:center; gap:2px; cursor:pointer; }
.zbtn{ width:16px; height:16px; background:#21262D; border:1px solid #30363D; color:#E6EDF3; border-radius:3px; cursor:pointer; font-size:10px; display:grid; place-items:center; }
.zval{ font-family:'JetBrains Mono',monospace; font-size:7px; color:#7D8590; min-width:24px; text-align:center; }
.g-toggle{ font-family:'JetBrains Mono',monospace; font-size:8px; color:#7D8590; margin-left:4px; }
.graph-canvas{ position:relative; height:180px; overflow:auto; background:#080A0F; }
.graph-canvas::-webkit-scrollbar{ width:5px; height:5px; } .graph-canvas::-webkit-scrollbar-thumb{ background:#21262D; border-radius:3px; }
.graph-inner{ position:relative; transition: transform 0.15s; padding-left:4px; }
.grid-bg{ position:absolute; inset:0; background-image: radial-gradient(#21262D 1px, transparent 1px); background-size:16px 16px; opacity:0.12; pointer-events:none; }
.edges-layer{ position:absolute; inset:0; pointer-events:none; overflow:visible; }
.ink-path{ fill:none; stroke-width:1.9px; opacity:0.9; }
.ink-path.active{ stroke-width:2.6px; opacity:1; filter: drop-shadow(0 0 6px currentColor); }
.badge-bg{ fill:#161B22; stroke-width:1px; }
.badge-text{ font-family:'JetBrains Mono',monospace; font-size:7px; font-weight:800; fill:#E6EDF3; }
.node{ position:absolute; width:105px; height:38px; background:#161B22; border:1px solid #30363D; border-left:3px solid #30363D; border-radius:7px; padding:4px 6px; cursor:pointer; display:flex; flex-direction:column; justify-content:center; transition: all 0.12s; box-shadow: 0 1px 6px rgba(0,0,0,0.4); z-index:2; }
.node:hover{ transform:translateY(-1px); z-index:5; }
.node.active{ background:#1E1A12; box-shadow: 0 0 0 1.5px currentColor; }
.node.end{ border-color:#fbbf24; }
.n-top{ display:flex; align-items:center; justify-content:space-between; }
.n-id{ font-family:'JetBrains Mono',monospace; font-size:8px; font-weight:800; color:#E6EDF3; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; max-width:70px; }
.n-count{ font-family:'JetBrains Mono',monospace; font-size:6px; color:#7D8590; margin-top:1px; }
.dot{ width:6px; height:6px; border-radius:50%; }
</style>
