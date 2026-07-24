import { FormEvent, ReactNode, useCallback, useEffect, useState } from 'react'
import {
  Bell,
  Check,
  Circle,
  ListTodo,
  LogIn,
  Plus,
  RefreshCw,
  WalletCards,
} from 'lucide-react'
import { api } from './api'
import { formatRubles, parseRubles } from './money'
import type { Notification, Task, Wallet } from './types'

type View = 'tasks' | 'wallet' | 'notifications'

export function App() {
  const [authenticated, setAuthenticated] = useState(false)
  if (!authenticated) {
    return <AuthScreen onAuthenticated={() => setAuthenticated(true)} />
  }
  return <Workspace />
}

function AuthScreen({ onAuthenticated }: { onAuthenticated: () => void }) {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [status, setStatus] = useState('')
  const [busy, setBusy] = useState(false)

  async function submit(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setStatus('')
    try {
      if (mode === 'register') {
        await api.register(fullName, email, password)
        setStatus('Проверьте email и подтвердите аккаунт')
        setMode('login')
      } else {
        await api.login(email, password)
        onAuthenticated()
      }
    } catch (error) {
      setStatus(error instanceof Error ? error.message : 'Не удалось выполнить вход')
    } finally {
      setBusy(false)
    }
  }

  return (
    <main className="auth-shell">
      <section className="auth-panel" aria-labelledby="auth-title">
        <div className="brand"><span>F</span><strong>FinTask</strong></div>
        <h1 id="auth-title">{mode === 'login' ? 'Вход' : 'Регистрация'}</h1>
        <div className="segmented" role="tablist">
          <button className={mode === 'login' ? 'active' : ''} onClick={() => setMode('login')}>Войти</button>
          <button className={mode === 'register' ? 'active' : ''} onClick={() => setMode('register')}>Создать аккаунт</button>
        </div>
        <form onSubmit={submit}>
          {mode === 'register' && (
            <label>Имя<input value={fullName} onChange={(e) => setFullName(e.target.value)} minLength={3} required /></label>
          )}
          <label>Email<input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required /></label>
          <label>Пароль<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} minLength={12} required /></label>
          {status && <p className="status" role="status">{status}</p>}
          <button className="primary command" disabled={busy}>
            <LogIn size={18} />{busy ? 'Подождите…' : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
          </button>
        </form>
        <div className="divider"><span>или</span></div>
        <a className="google-button" href="/api/v2/auth/google/start">Продолжить с Google</a>
      </section>
    </main>
  )
}

function Workspace() {
  const [view, setView] = useState<View>('tasks')
  return (
    <div className="workspace">
      <aside>
        <div className="brand"><span>F</span><strong>FinTask</strong></div>
        <nav>
          <NavButton icon={<ListTodo />} active={view === 'tasks'} onClick={() => setView('tasks')}>Задачи</NavButton>
          <NavButton icon={<WalletCards />} active={view === 'wallet'} onClick={() => setView('wallet')}>Кошелек</NavButton>
          <NavButton icon={<Bell />} active={view === 'notifications'} onClick={() => setView('notifications')}>Уведомления</NavButton>
        </nav>
      </aside>
      <main className="content">
        {view === 'tasks' && <TasksView />}
        {view === 'wallet' && <WalletView />}
        {view === 'notifications' && <NotificationsView />}
      </main>
    </div>
  )
}

function NavButton(props: { icon: ReactNode; active: boolean; onClick: () => void; children: ReactNode }) {
  return <button className={props.active ? 'nav-active' : ''} onClick={props.onClick}>{props.icon}{props.children}</button>
}

