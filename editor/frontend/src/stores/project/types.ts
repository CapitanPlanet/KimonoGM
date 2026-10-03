export interface StatDef { id: string; name: string; initial: number }
export interface Choice {
  id?: string; Text: string; Next: string; NextDayId?: string; NextDay?: string;
  Stats?: Record<string, number>; ReactionText?: string; ReactionImage?: string; SoundFile?: string;
  FlagsSet?: string[]; FlagsRequired?: string[]; MinPortfel?: number | null; KosztPortfel?: number | null; FailText?: string;
  [key: string]: any
}
export interface EndDayTransfers { keep_flags: string[]; keep_stats: string[]; summary_text?: string; nextDay?: string }
export interface Scene {
  Id: string; SceneTitle?: string; Background?: string; Text?: string; Choices?: Choice[];
  IsEndDay?: boolean; Type?: 'normal' | 'end_of_day'; Day?: number; NextDay?: string; NextDayId?: string;
  Transfers?: EndDayTransfers | null; [key: string]: any
}
export type ConditionOp = { gte?: number; lte?: number; gt?: number; lt?: number; eq?: number }
export interface AvatarRule { id: string; use: string; if: Record<string, ConditionOp>; priority: number; when?: any }
export interface AvatarSystem { default: string; rules: AvatarRule[] }
export interface StatsSystem { stats: StatDef[] }
export interface ProjectMeta { gameName: string; author: string; version: string; engineVersion: string; startDay: string; startScene: string }
