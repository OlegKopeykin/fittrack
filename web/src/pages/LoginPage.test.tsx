import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Routes, Route } from 'react-router-dom'
import { renderApp } from '../test/render'
import LoginPage from './LoginPage'

function setup(path = '/login') {
  return renderApp(
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div>дашборд</div>} />
      <Route path="/register" element={<div>регистрация</div>} />
    </Routes>,
    path,
  )
}

describe('LoginPage', () => {
  it('показывает форму входа', () => {
    setup()
    expect(screen.getByRole('heading', { name: /вход/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/логин/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/пароль/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /войти/i })).toBeInTheDocument()
  })

  it('при верных данных пускает на дашборд', async () => {
    const user = userEvent.setup()
    setup()
    await user.type(screen.getByLabelText(/логин/i), 'oleg')
    await user.type(screen.getByLabelText(/пароль/i), 'верный-пароль')
    await user.click(screen.getByRole('button', { name: /войти/i }))
    await waitFor(() => expect(screen.getByText('дашборд')).toBeInTheDocument())
  })

  it('при неверном пароле показывает ошибку и остаётся на странице', async () => {
    const user = userEvent.setup()
    setup()
    await user.type(screen.getByLabelText(/логин/i), 'oleg')
    await user.type(screen.getByLabelText(/пароль/i), 'не-тот')
    await user.click(screen.getByRole('button', { name: /войти/i }))
    expect(await screen.findByText(/неверный логин или пароль/i)).toBeInTheDocument()
    expect(screen.queryByText('дашборд')).not.toBeInTheDocument()
  })

  it('ведёт на регистрацию по ссылке', () => {
    setup()
    expect(screen.getByRole('link', { name: /регистрац/i })).toHaveAttribute('href', '/register')
  })
})
