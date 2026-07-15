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

  http.get('/api/v1/muscle-groups', () =>
    HttpResponse.json([
      { id: 1, slug: 'chest', name_ru: 'Грудь', weekly_mev: 8, weekly_mav: 18 },
      { id: 2, slug: 'quads', name_ru: 'Квадрицепс', weekly_mev: 8, weekly_mav: 18 },
      { id: 3, slug: 'lats', name_ru: 'Широчайшие', weekly_mev: 6, weekly_mav: 18 },
    ]),
  ),

  http.get('/api/v1/exercises', ({ request }) => {
    const url = new URL(request.url)
    const q = (url.searchParams.get('q') ?? '').toLowerCase()
    const group = url.searchParams.get('muscle_group') ?? ''
    let list = [
      { id: 10, name: 'Присед в Смите', muscle_group_id: 2, kind: 'compound', per_arm: false, technique_notes: '', global: true, archived: false, aliases: ['high-bar'] },
      { id: 11, name: 'Жим гантелей лёжа', muscle_group_id: 1, kind: 'compound', per_arm: true, technique_notes: '', global: true, archived: false },
      { id: 12, name: 'Тяга верхнего блока', muscle_group_id: 3, kind: 'compound', per_arm: false, technique_notes: '', global: true, archived: false, aliases: ['подтягивания'] },
    ]
    if (group === 'chest') list = list.filter((e) => e.muscle_group_id === 1)
    if (q === 'подтягивания') list = list.filter((e) => e.id === 12)
    else if (q) list = list.filter((e) => e.name.toLowerCase().includes(q))
    return HttpResponse.json(list)
  }),

  http.get('/api/v1/programs', () =>
    HttpResponse.json([
      { id: 1, name: 'Фул бади', description: 'A/B' },
      { id: 2, name: '5-дневный сплит' },
    ]),
  ),

  http.get('/api/v1/programs/1', () =>
    HttpResponse.json({
      id: 1,
      name: 'Фул бади',
      description: 'A/B',
      days: [
        {
          id: 11,
          position: 0,
          name: 'День A',
          exercises: [
            { id: 100, exercise_id: 10, position: 0, sets: 3, rep_min: 6, rep_max: 10, weight_min_kg: 70, weight_max_kg: 90, tempo: '3-0-1' },
          ],
        },
      ],
    }),
  ),

  http.get('/api/v1/workouts', () =>
    HttpResponse.json({
      items: [{ id: 500, date: '2026-05-10', feeling: 'бодро', bodyweight_kg: 86.5, notes: '' }],
      next_cursor: '',
    }),
  ),

  http.get('/api/v1/workouts/500', () =>
    HttpResponse.json({
      id: 500,
      date: '2026-05-10',
      feeling: 'бодро',
      bodyweight_kg: 86.5,
      notes: '',
      sets: [
        { id: 1, exercise_id: 10, position: 0, role: 'warmup', weight_kg: 40, reps: 12 },
        { id: 2, exercise_id: 10, position: 1, role: 'working', weight_kg: 60, reps: 12 },
      ],
    }),
  ),
]
