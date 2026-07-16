import { Link, useNavigate } from 'react-router-dom'
import { usePrograms, useStartWorkout } from '../training/useTraining'
import { PageHeader } from '../components/AppShell'

export default function TodayPage() {
  const programs = usePrograms()
  const start = useStartWorkout()
  const navigate = useNavigate()

  return (
    <>
      <PageHeader
        title="Тренировка"
        right={
          <Link
            to="/workouts"
            className="rounded-xl border border-slate-700 px-3 py-1.5 text-sm font-semibold text-slate-300"
          >
            История
          </Link>
        }
      />
      <div className="mx-auto max-w-3xl px-5 py-4">
        <button
          type="button"
          disabled={start.isPending}
          onClick={() =>
            start.mutate(
              {},
              { onSuccess: (wk) => navigate(`/workout/${wk.id}`) },
            )
          }
          className="mb-5 w-full rounded-2xl bg-indigo-500 px-4 py-3.5 text-[15px] font-extrabold text-white disabled:opacity-60"
        >
          Начать пустую тренировку
        </button>

        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-bold uppercase tracking-wide text-slate-500">Программы</h2>
          <Link to="/programs/new" className="text-sm font-semibold text-indigo-300">
            + Создать
          </Link>
        </div>
        {programs.data?.length === 0 && <p className="text-sm text-slate-500">Программ пока нет.</p>}
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
      </div>
    </>
  )
}
