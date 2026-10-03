import type { Scene } from './types'
import { genId } from './utils'
export function createEmptyDay(dayId: string, dayNum: number, statNames: string[]): Scene[] {
  const startId = `start_${dayId}`
  const endId = `koniec_${dayId}`
  const emptyStats = Object.fromEntries(statNames.map(n => [n, 0]))
  return [
    { Id: startId, SceneTitle: `${dayId} - Start`, Background:'images/bg_tutorial.webp', Text:`Początek ${dayId}`, Choices:[{ id: genId(), Text:'Dalej', Next:endId, Stats: emptyStats }], Type:'normal' as const, Day: dayNum },
    { Id: endId, SceneTitle:`KONIEC ${dayId}`, Background:'images/bg_tutorial.webp', Text:`Koniec ${dayId}. Idziesz spać.`, IsEndDay: true, Type:'end_of_day' as const, Day: dayNum, Transfers: { keep_flags: [], keep_stats: statNames, summary_text: '' }, Choices:[{ id: genId(), Text:'Zakończ dzień', Next: 'END', Stats: {} }] }
  ]
}
