import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import WorkoutDetailPage from '../pages/WorkoutDetailPage'
import { mockState, mockWorkouts } from '../test/handlers'

function authed() {
  mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
}

function seedActive(id: number, extra: Record<string, unknown> = {}) {
  mockWorkouts.set(id, {
    id,
    date: '2026-07-16',
    title: 'Фул бади A',
    program_day_id: 11,
    started_at: '2026-07-16T09:00:00Z',
    feeling: '',
    notes: '',
    sets: [],
    ...extra,
  })
}

function routes(path: string) {
  renderApp(
    <Routes>
      <Route path="/workout/:id" element={<WorkoutDetailPage />} />
      <Route path="/workouts" element={<div>История экран</div>} />
    </Routes>,
    path,
  )
}

describe('WorkoutLogger', () => {
  it('показывает упражнение дня и «Прошлый», логирует подход', async () => {
    authed()
    seedActive(901)
    const user = userEvent.setup()
    routes('/workout/901')

    expect(await screen.findByText('Присед в Смите')).toBeInTheDocument()
    // «Прошлый» из истории — кнопка автоподстановки
    expect(await screen.findByRole('button', { name: '60×10' })).toBeInTheDocument()

    await user.type(screen.getByLabelText('вес'), '80')
    await user.type(screen.getByLabelText('повторы'), '8')
    await user.click(screen.getByRole('button', { name: 'Записать подход' }))

    // подход стал залогированным: появилась кнопка снятия отметки
    expect(await screen.findByRole('button', { name: 'Снять отметку' })).toBeInTheDocument()
  })

  it('автоподстановка из «Прошлый» заполняет поля', async () => {
    authed()
    seedActive(904)
    const user = userEvent.setup()
    routes('/workout/904')

    await screen.findByText('Присед в Смите')
    await user.click(await screen.findByRole('button', { name: '60×10' }))
    expect(screen.getByLabelText('вес')).toHaveValue('60')
    expect(screen.getByLabelText('повторы')).toHaveValue('10')
  })

  it('степперы меняют вес на ±2.5', async () => {
    authed()
    seedActive(905)
    const user = userEvent.setup()
    routes('/workout/905')

    await screen.findByText('Присед в Смите')
    const weight = screen.getByLabelText('вес')
    await user.type(weight, '80')
    await user.click(screen.getByRole('button', { name: 'плюс' }))
    expect(weight).toHaveValue('82.5')
  })

  it('завершение открывает лист и уводит в историю', async () => {
    authed()
    seedActive(902)
    const user = userEvent.setup()
    routes('/workout/902')

    await screen.findByText('Присед в Смите')
    await user.click(screen.getByRole('button', { name: 'Завершить' }))
    expect(await screen.findByText('Завершить тренировку')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Сохранить и завершить' }))
    expect(await screen.findByText('История экран')).toBeInTheDocument()
  })

  it('добавление упражнения через поиск', async () => {
    authed()
    seedActive(903, { program_day_id: undefined })
    const user = userEvent.setup()
    routes('/workout/903')

    expect(await screen.findByText(/Пустая тренировка/)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: '+ Добавить упражнение' }))
    await user.type(screen.getByLabelText('Поиск упражнения'), 'Присед')
    await user.click(await screen.findByRole('button', { name: 'Присед в Смите' }))
    expect(await screen.findByRole('button', { name: 'Записать подход' })).toBeInTheDocument()
  })
})
