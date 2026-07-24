import { describe, expect, it } from 'vitest'
import { parseRubles } from './money'

describe('parseRubles', () => {
  it('converts decimal RUB to integer kopecks', () => {
    expect(parseRubles('125.50')).toBe(12550)
    expect(parseRubles('0.01')).toBe(1)
  })

  it('rejects more than two decimal places', () => {
    expect(() => parseRubles('1.001')).toThrow()
  })
})
