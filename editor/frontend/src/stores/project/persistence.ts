import { ReadJSON, WriteJSON, ListFiles, ListAssets, ListAudioAssets, CreateProject, DeleteFile } from '../../../wailsjs/go/main/App'
import type { Scene } from './types'
import { fixRelPath, toSafeName, genId } from './utils'
import { DEFAULT_STATS } from './constants'
export function createPersistenceActions(){
  return {
    async loadAssets(this:any){
      if(!this.projectPath){ this.assets.all=[]; this.assets.images=[]; this.assets.sounds=[]; return }
      try{
        const [images,sounds]=await Promise.all([ListAssets(this.projectPath).catch(()=>[] as string[]),ListAudioAssets(this.projectPath).catch(()=>[] as string[])])
        this.assets.images=(images||[]).map(fixRelPath); this.assets.sounds=(sounds||[]).map(fixRelPath); this.assets.all=[...this.assets.images,...this.assets.sounds]
      }catch{ this.assets.all=[]; this.assets.images=[]; this.assets.sounds=[] }
      this.ensureAvatarSystem()
    },
    async scanAssets(this:any){ await this.loadAssets() },
    async refreshAssets(this:any){ await this.loadAssets() },
    async createProjectAtPath(this:any,p:string,name:string){
      p=p.replace(/\\/g,'/').replace(/\/+/g,'/').replace(/\/$/,'')
      try{ await CreateProject(p,name) }catch{}
      await this.loadProjectFromPath(p); return p
    },
    async deleteDay(this:any,dayId:string){
      if(this.dayFileList.length<=1){ alert('Musisz zostawić minimum 1 dzień!'); return }
      if(this.projectPath){ try{ await DeleteFile(this.projectPath,`Data/${dayId}.json`) }catch{} }
      delete this.days[dayId]
      if(this.currentDay===dayId){ this.currentDay=Object.keys(this.days)[0]; this.currentSceneId=this.days[this.currentDay]?.[0]?.Id||null }
    },
    async loadProjectFromPath(this:any,path:string){
      path=path.replace(/\\/g,'/').replace(/\/+/g,'/').replace(/\/$/,'')
      this.projectPath=path
      try{
        const metaRaw=await ReadJSON(`${path}/project.janproj`)
        const parsed=JSON.parse(metaRaw)
        this.meta={gameName:parsed.gameName||'',author:parsed.author||'',version:parsed.version||'1.0.0',engineVersion:parsed.engineVersion||'2.0.0',startDay:parsed.startDay||'day1',startScene:parsed.startScene||'start'}
        let incomingStats=parsed.statsSystem?.stats||[...DEFAULT_STATS]
        const migrated=incomingStats.map((s:any)=>{const finalName=toSafeName((s.name||s.id||'').toString()); return {id:finalName,name:finalName,initial:s.initial??0}}).filter((s:any)=>s.name)
        this.statsSystem={stats:migrated.length?migrated:[...DEFAULT_STATS]}
        this.ensureStatsSystem()
        this.avatarSystem={default:fixRelPath(parsed.avatarSystem?.default||''),rules:Array.isArray(parsed.avatarSystem?.rules)?parsed.avatarSystem.rules:[]}
        this.avatarSystem.rules.forEach((r:any)=>{if(!r.id) r.id=genId(); if(r.priority==null) r.priority=0; if(r.use) r.use=fixRelPath(r.use)})
        let dayFiles:string[]=[]; try{ dayFiles=await ListFiles(`${path}/Data`,'.json')||[] }catch{ dayFiles=['day1.json'] }
        dayFiles=dayFiles.filter(f=>!f.startsWith('_')&&!f.startsWith('.'))
        this.days={}
        for(const file of dayFiles){
          const dayName=file.replace('.json','')
          try{ const raw=await ReadJSON(`${path}/Data/${file}`); const data=JSON.parse(raw) as Scene[]; data.forEach(s=>{this.normalizeScene(s); this.ensureChoiceIds(s)}); this.days[dayName]=data }catch{}
        }
        this.currentDay=this.meta.startDay||Object.keys(this.days)[0]||'day1'
        this.currentSceneId=this.meta.startScene||this.days[this.currentDay]?.[0]?.Id||null
        await this.loadAssets(); this.ensureAvatarSystem()
      }catch(e){ console.error(e); this.projectPath=null; throw e }
    },
    async loadProject(this:any,path:string){ return this.loadProjectFromPath(path) },
    async saveProject(this:any){
      if(!this.projectPath||!this.meta) return
      this.saveStatus='Zapisywanie...'
      try{
        this.ensureAvatarSystem(); this.ensureStatsSystem()
        const toSave={...this.meta,avatarSystem:{default:fixRelPath(this.avatarSystem.default),rules:this.avatarSystem.rules.map((r:any)=>({...r,use:fixRelPath(r.use)}))},statsSystem:{stats:this.statsSystem.stats}}
        await WriteJSON(`${this.projectPath}/project.janproj`,JSON.stringify(toSave,null,2))
        for(const dayFile of Object.keys(this.days)){
          const cleaned=this.days[dayFile].map((scene:any)=>({...scene,Background:fixRelPath(scene.Background||'images/bg_tutorial.webp'),Choices:scene.Choices?.map((c:any)=>{const {id,...rest}=c; let stats=rest.Stats; if(scene.IsEndDay) stats={}; return {...rest,Stats:stats,ReactionImage:fixRelPath(rest.ReactionImage||''),SoundFile:fixRelPath(rest.SoundFile||'')} })}))
          await WriteJSON(`${this.projectPath}/Data/${dayFile}.json`,JSON.stringify(cleaned,null,2))
        }
        const manifest={days:Object.keys(this.days).sort(),startDay:this.meta.startDay||Object.keys(this.days)[0]||'day1',startScene:this.meta.startScene||this.days[this.meta.startDay||Object.keys(this.days)[0]]?.[0]?.Id||'start',version:new Date().toISOString()}
        await WriteJSON(`${this.projectPath}/Data/_manifest.json`,JSON.stringify(manifest,null,2))
        this.saveStatus='Zapisano'; setTimeout(()=>{this.saveStatus=''},2500)
      }catch(e){ console.error(e); this.saveStatus='Błąd zapisu: '+e }
    },
    closeProject(this:any){ this.projectPath=null; this.meta=null; this.avatarSystem={default:'',rules:[]}; this.statsSystem={stats:[...DEFAULT_STATS]}; this.days={}; this.currentDay='day1'; this.currentSceneId=null; this.assets={all:[],images:[],sounds:[]} }
  }
}
