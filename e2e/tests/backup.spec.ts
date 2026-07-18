import { test, expect } from '@playwright/test'

async function login(page) {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка' })).toBeVisible()
}

test('профиль: восстановление лога из файла бэкапа', async ({ page }, testInfo) => {
  await login(page)
  await page.getByRole('link', { name: 'Профиль' }).click()
  await expect(page.getByRole('heading', { name: 'Резервная копия' })).toBeVisible()
  await expect(page.getByRole('link', { name: /Скачать бэкап/ })).toBeVisible()

  // уникальный на проект бэкап (общая e2e-БД: избегаем дедупа между проектами)
  const title = 'Импорт e2e ' + testInfo.project.name
  const payload = JSON.stringify({
    user: 'e2euser',
    workouts: [
      {
        date: '2026-03-03',
        title,
        finished_at: '2026-03-03T10:00:00Z',
        sets: [{ exercise: 'Присед в Смите', role: 'working', weight_kg: 70, reps: 5 }],
      },
    ],
  })
  await page.getByLabel('Файл бэкапа').setInputFiles({
    name: 'backup.json',
    mimeType: 'application/json',
    buffer: Buffer.from(payload),
  })
  await expect(page.getByText(/Восстановлено тренировок: 1/)).toBeVisible()

  // тренировка появилась в истории
  await page.getByRole('link', { name: 'Тренировка' }).click()
  await page.getByRole('link', { name: 'История' }).click()
  await expect(page.getByText(title)).toBeVisible()
})
