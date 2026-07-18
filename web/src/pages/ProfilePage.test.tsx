import { screen } from '@testing-library/react'
import { renderApp } from '../test/render'
import ProfilePage from './ProfilePage'
import { mockState } from '../test/handlers'

describe('ProfilePage', () => {
  it('владелец видит «Пригласить» и «Выйти» в шапке', async () => {
    mockState.me = { id: 1, username: 'oleg', display_name: '', role: 'owner' }
    renderApp(<ProfilePage />, '/profile')
    expect(await screen.findByRole('heading', { name: 'Пригласить' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Выйти' })).toBeInTheDocument()
  })

  it('обычный пользователь не видит «Пригласить»', async () => {
    mockState.me = { id: 2, username: 'petya', display_name: '', role: 'user' }
    renderApp(<ProfilePage />, '/profile')
    // дождёмся отрисовки карточки пользователя
    expect(await screen.findByText('petya')).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Пригласить' })).not.toBeInTheDocument()
  })
})
