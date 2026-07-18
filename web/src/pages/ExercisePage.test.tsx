import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import ExercisePage from './ExercisePage'
import { mockState } from '../test/handlers'

function authed() {
  mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
}

function render(path = '/exercises/10') {
  renderApp(
    <Routes>
      <Route path="/exercises/:id" element={<ExercisePage />} />
    </Routes>,
    path,
  )
}

describe('ExercisePage', () => {
  it('показывает упражнение, историю и позволяет оставить комментарий', async () => {
    authed()
    const user = userEvent.setup()
    render()

    expect(await screen.findByRole('heading', { name: 'Мой присед' })).toBeInTheDocument()
    // история из лога
    expect(await screen.findByText('60×10')).toBeInTheDocument()

    // комментарий: пустой → печатаем → «Сохранить»
    const ta = screen.getByLabelText('Комментарий к упражнению')
    await user.type(ta, 'тянуть лопатками')
    await user.click(screen.getByRole('button', { name: 'Сохранить' }))
    // после сохранения кнопка пропадает (нет несохранённых изменений)
    await waitFor(() =>
      expect(screen.queryByRole('button', { name: 'Сохранить' })).not.toBeInTheDocument(),
    )
  })
})
