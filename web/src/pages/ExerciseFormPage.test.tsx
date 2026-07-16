import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import ExerciseFormPage from './ExerciseFormPage'
import { mockState } from '../test/handlers'

function authed() {
  mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
}

function render(path: string) {
  renderApp(
    <Routes>
      <Route path="/exercises/new" element={<ExerciseFormPage />} />
      <Route path="/exercises/:id/edit" element={<ExerciseFormPage />} />
      <Route path="/exercises" element={<div>Каталог</div>} />
    </Routes>,
    path,
  )
}

describe('ExerciseFormPage', () => {
  it('создаёт упражнение', async () => {
    authed()
    const user = userEvent.setup()
    render('/exercises/new')

    expect(screen.getByRole('heading', { name: 'Новое упражнение' })).toBeInTheDocument()
    await user.type(screen.getByLabelText('Название упражнения'), 'Мой жим')
    // группа мышц обязательна
    expect(screen.getByRole('button', { name: 'Сохранить упражнение' })).toBeDisabled()
    await user.selectOptions(await screen.findByLabelText('Группа мышц'), 'chest')
    await user.click(screen.getByRole('button', { name: 'Сохранить упражнение' }))

    expect(await screen.findByText('Каталог')).toBeInTheDocument()
  })

  it('правка предзаполняет поля', async () => {
    authed()
    const user = userEvent.setup()
    render('/exercises/5/edit')

    expect(await screen.findByDisplayValue('Мой присед')).toBeInTheDocument()
    expect(await screen.findByDisplayValue('спина прямая')).toBeInTheDocument()
    // группа подтянулась (quads = id 2)
    expect(screen.getByLabelText('Группа мышц')).toHaveValue('quads')

    await user.click(screen.getByRole('button', { name: 'Сохранить упражнение' }))
    expect(await screen.findByText('Каталог')).toBeInTheDocument()
  })
})
