import { NavLink, Outlet } from 'react-router-dom'
import { type ReactNode } from 'react'
import { ChartIcon, DumbbellIcon, GridIcon, UserIcon } from './icons'

type Item = { to: string; label: string; icon: (p: { className?: string }) => ReactNode }

const items: Item[] = [
  { to: '/', label: 'Тренировка', icon: DumbbellIcon },
  { to: '/exercises', label: 'Упражнения', icon: GridIcon },
  { to: '/progress', label: 'Прогресс', icon: ChartIcon },
  { to: '/profile', label: 'Профиль', icon: UserIcon },
]

export default function AppShell() {
  return (
    <div className="min-h-dvh bg-slate-950 text-slate-50 lg:grid lg:grid-cols-[224px_1fr]">
      {/* Боковая навигация — десктоп */}
      <aside className="hidden border-r border-slate-800 p-3 lg:flex lg:flex-col lg:gap-1">
        <div className="px-3 pb-4 pt-2 text-lg font-extrabold tracking-tight">
          Fit<span className="text-indigo-400">Track</span>
        </div>
        {items.map((it) => (
          <NavLink
            key={it.to}
            to={it.to}
            end={it.to === '/'}
            className={({ isActive }) =>
              `flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-semibold ${
                isActive ? 'bg-indigo-500/15 text-indigo-300' : 'text-slate-400 hover:text-slate-200'
              }`
            }
          >
            <it.icon className="size-5" />
            {it.label}
          </NavLink>
        ))}
      </aside>

      {/* Контент */}
      <div className="min-w-0 pb-20 lg:pb-0">
        <Outlet />
      </div>

      {/* Нижний таб-бар — мобильный */}
      <nav
        className="fixed inset-x-0 bottom-0 z-10 grid grid-cols-4 border-t border-slate-800 bg-slate-900 lg:hidden"
        style={{ paddingBottom: 'env(safe-area-inset-bottom)' }}
      >
        {items.map((it) => (
          <NavLink
            key={it.to}
            to={it.to}
            end={it.to === '/'}
            className={({ isActive }) =>
              `flex flex-col items-center gap-1 py-2 text-[10px] font-semibold ${
                isActive ? 'text-indigo-300' : 'text-slate-500'
              }`
            }
          >
            <it.icon className="size-6" />
            {it.label}
          </NavLink>
        ))}
      </nav>
    </div>
  )
}

// PageHeader — единый sticky-заголовок страниц.
export function PageHeader({ title, right }: { title: string; right?: ReactNode }) {
  return (
    <header className="sticky top-0 z-[5] bg-slate-950/90 px-5 py-4 backdrop-blur">
      <div className="mx-auto flex max-w-3xl items-center justify-between gap-3">
        <h1 className="text-xl font-extrabold tracking-tight">{title}</h1>
        {right}
      </div>
    </header>
  )
}
