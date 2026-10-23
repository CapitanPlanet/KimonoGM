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
      <div class="obi-weave"></div>
      <div class="obi-stripes"></div>
      <div class="obi-stitch top"></div>
      <div class="obi-stitch bottom"></div>

      <div class="obi-content">
        <div class="left-cluster">
          <div class="k-frame"><button @click="emit('save')" class="k-btn save"><svg viewBox="0 0 20 20" class="k-ico"><path d="M3.5 2.5A1 1 0 0 1 4.5 1.5h9l3 3v12a1 1 0 0 1-1 1.5h-12a1 1 0 0 1-1-1.5v-13a1 1 0 0 1 1-1.5Zm9 0v3.5A1 1 0 0 0 13.5 7h3"/><path d="M6.5 12.5h7M6.5 15.5h7"/></svg><span>ZAPISZ</span></button></div>
          <div class="k-frame"><button @click="emit('menu')" class="k-btn menu"><svg viewBox="0 0 20 20" class="k-ico"><path d="M3 5h3.5a1 1 0 0 1 .8.4L8.5 7h8.5M3 10h14M3 15h14"/></svg><span>MENU</span></button></div>
          <div class="k-frame"><button @click="emit('tutorial')" class="k-btn tut"><svg viewBox="0 0 20 20" class="k-ico"><path d="M4 3.5A1 1 0 0 1 5 2.5h6v15H5a1 1 0 0 1-1-1v-13Z"/><path d="M11 2.5h4a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1h-4"/><path d="M6.5 6.5h2M6.5 9.5h2"/></svg><span>TUTORIAL</span></button></div>
        </div>
        <div class="center-cluster">
          <div class="k-frame"><button @click="emit('host')" :disabled="isBuildingHost" class="k-btn web" :class="{ running: isHosting }"><svg viewBox="0 0 20 20" class="k-ico"><circle cx="10" cy="10" r="6.5"/><path d="M3.5 10h13M10 3.5c1.5 2 2.5 4.2 2.5 6.5S11.5 14.5 10 16.5M10 3.5C8.5 5.5 7.5 7.7 7.5 10S8.5 14.5 10 16.5"/></svg><span v-if="!isBuildingHost && !isHosting">WEB-BUILD</span><span v-else-if="isBuildingHost">BUDUJĘ…</span><span v-else>● {{ hostUrl.replace('http://','') }}</span></button></div>
          <div class="k-frame"><button @click="emit('build-exe')" :disabled="isBuildingExe" class="k-btn exe" :class="{ running: isBuildingExe }"><svg viewBox="0 0 20 20" class="k-ico"><circle cx="10" cy="10" r="6.5"/><circle cx="10" cy="10" r="2.2"/><path d="M13.8 6.2 12 8M6.2 13.8 8 12"/></svg><span v-if="!isBuildingExe">EXE</span><span v-else>●</span></button></div>
          <div class="k-frame"><button @click="emit('parkiet')" :disabled="isRunning" class="k-btn parkiet" :class="{ running: isRunning }"><svg viewBox="0 0 20 20" class="k-ico"><path d="M7 4.5 14.5 10 7 15.5V4.5Z"/></svg><span v-if="!isRunning">WPUŚĆ NA PARKIET</span><span v-else>ODPALAM…</span></button></div>
        </div>
      </div>

      <!-- METKA - FINAL - MARTWY OBRAZEK INLINE - ZERO KLAS -->
      <div style="all:initial; position:absolute !important; right:12px !important; top:50% !important; transform:translateY(-50%) !important; width:86px !important; height:44px !important; pointer-events:none !important; user-select:none !important; z-index:9999 !important; display:block !important; contain:strict !important; isolation:isolate !important;">
        <div style="width:86px; height:44px; position:relative; box-shadow:2px 2px 0 #000; pointer-events:none;">
          <div style="position:absolute; left:0; top:0; bottom:0; width:10px; background:#0f0c0a; clip-path:polygon(0 0, 100% 0, 70% 50%, 100% 100%, 0 100%);"></div>
          <div style="margin-left:10px; height:100%; background:#ece6d6; border:1px solid #000; border-left:none; display:flex; flex-direction:column; align-items:center; justify-content:center; box-sizing:border-box;">
            <div style="font-size:7px; line-height:1; color:#111; font-family:'MS Gothic','Yu Gothic',sans-serif; font-weight:400; text-align:center; letter-spacing:0.03em;">先生、玉舐めて</div>
            <div style="font-family:'Bebas Neue',Impact,sans-serif; font-size:13px; line-height:1; color:#7a1525; letter-spacing:0.06em; margin-top:1px; font-weight:700;">KIMONO</div>
            <div style="font-family:'JetBrains Mono',monospace; font-size:5.5px; color:#5a4a3a; letter-spacing:0.08em; margin-top:1px;">MANIECKI • 02:17</div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<!-- BEZ SCOPED - ZEBY NIC PONIZEJ TEGO NIE NADPISALO -->
