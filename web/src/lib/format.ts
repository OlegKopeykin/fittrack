import type { WorkoutSet } from '../api/training'

const ruDate = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })

// parseDecimal — число из пользовательского ввода; точка и запятая
// равнозначны как десятичный разделитель ('26,5' → 26.5).
// Пустая строка → undefined, нечисловой ввод → NaN.
export function parseDecimal(raw: string): number | undefined {
  const s = raw.trim()
  if (s === '') return undefined
  return Number(s.replace(',', '.'))
}

export function formatDate(iso: string): string {
  // iso = YYYY-MM-DD
  const [y, m, d] = iso.split('-').map(Number)
  if (!y || !m || !d) return iso
  return ruDate.format(new Date(y, m - 1, d))
}

// formatSet — краткая запись подхода: вес×повторы, только повторы,
// длительность (изометрия) или дистанция×время (кардио).
export function formatSet(s: WorkoutSet): string {
  if (s.distance_km != null) {
    const t = s.duration_sec != null ? ` · ${Math.round(s.duration_sec / 60)} мин` : ''
    return `${s.distance_km} км${t}`
  }
  if (s.weight_kg != null && s.reps != null) return `${s.weight_kg}×${s.reps}`
  if (s.reps != null) return `${s.reps}`
  if (s.duration_sec != null) return `${s.duration_sec} с`
  return '—'
}

// formatReps — диапазон повторов для предписания.
export function formatReps(min?: number, max?: number): string {
  if (min == null && max == null) return ''
  if (min != null && max != null) return min === max ? `${min}` : `${min}–${max}`
  return `${min ?? max}`
}

export function formatWeightRange(min?: number, max?: number): string {
  if (min == null && max == null) return ''
  if (min != null && max != null) return min === max ? `${min} кг` : `${min}–${max} кг`
  return `${min ?? max} кг`
}
