import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  useExercise,
  useMuscleGroups,
  useExerciseNote,
  useSetExerciseNote,
} from '../exercises/useExercises'
import { useExerciseHistory } from '../training/useTraining'
import { kindLabel, exerciseType, exTypeLabel } from '../api/exercises'
import { PageHeader } from '../components/AppShell'
import { formatDate, formatSet } from '../lib/format'
import type { WorkoutSet } from '../api/training'

export default function ExercisePage() {
  const { id } = useParams()
  const exId = Number(id)
  const ex = useExercise(exId)
  const groups = useMuscleGroups()
  const history = useExerciseHistory(exId)
  const note = useExerciseNote(exId)
  const saveNote = useSetExerciseNote(exId)

  const [text, setText] = useState('')
  const [seeded, setSeeded] = useState(false)
  useEffect(() => {
    if (!seeded && note.data) {
      setText(note.data.note)
      setSeeded(true)
    }
  }, [seeded, note.data])

  const groupName = groups.data?.find((g) => g.id === ex.data?.muscle_group_id)?.name_ru
  const et = ex.data ? exerciseType(ex.data) : 'other'
  const dirty = note.data !== undefined && text !== note.data.note

  return (
    <>
      <PageHeader
        title={ex.data?.name ?? 'Упражнение'}
        right={
          <div className="flex items-center gap-2">
            {ex.data && !ex.data.global && (
              <Link
                to={`/exercises/${exId}/edit`}
                className="rounded-lg border border-slate-700 px-3 py-1.5 text-sm font-semibold text-slate-300"
              >
                Изменить
              </Link>
            )}
            <Link to="/exercises" className="text-sm text-slate-400">
              ‹ Назад
            </Link>
          </div>
        }
      />

      <div className="mx-auto max-w-3xl px-5 py-4">
        {ex.isLoading && <p className="text-slate-500">Загрузка…</p>}

        {ex.data && (
          <>
            {ex.data.has_image && (
              <img
                src={ex.data.image_url}
                alt={ex.data.name}
                className="mb-4 max-h-64 w-full rounded-2xl border border-slate-800 object-cover"
              />
            )}

            <div className="mb-5 flex flex-wrap gap-2 text-xs">
              {groupName && (
                <span className="rounded-full bg-slate-800 px-2.5 py-1 font-semibold text-slate-300">
                  {groupName}
                </span>
              )}
              <span className="rounded-full bg-slate-800 px-2.5 py-1 font-semibold text-slate-300">
                {kindLabel[ex.data.kind]}
              </span>
              {et !== 'other' && (
                <span className="rounded-full bg-indigo-500/15 px-2.5 py-1 font-semibold text-indigo-300">
                  {exTypeLabel[et]}
                </span>
              )}
              {ex.data.per_arm && (
                <span className="rounded-full bg-slate-800 px-2.5 py-1 font-semibold text-slate-300">
                  на руку
                </span>
              )}
            </div>

            {/* Комментарий */}
            <h2 className="mb-1 text-sm font-bold text-slate-100">Мой комментарий</h2>
            <textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Заметка по технике, настройке тренажёра, ощущениям…"
              aria-label="Комментарий к упражнению"
              rows={3}
              className="w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-sm text-slate-100"
            />
            {dirty && (
              <button
                type="button"
                disabled={saveNote.isPending}
                onClick={() => saveNote.mutate(text)}
                className="mt-2 rounded-xl bg-indigo-500 px-4 py-2 text-sm font-bold text-white disabled:opacity-50"
              >
                Сохранить
              </button>
            )}

            {/* Лог */}
            <h2 className="mb-2 mt-6 text-sm font-bold text-slate-100">История</h2>
            {history.isLoading && <p className="text-sm text-slate-500">Загрузка…</p>}
            {history.data?.length === 0 && (
              <p className="text-sm text-slate-500">Пока нет записей — упражнение ещё не логировалось.</p>
            )}
            <div className="flex flex-col gap-2">
              {history.data?.map((session, i) => (
                <div key={i} className="rounded-2xl border border-slate-800 bg-slate-900 p-3">
                  <div className="mb-1 text-xs font-semibold text-slate-500">
                    {formatDate(session.date)}
                  </div>
                  <div className="flex flex-wrap gap-x-4 gap-y-1 tabular-nums">
                    {session.sets.map((s, j) => (
                      <span
                        key={j}
                        className={s.role === 'warmup' ? 'text-slate-500' : 'font-semibold text-slate-100'}
                      >
                        {formatSet(s as WorkoutSet)}
                      </span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </>
  )
}
