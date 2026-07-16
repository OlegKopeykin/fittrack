import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  useProgram,
  useExerciseMap,
  useArchiveProgram,
  useStartWorkout,
} from '../training/useTraining'
import { PageHeader } from '../components/AppShell'
import { formatReps, formatWeightRange } from '../lib/format'

export default function ProgramDetailPage() {
  const { id } = useParams()
  const pid = Number(id)
  const program = useProgram(pid)
  const exMap = useExerciseMap()
  const archive = useArchiveProgram()
  const start = useStartWorkout()
  const navigate = useNavigate()

  return (
    <>
      <PageHeader
        title={program.data?.name ?? 'Программа'}
        right={
          <Link to="/" className="text-sm text-slate-400">
            ‹ Назад
          </Link>
        }
      />
      <div className="mx-auto max-w-3xl px-5 py-4">
        {program.isLoading && <p className="text-slate-500">Загрузка…</p>}
        {program.data?.description && (
          <p className="mb-4 text-sm text-slate-400">{program.data.description}</p>
        )}

        <div className="flex flex-col gap-4">
          {program.data?.days?.map((day) => (
            <section key={day.id} className="rounded-2xl border border-slate-800 bg-slate-900 p-3">
              <div className="mb-3 flex items-center justify-between gap-3">
                <h2 className="font-bold text-slate-100">{day.name}</h2>
                <button
                  type="button"
                  disabled={start.isPending}
                  onClick={() =>
                    start.mutate(
                      {
                        program_day_id: day.id,
                        title: `${program.data?.name ?? ''} · ${day.name}`.trim(),
                      },
                      { onSuccess: (wk) => navigate(`/workout/${wk.id}`) },
                    )
                  }
                  className="rounded-lg bg-indigo-500 px-3 py-1.5 text-sm font-bold text-white disabled:opacity-60"
                >
                  Начать
                </button>
              </div>
              <ul className="flex flex-col gap-2.5">
                {day.exercises.map((rx) => {
                  const ex = exMap.data?.get(rx.exercise_id)
                  const reps = formatReps(rx.rep_min, rx.rep_max)
                  const weight = formatWeightRange(rx.weight_min_kg, rx.weight_max_kg)
                  return (
                    <li key={rx.id} className="flex items-baseline justify-between gap-3">
                      <div className="min-w-0">
                        <div className="font-medium text-indigo-300">
                          {ex?.name ?? `Упражнение #${rx.exercise_id}`}
                        </div>
                        {(weight || rx.tempo || rx.rest_sec || rx.notes) && (
                          <div className="text-xs text-slate-500">
                            {[
                              weight,
                              rx.rest_sec ? `отдых ${rx.rest_sec} с` : '',
                              rx.tempo ? `темп ${rx.tempo}` : '',
                              rx.notes,
                            ]
                              .filter(Boolean)
                              .join(' · ')}
                          </div>
                        )}
                      </div>
                      <div className="whitespace-nowrap text-base font-bold tabular-nums text-slate-100">
                        {rx.sets}
                        {reps ? ` × ${reps}` : ''}
                      </div>
                    </li>
                  )
                })}
              </ul>
            </section>
          ))}
        </div>

        {program.data && (
          <button
            type="button"
            onClick={() =>
              archive.mutate(
                { id: pid, archived: !program.data!.archived },
                { onSuccess: () => navigate('/') },
              )
            }
            className="mt-6 w-full rounded-xl border border-slate-700 px-4 py-2.5 text-sm font-semibold text-slate-300"
          >
            {program.data.archived ? 'Вернуть из архива' : 'В архив'}
          </button>
        )}
      </div>
    </>
  )
}
