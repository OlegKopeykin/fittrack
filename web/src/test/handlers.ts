import { http, HttpResponse } from 'msw'

// Изменяемое состояние мок-бэкенда для тестов auth.
type State = {
  me: { id: number; username: string; display_name: string; role: string } | null
}
export const mockState: State = { me: null }

// Stateful-хранилище тренировок для тестов логирования.
type MockSet = Record<string, unknown> & { id: number; exercise_id: number }
type MockWorkout = Record<string, unknown> & { id: number; sets: MockSet[] }
export const mockWorkouts = new Map<number, MockWorkout>()
let nextWorkoutId = 900
let nextSetId = 700

function seedWorkouts() {
  mockWorkouts.clear()
  nextWorkoutId = 900
  nextSetId = 700
  // 500 — завершённая тренировка (для read-only экрана деталей).
  mockWorkouts.set(500, {
    id: 500,
    date: '2026-05-10',
    title: 'Full-A',
    feeling: 'бодро',
    bodyweight_kg: 86.5,
    notes: '',
    finished_at: '2026-05-10T12:00:00Z',
    sets: [
      { id: 1, exercise_id: 10, position: 0, role: 'warmup', weight_kg: 40, reps: 12 },
      { id: 2, exercise_id: 10, position: 1, role: 'working', weight_kg: 60, reps: 12 },
    ],
  })
}
seedWorkouts()

export function resetMock() {
  mockState.me = null
  seedWorkouts()
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

  http.post('/api/v1/programs', async ({ request }) => {
    const body = (await request.json()) as {
      name: string
      description?: string
      days?: { name: string; exercises?: { exercise_id: number }[] }[]
    }
    return HttpResponse.json(
      {
        id: 700,
        name: body.name,
        description: body.description ?? '',
        days: (body.days ?? []).map((d, i) => ({
          id: 7000 + i,
          position: i,
          name: d.name,
          exercises: (d.exercises ?? []).map((e, j) => ({
            id: 70000 + i * 100 + j,
            exercise_id: e.exercise_id,
            position: j,
            sets: 0,
          })),
        })),
      },
      { status: 201 },
    )
  }),

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

  http.get('/api/v1/workouts', () => {
    const items = [...mockWorkouts.values()]
      .map(({ sets, ...w }) => {
        void sets
        return w
      })
      .sort((a, b) => b.id - a.id)
    return HttpResponse.json({ items, next_cursor: '' })
  }),

  http.post('/api/v1/workouts', async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>
    const id = ++nextWorkoutId
    const wk: MockWorkout = {
      id,
      date: (body.date as string) ?? '2026-07-16',
      title: (body.title as string) ?? '',
      program_day_id: body.program_day_id ?? undefined,
      started_at: '2026-07-16T09:00:00Z',
      feeling: '',
      notes: '',
      sets: [],
    }
    mockWorkouts.set(id, wk)
    return HttpResponse.json(wk, { status: 201 })
  }),

  http.get('/api/v1/workouts/:id', ({ params }) => {
    const wk = mockWorkouts.get(Number(params.id))
    if (!wk) return HttpResponse.json({ error: { code: 'not_found', message: 'нет' } }, { status: 404 })
    return HttpResponse.json(wk)
  }),

  http.patch('/api/v1/workouts/:id', async ({ params, request }) => {
    const wk = mockWorkouts.get(Number(params.id))
    if (!wk) return HttpResponse.json({ error: { code: 'not_found', message: 'нет' } }, { status: 404 })
    Object.assign(wk, (await request.json()) as Record<string, unknown>)
    return HttpResponse.json(wk)
  }),

  http.post('/api/v1/workouts/:id/sets', async ({ params, request }) => {
    const wk = mockWorkouts.get(Number(params.id))
    if (!wk) return HttpResponse.json({ error: { code: 'not_found', message: 'нет' } }, { status: 404 })
    const body = (await request.json()) as Record<string, unknown>
    const set: MockSet = {
      id: ++nextSetId,
      exercise_id: body.exercise_id as number,
      position: wk.sets.length,
      role: (body.role as string) ?? 'working',
      ...body,
    }
    wk.sets.push(set)
    return HttpResponse.json(set, { status: 201 })
  }),

  http.patch('/api/v1/sets/:id', async ({ params, request }) => {
    const sid = Number(params.id)
    for (const wk of mockWorkouts.values()) {
      const s = wk.sets.find((x) => x.id === sid)
      if (s) {
        Object.assign(s, (await request.json()) as Record<string, unknown>)
        return HttpResponse.json(s)
      }
    }
    return HttpResponse.json({ error: { code: 'not_found', message: 'нет' } }, { status: 404 })
  }),

  http.delete('/api/v1/sets/:id', ({ params }) => {
    const sid = Number(params.id)
    for (const wk of mockWorkouts.values()) {
      const i = wk.sets.findIndex((x) => x.id === sid)
      if (i >= 0) {
        wk.sets.splice(i, 1)
        return new HttpResponse(null, { status: 204 })
      }
    }
    return HttpResponse.json({ error: { code: 'not_found', message: 'нет' } }, { status: 404 })
  }),

  http.get('/api/v1/exercises/:id/history', () =>
    HttpResponse.json([
      { date: '2026-05-05', sets: [{ role: 'working', weight_kg: 60, reps: 10 }] },
    ]),
  ),

  http.get('/api/v1/program-days/:id', ({ params }) =>
    HttpResponse.json({
      id: Number(params.id),
      program_id: 1,
      program_name: 'Фул бади',
      position: 0,
      name: 'День A',
      exercises: [
        { id: 100, exercise_id: 10, position: 0, sets: 3, rep_min: 6, rep_max: 10, weight_max_kg: 90 },
      ],
    }),
  ),
]
