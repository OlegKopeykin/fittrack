import { bodyweightSeries } from './progress'
import type { Workout } from '../api/training'

function wk(date: string, bw?: number): Workout {
  return { id: 0, date, feeling: '', notes: '', bodyweight_kg: bw }
}

describe('bodyweightSeries', () => {
  it('берёт только тренировки с весом, сортирует по дате', () => {
    const s = bodyweightSeries([wk('2026-05-10', 86.5), wk('2026-04-09', 87), wk('2026-04-13')])
    expect(s).toEqual([
      { date: '2026-04-09', kg: 87 },
      { date: '2026-05-10', kg: 86.5 },
    ])
  })

  it('дедуплицирует по дате (последнее значение)', () => {
    const s = bodyweightSeries([wk('2026-04-09', 87), wk('2026-04-09', 86)])
    expect(s).toEqual([{ date: '2026-04-09', kg: 86 }])
  })

  it('пустой список без веса → пусто', () => {
    expect(bodyweightSeries([wk('2026-04-09')])).toEqual([])
  })
})
