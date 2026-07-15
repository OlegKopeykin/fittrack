import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseQueryResult,
} from '@tanstack/react-query'
import { authApi, type User } from '../api/auth'
import { ApiError } from '../api/client'

const ME_KEY = ['auth', 'me'] as const

// useMe загружает текущего пользователя; 401 трактуется как «не залогинен»
// (data === null), а не как ошибка запроса.
export function useMe(): UseQueryResult<User | null> {
  return useQuery({
    queryKey: ME_KEY,
    queryFn: async () => {
      try {
        return await authApi.me()
      } catch (e) {
        if (e instanceof ApiError && e.status === 401) return null
        throw e
      }
    },
    staleTime: 5 * 60 * 1000,
  })
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (v: { username: string; password: string }) =>
      authApi.login(v.username, v.password),
    onSuccess: (user) => qc.setQueryData(ME_KEY, user),
  })
}

export function useRegister() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (v: { invite_code: string; username: string; password: string }) =>
      authApi.register(v.invite_code, v.username, v.password),
    onSuccess: (user) => qc.setQueryData(ME_KEY, user),
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      qc.setQueryData(ME_KEY, null)
      qc.clear()
    },
  })
}
