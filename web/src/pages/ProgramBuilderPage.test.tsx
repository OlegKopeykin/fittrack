import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import { server } from '../test/setup'
import ProgramBuilderPage from './ProgramBuilderPage'
import { mockState } from '../test/handlers'

function authed() {
  mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
}

function render(path = '/programs/new') {
  renderApp(
    <Routes>
      <Route path="/programs/new" element={<ProgramBuilderPage />} />
      <Route path="/programs/:id/edit" element={<ProgramBuilderPage />} />
      <Route path="/program/:id" element={<div>Экран программы 700</div>} />
    </Routes>,
    path,
  )
}

describe('ProgramBuilderPage', () => {
  it('создаёт программу из упражнений каталога', async () => {
    authed()
    const user = userEvent.setup()
    render()

    await user.type(screen.getByLabelText('Название программы'), 'Мой сплит')
    await user.click(screen.getByRole('button', { name: '+ Добавить упражнение' }))
    await user.type(screen.getByLabelText('Поиск упражнения'), 'Присед')
    await user.click(await screen.findByRole('button', { name: 'Присед в Смите' }))

    // упражнение добавлено чипом, пикер закрылся
    expect(screen.getByText('Присед в Смите')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Сохранить программу' }))
    expect(await screen.findByText('Экран программы 700')).toBeInTheDocument()
  })

  it('кнопка сохранения выключена без названия и упражнений', async () => {
    authed()
    render()
    expect(screen.getByRole('button', { name: 'Сохранить программу' })).toBeDisabled()
  })

  it('добавляет второй день для сплита', async () => {
    authed()
    const user = userEvent.setup()
    render()
    await user.click(screen.getByRole('button', { name: '+ Добавить день' }))
    expect(screen.getByLabelText('Название дня 2')).toBeInTheDocument()
  })

  it('редактирует существующую программу с предзаполнением', async () => {
    authed()
    const user = userEvent.setup()
    render('/programs/1/edit')

    // подтянулись имя и упражнение
    expect(await screen.findByDisplayValue('Фул бади')).toBeInTheDocument()
    expect(await screen.findByText('Присед в Смите')).toBeInTheDocument()

    const nameInput = screen.getByLabelText('Название программы')
    await user.clear(nameInput)
    await user.type(nameInput, 'Фул бади v2')
    await user.click(screen.getByRole('button', { name: 'Сохранить программу' }))
    expect(await screen.findByText('Экран программы 700')).toBeInTheDocument()
  })

  type CapturedProgram = {
    name: string
    description?: string
    days: {
      name: string
      notes?: string
      exercises: {
        exercise_id: number
        sets?: number
        rep_min?: number
        rep_max?: number
        weight_min_kg?: number
        weight_max_kg?: number
        rest_sec?: number
        tempo?: string
        notes?: string
      }[]
    }[]
  }

  it('сохраняет описание, заметки дня и предписания без потерь при правке', async () => {
    authed()
    const user = userEvent.setup()

    let captured: CapturedProgram | null = null
    server.use(
      http.put('/api/v1/programs/:id', async ({ request }) => {
        captured = (await request.json()) as CapturedProgram
        return HttpResponse.json({ id: 1, name: 'Фул бади', days: [] })
      }),
    )

    render('/programs/1/edit')
    // дожидаемся предзаполнения (имя и упражнение подтянулись из GET)
    expect(await screen.findByDisplayValue('Фул бади')).toBeInTheDocument()
    await screen.findByText('Присед в Смите')

    // сохраняем без изменений — весь набор полей должен уйти как есть
    await user.click(screen.getByRole('button', { name: 'Сохранить программу' }))
    expect(await screen.findByText('Экран программы 700')).toBeInTheDocument()

    expect(captured).toMatchObject({
      name: 'Фул бади',
      description: 'A/B',
      days: [
        {
          name: 'День A',
          notes: 'разминка 5 минут',
          exercises: [
            {
              exercise_id: 10,
              sets: 3,
              rep_min: 6,
              rep_max: 10,
              weight_min_kg: 70,
              weight_max_kg: 90,
              rest_sec: 90,
              tempo: '3-0-1',
              notes: 'делать медленно',
            },
          ],
        },
      ],
    })
  })

  it('позволяет править описание программы, заметку дня, подходы/повторы и заметку упражнения', async () => {
    authed()
    const user = userEvent.setup()

    let captured: CapturedProgram | null = null
    server.use(
      http.put('/api/v1/programs/:id', async ({ request }) => {
        captured = (await request.json()) as CapturedProgram
        return HttpResponse.json({ id: 1, name: 'Фул бади', days: [] })
      }),
    )

    render('/programs/1/edit')
    await screen.findByText('Присед в Смите')

    const description = screen.getByLabelText('Описание программы')
    await user.clear(description)
    await user.type(description, 'новое описание')

    const dayNote = screen.getByLabelText('Заметка дня 1')
    await user.clear(dayNote)
    await user.type(dayNote, 'новая заметка дня')

    const sets = screen.getByLabelText('Подходы: Присед в Смите')
    await user.clear(sets)
    await user.type(sets, '5')

    const repMin = screen.getByLabelText('Повторения от: Присед в Смите')
    await user.clear(repMin)
    await user.type(repMin, '4')

    const repMax = screen.getByLabelText('Повторения до: Присед в Смите')
    await user.clear(repMax)
    await user.type(repMax, '8')

    const note = screen.getByLabelText('Заметка: Присед в Смите')
    await user.clear(note)
    await user.type(note, 'делать быстрее')

    await user.click(screen.getByRole('button', { name: 'Сохранить программу' }))
    expect(await screen.findByText('Экран программы 700')).toBeInTheDocument()

    expect(captured).toMatchObject({
      description: 'новое описание',
      days: [
        {
          notes: 'новая заметка дня',
          exercises: [{ sets: 5, rep_min: 4, rep_max: 8, notes: 'делать быстрее' }],
        },
      ],
    })
  })
})
