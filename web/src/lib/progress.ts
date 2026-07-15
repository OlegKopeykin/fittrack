import type { Workout } from '../api/training'

export type Point = { date: string; kg: number }

// bodyweightSeries — точки веса тела по датам (последнее значение за дату),
// отсортированные по возрастанию даты.
export function bodyweightSeries(workouts: Workout[]): Point[] {
  const byDate = new Map<string, number>()
  for (const w of workouts) {
    if (w.bodyweight_kg != null) byDate.set(w.date, w.bodyweight_kg)
  }
  return [...byDate.entries()]
    .map(([date, kg]) => ({ date, kg }))
    .sort((a, b) => (a.date < b.date ? -1 : 1))
}
