import { test, expect } from '@playwright/test'

async function login(page) {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByText(/привет, e2euser/i)).toBeVisible()
}

test('вкладка «Тренировка» показывает программы и историю', async ({ page }) => {
  await login(page)
  await expect(page.getByText('Демо-программа')).toBeVisible()
  await expect(page.getByText(/10 мая 2026/)).toBeVisible()
})

test('детали программы: дни и предписания', async ({ page }) => {
  await login(page)
  await page.getByText('Демо-программа').click()
  await expect(page.getByRole('heading', { name: 'Демо-программа' })).toBeVisible()
  await expect(page.getByText('День 1')).toBeVisible()
  await expect(page.getByText('Присед в Смите')).toBeVisible()
  await expect(page.getByText('3 × 6–10')).toBeVisible()
})

test('детали тренировки: подходы по упражнению', async ({ page }) => {
  await login(page)
  await page.getByText(/10 мая 2026/).click()
  await expect(page.getByText('Присед в Смите')).toBeVisible()
  await expect(page.getByText('60×12')).toBeVisible()
})
