import { parseDecimal } from './format'

describe('parseDecimal', () => {
  it('точка как разделитель', () => {
    expect(parseDecimal('25.5')).toBe(25.5)
  })

  it('запятая как разделитель', () => {
    expect(parseDecimal('26,5')).toBe(26.5)
  })

  it('целое число', () => {
    expect(parseDecimal('80')).toBe(80)
  })

  it('пробелы по краям игнорируются', () => {
    expect(parseDecimal(' 26,5 ')).toBe(26.5)
  })

  it('пустая строка → undefined', () => {
    expect(parseDecimal('')).toBeUndefined()
    expect(parseDecimal('   ')).toBeUndefined()
  })

  it('мусор → NaN', () => {
    expect(parseDecimal('abc')).toBeNaN()
    expect(parseDecimal('2,,5')).toBeNaN()
    expect(parseDecimal('2.5.5')).toBeNaN()
  })
})
