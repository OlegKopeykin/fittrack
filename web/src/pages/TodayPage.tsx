import { useMe } from '../auth/useAuth'
import { PageHeader } from '../components/AppShell'

export default function TodayPage() {
  const { data: user } = useMe()
  return (
    <>
      <PageHeader title="Тренировка" />
      <div className="px-5 py-6">
        <p className="text-slate-400">
          Привет, {user?.display_name || user?.username}. Здесь будет старт тренировки.
        </p>
        <button
          type="button"
          disabled
          className="mt-6 w-full rounded-xl bg-indigo-600/50 px-4 py-3 text-base font-semibold text-white/70"
        >
          Начать тренировку (скоро)
        </button>
      </div>
    </>
  )
}
