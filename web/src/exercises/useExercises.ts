import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { exercisesApi, type ExerciseQuery, type NewExercise } from '../api/exercises'

export function useMuscleGroups() {
  return useQuery({
    queryKey: ['muscle-groups'],
    queryFn: () => exercisesApi.muscleGroups(),
    staleTime: 60 * 60 * 1000,
  })
}

export function useExercises(params: ExerciseQuery) {
  return useQuery({
    queryKey: ['exercises', params.q ?? '', params.muscleGroup ?? ''],
    queryFn: () => exercisesApi.list(params),
    staleTime: 5 * 60 * 1000,
  })
}

export function useExercise(id: number, enabled = true) {
  return useQuery({
    queryKey: ['exercise', id],
    queryFn: () => exercisesApi.get(id),
    enabled,
  })
}

// Мутации упражнений сбрасывают и список каталога, и карту имён логгера.
function invalidateExercises(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['exercises'] })
  qc.invalidateQueries({ queryKey: ['exercise-map'] })
}

export function useCreateExercise() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: NewExercise) => exercisesApi.create(body),
    onSuccess: () => invalidateExercises(qc),
  })
}

export function useExerciseNote(id: number) {
  return useQuery({
    queryKey: ['exercise-note', id],
    queryFn: () => exercisesApi.getNote(id),
  })
}

export function useSetExerciseNote(id: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (note: string) => exercisesApi.setNote(id, note),
    onSuccess: (data) => qc.setQueryData(['exercise-note', id], data),
  })
}

export function useUpdateExercise(id: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: NewExercise) => exercisesApi.update(id, body),
    onSuccess: () => {
      invalidateExercises(qc)
      qc.invalidateQueries({ queryKey: ['exercise', id] })
    },
  })
}
