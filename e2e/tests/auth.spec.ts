import { test, expect } from '@playwright/test'

test('незалогиненного редиректит на форму входа', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Вход' })).toBeVisible()
  await expect(page.getByLabel('Логин')).toBeVisible()
})

test('вход существующим пользователем ведёт в приложение', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByText(/привет, e2euser/i)).toBeVisible()

  // выход — на вкладке «Профиль»
  await page.getByRole('link', { name: 'Профиль' }).click()
  await page.getByRole('button', { name: 'Выйти' }).click()
  await expect(page.getByRole('heading', { name: 'Вход' })).toBeVisible()
})

test('каталог упражнений: поиск и переход', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByText(/привет, e2euser/i)).toBeVisible()

  await page.getByRole('link', { name: 'Упражнения' }).click()
  await expect(page.getByRole('heading', { name: 'Упражнения' })).toBeVisible()
  await expect(page.getByText('Присед в Смите')).toBeVisible()

  // поиск по синониму «РДЛ» → «Румынская тяга с гантелями»
  await page.getByPlaceholder(/поиск/i).fill('РДЛ')
  await expect(page.getByText('Румынская тяга с гантелями')).toBeVisible()
  await expect(page.getByText('Присед в Смите')).toHaveCount(0)
})

test('неверный пароль показывает ошибку', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('не-тот-пароль')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByText(/неверный логин или пароль/i)).toBeVisible()
})

test('форма регистрации подставляет код инвайта из ссылки', async ({ page }) => {
  await page.goto('/register?code=SOMEINVITECODE')
  await expect(page.getByLabel(/код инвайта/i)).toHaveValue('SOMEINVITECODE')
})
