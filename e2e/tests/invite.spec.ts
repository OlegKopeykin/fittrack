import { test, expect } from '@playwright/test'

async function login(page) {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка' })).toBeVisible()
}

test('профиль: владелец создаёт ссылку-приглашение', async ({ page }) => {
  await login(page)
  await page.getByRole('link', { name: 'Профиль' }).click()
  await expect(page.getByRole('heading', { name: 'Пригласить' })).toBeVisible()

  await page.getByRole('button', { name: 'Создать ссылку-приглашение' }).click()
  const link = page.getByLabel('Ссылка приглашения')
  await expect(link).toBeVisible()
  await expect(link).toHaveValue(/\/register\?code=[A-Za-z0-9]+/)
})
