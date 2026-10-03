import { defineStore } from 'pinia'
import type { Scene, AvatarRule, ProjectMeta, StatDef } from './project/types'
import { DEFAULT_STATS } from './project/constants'
import { isImageOfType, isEndDayId } from './project/utils'
import { createPersistenceActions } from './project/persistence'
import { createSceneActions } from './project/sceneActions'
import { createStatsActions } from './project/statsAvatarActions'

export type { Scene, StatDef, ProjectMeta, AvatarRule }

export const useProjectStore = defineStore('project', {
  state: () => ({
    projectPath: null as string | null,
    meta: null as ProjectMeta | null,
    avatarSystem: { default: '', rules: [] } as { default: string; rules: AvatarRule[] },
    statsSystem: { stats: [...DEFAULT_STATS] } as { stats: StatDef[] },
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
    currentScene: (s): Scene | null => { const scenes = s.days[s.currentDay] || []; return scenes.find(x => x.Id === s.currentSceneId) || null },
    sceneIdsInCurrentDay: (s) => (s.days[s.currentDay] || []).map(x => x.Id),
    allSceneIds: (s) => { const ids: string[]=[]; Object.values(s.days).forEach((sc: any)=>sc.forEach((c: Scene)=>ids.push(c.Id))); return ids },
    allScenes: (s) => { const all: Scene[]=[]; Object.values(s.days).forEach((sc: any)=>all.push(...sc)); return all },
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
   ...createPersistenceActions() as any,
   ...createSceneActions() as any,
   ...createStatsActions() as any,
  }
})
