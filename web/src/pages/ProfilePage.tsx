import { useNavigate } from 'react-router-dom'
import { useLogout, useMe } from '../auth/useAuth'
import { PageHeader } from '../components/AppShell'

export default function ProfilePage() {
  const { data: user } = useMe()
  const logout = useLogout()
  const navigate = useNavigate()

  return (
    <>
      <PageHeader title="Профиль" />
      <div className="px-5 py-6">
        <div className="rounded-2xl border border-slate-800 bg-slate-900 p-4">
          <div className="text-lg font-bold">{user?.display_name || user?.username}</div>
          <div className="mt-1 text-sm text-slate-500">
            {user?.role === 'owner' ? 'Владелец' : 'Пользователь'}
          </div>
        </div>
        <button
          type="button"
          onClick={() => logout.mutate(undefined, { onSuccess: () => navigate('/login') })}
          className="mt-6 w-full rounded-xl border border-slate-700 px-4 py-3 text-sm font-semibold text-slate-300"
        >
          Выйти
        </button>
      </div>
    </>
  )
}
