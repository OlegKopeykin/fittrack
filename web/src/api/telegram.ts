import { api } from './client'

export type TelegramFrequency = 'daily' | 'weekly' | 'monthly'

export type TelegramSettings = {
  configured: boolean
  bot_username?: string
  chat_linked: boolean
  enabled: boolean
  frequency: TelegramFrequency
  last_sent_at?: string
}

export type TelegramPatch = {
  bot_token?: string
  frequency?: TelegramFrequency
  enabled?: boolean
}

export const telegramApi = {
  get: () => api.get<TelegramSettings>('/api/v1/profile/telegram'),
  set: (body: TelegramPatch) => api.put<TelegramSettings>('/api/v1/profile/telegram', body),
  link: () => api.post<TelegramSettings>('/api/v1/profile/telegram/link'),
  test: () => api.post<void>('/api/v1/profile/telegram/test'),
  remove: () => api.del<void>('/api/v1/profile/telegram'),
}
