import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { renderApp } from '../test/render'
import BackupRestore from './BackupRestore'

describe('BackupRestore', () => {
  it('даёт ссылку скачивания и восстанавливает из файла', async () => {
    const user = userEvent.setup()
    renderApp(<BackupRestore />)

    expect(screen.getByRole('link', { name: /Скачать бэкап/ })).toHaveAttribute(
      'href',
      '/api/v1/profile/export',
    )

    const file = new File(
      [JSON.stringify({ workouts: [{ date: '2026-05-10', sets: [] }, { date: '2026-05-12', sets: [] }] })],
      'backup.json',
      { type: 'application/json' },
    )
    await user.upload(screen.getByLabelText('Файл бэкапа'), file)
    expect(await screen.findByText(/Восстановлено тренировок: 2/)).toBeInTheDocument()
  })

  it('сообщает про не-JSON файл', async () => {
    const user = userEvent.setup()
    renderApp(<BackupRestore />)
    const bad = new File(['это не json'], 'x.json', { type: 'application/json' })
    await user.upload(screen.getByLabelText('Файл бэкапа'), bad)
    expect(await screen.findByText(/не JSON/)).toBeInTheDocument()
  })
})
