import { screen } from '@testing-library/react'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import TodayPage from './TodayPage'
import WorkoutHistoryPage from './WorkoutHistoryPage'
import WorkoutDetailPage from './WorkoutDetailPage'
import ProgramDetailPage from './ProgramDetailPage'
import ProgressPage from './ProgressPage'
import { mockState } from '../test/handlers'

function authed() {
  mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
}

describe('TodayPage', () => {
  it('показывает программы и кнопку истории, без приветствия', async () => {
    authed()
    renderApp(<TodayPage />, '/')
    expect(await screen.findByText('Фул бади')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'История' })).toHaveAttribute('href', '/workouts')
    expect(screen.queryByText(/привет/i)).not.toBeInTheDocument()
  })
})

describe('WorkoutHistoryPage', () => {
  it('показывает список тренировок с датой и названием', async () => {
    authed()
    renderApp(<WorkoutHistoryPage />, '/workouts')
    expect(await screen.findByText(/10 мая 2026/)).toBeInTheDocument()
    expect(screen.getByText(/Full-A/)).toBeInTheDocument()
  })
})

describe('WorkoutDetailPage', () => {
  it('группирует подходы по упражнению, без веса тела', async () => {
    authed()
    renderApp(
      <Routes>
        <Route path="/workout/:id" element={<WorkoutDetailPage />} />
      </Routes>,
      '/workout/500',
    )
    expect(await screen.findByText('Присед в Смите')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Full-A' })).toBeInTheDocument()
    expect(screen.getByText('60×12')).toBeInTheDocument()
    expect(screen.queryByText(/вес тела/i)).not.toBeInTheDocument()
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

describe('ProgressPage', () => {
  it('показывает вес тела графиком с текущим значением', async () => {
    authed()
    renderApp(<ProgressPage />, '/progress')
    expect(await screen.findByText('Вес тела')).toBeInTheDocument()
    expect(await screen.findByText('86.5')).toBeInTheDocument()
    expect(screen.getByRole('img', { name: /вес/i })).toBeInTheDocument()
  })
})
