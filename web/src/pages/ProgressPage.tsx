import { PageHeader } from '../components/AppShell'

export default function ProgressPage() {
  return (
    <>
      <PageHeader title="Прогресс" />
      <div className="px-5 py-6 text-slate-400">
        Графики прогресса появятся, когда добавим дневник тренировок.
      </div>
    </>
  )
}
