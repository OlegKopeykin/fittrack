import { api } from './client'

export type Invite = {
  id: number
  code: string
  role: string
  created_at: string
  expires_at?: string
  used_at?: string
}

export const invitesApi = {
  create: (expiresInDays = 14) =>
    api.post<Invite>('/api/v1/invites', { role: 'user', expires_in_days: expiresInDays }),
}
