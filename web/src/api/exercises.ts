import { api } from './client'

export type MuscleGroup = {
  id: number
  slug: string
  name_ru: string
  weekly_mev?: number
  weekly_mav?: number
}

export type ExerciseKind = 'compound' | 'isolation' | 'isometric' | 'bodyweight' | 'cardio'

export type Exercise = {
  id: number
  name: string
  muscle_group_id: number
  kind: ExerciseKind
  per_arm: boolean
  technique_notes: string
  equipment?: string
  global: boolean
  archived: boolean
  aliases?: string[]
}

export type ExerciseQuery = {
  q?: string
  muscleGroup?: string
}

export type NewExercise = {
  name: string
  muscle_group: string // slug
  kind: ExerciseKind
  per_arm: boolean
  equipment?: string
  technique_notes?: string
}

export const exercisesApi = {
  muscleGroups: () => api.get<MuscleGroup[]>('/api/v1/muscle-groups'),
  list: (params: ExerciseQuery = {}) => {
    const qs = new URLSearchParams()
    if (params.q) qs.set('q', params.q)
    if (params.muscleGroup) qs.set('muscle_group', params.muscleGroup)
    const suffix = qs.toString() ? `?${qs}` : ''
    return api.get<Exercise[]>(`/api/v1/exercises${suffix}`)
  },
  get: (id: number) => api.get<Exercise>(`/api/v1/exercises/${id}`),
  create: (body: NewExercise) => api.post<Exercise>('/api/v1/exercises', body),
  update: (id: number, body: NewExercise) => api.patch<Exercise>(`/api/v1/exercises/${id}`, body),
}

export const equipmentLabel: Record<string, string> = {
  '': '—',
  barbell: 'штанга',
  dumbbell: 'гантели',
  machine: 'тренажёр',
  cable: 'блок',
  bodyweight: 'свой вес',
  band: 'резина',
  kettlebell: 'гиря',
  other: 'другое',
  none: 'без снаряда',
}

export const kindLabel: Record<ExerciseKind, string> = {
  compound: 'базовое',
  isolation: 'изолирующее',
  isometric: 'статика',
  bodyweight: 'свой вес',
  cardio: 'кардио',
}
