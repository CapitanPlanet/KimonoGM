import { LEGACY_MAP } from './constants'
export function toSafeName(raw: string): string { const up = raw.toString().trim().toUpperCase(); return LEGACY_MAP[up] || up }
export function getBasename(path: string): string { return path.replace(/\\/g,'/').split('/').pop() || '' }
export function isImageOfType(path: string, prefix: 'bg_'|'re_'|'av_'): boolean { const base = getBasename(path).toLowerCase(); return base.startsWith(prefix) }
export function fixRelPath(input: string): string {
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
  if (p === 'images/bg_tutorial.jpg') return 'images/bg_tutorial.webp'
  if (p === 'images/bg_front.jpg') return 'images/bg_tutorial.webp'
  if (p === 'images/bg_sen_ jaanusza.jpg') return 'images/bg_tutorial.webp'
  return p
}
export function isEndDayId(id: string): boolean { return id.startsWith('koniec_') || id.startsWith('koniec_dnia_') || id.startsWith('koniec_dnia') }
export function genId(): string { return (crypto as any).randomUUID?.() || Math.random().toString(36).slice(2) }
export function toSafeDayId(raw: string): string { let s = raw.trim().toLowerCase().replace(/\s+/g,'_').replace(/[^a-z0-9_\-]/g,'').replace(/_+/g,'_').replace(/-+/g,'-').replace(/^_+|_+$/g,''); return s.slice(0,40) }
