import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { renderApp } from '../test/render'
import InviteFriend from './InviteFriend'

describe('InviteFriend', () => {
  it('создаёт ссылку-приглашение с кодом', async () => {
    const user = userEvent.setup()
    renderApp(<InviteFriend />)

    await user.click(screen.getByRole('button', { name: 'Создать ссылку-приглашение' }))
    const input = (await screen.findByLabelText('Ссылка приглашения')) as HTMLInputElement
    expect(input.value).toContain('/register?code=NEWINVITE01')
  })
})
