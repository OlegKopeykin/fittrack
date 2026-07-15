import { test, expect } from '@playwright/test'

test('healthz отвечает ok', async ({ request }) => {
  const res = await request.get('/healthz')
  expect(res.status()).toBe(200)
  expect(await res.json()).toEqual({ status: 'ok' })
})

test('SPA открывается и показывает FitTrack', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'FitTrack' })).toBeVisible()
})

test('клиентский роут отдаёт SPA (fallback)', async ({ page }) => {
  await page.goto('/workouts')
  await expect(page.getByRole('heading', { name: 'FitTrack' })).toBeVisible()
})

test('неизвестный API-роут отдаёт JSON-ошибку', async ({ request }) => {
  const res = await request.get('/api/v1/nope')
  expect(res.status()).toBe(404)
  const body = await res.json()
  expect(body.error.code).toBe('not_found')
})
