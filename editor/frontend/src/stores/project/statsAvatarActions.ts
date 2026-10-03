import { toSafeName, fixRelPath } from './utils'
import { DEFAULT_STATS } from './constants'
import type { StatDef, AvatarRule } from './types'

export function createStatsActions(){
  return {
    ensureStatsSystem(this: any){
      if(!this.statsSystem||!Array.isArray(this.statsSystem.stats)||!this.statsSystem.stats.length){ this.statsSystem={stats:[...DEFAULT_STATS]}; return }
      const used=new Set<string>(); const fixed: StatDef[]=[]
      for(const s of this.statsSystem.stats as StatDef[]){ if(!s.name&&!(s as any).id) continue; const finalName=toSafeName((s.name||(s as any).id) as string); if(!finalName) continue; if(used.has(finalName)) continue; used.add(finalName); fixed.push({id:finalName,name:finalName,initial:s.initial??0}) }
      this.statsSystem.stats=fixed.length?fixed:[...DEFAULT_STATS]
    },
    addStat(this: any, name: string){ const safe=toSafeName(name); if(!safe) return; if(this.statsSystem.stats.some((s:StatDef)=>s.name===safe)) return; this.statsSystem.stats.push({id:safe,name:safe,initial:0}) },
    deleteStat(this: any, name: string){ this.statsSystem.stats=this.statsSystem.stats.filter((s:StatDef)=>s.name!==name) },
    ensureAvatarSystem(this: any){
      this.ensureStatsSystem()
      if(!this.avatarSystem) this.avatarSystem={default:'',rules:[]}
      if(!this.avatarSystem.rules) this.avatarSystem.rules=[]
      this.avatarSystem.default=fixRelPath(this.avatarSystem.default||'')
    },
    addAvatarRule(this: any, rule: AvatarRule){ this.avatarSystem.rules.push(rule) },
    updateAvatarRule(this: any, id: string, patch: Partial<AvatarRule>){ const r=this.avatarSystem.rules.find((x:AvatarRule)=>x.id===id); if(r) Object.assign(r,patch) },
    deleteAvatarRule(this: any, id: string){ this.avatarSystem.rules=this.avatarSystem.rules.filter((x:AvatarRule)=>x.id!==id) },
    setDefaultAvatar(this: any, path: string){ this.avatarSystem.default=fixRelPath(path) }
  }
}