<style>
.obi-root { flex-shrink:0 !important; padding:0 !important; background:#0e0c0f !important; border-top:1.5px solid #000 !important; }
.obi-belt { position:relative !important; height:56px !important; display:flex !important; align-items:center !important; background:#1d1917 !important; overflow:hidden !important; }
.obi-weave { position:absolute !important; inset:0 !important; z-index:0 !important; background: repeating-linear-gradient(90deg, rgba(0,0,0,0.04) 0 1px, transparent 1px 3px), repeating-linear-gradient(0deg, rgba(0,0,0,0.05) 0 1px, transparent 1px 4px) !important; opacity:0.6 !important; pointer-events:none !important; }
.obi-stripes { position:absolute !important; inset:0 !important; z-index:1 !important; background: repeating-linear-gradient(90deg, #e7ddd0 0 84px, #8a1e2f 84px 86px, #7d1b2b 86px 170px, #8a1e2f 170px 172px) !important; pointer-events:none !important; }
.obi-stitch { position:absolute !important; left:0 !important; right:0 !important; height:1px !important; background: repeating-linear-gradient(90deg, rgba(255,255,255,0.12) 0 8px, transparent 8px 16px) !important; z-index:3 !important; pointer-events:none !important; }
.obi-stitch.top { top:4px !important; } .obi-stitch.bottom { bottom:4px !important; }
.obi-content { position:relative !important; z-index:5 !important; width:100% !important; height:100% !important; display:flex !important; align-items:center !important; padding:0 12px !important; }
.left-cluster { display:flex !important; align-items:center !important; gap:8px !important; }
.center-cluster { position:absolute !important; left:50% !important; top:50% !important; transform:translate(-50%, -50%) !important; display:flex !important; align-items:center !important; gap:8px !important; }

/* WSZYSTKIE PRZYCISKI JAK MENU - ZABIJAMY HOVERY GLOBALNIE */
.k-frame { background:#0a0a0a !important; border:1.5px solid #000 !important; padding:2px !important; box-shadow:2px 2px 0 #000 !important; border-radius:6px !important; display:flex !important; flex-shrink:0 !important; }
.k-frame:hover, .k-frame:active, .k-frame:focus, .k-frame:focus-within { background:#0a0a0a !important; border:1.5px solid #000 !important; box-shadow:2px 2px 0 #000 !important; transform:none !important; }

.k-btn { height:28px !important; padding:0 14px !important; border-radius:4px !important; border:none !important; font-family:'JetBrains Mono', monospace !important; font-size:11px !important; font-weight:700 !important; display:inline-flex !important; align-items:center !important; gap:7px !important; white-space:nowrap !important; cursor:pointer !important; outline:none !important; box-shadow:none !important; transform:none !important; transition:none !important; filter:none !important; }
.k-btn:hover, .k-btn:active, .k-btn:focus, .k-btn:focus-visible { filter:none !important; transform:none !important; box-shadow:none !important; outline:none !important; border:none !important; font-weight:700 !important; }
.k-btn:disabled { opacity:0.6 !important; cursor:not-allowed !important; }
.k-ico { width:14px !important; height:14px !important; flex-shrink:0 !important; stroke:currentColor !important; fill:none !important; stroke-width:1.6 !important; stroke-linecap:round !important; stroke-linejoin:round !important; }

.k-btn.save { background:#162032 !important; color:#7DD3FC !important; min-width:96px !important; }
.k-btn.menu { background:#16131F !important; color:#9CA3AF !important; min-width:82px !important; }
.k-btn.tut { background:#201c15 !important; color:#D4A574 !important; min-width:108px !important; }
.k-btn.web { background:#102a30 !important; color:#7defff !important; min-width:124px !important; }
.k-btn.web.running { background:#00d9ff !important; color:#001a1e !important; }
.k-btn.exe { background:#2a2010 !important; color:#E7C36A !important; width:84px !important; }
.k-btn.parkiet { background:#3d4f20 !important; color:#d9f99d !important; min-width:184px !important; }
</style>
