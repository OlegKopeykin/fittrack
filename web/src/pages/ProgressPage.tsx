import { PageHeader } from '../components/AppShell'
import LineChart from '../components/LineChart'
import { useWorkouts } from '../training/useTraining'
import { bodyweightSeries } from '../lib/progress'

export default function ProgressPage() {
  const workouts = useWorkouts()
  const series = bodyweightSeries(workouts.data?.items ?? [])
  const first = series[0]
  const last = series[series.length - 1]
  const delta = first && last ? +(last.kg - first.kg).toFixed(1) : 0

  return (
    <>
      <PageHeader title="Прогресс" />
      <div className="mx-auto max-w-3xl px-5 py-4">
        <section className="rounded-2xl border border-slate-800 bg-slate-900 p-4">
          <div className="mb-3 flex items-baseline justify-between">
            <h2 className="text-sm font-bold uppercase tracking-wide text-slate-500">Вес тела</h2>
            {last && (
              <div className="text-right">
                <span className="text-2xl font-bold tabular-nums text-slate-50">{last.kg}</span>
                <span className="ml-1 text-sm text-slate-500">кг</span>
                {series.length > 1 && (
                  <div
                    className={`text-xs tabular-nums ${
                      delta > 0 ? 'text-amber-400' : delta < 0 ? 'text-emerald-400' : 'text-slate-500'
                    }`}
                  >
                    {delta > 0 ? '+' : ''}
                    {delta} кг за период
                  </div>
                )}
              </div>
            )}
          </div>
          {series.length === 0 ? (
            <p className="text-sm text-slate-500">
              Вес тела появится здесь, когда будет записан в тренировках.
            </p>
          ) : (
            <LineChart points={series} />
          )}
        </section>
      </div>
    </>
  )
}
