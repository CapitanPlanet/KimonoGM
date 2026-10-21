<script setup lang="ts">
defineProps<{
  isHosting: boolean
  hostUrl: string
  isBuildingHost: boolean
  isBuildingExe: boolean
  isRunning: boolean
}>()
const emit = defineEmits(['save', 'menu', 'tutorial', 'host', 'build-exe', 'parkiet'])
</script>

<template>
  <div class="obi-root">
    <div class="obi-belt">
      <!-- tkanina base -->
      <div class="obi-weave"></div>
      <div class="obi-stripes"></div>
      <div class="obi-noise"></div>
      <div class="obi-stitch top"></div>
      <div class="obi-stitch bottom"></div>
      <div class="obi-vignette"></div>

      <div class="obi-content">
        <div class="left-cluster">
          <button @click="emit('save')" class="k-btn save" title="Ctrl+S">
            <svg viewBox="0 0 20 20" class="k-ico"><path d="M3.5 2.5A1 1 0 0 1 4.5 1.5h9l3 3v12a1 1 0 0 1-1 1.5h-12a1 1 0 0 1-1-1.5v-13a1 1 0 0 1 1-1.5Zm9 0v3.5A1 1 0 0 0 13.5 7h3"/><path d="M6.5 12.5h7M6.5 15.5h7"/></svg>
            <span>ZAPISZ</span>
          </button>
          <div class="k-frame">
            <button @click="emit('menu')" class="k-btn menu">
              <svg viewBox="0 0 20 20" class="k-ico"><path d="M3 5h3.5a1 1 0 0 1 .8.4L8.5 7h8.5M3 10h14M3 15h14"/></svg>
              <span>MENU</span>
            </button>
          </div>
          <button @click="emit('tutorial')" class="k-btn tut">
            <svg viewBox="0 0 20 20" class="k-ico"><path d="M4 3.5A1 1 0 0 1 5 2.5h6v15H5a1 1 0 0 1-1-1v-13Z"/><path d="M11 2.5h4a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1h-4"/><path d="M6.5 6.5h2M6.5 9.5h2"/></svg>
            <span>TUTORIAL</span>
          </button>
        </div>

        <div class="spacer"></div>

        <div class="right-cluster">
          <button @click="emit('host')" :disabled="isBuildingHost" class="k-btn web" :class="{ running: isHosting }">
            <svg viewBox="0 0 20 20" class="k-ico"><circle cx="10" cy="10" r="6.5"/><path d="M3.5 10h13M10 3.5c1.5 2 2.5 4.2 2.5 6.5S11.5 14.5 10 16.5M10 3.5C8.5 5.5 7.5 7.7 7.5 10S8.5 14.5 10 16.5"/></svg>
            <span v-if="!isBuildingHost && !isHosting">WEB-BUILD</span>
            <span v-else-if="isBuildingHost">BUDUJĘ…</span>
            <span v-else>● {{ hostUrl.replace('http://','') }}</span>
          </button>

          <button @click="emit('build-exe')" :disabled="isBuildingExe" class="k-btn exe" :class="{ running: isBuildingExe }">
            <svg viewBox="0 0 20 20" class="k-ico"><circle cx="10" cy="10" r="6.5"/><circle cx="10" cy="10" r="2.2"/><path d="M13.8 6.2 12 8M6.2 13.8 8 12"/></svg>
            <span v-if="!isBuildingExe">EXE</span>
            <span v-else>●</span>
          </button>

          <button @click="emit('parkiet')" :disabled="isRunning" class="k-btn parkiet" :class="{ running: isRunning }">
            <svg viewBox="0 0 20 20" class="k-ico"><path d="M7 4.5 14.5 10 7 15.5V4.5Z"/></svg>
            <span v-if="!isRunning">WPUŚĆ NA PARKIET</span>
            <span v-else>ODPALAM…</span>
          </button>
        </div>

        <div class="maker-tag">
          <div class="tag-fold"></div>
          <div class="tag-body">
            <span class="tag-kanji">先生、玉舐めて</span>
            <span class="tag-logo">KIMONO</span>
            <span class="tag-sub">MANIECKI • 02:17</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.obi-root { flex-shrink:0; padding: 0; background:#0e0c0f; border-top:1.5px solid #000; }
.obi-belt {
  position:relative; height:56px; display:flex; align-items:center;
  background:#1d1917; overflow:hidden; isolation:isolate;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.06), inset 0 -1px 0 rgba(0,0,0,0.6);
}
.obi-weave {
  position:absolute; inset:0; z-index:0;
  background:
    repeating-linear-gradient(90deg, rgba(0,0,0,0.04) 0 1px, transparent 1px 3px),
    repeating-linear-gradient(0deg, rgba(0,0,0,0.05) 0 1px, transparent 1px 4px);
  opacity:0.6;
}
.obi-stripes {
  position:absolute; inset:0; z-index:1;
  background: repeating-linear-gradient(90deg,
    #e7ddd0 0px, #e7ddd0 84px,
    #8a1e2f 84px, #8a1e2f 86px,
    #7d1b2b 86px, #7d1b2b 170px,
    #8a1e2f 170px, #8a1e2f 172px
  );
  filter: brightness(0.93) contrast(1.02) saturate(0.92);
}
.obi-noise {
  position:absolute; inset:0; z-index:2; opacity:0.18;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
  mix-blend-mode: multiply;
}
.obi-stitch {
  position:absolute; left:0; right:0; height:1px; z-index:3;
  background: repeating-linear-gradient(90deg, transparent 0 6px, rgba(0,0,0,0.35) 6px 10px);
}
.obi-stitch.top { top:6px; }
.obi-stitch.bottom { bottom:6px; }
.obi-vignette {
  position:absolute; inset:0; z-index:3; pointer-events:none;
  background: linear-gradient(180deg, rgba(0,0,0,0.22), transparent 30%, transparent 70%, rgba(0,0,0,0.28));
}

.obi-content {
  position:relative; z-index:10; display:flex; align-items:center; gap:10px;
  width:100%; padding:0 12px; flex-wrap:nowrap;
}
.left-cluster, .right-cluster { display:flex; align-items:center; gap:8px; flex-shrink:0; }
.spacer { flex:1; }

.k-btn {
  height:34px; padding:0 14px; border-radius:5px; border:1.5px solid #000;
  font-family:'JetBrains Mono', monospace; font-size:11px; font-weight:700; letter-spacing:0.02em;
  display:inline-flex; align-items:center; gap:7px; white-space:nowrap;
  box-shadow: 2px 2px 0 #000; transition: transform 0.1s, box-shadow 0.1s, background 0.12s;
  cursor:pointer; flex-shrink:0;
}
.k-btn:hover { transform: translate(-1px,-1px); box-shadow: 3px 3px 0 #000; }
.k-btn:active { transform: translate(1px,1px); box-shadow: 1px 1px 0 #000; }
.k-btn:disabled { opacity:0.6; cursor:not-allowed; transform:none; }

.k-ico { width:14px; height:14px; flex-shrink:0; stroke: currentColor; fill:none; stroke-width:1.6; stroke-linecap:round; stroke-linejoin:round; }

.k-btn.save { background:#162032; border-color:#23344d; color:#7DD3FC; min-width:96px; }
.k-btn.menu { background:#16131F; border-color:#2A2440; color:#9CA3AF; min-width:82px; }
.k-frame { background:#0a0a0a; border:1.5px solid #000; padding:2px; box-shadow:2px 2px 0 #000; transform: rotate(0.3deg); }
.k-frame .menu { box-shadow:none; border:none; height:28px; }

.k-btn.tut { background:#201c15; border-color:#342e20; color:#D4A574; min-width:108px; }

.k-btn.web { background: linear-gradient(180deg, #102a30 0%, #0c2026 100%); border-color:#0e4a52; color:#7defff; min-width:124px; box-shadow: 2px 2px 0 #000, 0 0 10px rgba(0,217,255,0.18); }
.k-btn.web.running { background:#00d9ff; color:#001a1e; border-color:#7defff; animation: web-pulse 1.2s infinite; }
.k-btn.exe { background: linear-gradient(180deg, #2a2010 0%, #1a150a 100%); border-color:#3d2e15; color:#E7C36A; width:84px; }
.k-btn.exe.running { animation: gold-pulse 1s infinite; }
.k-btn.parkiet { background: linear-gradient(180deg, #3d4f20 0%, #2a3515 100%); border-color:#4a5e2a; color:#d9f99d; min-width:184px; }
.k-btn.parkiet.running { animation: green-pulse 1s infinite; }

@keyframes web-pulse { 0%,100% { box-shadow: 2px 2px 0 #000, 0 0 10px rgba(0,217,255,0.3); } 50% { box-shadow: 2px 2px 0 #000, 0 0 18px rgba(0,217,255,0.6); } }
@keyframes gold-pulse { 0%,100% { box-shadow: 2px 2px 0 #000, 0 0 8px rgba(231,195,106,0.35); } 50% { box-shadow: 2px 2px 0 #000, 0 0 16px rgba(231,195,106,0.65); } }
@keyframes green-pulse { 0%,100% { box-shadow: 2px 2px 0 #000, 0 0 8px rgba(163,230,53,0.35); } 50% { box-shadow: 2px 2px 0 #000, 0 0 16px rgba(163,230,53,0.65); } }

.maker-tag {
  position:relative; width:86px; height:44px; flex-shrink:0; flex-grow:0; margin-left:8px;
  filter: drop-shadow(2px 2px 0 #000);
}
.tag-fold { position:absolute; left:0; top:0; bottom:0; width:10px; background:#0f0c0a; clip-path: polygon(0 0, 100% 0, 70% 50%, 100% 100%, 0 100%); }
.tag-body {
  margin-left:10px; height:100%; background:#ece6d6; border:1px solid #000; border-left:none;
  display:flex; flex-direction:column; align-items:center; justify-content:center; padding:2px 4px;
}
.tag-kanji { font-size:7px; line-height:1; color:#111; font-weight:900; letter-spacing:0.02em; }
.tag-logo { font-family:'Bebas Neue', sans-serif; font-size:13px; line-height:1; color:#7a1525; letter-spacing:0.06em; margin-top:1px; }
.tag-sub { font-family:'JetBrains Mono', monospace; font-size:5.5px; color:#5a4a3a; letter-spacing:0.08em; margin-top:1px; }
</style>
