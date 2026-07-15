import { Link } from 'react-router-dom'
import { usePrograms } from '../training/useTraining'
import { PageHeader } from '../components/AppShell'

export default function TodayPage() {
  const programs = usePrograms()

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
        <h2 className="mb-3 text-sm font-bold uppercase tracking-wide text-slate-500">Программы</h2>
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
