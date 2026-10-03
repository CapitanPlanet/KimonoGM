import type { StatDef } from './types'
export const DEFAULT_STATS: StatDef[] = [
  { id: 'CEBULA', name: 'CEBULA', initial: 0 },
  { id: 'WSTYD', name: 'WSTYD', initial: 0 },
  { id: 'PORTFEL', name: 'PORTFEL', initial: 0 },
  { id: 'REPUTACJA', name: 'REPUTACJA', initial: 0 },
]
export const LEGACY_MAP: Record<string,string> = { A:'CEBULA', B:'WSTYD', C:'PORTFEL', D:'REPUTACJA', PORTFEL1:'PORTFEL' }
