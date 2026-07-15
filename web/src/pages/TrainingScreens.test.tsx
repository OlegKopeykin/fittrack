import { screen } from '@testing-library/react'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import TodayPage from './TodayPage'
import WorkoutDetailPage from './WorkoutDetailPage'
import ProgramDetailPage from './ProgramDetailPage'
import { mockState } from '../test/handlers'

function authed() {
  mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
}

describe('TodayPage', () => {
  it('показывает программы и историю тренировок', async () => {
    authed()
    renderApp(<TodayPage />, '/')
    expect(await screen.findByText('Фул бади')).toBeInTheDocument()
    expect(screen.getByText('5-дневный сплит')).toBeInTheDocument()
    // история — дата тренировки
    expect(await screen.findByText(/10 мая 2026/)).toBeInTheDocument()
  })
})

describe('WorkoutDetailPage', () => {
  it('группирует подходы по упражнению с именами и весами', async () => {
    authed()
    renderApp(
      <Routes>
        <Route path="/workout/:id" element={<WorkoutDetailPage />} />
      </Routes>,
      '/workout/500',
    )
    expect(await screen.findByText('Присед в Смите')).toBeInTheDocument()
    expect(screen.getByText('40×12')).toBeInTheDocument()
    expect(screen.getByText('60×12')).toBeInTheDocument()
  })
})

describe('ProgramDetailPage', () => {
  it('показывает дни и предписания с именами упражнений', async () => {
    authed()
    renderApp(
      <Routes>
        <Route path="/program/:id" element={<ProgramDetailPage />} />
      </Routes>,
      '/program/1',
    )
    expect(await screen.findByText('День A')).toBeInTheDocument()
    expect(screen.getByText('Присед в Смите')).toBeInTheDocument()
    expect(screen.getByText('3 × 6–10')).toBeInTheDocument()
  })
})
