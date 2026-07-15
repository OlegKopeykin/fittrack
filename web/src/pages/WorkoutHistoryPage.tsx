import { Link } from 'react-router-dom'
import { useWorkouts } from '../training/useTraining'
import { PageHeader } from '../components/AppShell'
import { formatDate } from '../lib/format'

export default function WorkoutHistoryPage() {
  const workouts = useWorkouts()

  return (
    <>
      <PageHeader
        title="История"
        right={
          <Link to="/" className="text-sm text-slate-400">
            ‹ Назад
          </Link>
        }
      />
      <div className="mx-auto max-w-3xl px-5 py-4">
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
                  {wk.feeling && <div className="text-xs text-slate-500">{wk.feeling}</div>}
                </div>
                <span className="text-slate-600">›</span>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </>
  )
}
