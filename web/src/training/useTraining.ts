import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { trainingApi } from '../api/training'
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
