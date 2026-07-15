import { useQuery } from '@tanstack/react-query'
import { exercisesApi, type ExerciseQuery } from '../api/exercises'

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
