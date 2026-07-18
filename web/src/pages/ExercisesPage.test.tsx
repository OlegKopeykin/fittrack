import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import ExercisesPage from './ExercisesPage'

function setup() {
  return renderApp(
    <Routes>
      <Route path="/" element={<ExercisesPage />} />
    </Routes>,
    '/',
  )
}

describe('ExercisesPage', () => {
  it('показывает каталог упражнений', async () => {
    setup()
    expect(await screen.findByText('Присед в Смите')).toBeInTheDocument()
    expect(screen.getByText('Жим гантелей лёжа')).toBeInTheDocument()
    expect(screen.getByText('Тяга верхнего блока')).toBeInTheDocument()
  })

  it('фильтрует по группе мышц через чип', async () => {
    const user = userEvent.setup()
    setup()
    await screen.findByText('Присед в Смите')
    await user.click(screen.getByRole('button', { name: 'Грудь' }))
    await waitFor(() => expect(screen.queryByText('Присед в Смите')).not.toBeInTheDocument())
    expect(screen.getByText('Жим гантелей лёжа')).toBeInTheDocument()
  })

  it('ищет по алиасу', async () => {
    const user = userEvent.setup()
    setup()
    await screen.findByText('Присед в Смите')
    await user.type(screen.getByPlaceholderText(/поиск/i), 'подтягивания')
    await waitFor(() => expect(screen.getByText('Тяга верхнего блока')).toBeInTheDocument())
    expect(screen.queryByText('Присед в Смите')).not.toBeInTheDocument()
  })

  it('фильтрует по типу: свободные веса', async () => {
    const user = userEvent.setup()
    setup()
    await screen.findByText('Присед в Смите')
    await user.click(screen.getByRole('button', { name: 'Свободные веса' }))
    // остаётся только гантельное; тренажёрные (Присед в Смите=machine, Тяга=cable) уходят
    await waitFor(() => expect(screen.queryByText('Присед в Смите')).not.toBeInTheDocument())
    expect(screen.getByText('Жим гантелей лёжа')).toBeInTheDocument()
    expect(screen.queryByText('Тяга верхнего блока')).not.toBeInTheDocument()
  })

  it('фильтрует по типу: тренажёр', async () => {
    const user = userEvent.setup()
    setup()
    await screen.findByText('Присед в Смите')
    await user.click(screen.getByRole('button', { name: 'Тренажёр' }))
    await waitFor(() => expect(screen.queryByText('Жим гантелей лёжа')).not.toBeInTheDocument())
    expect(screen.getByText('Присед в Смите')).toBeInTheDocument()
    expect(screen.getByText('Тяга верхнего блока')).toBeInTheDocument()
  })

  it('показывает подпись группы у упражнения', async () => {
    setup()
    await screen.findByText('Присед в Смите')
    expect(screen.getAllByText(/Квадрицепс/).length).toBeGreaterThan(0)
  })
})
