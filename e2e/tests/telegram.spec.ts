import { test, expect } from '@playwright/test'

async function login(page) {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка' })).toBeVisible()
}

// Настройки Telegram — синглтон на пользователя, а оба проекта делят одну
// e2e-БД. Гоняем поток на одном проекте, чтобы не было гонки; мобильную
// вёрстку секции покрывает Vitest.
test('профиль: подключение Telegram, частота и тест-отправка', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chrome', 'синглтон-настройка — один проект')
  await login(page)
  await page.getByRole('link', { name: 'Профиль' }).click()
  await expect(page.getByRole('heading', { name: 'Экспорт в Telegram' })).toBeVisible()

  // токен (в e2e-сборке Bot API фейковый → принимается)
  await page.getByLabel('Токен бота').fill('123456:FAKEE2E')
  await page.getByRole('button', { name: 'Подключить' }).click()

  // связать чат (fake ResolveChatID)
  await page.getByRole('button', { name: 'Проверить связь' }).click()
  await expect(page.getByText(/Подключено/)).toBeVisible()

  // включить, сменить частоту, тест-отправка
  await page.getByRole('switch', { name: 'Автовыгрузка' }).click()
  await page.getByRole('button', { name: 'Раз в неделю' }).click()
  await page.getByRole('button', { name: /Отправить сейчас/ }).click()
  await expect(page.getByText('Отправлено ✓')).toBeVisible()
})
