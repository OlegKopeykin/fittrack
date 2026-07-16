import { test, expect } from '@playwright/test'

async function login(page) {
  await page.goto('/login')
  await page.getByLabel('Логин').fill('e2euser')
  await page.getByLabel('Пароль').fill('e2e-password-123')
  await page.getByRole('button', { name: 'Войти' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка' })).toBeVisible()
}

test('каталог: добавить своё упражнение и отредактировать', async ({ page }) => {
  await login(page)
  await page.getByRole('link', { name: 'Упражнения' }).click()
  await page.getByRole('link', { name: '+ Добавить' }).click()
  await expect(page.getByRole('heading', { name: 'Новое упражнение' })).toBeVisible()

  const uniq = 'Тяга e2e ' + Date.now().toString().slice(-6)
  await page.getByLabel('Название упражнения').fill(uniq)
  await page.getByLabel('Группа мышц').selectOption({ index: 1 })
  await page.getByLabel('Снаряд').selectOption('cable')
  await page.getByRole('button', { name: 'Сохранить упражнение' }).click()

  // вернулись в каталог, упражнение появилось
  await expect(page.getByRole('heading', { name: 'Упражнения' })).toBeVisible()
  await expect(page.getByText(uniq)).toBeVisible()

  // сужаем поиском, чтобы «Изменить» не перекрывался липкой панелью
  await page.getByPlaceholder('Поиск по названию и синонимам').fill(uniq)
  await page.getByRole('listitem').filter({ hasText: uniq }).getByRole('link', { name: 'Изменить' }).click()
  await expect(page.getByRole('heading', { name: 'Изменить упражнение' })).toBeVisible()
  await expect(page.getByLabel('Название упражнения')).toHaveValue(uniq)
  await page.getByLabel('Название упражнения').fill(uniq + ' ✓')
  await page.getByRole('button', { name: 'Сохранить упражнение' }).click()
  await expect(page.getByText(uniq + ' ✓')).toBeVisible()
})
