import { useNavigate } from 'react-router-dom'
import { useMe, useLogout } from '../auth/useAuth'

export default function DashboardPage() {
  const { data: user } = useMe()
  const logout = useLogout()
  const navigate = useNavigate()

  return (
    <main className="min-h-dvh bg-slate-950 px-6 py-8 text-slate-100">
      <div className="mx-auto max-w-md">
        <h1 className="text-2xl font-bold">FitTrack</h1>
        <p className="mt-2 text-slate-400">
          Привет, {user?.display_name || user?.username}!
        </p>
        <p className="mt-6 text-slate-500">
          Здесь появятся упражнения, тренировки и прогресс.
        </p>
        <button
          type="button"
          onClick={() => logout.mutate(undefined, { onSuccess: () => navigate('/login') })}
          className="mt-8 rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300"
        >
          Выйти
        </button>
      </div>
    </main>
  )
}
