import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { exercisesApi, type Exercise } from '../api/exercises'

// ExercisePicker — поиск и выбор упражнения из каталога. Общий для логгера
// и конструктора программ.
export default function ExercisePicker({
  onPick,
  onClose,
}: {
  onPick: (ex: Exercise) => void
  onClose: () => void
}) {
  const [q, setQ] = useState('')
  const list = useQuery({
    queryKey: ['ex-search', q],
    queryFn: () => exercisesApi.list(q ? { q } : {}),
  })
  return (
    <div className="mt-3 rounded-2xl border border-slate-800 bg-slate-900 p-3">
      <div className="mb-2 flex items-center gap-2">
        <input
          autoFocus
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Поиск упражнения"
          aria-label="Поиск упражнения"
          className="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-100"
        />
        <button type="button" onClick={onClose} className="text-sm text-slate-400">
          Отмена
        </button>
      </div>
      <ul className="max-h-64 overflow-auto">
        {list.data?.slice(0, 30).map((e) => (
          <li key={e.id}>
            <button
              type="button"
              onClick={() => onPick(e)}
              className="w-full rounded-lg px-2 py-2 text-left text-sm text-slate-200 hover:bg-slate-800"
            >
              {e.name}
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
