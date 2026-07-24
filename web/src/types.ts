export type Task = {
  id: number
  version: number
  title: string
  description: string | null
  completed: boolean
  created_at: string
  completed_at: string | null
}

export type Wallet = {
  version: number
  balance_minor: number
  currency: 'RUB'
}

export type Notification = {
  id: number
  title: string
  body: string
  read_at: string | null
  created_at: string
}

export type Payment = {
  id: string
  provider_payment_id: string
  status: string
  amount_minor: number
  currency: string
  confirmation_url: string
}