function TasksView() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const load = useCallback(() => api.tasks().then(setTasks).catch((e) => setError(e.message)), [])
  useEffect(() => { void load() }, [load])

  async function create(event: FormEvent) {
    event.preventDefault()
    try {
      const task = await api.createTask(title, description)
      setTasks((items) => [task, ...items])
      setTitle('')
      setDescription('')
    } catch (e) { setError(e instanceof Error ? e.message : 'Ошибка') }
  }
  async function toggle(task: Task) {
    try {
      const updated = await api.toggleTask(task)
      setTasks((items) => items.map((item) => item.id === updated.id ? updated : item))
    } catch (e) { setError(e instanceof Error ? e.message : 'Ошибка') }
  }

  return (
    <>
      <Header title="Задачи" action={<button className="icon-button" title="Обновить" onClick={() => void load()}><RefreshCw /></button>} />
      <form className="task-create" onSubmit={create}>
        <input aria-label="Название задачи" placeholder="Новая задача" value={title} onChange={(e) => setTitle(e.target.value)} minLength={3} required />
        <input aria-label="Описание" placeholder="Описание" value={description} onChange={(e) => setDescription(e.target.value)} />
        <button className="primary icon-button" title="Добавить задачу"><Plus /></button>
      </form>
      {error && <p className="status error">{error}</p>}
      <div className="task-list">
        {tasks.map((task) => (
          <article className={task.completed ? 'task completed' : 'task'} key={task.id}>
            <button className="check-button" title={task.completed ? 'Вернуть в работу' : 'Завершить'} onClick={() => void toggle(task)}>
              {task.completed ? <Check /> : <Circle />}
            </button>
            <div><h2>{task.title}</h2>{task.description && <p>{task.description}</p>}</div>
            <time>{new Date(task.created_at).toLocaleDateString('ru-RU')}</time>
          </article>
        ))}
        {!tasks.length && !error && <div className="empty">Задач пока нет</div>}
      </div>
    </>
  )
}

function WalletView() {
  const [wallet, setWallet] = useState<Wallet | null>(null)
  const [amount, setAmount] = useState('100.00')
  const [error, setError] = useState('')
  const load = useCallback(() => api.wallet().then(setWallet).catch((e) => setError(e.message)), [])
  useEffect(() => { void load() }, [load])

  async function pay(event: FormEvent) {
    event.preventDefault()
    try {
      const payment = await api.createPayment(parseRubles(amount))
      window.location.assign(payment.confirmation_url)
    } catch (e) { setError(e instanceof Error ? e.message : 'Ошибка платежа') }
  }
  return (
    <>
      <Header title="Кошелек" action={<button className="icon-button" title="Обновить" onClick={() => void load()}><RefreshCw /></button>} />
      <section className="balance-band">
        <span>Доступно</span><strong>{wallet ? formatRubles(wallet.balance_minor) : '—'}</strong><small>RUB · версия {wallet?.version ?? '—'}</small>
      </section>
      <form className="payment-form" onSubmit={pay}>
        <label>Сумма пополнения<input inputMode="decimal" value={amount} onChange={(e) => setAmount(e.target.value)} required /></label>
        <button className="primary command"><WalletCards size={18} />Перейти к оплате</button>
      </form>
      {error && <p className="status error">{error}</p>}
    </>
  )
}

function NotificationsView() {
  const [items, setItems] = useState<Notification[]>([])
  const [error, setError] = useState('')
  const load = useCallback(() => api.notifications().then(setItems).catch((e) => setError(e.message)), [])
  useEffect(() => { void load() }, [load])
  return (
    <>
      <Header title="Уведомления" action={<button className="icon-button" title="Обновить" onClick={() => void load()}><RefreshCw /></button>} />
      {error && <p className="status error">{error}</p>}
      <div className="notification-list">
        {items.map((item) => <article key={item.id}><Bell /><div><h2>{item.title}</h2><p>{item.body}</p></div><time>{new Date(item.created_at).toLocaleString('ru-RU')}</time></article>)}
        {!items.length && !error && <div className="empty">Новых уведомлений нет</div>}
      </div>
    </>
  )
}

function Header({ title, action }: { title: string; action: ReactNode }) {
  return <header className="page-header"><div><span>Рабочее пространство</span><h1>{title}</h1></div>{action}</header>
}
