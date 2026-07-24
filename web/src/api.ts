import type { Notification, Payment, Task, Wallet } from './types'

function csrfToken(): string {
  const cookie = document.cookie
    .split('; ')
    .find((item) => item.startsWith('__Host-fintask_csrf='))
  return cookie ? decodeURIComponent(cookie.split('=').slice(1).join('=')) : ''
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body) headers.set('Content-Type', 'application/json')
  if (init.method && !['GET', 'HEAD'].includes(init.method)) {
    headers.set('X-CSRF-Token', csrfToken())
  }
  const response = await fetch(`/api/v2${path}`, {
    ...init,
    headers,
    credentials: 'include',
  })
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Ошибка запроса' }))
    throw new Error(error.message || `HTTP ${response.status}`)
  }
  return response.json() as Promise<T>
}

export const api = {
  login: (email: string, password: string) =>
    request<{ user_id: number; email: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),
  register: (fullName: string, email: string, password: string) =>
    request('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ full_name: fullName, email, password }),
    }),
  tasks: () => request<Task[]>('/tasks?limit=100&offset=0'),
  createTask: (title: string, description: string) =>
    request<Task>('/tasks', {
      method: 'POST',
      body: JSON.stringify({ title, description: description || null }),
    }),
  toggleTask: (task: Task) =>
    request<Task>(`/tasks/${task.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ version: task.version, completed: !task.completed }),
    }),
  wallet: () => request<Wallet>('/wallet'),
  notifications: () => request<Notification[]>('/notifications'),
  createPayment: (amountMinor: number) =>
    request<Payment>('/payments', {
      method: 'POST',
      headers: { 'Idempotency-Key': crypto.randomUUID() },
      body: JSON.stringify({ amount_minor: amountMinor }),
    }),
}
