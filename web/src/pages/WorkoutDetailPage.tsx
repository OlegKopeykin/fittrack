import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useWorkout, useExerciseMap } from '../training/useTraining'
import { PageHeader } from '../components/AppShell'
import { formatDate, formatSet } from '../lib/format'
import type { WorkoutSet } from '../api/training'
import WorkoutLogger from '../training/WorkoutLogger'

const roleLabel: Record<string, string> = { warmup: 'разминка', ramp: 'подводящий', working: '' }

export default function WorkoutDetailPage() {
  const { id } = useParams()
  const workout = useWorkout(Number(id))
  const exMap = useExerciseMap()
  const [editing, setEditing] = useState(false)

  // Активная (не завершённая) или явно открытая на правку — интерактивный логгер.
  if (workout.data && (!workout.data.finished_at || editing)) {
    return <WorkoutLogger workout={workout.data} />
  }

  const sets = workout.data?.sets ?? []
  // группируем подряд идущие подходы одного упражнения
  const groups: { exerciseId: number; sets: WorkoutSet[] }[] = []
  for (const s of sets) {
    const last = groups[groups.length - 1]
    if (last && last.exerciseId === s.exercise_id) last.sets.push(s)
    else groups.push({ exerciseId: s.exercise_id, sets: [s] })
  }

  return (
    <>
      <PageHeader
        title={workout.data?.title || (workout.data ? formatDate(workout.data.date) : 'Тренировка')}
        right={
          <div className="flex items-center gap-2">
            {workout.data && (
              <button
                type="button"
                onClick={() => setEditing(true)}
                className="rounded-lg border border-slate-700 px-3 py-1.5 text-sm font-semibold text-slate-300"
              >
                Редактировать
              </button>
            )}
            <Link to="/workouts" className="text-sm text-slate-400">
              ‹ Назад
            </Link>
          </div>
        }
      />
      <div className="mx-auto max-w-3xl px-5 py-4">
        {workout.isLoading && <p className="text-slate-500">Загрузка…</p>}
        {workout.data && (
          <p className="mb-4 text-sm text-slate-400">
            {workout.data.title ? `${formatDate(workout.data.date)}` : ''}
            {workout.data.title && workout.data.feeling ? ' · ' : ''}
            {workout.data.feeling}
          </p>
        )}

        <div className="flex flex-col gap-3">
          {groups.map((g, i) => {
            const ex = exMap.data?.get(g.exerciseId)
            return (
              <div key={i} className="rounded-2xl border border-slate-800 bg-slate-900 p-3">
                <div className="mb-2 font-semibold text-indigo-300">
                  {ex?.name ?? `Упражнение #${g.exerciseId}`}
                </div>
                <div className="flex flex-wrap gap-x-4 gap-y-1 tabular-nums">
                  {g.sets.map((s) => (
                    <span
                      key={s.id}
                      className={s.role === 'warmup' ? 'text-slate-500' : 'font-semibold text-slate-100'}
                      title={[roleLabel[s.role], s.note].filter(Boolean).join(' · ')}
                    >
                      {formatSet(s)}
                    </span>
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </>
  )
}
