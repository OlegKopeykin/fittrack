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
  program_day_id?: number | null
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

export type NewProgram = {
  name: string
  description?: string
  days: { name: string; exercises: { exercise_id: number }[] }[]
}

export type ProgramDayDetail = {
  id: number
  program_id: number
  program_name: string
  position: number
  name: string
  notes?: string
  exercises: Prescription[]
}

export type ExerciseSession = {
  date: string
  sets: {
    role: string
    weight_kg?: number
    reps?: number
    distance_km?: number
    duration_sec?: number
  }[]
}

export type NewSet = {
  exercise_id: number
  role?: 'warmup' | 'ramp' | 'working'
  weight_kg?: number
  reps?: number
  distance_km?: number
  duration_sec?: number
  client_id?: string
}

export type FinishWorkout = {
  finished_at?: string
  bodyweight_kg?: number
  feeling?: string
}

export const trainingApi = {
  workouts: (limit = 50) => api.get<WorkoutPage>(`/api/v1/workouts?limit=${limit}`),
  workout: (id: number) => api.get<Workout>(`/api/v1/workouts/${id}`),
  createWorkout: (body: { date?: string; title?: string; program_day_id?: number }) =>
    api.post<Workout>('/api/v1/workouts', body),
  finishWorkout: (id: number, body: FinishWorkout) =>
    api.patch<Workout>(`/api/v1/workouts/${id}`, body),
  deleteWorkout: (id: number) => api.del<void>(`/api/v1/workouts/${id}`),
  addSet: (workoutId: number, body: NewSet) =>
    api.post<WorkoutSet>(`/api/v1/workouts/${workoutId}/sets`, body),
  updateSet: (setId: number, body: Partial<NewSet>) =>
    api.patch<WorkoutSet>(`/api/v1/sets/${setId}`, body),
  deleteSet: (setId: number) => api.del<void>(`/api/v1/sets/${setId}`),
  exerciseHistory: (exerciseId: number) =>
    api.get<ExerciseSession[]>(`/api/v1/exercises/${exerciseId}/history`),
  programs: (archived = false) =>
    api.get<Program[]>(`/api/v1/programs${archived ? '?archived=1' : ''}`),
  program: (id: number) => api.get<Program>(`/api/v1/programs/${id}`),
  createProgram: (body: NewProgram) => api.post<Program>('/api/v1/programs', body),
  updateProgram: (id: number, body: NewProgram) =>
    api.put<Program>(`/api/v1/programs/${id}`, body),
  programDay: (dayId: number) => api.get<ProgramDayDetail>(`/api/v1/program-days/${dayId}`),
  archiveProgram: (id: number) => api.post<void>(`/api/v1/programs/${id}/archive`),
  unarchiveProgram: (id: number) => api.post<void>(`/api/v1/programs/${id}/unarchive`),
}
