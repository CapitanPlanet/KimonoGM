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
      
      <!-- 5 MISTRZOWSKICH CZARNYCH PRZESZYĆ PASA KOHAKU -->
      <div class="obi-stitch line-1"></div>
      <div class="obi-stitch line-2"></div>
      <div class="obi-stitch line-3"></div>
      <div class="obi-stitch line-4"></div>
      <div class="obi-stitch line-5"></div>
      
      <div class="obi-vignette"></div>

      <div class="obi-content">
        <div class="left-cluster">
          <div class="k-frame">
            <button @click="emit('save')" class="k-btn save" title="Ctrl+S">
              <svg viewBox="0 0 20 20" class="k-ico"><path d="M3.5 2.5A1 1 0 0 1 4.5 1.5h9l3 3v12a1 1 0 0 1-1 1.5h-12a1 1 0 0 1-1-1.5v-13a1 1 0 0 1 1-1.5Zm9 0v3.5A1 1 0 0 0 13.5 7h3"/><path d="M6.5 12.5h7M6.5 15.5h7"/></svg>
              <span>ZAPISZ</span>
            </button>
          </div>
          <div class="k-frame">
            <button @click="emit('menu')" class="k-btn menu">
              <svg viewBox="0 0 20 20" class="k-ico"><path d="M3 5h3.5a1 1 0 0 1 .8.4L8.5 7h8.5M3 10h14M3 15h14"/></svg>
              <span>MENU</span>
            </button>
          </div>
          <div class="k-frame">
            <button @click="emit('tutorial')" class="k-btn tut">
              <svg viewBox="0 0 20 20" class="k-ico"><path d="M4 3.5A1 1 0 0 1 5 2.5h6v15H5a1 1 0 0 1-1-1v-13Z"/><path d="M11 2.5h4a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1h-4"/><path d="M6.5 6.5h2M6.5 9.5h2"/></svg>
              <span>TUTORIAL</span>
            </button>
          </div>
        </div>

        <div class="spacer"></div>

        <div class="right-cluster">
          <div class="k-frame">
            <button @click="emit('host')" :disabled="isBuildingHost" class="k-btn web" :class="{ running: isHosting }">
              <svg viewBox="0 0 20 20" class="k-ico"><circle cx="10" cy="10" r="6.5"/><path d="M3.5 10h13M10 3.5c1.5 2 2.5 4.2 2.5 6.5S11.5 14.5 10 16.5M10 3.5C8.5 5.5 7.5 7.7 7.5 10S8.5 14.5 10 16.5"/></svg>
              <span v-if="!isBuildingHost && !isHosting">WEB-BUILD</span>
              <span v-else-if="isBuildingHost">BUDUJĘ…</span>
              <span v-else>● {{ hostUrl.replace('http://','') }}</span>
            </button>
          </div>

          <div class="k-frame">
            <button @click="emit('build-exe')" :disabled="isBuildingExe" class="k-btn exe" :class="{ running: isBuildingExe }">
              <svg viewBox="0 0 20 20" class="k-ico"><circle cx="10" cy="10" r="6.5"/><circle cx="10" cy="10" r="2.2"/><path d="M13.8 6.2 12 8M6.2 13.8 8 12"/></svg>
              <span v-if="!isBuildingExe">EXE</span>
              <span v-else>●</span>
            </button>
          </div>

          <div class="k-frame">
            <button @click="emit('parkiet')" :disabled="isRunning" class="k-btn parkiet" :class="{ running: isRunning }">
              <svg viewBox="0 0 20 20" class="k-ico"><path d="M7 4.5 14.5 10 7 15.5V4.5Z"/></svg>
              <span v-if="!isRunning">WPUŚĆ NA PARKIET</span>
              <span v-else>ODPALAM…</span>
            </button>
          </div>
        </div>

        <div class="maker-tag">
          <div class="tag-fold"></div>
            <div class="tag-body">
              <span class="tag-kanji">「先生、玉舐めて」</span>
                <div class="tag-red-dot"></div>
                <span class="tag-sub">MANIECKI • 02:17</span>
            </div>
          </div>
        </div>
      </div>
    </div>
</template>
