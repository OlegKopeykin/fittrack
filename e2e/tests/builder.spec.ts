import { test, expect } from '@playwright/test'

async function login(page) {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка' })).toBeVisible()
}

test('конструктор: создать программу из упражнений каталога', async ({ page }) => {
  await login(page)
  await page.getByRole('link', { name: '+ Создать' }).click()
  await expect(page.getByRole('heading', { name: 'Новая программа' })).toBeVisible()

  await page.getByLabel('Название программы').fill('Мой сплит e2e')
  await page.getByRole('button', { name: '+ Добавить упражнение' }).click()
  await page.getByLabel('Поиск упражнения').fill('Присед')
  await page.getByRole('button', { name: 'Присед в Смите' }).click()
  await expect(page.getByText('Присед в Смите')).toBeVisible()

  await page.getByRole('button', { name: 'Сохранить программу' }).click()

  // перешли на страницу созданной программы
  await expect(page.getByRole('heading', { name: 'Мой сплит e2e' })).toBeVisible()
  await expect(page.getByText('Присед в Смите')).toBeVisible()
})

test('конструктор: правка существующей программы', async ({ page }) => {
  await login(page)
  // создаём
  await page.getByRole('link', { name: '+ Создать' }).click()
  await page.getByLabel('Название программы').fill('Правлю e2e')
  await page.getByRole('button', { name: '+ Добавить упражнение' }).click()
  await page.getByLabel('Поиск упражнения').fill('Присед')
  await page.getByRole('button', { name: 'Присед в Смите' }).click()
  await page.getByRole('button', { name: 'Сохранить программу' }).click()
  await expect(page.getByRole('heading', { name: 'Правлю e2e' })).toBeVisible()

  // редактируем: меняем имя и добавляем упражнение
  await page.getByRole('link', { name: 'Редактировать' }).click()
  await expect(page.getByRole('heading', { name: 'Редактировать программу' })).toBeVisible()
  await expect(page.getByLabel('Название программы')).toHaveValue('Правлю e2e')
  await page.getByLabel('Название программы').fill('Правлю e2e — обновлено')
  await page.getByRole('button', { name: '+ Добавить упражнение' }).click()
  await page.getByLabel('Поиск упражнения').fill('Жим гантелей лёжа')
  await page.getByRole('button', { name: 'Жим гантелей лёжа' }).click()
  await page.getByRole('button', { name: 'Сохранить программу' }).click()

  await expect(page.getByRole('heading', { name: 'Правлю e2e — обновлено' })).toBeVisible()
  await expect(page.getByText('Присед в Смите')).toBeVisible()
  await expect(page.getByText('Жим гантелей лёжа')).toBeVisible()
})

test('конструктор: второй день делает сплит', async ({ page }) => {
  await login(page)
  await page.getByRole('link', { name: '+ Создать' }).click()
  await page.getByLabel('Название программы').fill('Двухдневный e2e')

  // день 1
  await page.getByRole('button', { name: '+ Добавить упражнение' }).first().click()
  await page.getByLabel('Поиск упражнения').fill('Присед')
  await page.getByRole('button', { name: 'Присед в Смите' }).click()

  // добавляем второй день
  await page.getByRole('button', { name: '+ Добавить день' }).click()
  await expect(page.getByLabel('Название дня 2')).toBeVisible()
  await page.getByRole('button', { name: '+ Добавить упражнение' }).nth(1).click()
  await page.getByLabel('Поиск упражнения').fill('Жим')
  await page.getByRole('button', { name: 'Жим гантелей лёжа' }).click()

  await page.getByRole('button', { name: 'Сохранить программу' }).click()
  await expect(page.getByRole('heading', { name: 'Двухдневный e2e' })).toBeVisible()
  await expect(page.getByText('День 1')).toBeVisible()
  await expect(page.getByText('День 2')).toBeVisible()
})
