export function parseRubles(value: string): number {
  const normalized = value.trim().replace(',', '.')
  if (!/^\d+(\.\d{1,2})?$/.test(normalized)) {
    throw new Error('Введите сумму с точностью до копеек')
  }
  const [whole, fraction = ''] = normalized.split('.')
  const minor = Number(whole) * 100 + Number(fraction.padEnd(2, '0'))
  if (!Number.isSafeInteger(minor) || minor <= 0) {
    throw new Error('Сумма должна быть больше нуля')
  }
  return minor
}

export function formatRubles(minor: number): string {
  return `${Math.floor(minor / 100).toLocaleString('ru-RU')},${String(minor % 100).padStart(2, '0')} ₽`
}
