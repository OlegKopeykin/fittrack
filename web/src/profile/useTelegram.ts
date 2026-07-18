import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { telegramApi, type TelegramPatch } from '../api/telegram'

export function useTelegram() {
  return useQuery({ queryKey: ['telegram'], queryFn: () => telegramApi.get() })
}

export function useSetTelegram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: TelegramPatch) => telegramApi.set(body),
    onSuccess: (data) => qc.setQueryData(['telegram'], data),
  })
}

export function useLinkTelegram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => telegramApi.link(),
    onSuccess: (data) => qc.setQueryData(['telegram'], data),
  })
}

export function useTestTelegram() {
  return useMutation({ mutationFn: () => telegramApi.test() })
}

export function useRemoveTelegram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => telegramApi.remove(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['telegram'] }),
  })
}
