import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { trainingApi, type NewSet, type FinishWorkout } from '../api/training'
import { exercisesApi, type Exercise } from '../api/exercises'

export function useWorkouts() {
  return useQuery({ queryKey: ['workouts'], queryFn: () => trainingApi.workouts() })
}

export function useWorkout(id: number) {
  return useQuery({ queryKey: ['workout', id], queryFn: () => trainingApi.workout(id) })
}

export function usePrograms(archived = false) {
  return useQuery({
    queryKey: ['programs', archived],
    queryFn: () => trainingApi.programs(archived),
  })
}

export function useProgram(id: number) {
  return useQuery({ queryKey: ['program', id], queryFn: () => trainingApi.program(id) })
}

// useExerciseMap — id -> упражнение (для подписей подходов/предписаний).
export function useExerciseMap() {
  return useQuery({
    queryKey: ['exercise-map'],
    queryFn: async () => {
      const list = await exercisesApi.list({})
      return new Map<number, Exercise>(list.map((e) => [e.id, e]))
    },
    staleTime: 10 * 60 * 1000,
  })
}

export function useArchiveProgram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (v: { id: number; archived: boolean }) =>
      v.archived ? trainingApi.archiveProgram(v.id) : trainingApi.unarchiveProgram(v.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['programs'] }),
  })
}

// useProgramDay — день программы с предписаниями (цели по подходам логгера).
export function useProgramDay(dayId?: number) {
  return useQuery({
    queryKey: ['program-day', dayId],
    queryFn: () => trainingApi.programDay(dayId!),
    enabled: !!dayId,
  })
}

// useExerciseHistory — прошлые сессии упражнения (столбец «Прошлый»).
export function useExerciseHistory(exerciseId?: number) {
  return useQuery({
    queryKey: ['exercise-history', exerciseId],
    queryFn: () => trainingApi.exerciseHistory(exerciseId!),
    enabled: !!exerciseId,
    staleTime: 60 * 1000,
  })
}

export function useStartWorkout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { date?: string; title?: string; program_day_id?: number }) =>
      trainingApi.createWorkout(body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workouts'] }),
  })
}

// Мутации подходов инвалидируют конкретную тренировку.
export function useAddSet(workoutId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: NewSet) => trainingApi.addSet(workoutId, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workout', workoutId] }),
  })
}

export function useUpdateSet(workoutId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (v: { setId: number; body: Partial<NewSet> }) =>
      trainingApi.updateSet(v.setId, v.body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workout', workoutId] }),
  })
}

export function useDeleteSet(workoutId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (setId: number) => trainingApi.deleteSet(setId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workout', workoutId] }),
  })
}

export function useFinishWorkout(workoutId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: FinishWorkout) => trainingApi.finishWorkout(workoutId, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workout', workoutId] })
      qc.invalidateQueries({ queryKey: ['workouts'] })
    },
  })
}
