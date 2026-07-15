import { api } from './client'

export type User = {
  id: number
  username: string
  display_name: string
  role: 'owner' | 'user'
}

export const authApi = {
  me: () => api.get<User>('/api/v1/auth/me'),
  login: (username: string, password: string) =>
    api.post<User>('/api/v1/auth/login', { username, password }),
  register: (invite_code: string, username: string, password: string) =>
    api.post<User>('/api/v1/auth/register', { invite_code, username, password }),
  logout: () => api.post<void>('/api/v1/auth/logout'),
}
