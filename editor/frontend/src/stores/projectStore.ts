import { defineStore } from 'pinia'
import { CreateProject, ReadJSON, WriteJSON, ListFiles, ListAssets, ListAudioAssets, DeleteFile } from '../../wailsjs/go/main/App'

export interface StatDef { id: string; name: string; initial: number }
interface Choice {
  id?: string;
  Text: string;
  Next: string;
  NextDayId?: string;
  NextDay?: string;
  Stats?: Record<string, number>;
  ReactionText?: string;
  ReactionImage?: string;
  SoundFile?: string;
  FlagsSet?: string[];
  FlagsRequired?: string[];
  MinPortfel?: number | null;
  KosztPortfel?: number | null;
  FailText?: string;
  [key: string]: any
}

export interface EndDayTransfers {
  keep_flags: string[]
  keep_stats: string[]
  summary_text?: string
  nextDay?: string
}

interface Scene {
  Id: string;
  SceneTitle?: string;
  Background?: string;
  Text?: string;
  Choices?: Choice[];
  IsEndDay?: boolean;
  Type?: 'normal' | 'end_of_day';
  Day?: number;
  NextDay?: string;
  NextDayId?: string;
  Transfers?: EndDayTransfers | null;
  [key: string]: any
}

type ConditionOp = { gte?: number; lte?: number; gt?: number; lt?: number; eq?: number }
export interface AvatarRule { id: string; use: string; if: Record<string, ConditionOp>; priority: number; when?: any }
export interface AvatarSystem { default: string; rules: AvatarRule[] }
export interface StatsSystem { stats: StatDef[] }
interface ProjectMeta { gameName: string; author: string; version: string; engineVersion: string; startDay: string; startScene: string }

const DEFAULT_STATS: StatDef[] = [
  { id: 'CEBULA', name: 'CEBULA', initial: 0 },
  { id: 'WSTYD', name: 'WSTYD', initial: 0 },
  { id: 'PORTFEL', name: 'PORTFEL', initial: 0 },
  { id: 'REPUTACJA', name: 'REPUTACJA', initial: 0 },
]

const LEGACY_MAP: Record<string,string> = { A:'CEBULA', B:'WSTYD', C:'PORTFEL', D:'REPUTACJA', PORTFEL1:'PORTFEL' }

function toSafeName(raw: string): string {
  const up = raw.toString().trim().toUpperCase()
  return LEGACY_MAP[up] || up
}
function getBasename(path: string): string {
  return path.replace(/\\/g,'/').split('/').pop() || ''
}
function isImageOfType(path: string, prefix: 'bg_'|'re_'|'av_'): boolean {
  const base = getBasename(path).toLowerCase()
  return base.startsWith(prefix)
}
function fixRelPath(input: string): string {
  if (!input) return ''
  let p = input.replace(/\\/g,'/').trim()
  if (p.includes(':') || p.startsWith('/') || p.startsWith('C:')) {
    const idxImg = p.toLowerCase().lastIndexOf('images/')
    const idxSnd = p.toLowerCase().lastIndexOf('sounds/')
    if (idxImg >= 0) p = p.substring(idxImg)
    else if (idxSnd >= 0) p = p.substring(idxSnd)
    else p = 'images/' + p.split('/').pop()!
  }
  p = p.replace(/^\/+/, '')
  // FIX: jpg -> webp fallback dla tutorial
  if (p === 'images/bg_tutorial.jpg') return 'images/bg_tutorial.webp'
  if (p === 'images/bg_front.jpg') return 'images/bg_tutorial.webp'
  if (p === 'images/bg_sen_ jaanusza.jpg') return 'images/bg_tutorial.webp'
  return p
}
function isEndDayId(id: string): boolean {
  return id.startsWith('koniec_') || id.startsWith('koniec_dnia_') || id.startsWith('koniec_dnia')
}
function genId(): string {
  return (crypto as any).randomUUID?.() || Math.random().toString(36).slice(2)
}
function toSafeDayId(raw: string): string {
  let s = raw.trim().toLowerCase().replace(/\s+/g,'_').replace(/[^a-z0-9_\-]/g,'').replace(/_+/g,'_').replace(/-+/g,'-').replace(/^_+|_+$/g,'')
  return s.slice(0,40)
}

