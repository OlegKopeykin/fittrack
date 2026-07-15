import { screen, waitFor } from '@testing-library/react'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import RequireAuth from './RequireAuth'
import { mockState } from '../test/handlers'

function setup(path = '/') {
  return renderApp(
    <Routes>
      <Route
        path="/"
        element={
          <RequireAuth>
            <div>защищённый контент</div>
          </RequireAuth>
        }
      />
      <Route path="/login" element={<div>страница входа</div>} />
    </Routes>,
    path,
  )
}

describe('RequireAuth', () => {
  it('редиректит неавторизованного на /login', async () => {
    setup()
    await waitFor(() => expect(screen.getByText('страница входа')).toBeInTheDocument())
    expect(screen.queryByText('защищённый контент')).not.toBeInTheDocument()
  })

  it('пропускает авторизованного к контенту', async () => {
    mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
    setup()
    await waitFor(() => expect(screen.getByText('защищённый контент')).toBeInTheDocument())
  })
})
