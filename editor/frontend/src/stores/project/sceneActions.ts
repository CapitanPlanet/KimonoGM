
import type { Scene, Choice } from './types'
import { genId, fixRelPath, toSafeDayId, isEndDayId } from './utils'

export function createSceneActions() {
  return {
    getDayStartSceneId(this: any, day: string): string | null {
      const scenes = this.days[day] as Scene[]; if(!scenes?.length) return null
      const start = scenes.find((s:Scene)=>!s.IsEndDay) || scenes[0]
      return start?.Id || null
    },
    selectScene(this: any, id: string){ this.currentSceneId=id },
    ensureChoiceIds(this: any, scene: Scene){ scene.Choices?.forEach(c=>{ if(!c.id) c.id=genId() }) },
    createEmptyStats(this: any): Record<string, number> { const o: Record<string, number>={}; this.statsSystem.stats.forEach((s:any)=>{ o[s.name]=0 }); return o },
    normalizeScene(this: any, scene: Scene){
      if(scene.Background) scene.Background=fixRelPath(scene.Background)
      scene.Choices?.forEach(c=>{ if(c.ReactionImage) c.ReactionImage=fixRelPath(c.ReactionImage) })
      this.ensureChoiceIds(scene)
    },
    addSceneToCurrentDay(this: any, preset: Partial<Scene> & { Id?: string }){
      if(!this.days[this.currentDay]) this.days[this.currentDay]=[]
      const scenes = this.days[this.currentDay] as Scene[]
      const newId = preset.Id || `scene_${Date.now()}`
      const merged: Scene = { Id: newId, SceneTitle: preset.SceneTitle || 'Nowa scena', Background: fixRelPath(preset.Background||'images/bg_tutorial.webp'), Text: preset.Text||'', Choices: preset.Choices||[], Type: (preset as any).Type||'normal', Day: scenes.length+1 } as Scene
      if((preset as any).IsEndDay) { merged.IsEndDay=true; merged.Type='end_of_day' }
      if(merged.IsEndDay){
        scenes.push(merged)
      } else {
        const firstEnd = scenes.findIndex((s:Scene)=>s.IsEndDay)
        if(firstEnd!==-1) scenes.splice(firstEnd,0,merged); else scenes.push(merged)
      }
      this.currentSceneId=newId
      this.ensureChoiceIds(merged)
      this.normalizeScene(merged)
    },
    moveSceneInCurrentDay(this: any, fromIndex: number, dir: -1|1){
      const scenes = this.days[this.currentDay] as Scene[]; if(!scenes) return
      const toIndex = fromIndex+dir; if(toIndex<0||toIndex>=scenes.length) return
      if(scenes[fromIndex]?.IsEndDay!==scenes[toIndex]?.IsEndDay) return
      const [moved]=scenes.splice(fromIndex,1); scenes.splice(toIndex,0,moved)
    },
    duplicateScene(this: any, sceneId: string){
      const scenes = this.days[this.currentDay] as Scene[]; if(!scenes) return
      const toCopy = scenes.find((s:Scene)=>s.Id===sceneId); if(!toCopy) return
      const copy = JSON.parse(JSON.stringify(toCopy)); copy.Id=`${toCopy.Id}_copy_${Date.now()}`; copy.SceneTitle=`${toCopy.SceneTitle||sceneId} - Kopia`
      this.ensureChoiceIds(copy)
      const idx = scenes.findIndex((s:Scene)=>s.Id===sceneId)
      if(copy.IsEndDay){ const lastEnd = scenes.map((s:Scene,i:number)=>s.IsEndDay?i:-1).filter((i:number)=>i!==-1).pop()??scenes.length-1; scenes.splice(lastEnd+1,0,copy) }
      else scenes.splice(idx+1,0,copy)
      this.currentSceneId=copy.Id
    },
    deleteScene(this: any, sceneId: string){
      const scenes = this.days[this.currentDay] as Scene[]; if(!scenes||scenes.length<=1){ alert('Nie możesz usunąć ostatniej sceny'); return }
      const idx = scenes.findIndex((s:Scene)=>s.Id===sceneId); if(idx>-1){ scenes.splice(idx,1); if(this.currentSceneId===sceneId) this.currentSceneId=scenes[0]?.Id||null }
    },
    updateCurrentScene(this: any, field: keyof Scene, value: any){
      if(!this.currentScene) return
      if(field==='Background') value=fixRelPath(value as string)
      if(field==='NextDayId'||field==='NextDay'){ value=(value as string).replace('.json','').trim(); if(value==='') value=undefined; else value=toSafeDayId(value)||value }
      ;(this.currentScene as any)[field]=value
    },
    addDay(this: any, dayId: string){
      const { toSafeDayId } = require('./utils')
      const safe = toSafeDayId(dayId); if(!safe){ alert('Nieprawidłowa nazwa dnia'); return }
      if(this.days[safe]){ alert(`Dzień "${safe}" już istnieje`); return }
      const { createEmptyDay } = require('./factory')
      const dayNum = Object.keys(this.days).length+1
      this.days[safe]=createEmptyDay(safe, dayNum, this.statsSystem.stats.map((s:any)=>s.name))
      this.currentDay=safe; this.currentSceneId=this.days[safe][0].Id
    },
    addChoiceToCurrentScene(this: any){
      if(!this.currentScene) return; if(!this.currentScene.Choices) this.currentScene.Choices=[]
      const isEnd = this.currentScene.IsEndDay
      this.currentScene.Choices.push({ id: genId(), Text: isEnd?'Przejdź dalej':'Nowy wybór', Next: isEnd?(this.currentScene.NextDayId||'END'):this.currentScene.Id, Stats: isEnd?{}:this.createEmptyStats(), ReactionText:'', ReactionImage:'', SoundFile:'' })
    }
  }
}