export const useProjectStore = defineStore('project', {
  state: () => ({
    projectPath: null as string | null,
    meta: null as ProjectMeta | null,
    avatarSystem: { default: '', rules: [] } as AvatarSystem,
    statsSystem: { stats: [...DEFAULT_STATS] } as StatsSystem,
    days: {} as Record<string, Scene[]>,
    currentDay: 'day1',
    currentSceneId: null as string | null,
    assets: { all: [] as string[], images: [] as string[], sounds: [] as string[] },
    saveStatus: '',
    ui: { showAvatarEditor: false }
  }),
  getters: {
    isProjectLoaded: (s) =>!!s.projectPath,
    currentDayScenes: (s) => s.days[s.currentDay] || [],
    currentScene: (s) => {
      const scenes = s.days[s.currentDay] || []
      return scenes.find(x => x.Id === s.currentSceneId) || null
    },
    sceneIdsInCurrentDay: (s) => (s.days[s.currentDay] || []).map(x => x.Id),
    allSceneIds: (s) => {
      const ids: string[] = []
      Object.values(s.days).forEach((scenes: any) => scenes.forEach((sc: Scene) => ids.push(sc.Id)))
      return ids
    },
    allScenes: (s) => {
      const all: Scene[] = []
      Object.values(s.days).forEach((scenes: any) => all.push(...scenes))
      return all
    },
    daysList: (s) => Object.keys(s.days).sort(),
    dayFileList: (s) => Object.keys(s.days),
    backgroundAssets: (s) => s.assets.images.filter(a => isImageOfType(a, 'bg_')),
    reactionAssets: (s) => s.assets.images.filter(a => isImageOfType(a, 're_')),
    avatarAssets: (s) => s.assets.images.filter(a => isImageOfType(a, 'av_')),
    availableBackgrounds: (s) => s.assets.images.filter(a => isImageOfType(a, 'bg_')),
    availableSounds: (s) => s.assets.sounds,
    sortedAvatarRules(s): AvatarRule[] { return [...s.avatarSystem.rules].sort((a,b) => b.priority - a.priority) },
    statList: (s) => s.statsSystem.stats,
    endDaySceneInCurrentDay: (s) => (s.days[s.currentDay] || []).find(x => x.IsEndDay || x.Type === 'end_of_day' || isEndDayId(x.Id))
  },
  actions: {
    getDayStartSceneId(day: string): string | null {
      const scenes = this.days[day]
      if (!scenes ||!scenes.length) return null
      // pierwszy nie-end
      const start = scenes.find(s =>!s.IsEndDay) || scenes[0]
      return start?.Id || null
    },

    selectScene(id: string) { this.currentSceneId = id },

    ensureChoiceIds(scene: Scene) {
      scene.Choices?.forEach(c => { if(!c.id) c.id = genId() })
    },

    createEmptyStats(): Record<string, number> {
      const obj: Record<string, number> = {}
      this.statsSystem.stats.forEach(s => { obj[s.name] = 0 })
      return obj
    },

    ensureStatsSystem() {
      if (!this.statsSystem ||!Array.isArray(this.statsSystem.stats) ||!this.statsSystem.stats.length) {
        this.statsSystem = { stats: [...DEFAULT_STATS] }
        return
      }
      const used = new Set<string>()
      const fixed: StatDef[] = []
      for(const s of this.statsSystem.stats){
        if(!s.name &&!s.id) continue
        const finalName = toSafeName((s.name || s.id) as string)
        if(!finalName) continue
        if(used.has(finalName)) continue
        used.add(finalName)
        fixed.push({ id: finalName, name: finalName, initial: s.initial?? 0 })
      }
      this.statsSystem.stats = fixed.length? fixed : [...DEFAULT_STATS]
    },

    ensureAvatarSystem() {
      this.ensureStatsSystem()
      if (!this.avatarSystem) this.avatarSystem = { default: '', rules: [] }
      if (!this.avatarSystem.rules) this.avatarSystem.rules = []
      this.avatarSystem.default = fixRelPath(this.avatarSystem.default || '')
      if(!this.avatarSystem.default && this.avatarAssets.length > 0){
        this.avatarSystem.default = this.avatarAssets[0]
      }
      this.avatarSystem.rules.forEach(r=>{
        if(r.if){
          const newIf: Record<string, ConditionOp> = {}
          for(const [k,v] of Object.entries(r.if)){
            const final = toSafeName(k)
            if(this.statsSystem.stats.some(s=>s.name===final)){
              newIf[final] = v as ConditionOp
            }
          }
          r.if = newIf
        }
        if(r.use) r.use = fixRelPath(r.use)
      })
    },

    migrateChoiceStats(choice: any) {
      if (!choice.Stats) choice.Stats = {}
      const newStats: Record<string, number> = {}
      for (const [k,v] of Object.entries(choice.Stats as Record<string,any>)) {
        if(typeof v!== 'number') continue
        const finalKey = toSafeName(k)
        if(this.statsSystem.stats.some(s=>s.name===finalKey)){
          newStats[finalKey] = v as number
        }
      }
      const toDelete = ['Cebula','Wstyd','Portfel','Reputacja','cebula','wstyd','portfel','reputacja','a','b','c','d','A','B','C','D']
      toDelete.forEach(k => delete choice[k])
      choice.Stats = newStats
      this.statsSystem.stats.forEach(s=>{ if(choice.Stats[s.name]==null) choice.Stats[s.name]=0 })
      if(choice.ReactionImage) choice.ReactionImage = fixRelPath(choice.ReactionImage)
      if(choice.SoundFile) choice.SoundFile = fixRelPath(choice.SoundFile)
    },

    normalizeScene(scene: Scene): Scene {
      if(scene.Background) scene.Background = fixRelPath(scene.Background)
      const isEnd = isEndDayId(scene.Id) || scene.IsEndDay || scene.Type === 'end_of_day'
      if (isEnd) {
        scene.IsEndDay = true
        scene.Type = 'end_of_day'
        if (!scene.Transfers) {
          scene.Transfers = { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' }
        } else {
          if(!scene.Transfers.keep_flags) scene.Transfers.keep_flags = []
          if(!scene.Transfers.keep_stats ||!scene.Transfers.keep_stats.length) {
            scene.Transfers.keep_stats = this.statsSystem.stats.map(s=>s.name)
          }
        }
        if(!scene.Choices) scene.Choices = []
        if(scene.NextDay) scene.NextDayId = scene.NextDay
        // FIX: auto-napraw pusty Choices w końcu dnia
        if (scene.Choices.length === 0) {
          let nextTarget = 'END'
          if (scene.NextDayId && scene.NextDayId !== 'END') {
            const startId = this.getDayStartSceneId(scene.NextDayId)
            nextTarget = startId || 'END'
          }
          scene.Choices = [{ id: genId(), Text: 'Zakończ', Next: nextTarget, Stats: {}, ReactionText:'', ReactionImage:'', SoundFile:'' } as any]
        }
        // FIX: jesli Next to nazwa dnia a nie scena, popraw
        if (scene.Choices[0] && scene.NextDayId && scene.NextDayId !== 'END' && scene.Choices[0].Next === scene.NextDayId) {
          const startId = this.getDayStartSceneId(scene.NextDayId)
          if (startId) scene.Choices[0].Next = startId
        }
      }
      if (scene.IsEndDay &&!scene.Type) scene.Type = 'end_of_day'
      return scene
    },

    async loadAssets() {
      if (!this.projectPath) { this.assets.all=[]; this.assets.images=[]; this.assets.sounds=[]; return }
      try {
        const [images, sounds] = await Promise.all([
          ListAssets(this.projectPath).catch(()=>[] as string[]),
          ListAudioAssets(this.projectPath).catch(()=>[] as string[])
        ])
        this.assets.images = (images || []).map(fixRelPath)
        this.assets.sounds = (sounds || []).map(fixRelPath)
        this.assets.all = [...this.assets.images,...this.assets.sounds]
      } catch {
        this.assets.all=[]; this.assets.images=[]; this.assets.sounds=[]
      }
      this.ensureAvatarSystem()
    },

    async createProjectAtPath(p: string, name: string) {
      p = p.replace(/\\/g,'/').replace(/\/+/g,'/').replace(/\/$/,'')
      try { await CreateProject(p, name) } catch {}
      await this.loadProjectFromPath(p)
      return p
    },

    async deleteDay(dayId: string) {
      if (this.dayFileList.length <= 1) { alert('Musisz zostawić minimum 1 dzień!'); return }
      if (this.projectPath) { try { await DeleteFile(this.projectPath, `Data/${dayId}.json`) } catch {} }
      delete this.days[dayId]
      if (this.currentDay === dayId) { this.currentDay = Object.keys(this.days)[0]; this.currentSceneId = this.days[this.currentDay]?.[0]?.Id || null }
    },

    async loadProjectFromPath(path: string) {
      path = path.replace(/\\/g,'/').replace(/\/+/g,'/').replace(/\/$/,'')
      this.projectPath = path
      try {
        const metaRaw = await ReadJSON(`${path}/project.janproj`)
        const parsed = JSON.parse(metaRaw)
        this.meta = {
          gameName: parsed.gameName||'',
          author: parsed.author||'',
          version: parsed.version||'1.0.0',
          engineVersion: parsed.engineVersion||'2.0.0',
          startDay: parsed.startDay||'day1',
          startScene: parsed.startScene||'start'
        }
        let incomingStats: StatDef[] = parsed.statsSystem?.stats || [...DEFAULT_STATS]
        const migrated: StatDef[] = incomingStats.map(s=>{
          const finalName = toSafeName((s.name||s.id||'').toString())
          return { id: finalName, name: finalName, initial: s.initial??0 }
        }).filter(s=>s.name)
        this.statsSystem = { stats: migrated.length? migrated : [...DEFAULT_STATS] }
        this.ensureStatsSystem()
        this.avatarSystem = {
          default: fixRelPath(parsed.avatarSystem?.default||''),
          rules: Array.isArray(parsed.avatarSystem?.rules)?parsed.avatarSystem.rules:[]
        }
        this.avatarSystem.rules.forEach((r:any)=>{
          if(!r.id) r.id = genId()
          if(r.priority==null) r.priority = 0
          if(r.use) r.use = fixRelPath(r.use)
        })
        let dayFiles: string[] = []
        try { dayFiles = await ListFiles(`${path}/Data`, '.json') || [] } catch { dayFiles = ['day1.json'] }
        dayFiles = dayFiles.filter(f =>!f.startsWith('_') &&!f.startsWith('.'))
        this.days = {}
        for (const file of dayFiles) {
          const dayName = file.replace('.json','')
          try {
            const raw = await ReadJSON(`${path}/Data/${file}`)
            const data = JSON.parse(raw) as Scene[]
            data.forEach(s => { this.normalizeScene(s); this.ensureChoiceIds(s); s.Choices?.forEach(c => this.migrateChoiceStats(c)) })
            this.days[dayName] = data
          } catch {}
        }
        this.currentDay = this.meta.startDay || Object.keys(this.days)[0] || 'day1'
        this.currentSceneId = this.meta.startScene || this.days[this.currentDay]?.[0]?.Id || null
        await this.loadAssets()
        this.ensureAvatarSystem()
      } catch (e) { console.error(e); this.projectPath = null; throw e }
    },

    async scanAssets() { await this.loadAssets() },

    async saveProject() {
      if (!this.projectPath ||!this.meta) return
      this.saveStatus = 'Zapisywanie...'
      try {
        this.ensureAvatarSystem(); this.ensureStatsSystem()
        Object.values(this.days).forEach(scenes => {
          scenes.forEach(scene => {
            this.normalizeScene(scene)
            scene.Choices?.forEach((c:any) => this.migrateChoiceStats(c))
          })
        })

        // Auto-twórz dni na które są linki - dowolna nazwa
        const referencedDays = new Set<string>()
        Object.values(this.days).forEach(scenes => {
          scenes.forEach(s => {
            if (s.NextDayId && s.NextDayId!=='END') referencedDays.add(s.NextDayId.replace('.json','').trim())
            if (s.NextDay && s.NextDay!=='END') referencedDays.add(s.NextDay.replace('.json','').trim())
          })
        })
        for (const ref of referencedDays) {
          if (!ref || ref === 'END') continue
          const safeRef = toSafeDayId(ref)
          if (!safeRef) continue
          if (!this.days[safeRef] && !this.days[ref]) {
            const dayNum = Object.keys(this.days).length+1
            const startId = `start_${safeRef}`
            const endId = `koniec_${safeRef}`
            this.days[safeRef] = [{
              Id: startId,
              SceneTitle: `${safeRef} - Auto`,
              Background: 'images/bg_tutorial.webp',
              Text: `Dzień ${safeRef} - auto-utworzony bo był na niego link. Podmień treść.`,
              Choices: [{ id: genId(), Text: 'Dalej', Next: endId, Stats: this.createEmptyStats() }],
              Type: 'normal' as const,
              Day: dayNum
            },{
              Id: endId,
              SceneTitle: `KONIEC ${safeRef}`,
              Background: 'images/bg_tutorial.webp',
              Text: `Koniec ${safeRef}.`,
              IsEndDay: true,
              Type: 'end_of_day' as const,
              Day: dayNum,
              NextDayId: undefined,
              NextDay: undefined,
              Transfers: { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' },
              Choices: [{ id: genId(), Text:'Zakończ dzień', Next:'END', Stats: {} }]
            }]
          }
        }

        const toSave = {
        ...this.meta,
          avatarSystem: { default: fixRelPath(this.avatarSystem.default), rules: this.avatarSystem.rules.map(r=>({...r, use: fixRelPath(r.use)})) },
          statsSystem: { stats: this.statsSystem.stats }
        }
        await WriteJSON(`${this.projectPath}/project.janproj`, JSON.stringify(toSave, null, 2))
        for (const dayFile of Object.keys(this.days)) {
          const cleaned = this.days[dayFile].map(scene => {
            // FIX: upewnij się że koniec dnia ma Choices
            let s = {...scene}
            if (s.IsEndDay && (!s.Choices || s.Choices.length===0)) {
              s.Choices = [{ Text:'Zakończ', Next: s.NextDayId ? (this.getDayStartSceneId(s.NextDayId) || 'END') : 'END', Stats: {} } as any]
            }
            return {
            ...s,
              Background: fixRelPath(s.Background || 'images/bg_tutorial.webp'),
              Choices: s.Choices?.map((c:any) => {
                const {id,...cleanChoice}=c
                // FIX: nie zapisuj pustych Stats dla END
                let stats = cleanChoice.Stats
                if (s.IsEndDay) stats = {}
                return {
                ...cleanChoice,
                  Stats: stats,
                  ReactionImage: fixRelPath(cleanChoice.ReactionImage||''),
                  SoundFile: fixRelPath(cleanChoice.SoundFile||'')
                }
              })
            }
          })
          await WriteJSON(`${this.projectPath}/Data/${dayFile}.json`, JSON.stringify(cleaned, null, 2))
        }
        const manifest = {
          days: Object.keys(this.days).sort(),
          startDay: this.meta.startDay || Object.keys(this.days)[0] || 'day1',
          startScene: this.meta.startScene || this.days[this.meta.startDay || Object.keys(this.days)[0]]?.[0]?.Id || 'start',
          version: new Date().toISOString()
        }
        await WriteJSON(`${this.projectPath}/Data/_manifest.json`, JSON.stringify(manifest, null, 2))
        this.saveStatus='Zapisano'; setTimeout(()=>{this.saveStatus=''},2500)
      } catch (e) { console.error(e); this.saveStatus='Błąd zapisu: '+e }
    },

    addStat() {
      const used = new Set(this.statsSystem.stats.map(s=>s.name))
      let newName = `STAT_${Date.now().toString(36).toUpperCase()}`
      let counter = 0
      while(used.has(newName) && counter < 10){ newName = `STAT_${Date.now().toString(36).toUpperCase()}_${counter++}` }
      if(!used.has(newName)){
        this.statsSystem.stats.push({ id: newName, name: newName, initial: 0 })
      }
      Object.values(this.days).forEach(scenes => scenes.forEach(scene => scene.Choices?.forEach(c => { if(!c.Stats) c.Stats={}; if(c.Stats[newName]==null) c.Stats[newName]=0 })))
    },

    deleteStat(id: string) {
      if (this.statsSystem.stats.length <= 1) { alert('Musi zostać przynajmniej 1 statystyka'); return }
      this.statsSystem.stats = this.statsSystem.stats.filter(s => s.id!==id && s.name!==id)
      Object.values(this.days).forEach(scenes => scenes.forEach(scene => scene.Choices?.forEach((c:any) => { if(c.Stats && c.Stats[id]!=null) delete c.Stats[id] })))
    },

    setDefaultAvatar(path: string) { this.avatarSystem.default = fixRelPath(path) },

    addAvatarRule(imagePath?: string) {
      const maxPrio = Math.max(0,...this.avatarSystem.rules.map(r => r.priority))
      const firstStat = this.statsSystem.stats[0]?.name || 'CEBULA'
      this.avatarSystem.rules.push({
        id: genId(),
        use: fixRelPath(imagePath || this.avatarAssets[0] || this.avatarSystem.default || ''),
        if: { [firstStat]: { gte: 50 } },
        priority: maxPrio+10
      })
    },
    updateAvatarRule(ruleId: string, patch: Partial<AvatarRule>) { const rule = this.avatarSystem.rules.find(r => r.id===ruleId); if(rule) Object.assign(rule, patch) },
    deleteAvatarRule(ruleId: string) { this.avatarSystem.rules = this.avatarSystem.rules.filter(r => r.id!==ruleId) },

    closeProject() {
      this.projectPath=null; this.meta=null; this.avatarSystem={default:'',rules:[]};
      this.statsSystem={stats:[...DEFAULT_STATS]}; this.days={}; this.currentDay='day1'; this.currentSceneId=null;
      this.assets={all:[],images:[],sounds:[]}
    },

    addSceneToCurrentDay(sceneId?: string, preset: Partial<Scene> = {}) {
      const rawId = sceneId || `scene_${Date.now()}`
      const newId = toSafeDayId(rawId) || rawId.toLowerCase().replace(/[^a-z0-9_]/g,'_')
      const currentDayNum = Object.keys(this.days).indexOf(this.currentDay)+1 || 1
      const safeDay = toSafeDayId(this.currentDay)

      const base: Scene = {
        Id: newId,
        SceneTitle: preset.SceneTitle || newId,
        Background: fixRelPath(preset.Background || 'images/bg_tutorial.webp'),
        Text: preset.Text || '',
        Choices: preset.Choices && preset.Choices.length ? preset.Choices : [],
        IsEndDay: preset.IsEndDay || false,
        Type: preset.Type || (preset.IsEndDay? 'end_of_day' : 'normal'),
        Day: preset.Day || currentDayNum,
        NextDayId: preset.NextDayId || (preset as any).NextDay || undefined,
        NextDay: (preset as any).NextDay || preset.NextDayId || undefined,
        Transfers: preset.Transfers || { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' }
      }
      // merge ale nie nadpisuj pustym Choices
      const merged: Scene = { ...base }
      if (preset.SceneTitle) merged.SceneTitle = preset.SceneTitle
      if (preset.Text) merged.Text = preset.Text
      if (preset.Background) merged.Background = fixRelPath(preset.Background)
      if (preset.NextDayId) merged.NextDayId = toSafeDayId(preset.NextDayId) || preset.NextDayId
      if ((preset as any).NextDay) merged.NextDay = toSafeDayId((preset as any).NextDay) || (preset as any).NextDay
      if (preset.Transfers) merged.Transfers = preset.Transfers
      if (preset.Type) merged.Type = preset.Type
      if (preset.Day) merged.Day = preset.Day
      merged.Id = newId
      merged.IsEndDay = preset.IsEndDay || base.IsEndDay
      if (preset.Choices && preset.Choices.length) merged.Choices = preset.Choices

      if (merged.IsEndDay) {
        merged.Type = 'end_of_day'
        if (!merged.Transfers) merged.Transfers = { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' }
        if (!merged.Choices || merged.Choices.length===0) {
          const target = merged.NextDayId || 'END'
          let nextScene = 'END'
          if (target !== 'END') {
            nextScene = this.getDayStartSceneId(target) || 'END'
          }
          merged.Choices = [{ id: genId(), Text: 'Zakończ', Next: nextScene, Stats: {} } as any]
        }
      } else {
        if (!merged.Choices?.length) merged.Choices = []
      }
      if(!this.days[this.currentDay]) this.days[this.currentDay]=[]
      const scenes = this.days[this.currentDay]

      if (merged.IsEndDay) {
        scenes.push(merged)
      } else {
        const firstEndIdx = scenes.findIndex(s => s.IsEndDay)
        if (firstEndIdx!== -1) {
          scenes.splice(firstEndIdx, 0, merged)
        } else {
          scenes.push(merged)
        }
      }
      this.currentSceneId=newId
      this.ensureChoiceIds(merged)
      this.normalizeScene(merged)
    },

    moveSceneInCurrentDay(fromIndex: number, dir: -1 | 1) {
      const scenes = this.days[this.currentDay]
      if (!scenes) return
      const toIndex = fromIndex + dir
      if (toIndex < 0 || toIndex >= scenes.length) return
      if (scenes[fromIndex]?.IsEndDay!== scenes[toIndex]?.IsEndDay) return
      const [moved] = scenes.splice(fromIndex, 1)
      scenes.splice(toIndex, 0, moved)
    },

    duplicateScene(sceneId: string) {
      const scenes = this.days[this.currentDay]; if(!scenes) return;
      const toCopy = scenes.find(s => s.Id===sceneId); if(!toCopy) return;
      const copy = JSON.parse(JSON.stringify(toCopy));
      copy.Id=`${toCopy.Id}_copy_${Date.now()}`;
      copy.SceneTitle=`${toCopy.SceneTitle||sceneId} - Kopia`;
      this.ensureChoiceIds(copy);
      const idx = scenes.findIndex(s=>s.Id===sceneId)
      if (copy.IsEndDay) {
        const lastEndIdx = scenes.map((s,i)=> s.IsEndDay? i : -1).filter(i=>i!==-1).pop()?? scenes.length-1
        scenes.splice(lastEndIdx+1, 0, copy)
      } else {
        scenes.splice(idx+1,0,copy);
      }
      this.currentSceneId=copy.Id
    },

    deleteScene(sceneId: string) {
      const scenes=this.days[this.currentDay];
      if(!scenes||scenes.length<=1){alert('Nie możesz usunąć ostatniej sceny');return}
      const target = scenes.find(s => s.Id===sceneId)
      if (target?.IsEndDay) {
        if (!confirm(`Usuwasz KONIEC DNIA (${sceneId}). Na pewno?`)) return
      }
      const idx=scenes.findIndex(s=>s.Id===sceneId);
      if(idx>-1){scenes.splice(idx,1); if(this.currentSceneId===sceneId) this.currentSceneId=scenes[0]?.Id||null}
    },

    updateCurrentScene(field: keyof Scene, value: any) {
      if(!this.currentScene) return;
      if(field==='Background'){
        value = fixRelPath(value as string)
      }
      if(field==='NextDayId' || field==='NextDay'){
        value = (value as string).replace('.json','').trim()
        if(value==='') value = undefined
        else value = toSafeDayId(value) || value
      }
      (this.currentScene as any)[field]=value
      if (this.currentScene.IsEndDay && (field === 'NextDayId' || field === 'NextDay')) {
        const clean = value as string | undefined
        this.currentScene.NextDayId = clean
        this.currentScene.NextDay = clean
        if (clean && clean!=='END') {
          const startId = this.getDayStartSceneId(clean)
          if (startId && this.currentScene.Choices && this.currentScene.Choices[0]) {
            this.currentScene.Choices[0].Next = startId
          }
        } else {
          if (this.currentScene.Choices && this.currentScene.Choices[0]) {
            this.currentScene.Choices[0].Next = 'END'
          }
        }
      }
    },

    addDay(dayId: string) {
      const safeInput = toSafeDayId(dayId)
      if(!safeInput){ alert('Nieprawidłowa nazwa dnia'); return }
      if(this.days[safeInput]){alert(`Dzień "${safeInput}" już istnieje`);return}
      const dayNum = Object.keys(this.days).length+1
      const startId = `start_${safeInput}`
      const endId = `koniec_${safeInput}`
      this.days[safeInput]=[
        {
          Id: startId,
          SceneTitle:`${safeInput} - Start`,
          Background:'images/bg_tutorial.webp',
          Text:`Początek ${safeInput}`,
          Choices:[{ id: genId(), Text:'Dalej', Next:endId, Stats: this.createEmptyStats() }],
          Type:'normal' as const,
          Day: dayNum
        },
        {
          Id:endId,
          SceneTitle:`KONIEC ${safeInput}`,
          Background:'images/bg_tutorial.webp',
          Text:`Koniec ${safeInput}. Idziesz spać.`,
          IsEndDay: true,
          Type:'end_of_day' as const,
          Day: dayNum,
          NextDayId: undefined,
          NextDay: undefined,
          Transfers: { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' },
          Choices:[{ id: genId(), Text:'Zakończ dzień', Next: 'END', Stats: {} }]
        }
      ];
      this.currentDay=safeInput; this.currentSceneId=startId
    },

    addChoiceToCurrentScene() {
      if(!this.currentScene) return;
      if(!this.currentScene.Choices) this.currentScene.Choices=[];
      const isEnd = this.currentScene.IsEndDay
      this.currentScene.Choices.push({
        id: genId(),
        Text: isEnd? 'Przejdź dalej' : 'Nowy wybór',
        Next: isEnd? (this.currentScene.NextDayId || 'END') : this.currentScene.Id,
        Stats: isEnd? {} : this.createEmptyStats(),
        ReactionText:'',
        ReactionImage:'',
        SoundFile:''
      })
    }
  }
})
