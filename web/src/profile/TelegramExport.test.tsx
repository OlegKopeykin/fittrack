import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { renderApp } from '../test/render'
import TelegramExport from './TelegramExport'

describe('TelegramExport', () => {
  it('поток: токен → связь чата → включение → частота → тест-отправка', async () => {
    const user = userEvent.setup()
    renderApp(<TelegramExport />)

    // шаг 1: инструкция и поле токена
    expect(await screen.findByLabelText('Токен бота')).toBeInTheDocument()
    expect(screen.getByText(/BotFather/)).toBeInTheDocument()
    await user.type(screen.getByLabelText('Токен бота'), '111:AAA')
    await user.click(screen.getByRole('button', { name: 'Подключить' }))

    // шаг 2: связать чат
    await user.click(await screen.findByRole('button', { name: 'Проверить связь' }))

    // шаг 3: подключено
    expect(await screen.findByText(/Подключено/)).toBeInTheDocument()

    // включаем автовыгрузку
    const sw = screen.getByRole('switch', { name: 'Автовыгрузка' })
    expect(sw).toHaveAttribute('aria-checked', 'false')
    await user.click(sw)
    expect(sw).toHaveAttribute('aria-checked', 'true')

    // меняем частоту
    await user.click(screen.getByRole('button', { name: 'Раз в неделю' }))

    // тест-отправка
    await user.click(screen.getByRole('button', { name: /Отправить сейчас/ }))
    expect(await screen.findByText('Отправлено ✓')).toBeInTheDocument()
  })

  it('отключение возвращает форму токена', async () => {
    const user = userEvent.setup()
    renderApp(<TelegramExport />)
    await user.type(await screen.findByLabelText('Токен бота'), '111:AAA')
    await user.click(screen.getByRole('button', { name: 'Подключить' }))
    await user.click(await screen.findByRole('button', { name: 'Сменить токен' }))
    expect(await screen.findByLabelText('Токен бота')).toBeInTheDocument()
  })
})
