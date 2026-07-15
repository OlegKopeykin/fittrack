import { http, HttpResponse } from 'msw'

// Изменяемое состояние мок-бэкенда для тестов auth.
type State = {
  me: { id: number; username: string; display_name: string; role: string } | null
}
export const mockState: State = { me: null }

export function resetMock() {
  mockState.me = null
}

function errorBody(code: string, message: string, fields?: Record<string, string>) {
  return { error: { code, message, ...(fields ? { fields } : {}) } }
}

export const handlers = [
  http.get('/api/v1/auth/me', () => {
    if (!mockState.me) {
      return HttpResponse.json(errorBody('unauthorized', 'Unauthorized'), { status: 401 })
    }
    return HttpResponse.json(mockState.me)
  }),

  http.post('/api/v1/auth/login', async ({ request }) => {
    const { username, password } = (await request.json()) as {
      username: string
      password: string
    }
    if (username === 'oleg' && password === 'верный-пароль') {
      mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
      return HttpResponse.json(mockState.me)
    }
    return HttpResponse.json(errorBody('invalid_credentials', 'Unauthorized'), { status: 401 })
  }),

  http.post('/api/v1/auth/register', async ({ request }) => {
    const body = (await request.json()) as {
      invite_code: string
      username: string
      password: string
    }
    if (body.invite_code !== 'GOODINVITE01') {
      return HttpResponse.json(errorBody('invalid_invite', 'Bad Request'), { status: 400 })
    }
    mockState.me = { id: 2, username: body.username, display_name: '', role: 'user' }
    return HttpResponse.json(mockState.me, { status: 201 })
  }),

  http.post('/api/v1/auth/logout', () => {
    mockState.me = null
    return new HttpResponse(null, { status: 204 })
  }),
]
