import { test, expect } from '@playwright/test'

async function login(page) {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка' })).toBeVisible()
}

test('логирование: старт дня программы, запись подхода, финиш', async ({ page }) => {
  await login(page)
  await page.getByText('Демо-программа').click()
  await expect(page.getByRole('heading', { name: 'Демо-программа' })).toBeVisible()
  await page.getByRole('button', { name: 'Начать' }).click()

  // логгер: упражнение дня и запись подхода
  await expect(page.getByText('Присед в Смите')).toBeVisible()
  await page.getByLabel('вес').fill('80')
  await page.getByLabel('повторы').fill('8')
  await page.getByRole('button', { name: 'Записать подход' }).click()
  await expect(page.getByRole('button', { name: 'Снять отметку' })).toBeVisible()

  // финиш
  await page.getByRole('button', { name: 'Завершить' }).click()
  await expect(page.getByText('Завершить тренировку')).toBeVisible()
  await page.getByRole('button', { name: 'Сохранить и завершить' }).click()

  // тренировка дня появилась в истории
  await expect(page.getByRole('heading', { name: 'История' })).toBeVisible()
  await expect(page.getByText('Демо-программа · День 1').first()).toBeVisible()
})

test('пустая тренировка: быстрый старт и добавление упражнения', async ({ page }) => {
  await login(page)
  await page.getByRole('button', { name: 'Начать пустую тренировку' }).click()
  await expect(page.getByText(/Пустая тренировка/)).toBeVisible()
  await page.getByRole('button', { name: '+ Добавить упражнение' }).click()
  await page.getByLabel('Поиск упражнения').fill('Присед')
  await page.getByRole('button', { name: 'Присед в Смите' }).click()
  await expect(page.getByRole('button', { name: 'Записать подход' })).toBeVisible()
})

test('завершённая тренировка read-only, кнопка «Редактировать» включает логгер', async ({ page }) => {
  await login(page)
  await page.getByRole('link', { name: 'История' }).click()
  await page.getByText(/10 мая 2026/).click()
  // read-only: подходы показаны текстом, есть кнопка редактирования
  await expect(page.getByText('60×12')).toBeVisible()
  await page.getByRole('button', { name: 'Редактировать' }).click()
  // логгер: появился ввод подхода
  await expect(page.getByRole('button', { name: 'Записать подход' })).toBeVisible()
})
