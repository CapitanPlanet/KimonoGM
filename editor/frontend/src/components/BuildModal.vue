<script setup lang="ts">
defineProps<{ step: string; log: string }>()
const emit = defineEmits(['close'])
</script>

<template>
  <div class="ov" @click.self="emit('close')">
    <div class="modal">
      <div class="scan"></div><div class="noise"></div>
      <div class="head">
        <div class="ttl"><span>💿</span><h2>BUDUJĘ EXE</h2><span class="vhs">VHS • REC • 02:17</span></div>
        <button v-if="step.includes('BŁĄD') || step.includes('GOTOWE')" @click="emit('close')" class="x">✕</button>
        <div v-else class="rec"><span class="dot"></span>REC</div>
      </div>
      <div class="body">
        <div class="step" :class="{ err: step.includes('BŁĄD'), ok: step.includes('GOTOWE') }">{{ step }}</div>
        <div class="bar"><div class="fill" :class="{ ok: step.includes('GOTOWE'), err: step.includes('BŁĄD') }"></div></div>
        <div class="log"><pre>{{ log || 'Uruchamiam builder…' }}</pre></div>
        <div class="foot">
          <div class="tip"><b>TIP:</b> Argus + Kustosz + 6 rano = elegancki build ✨</div>
          <div v-if="!step.includes('BŁĄD') && !step.includes('GOTOWE')" class="wait">Nie zamykaj edytora, ~15s… dotnet publish mieli…</div>
          <div v-if="step.includes('GOTOWE')" class="ok-msg">🎉 Maniecki Parkiet gotowy w Downloads! 🎉</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ov { position:fixed; inset:0; background: rgba(5,3,10,0.92); backdrop-filter: blur(12px) saturate(0.8); display:flex; align-items:center; justify-content:center; z-index:9999; padding:20px; }
.modal { position:relative; width:100%; max-width:560px; background: linear-gradient(180deg,#1a1214 0%,#121010 100%); border:2px solid #000; border-radius:8px; box-shadow:0 0 0 1px #2a2020,0 8px 32px rgba(0,0,0,0.8),0 0 80px rgba(122,21,37,0.15); overflow:hidden; }
.scan { position:absolute; inset:0; background: repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.15) 2px, rgba(0,0,0,0.15) 3px); pointer-events:none; z-index:5; }
.noise { position:absolute; inset:0; opacity:0.03; pointer-events:none; z-index:5; background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 100 100' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence baseFrequency='0.9'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E"); }
.head { position:relative; z-index:10; display:flex; align-items:center; justify-content:space-between; padding:12px 16px; border-bottom:1.5px solid #2a1a1e; background: linear-gradient(90deg,#1a1214 0%,#21181a 100%); }
.ttl { display:flex; align-items:center; gap:10px; }
.ttl h2 { font-family:'Bebas Neue',sans-serif; font-size:18px; letter-spacing:0.06em; margin:0; color:#E7C36A; }
.vhs { font-family:'JetBrains Mono',monospace; font-size:9px; color:#7a1525; background:#000; padding:2px 6px; border-radius:3px; }
.rec { display:flex; align-items:center; gap:6px; font-family:'JetBrains Mono',monospace; font-size:11px; color:#ff2e63; font-weight:700; }
.dot { width:8px; height:8px; background:#ff2e63; border-radius:50%; box-shadow:0 0 8px #ff2e63; animation: blink 0.8s infinite; }
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:0.3} }
.x { background:#1a1214; border:1.5px solid #2a1a1e; color:#D4A574; width:28px; height:28px; border-radius:4px; font-size:14px; cursor:pointer; }
.body { position:relative; z-index:10; padding:18px; }
.step { font-family:'JetBrains Mono',monospace; font-size:13px; font-weight:700; color:#E7C36A; background: rgba(231,195,106,0.08); border:1px solid rgba(231,195,106,0.2); padding:10px 12px; border-radius:4px; margin-bottom:14px; }
.step.ok { color:#4ade80; background: rgba(74,222,128,0.08); border-color: rgba(74,222,128,0.25); }
.step.err { color:#ff6b6b; background: rgba(255,107,107,0.08); border-color: rgba(255,107,107,0.25); }
.bar { height:6px; background:#0a0808; border:1px solid #1a1214; border-radius:3px; overflow:hidden; margin-bottom:16px; }
.fill { height:100%; width:0%; background: linear-gradient(90deg,#7a1525 0%,#D4A574 50%,#E7C36A 100%); border-radius:2px; animation: move 2s ease-in-out infinite; }
.fill.ok { width:100%!important; animation:none; background:#4ade80; }
.fill.err { width:100%!important; animation:none; background:#ff6b6b; }
@keyframes move { 0%{width:10%;transform:translateX(0)} 50%{width:70%;transform:translateX(20%)} 100%{width:10%;transform:translateX(90%)} }
.log { background:#050304; border:1.5px solid #1a1214; border-radius:4px; padding:12px; max-height:220px; overflow-y:auto; font-family:'JetBrains Mono',monospace; font-size:10px; color:#9CA3AF; }
.log pre { margin:0; white-space:pre-wrap; word-break:break-all; }
.foot { margin-top:14px; display:flex; flex-direction:column; gap:8px; }
.tip { font-family:'JetBrains Mono',monospace; font-size:10px; color:#6B7280; }
.wait { font-family:'JetBrains Mono',monospace; font-size:11px; color:#D4A574; background: rgba(212,165,116,0.08); border:1px dashed rgba(212,165,116,0.2); padding:8px 10px; border-radius:4px; }
.ok-msg { font-family:'Bebas Neue',sans-serif; font-size:16px; color:#4ade80; text-align:center; }
</style>
