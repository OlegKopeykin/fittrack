import { useMemo, useState } from 'react'
import { useExercises, useMuscleGroups } from '../exercises/useExercises'
import { kindLabel, type MuscleGroup } from '../api/exercises'
import { PageHeader } from '../components/AppShell'
import { SearchIcon } from '../components/icons'

// Палитра точек по группе мышц (по порядку slug'ов).
const dotColors = [
  'bg-rose-400', 'bg-indigo-400', 'bg-teal-400', 'bg-amber-400',
  'bg-emerald-400', 'bg-sky-400', 'bg-fuchsia-400', 'bg-orange-400',
  'bg-lime-400', 'bg-cyan-400', 'bg-violet-400', 'bg-pink-400',
]

export default function ExercisesPage() {
  const [q, setQ] = useState('')
  const [group, setGroup] = useState('')

  const groupsQ = useMuscleGroups()
  const exercisesQ = useExercises({ q: q.trim() || undefined, muscleGroup: group || undefined })

  const groupById = useMemo(() => {
    const m = new Map<number, MuscleGroup & { color: string }>()
    groupsQ.data?.forEach((g, i) => m.set(g.id, { ...g, color: dotColors[i % dotColors.length] }))
    return m
  }, [groupsQ.data])

  return (
    <>
      <PageHeader title="Упражнения" />

      <div className="mx-auto max-w-3xl">
      <div className="sticky top-[60px] z-[4] bg-slate-950/90 px-5 pb-3 backdrop-blur">
        <label className="flex items-center gap-2 rounded-xl border border-slate-800 bg-slate-900 px-3 py-2.5 text-slate-300">
          <SearchIcon className="size-4 text-slate-500" />
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Поиск по названию и синонимам"
            className="w-full bg-transparent text-base text-slate-100 outline-none placeholder:text-slate-500"
          />
        </label>

        <div className="mt-3 flex gap-2 overflow-x-auto pb-1">
          {groupsQ.data?.map((g) => {
            const active = group === g.slug
            return (
              <button
                key={g.slug}
                type="button"
                onClick={() => setGroup(active ? '' : g.slug)}
                className={`whitespace-nowrap rounded-full border px-3 py-1.5 text-sm font-semibold ${
                  active
                    ? 'border-indigo-500/40 bg-indigo-500/15 text-indigo-300'
                    : 'border-slate-800 bg-slate-900 text-slate-400'
                }`}
              >
                {g.name_ru}
              </button>
            )
          })}
        </div>
      </div>

      <div>
        {exercisesQ.isLoading && <p className="px-5 py-6 text-slate-500">Загрузка…</p>}
        {exercisesQ.isError && (
          <p className="px-5 py-6 text-red-400">Не удалось загрузить упражнения.</p>
        )}
        {exercisesQ.data?.length === 0 && (
          <p className="px-5 py-6 text-slate-500">Ничего не найдено.</p>
        )}

        <ul>
          {exercisesQ.data?.map((ex) => {
            const g = groupById.get(ex.muscle_group_id)
            return (
              <li
                key={ex.id}
                className="flex items-center gap-3 border-b border-slate-800/60 px-5 py-3"
              >
                <span className={`size-2.5 flex-none rounded-full ${g?.color ?? 'bg-slate-600'}`} />
                <div className="min-w-0">
                  <div className="font-medium text-slate-100">{ex.name}</div>
                  <div className="text-xs text-slate-500">
                    {g?.name_ru ?? '—'} · {kindLabel[ex.kind]}
                    {ex.per_arm ? ' · на руку' : ''}
                  </div>
                </div>
              </li>
            )
          })}
        </ul>
      </div>
      </div>
    </>
  )
}
