<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{ projectPath?: string; statusMsg?: string; absurd: string }>()
const emit = defineEmits(['open-folder', 'cycle-absurd'])
const shortPath = computed(() => {
  if (!props.projectPath) return 'brak projektu'
  const parts = props.projectPath.split(/[\\/]/)
  if (parts.length > 3) return '…/' + parts.slice(-3).join('/')
  return props.projectPath
})
</script>

<template>
  <header class="top-neon">
    <div class="tube"></div>
    <div class="inner">
      <div class="left">
        <button class="path-chip" @click="emit('open-folder')" :title="projectPath">
          <svg viewBox="0 0 16 16" class="ico"><path d="M2 3.5A1.5 1.5 0 0 1 3.5 2h3.2a1 1 0 0 1 .8.4L8.5 4H12.5A1.5 1.5 0 0 1 14 5.5v7A1.5 1.5 0 0 1 12.5 14h-9A1.5 1.5 0 0 1 2 12.5v-9Z"/></svg>
          <span class="p">{{ shortPath }}</span>
        </button>
        <div class="logo">
          <button class="badge" @click="emit('cycle-absurd')">
            <span class="b-label">SEKCJA</span>
            <span class="b-text">{{ absurd }}</span>
          </button>
          <h1>KIMONO<span class="ver">editor 1.0</span><span class="vhs">VHS • 1994</span></h1>
        </div>
      </div>
      <div class="center">
        <div class="fight"><span class="dot"></span>MANIECKI • 02:17 • PARKIET</div>
        <transition name="status-pop">
          <div v-if="statusMsg" class="status">{{ statusMsg }}</div>
        </transition>
      </div>
      <div class="right"></div>
    </div>
  </header>
</template>

<style scoped>
.top-neon { position:relative; background:#0F0D18; border-bottom:1px solid #1E1B2E; flex-shrink:0; overflow:visible; }
.tube { position:absolute; top:0; left:0; right:0; height:2px; background:linear-gradient(90deg,#ff0080,#D4A574,#7C3AED,#ff0080); background-size:200% 100%; animation: flow 3s linear infinite; box-shadow:0 0 8px rgba(212,165,116,.6); }
@keyframes flow { 0%{background-position:0% 0%} 100%{background-position:200% 0%} }
.inner { padding:10px 14px; display:flex; align-items:center; justify-content:space-between; position:relative; }
.left { display:flex; align-items:center; gap:14px; flex-shrink:0; }
.path-chip {
  display:flex; align-items:center; gap:6px; background: rgba(34,197,94,0.08); border:1.5px solid rgba(34,197,94,0.28);
  padding:5px 10px; border-radius:5px; font-family:'JetBrains Mono',monospace; font-size:10px; color:#22c55e; font-weight:700;
  max-width:300px; cursor:pointer; transition: all 0.12s;
}
.path-chip:hover { background: rgba(34,197,94,0.14); border-color:#22c55e; box-shadow:0 0 10px rgba(34,197,94,0.2), 0 2px 0 #000; transform:translateY(-1px); }
.path-chip .ico { width:12px; height:12px; fill: currentColor; flex-shrink:0; }
.path-chip .p { white-space:nowrap; overflow:hidden; text-overflow:ellipsis; text-decoration: underline; text-decoration-color: rgba(34,197,94,0.3); text-underline-offset:3px; }
.logo { display:flex; align-items:center; gap:10px; flex-shrink:0; }
.badge { background:#16131F; border:1px solid #2A2440; padding:3px 7px; border-radius:4px; cursor:pointer; text-align:left; }
.b-label { display:block; font-size:6px; color:#9CA3AF; font-family:'JetBrains Mono',monospace; line-height:1; }
.b-text { display:block; font-size:9px; color:#D4A574; font-weight:800; font-family:'JetBrains Mono',monospace; line-height:1.2; margin-top:1px; }
h1 { font-family:'Bebas Neue',sans-serif; font-size:20px; margin:0; letter-spacing:0.06em; display:flex; align-items:baseline; gap:6px; }
.ver { font-size:9px; color:#D4A574; font-family:'JetBrains Mono',monospace; }
.vhs { font-size:8px; color:#6B7280; font-family:'JetBrains Mono',monospace; }
.center { position:relative; display:flex; align-items:center; }
.fight { font-size:9px; background:#1a1626; border:1px solid #2A2440; padding:4px 10px; border-radius:12px; display:flex; align-items:center; gap:6px; font-family:'JetBrains Mono',monospace; color:#9CA3AF; }
.dot { width:6px; height:6px; background:#ff2e63; border-radius:50%; box-shadow:0 0 6px #ff2e63; animation: blink 1.2s infinite; }
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:0.4} }
.status {
  position:absolute; left:50%; top:calc(100% + 10px); transform:translateX(-50%);
  background:#1a1510; border:1.5px solid #3d2e15; color:#E7C36A; padding:6px 12px; border-radius:5px;
  font-family:'JetBrains Mono',monospace; font-size:11px; white-space:nowrap; box-shadow:0 6px 18px rgba(0,0,0,0.6), 2px 2px 0 #000; z-index:60; pointer-events:none;
}
.status-pop-enter-active, .status-pop-leave-active { transition: all 0.18s ease; }
.status-pop-enter-from, .status-pop-leave-to { opacity:0; transform:translateX(-50%) translateY(-6px) scale(0.98); }
.right { width:40px; }
</style>
