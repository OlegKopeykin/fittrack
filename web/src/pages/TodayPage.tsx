import { Link } from 'react-router-dom'
import { useMe } from '../auth/useAuth'
import { usePrograms, useWorkouts } from '../training/useTraining'
import { PageHeader } from '../components/AppShell'
import { formatDate } from '../lib/format'

export default function TodayPage() {
  const { data: user } = useMe()
  const programs = usePrograms()
  const workouts = useWorkouts()

  return (
    <>
      <PageHeader title="Тренировка" />
      <div className="mx-auto max-w-3xl px-5 py-4">
        <p className="text-slate-400">
          Привет, {user?.display_name || user?.username}.
        </p>

        <section className="mt-6">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-sm font-bold uppercase tracking-wide text-slate-500">Программы</h2>
          </div>
          {programs.data?.length === 0 && (
            <p className="text-sm text-slate-500">Программ пока нет.</p>
          )}
          <div className="flex flex-col gap-2">
            {programs.data?.map((p) => (
              <Link
                key={p.id}
                to={`/program/${p.id}`}
                className="flex items-center justify-between rounded-2xl border border-slate-800 bg-slate-900 px-4 py-3"
              >
                <span className="font-semibold text-indigo-300">{p.name}</span>
                <span className="text-slate-600">›</span>
              </Link>
            ))}
          </div>
        </section>

        <section className="mt-8">
          <h2 className="mb-3 text-sm font-bold uppercase tracking-wide text-slate-500">
            История тренировок
          </h2>
          {workouts.data?.items.length === 0 && (
            <p className="text-sm text-slate-500">Тренировок пока нет.</p>
          )}
          <ul className="flex flex-col">
            {workouts.data?.items.map((wk) => (
              <li key={wk.id}>
                <Link
                  to={`/workout/${wk.id}`}
                  className="flex items-center justify-between border-b border-slate-800/60 py-3"
                >
                  <div>
                    <div className="font-medium text-slate-100">{formatDate(wk.date)}</div>
                    {(wk.feeling || wk.bodyweight_kg) && (
                      <div className="text-xs text-slate-500">
                        {wk.bodyweight_kg ? `${wk.bodyweight_kg} кг` : ''}
                        {wk.bodyweight_kg && wk.feeling ? ' · ' : ''}
                        {wk.feeling}
                      </div>
                    )}
                  </div>
                  <span className="text-slate-600">›</span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      </div>
    </>
  )
}
