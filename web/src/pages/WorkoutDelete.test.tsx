import { afterEach, describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import WorkoutDetailPage from './WorkoutDetailPage'
import { mockState, mockWorkouts } from '../test/handlers'

function authed() {
  mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
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

afterEach(() => vi.restoreAllMocks())

describe('удаление тренировки', () => {
  it('завершённая: «Удалить» с подтверждением уводит в историю', async () => {
    authed()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    const user = userEvent.setup()
    routes('/workout/500')

    await screen.findByText('Присед в Смите')
    await user.click(screen.getByRole('button', { name: 'Удалить' }))
    expect(await screen.findByText('История экран')).toBeInTheDocument()
    expect(mockWorkouts.has(500)).toBe(false)
  })

  it('отмена подтверждения не удаляет', async () => {
    authed()
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    const user = userEvent.setup()
    routes('/workout/500')

    await screen.findByText('Присед в Смите')
    await user.click(screen.getByRole('button', { name: 'Удалить' }))
    expect(mockWorkouts.has(500)).toBe(true)
    expect(screen.queryByText('История экран')).not.toBeInTheDocument()
  })
})
