import { api } from './client'

export type WorkoutSet = {
  id: number
  exercise_id: number
  position: number
  role: 'warmup' | 'ramp' | 'working'
  weight_kg?: number
  reps?: number
  distance_km?: number
  duration_sec?: number
  note?: string
}

export type Workout = {
  id: number
  date: string
  title?: string
  started_at?: string
  finished_at?: string
  bodyweight_kg?: number
  feeling: string
  notes: string
  sets?: WorkoutSet[]
}

export type WorkoutPage = { items: Workout[]; next_cursor: string }

export type Prescription = {
  id: number
  exercise_id: number
  position: number
  sets: number
  rep_min?: number
  rep_max?: number
  weight_min_kg?: number
  weight_max_kg?: number
  rest_sec?: number
  tempo?: string
  notes?: string
}

export type ProgramDay = {
  id: number
  position: number
  name: string
  notes?: string
  exercises: Prescription[]
}

export type Program = {
  id: number
  name: string
  description?: string
  archived?: boolean
  days?: ProgramDay[]
}

export const trainingApi = {
  workouts: (limit = 50) => api.get<WorkoutPage>(`/api/v1/workouts?limit=${limit}`),
  workout: (id: number) => api.get<Workout>(`/api/v1/workouts/${id}`),
  programs: (archived = false) =>
    api.get<Program[]>(`/api/v1/programs${archived ? '?archived=1' : ''}`),
  program: (id: number) => api.get<Program>(`/api/v1/programs/${id}`),
  archiveProgram: (id: number) => api.post<void>(`/api/v1/programs/${id}/archive`),
  unarchiveProgram: (id: number) => api.post<void>(`/api/v1/programs/${id}/unarchive`),
}
