import { type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useMe } from '../auth/useAuth'

// RequireAuth пускает к контенту только залогиненного; пока грузится /me —
// показывает индикатор, при отсутствии сессии редиректит на /login.
export default function RequireAuth({ children }: { children: ReactNode }) {
  const { data: user, isLoading } = useMe()

  if (isLoading) {
    return (
      <div className="flex min-h-dvh items-center justify-center bg-slate-950 text-slate-400">
        Загрузка…
      </div>
    )
  }
  if (!user) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}
