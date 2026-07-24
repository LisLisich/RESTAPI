import { expect, test } from '@playwright/test'

test('login opens owner task workspace', async ({ page }) => {
  await page.route('**/api/v2/auth/login', (route) =>
    route.fulfill({ json: { user_id: 42, email: 'user@example.com' } }),
  )
  await page.route('**/api/v2/tasks?*', (route) =>
    route.fulfill({
      json: [{
        id: 1,
        version: 1,
        title: 'Подготовить рассказ о FinTask',
        description: 'Три минуты без ручного исправления данных',
        completed: false,
        created_at: '2026-07-24T12:00:00Z',
        completed_at: null,
      }],
    }),
  )

  await page.goto('/')
  await page.getByLabel('Email').fill('user@example.com')
  await page.getByLabel('Пароль').fill('very-strong-password')
  await page.locator('form').getByRole('button', { name: 'Войти', exact: true }).click()

  await expect(page.getByRole('heading', { name: 'Задачи' })).toBeVisible()
  await expect(page.getByText('Подготовить рассказ о FinTask')).toBeVisible()
  await page.screenshot({ path: `test-results/workspace-${test.info().project.name}.png`, fullPage: true })
})
