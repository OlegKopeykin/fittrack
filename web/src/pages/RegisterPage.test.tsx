import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import RegisterPage from './RegisterPage'

function setup(path = '/register') {
  return renderApp(
    <Routes>
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/" element={<div>дашборд</div>} />
      <Route path="/login" element={<div>вход</div>} />
    </Routes>,
    path,
  )
}

describe('RegisterPage', () => {
  it('подставляет код инвайта из query-параметра', () => {
    setup('/register?code=GOODINVITE01')
    expect(screen.getByLabelText(/инвайт|код/i)).toHaveValue('GOODINVITE01')
  })

  it('валидирует короткий пароль до отправки', async () => {
    const user = userEvent.setup()
    setup('/register?code=GOODINVITE01')
    await user.type(screen.getByLabelText(/логин/i), 'newuser')
    await user.type(screen.getByLabelText(/пароль/i), '123')
    await user.click(screen.getByRole('button', { name: /зарегистр/i }))
    expect(await screen.findByText(/минимум 8/i)).toBeInTheDocument()
    expect(screen.queryByText('дашборд')).not.toBeInTheDocument()
  })

  it('регистрирует по валидному инвайту и пускает на дашборд', async () => {
    const user = userEvent.setup()
    setup('/register?code=GOODINVITE01')
    await user.type(screen.getByLabelText(/логин/i), 'newuser')
    await user.type(screen.getByLabelText(/пароль/i), 'надёжный-пароль')
    await user.click(screen.getByRole('button', { name: /зарегистр/i }))
    await waitFor(() => expect(screen.getByText('дашборд')).toBeInTheDocument())
  })

  it('показывает ошибку неверного инвайта', async () => {
    const user = userEvent.setup()
    setup('/register?code=WRONGCODE999')
    await user.type(screen.getByLabelText(/логин/i), 'newuser')
    await user.type(screen.getByLabelText(/пароль/i), 'надёжный-пароль')
    await user.click(screen.getByRole('button', { name: /зарегистр/i }))
    expect(await screen.findByText(/инвайт.*(недействит|неверн)/i)).toBeInTheDocument()
  })
})
