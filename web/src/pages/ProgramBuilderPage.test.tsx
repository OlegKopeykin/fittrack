import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import ProgramBuilderPage from './ProgramBuilderPage'
import { mockState } from '../test/handlers'

function authed() {
  mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
}

function render(path = '/programs/new') {
  renderApp(
    <Routes>
      <Route path="/programs/new" element={<ProgramBuilderPage />} />
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
})
