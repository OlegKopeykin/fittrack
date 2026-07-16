import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  useCreateProgram,
  useUpdateProgram,
  useProgram,
  useExerciseMap,
} from '../training/useTraining'
import { PageHeader } from '../components/AppShell'
import ExercisePicker from '../training/ExercisePicker'

type BExercise = { id: number; name: string }
type BDay = { name: string; exercises: BExercise[] }

export default function ProgramBuilderPage() {
  const navigate = useNavigate()
  const { id } = useParams()
  const editId = id ? Number(id) : undefined
  const editing = editId !== undefined

  const create = useCreateProgram()
  const update = useUpdateProgram(editId ?? 0)
  const program = useProgram(editId ?? 0, editing)
  const exMap = useExerciseMap()

  const [name, setName] = useState('')
  const [days, setDays] = useState<BDay[]>([{ name: 'День 1', exercises: [] }])
  const [pickingDay, setPickingDay] = useState<number | null>(null)

  // Предзаполнение при правке — один раз, когда пришли программа и карта имён.
  const seeded = useRef(false)
  useEffect(() => {
    if (editing && !seeded.current && program.data && exMap.data) {
      setName(program.data.name)
      setDays(
        (program.data.days ?? []).map((d) => ({
          name: d.name,
          exercises: d.exercises.map((rx) => ({
            id: rx.exercise_id,
            name: exMap.data!.get(rx.exercise_id)?.name ?? `Упражнение #${rx.exercise_id}`,
          })),
        })),
      )
      seeded.current = true
    }
  }, [editing, program.data, exMap.data])

  const pending = create.isPending || update.isPending
  const canSave = name.trim() !== '' && days.some((d) => d.exercises.length > 0)

  function updateDay(i: number, patch: Partial<BDay>) {
    setDays((ds) => ds.map((d, j) => (j === i ? { ...d, ...patch } : d)))
  }
  function addExercise(dayIdx: number, ex: BExercise) {
    setDays((ds) =>
      ds.map((d, j) =>
        j === dayIdx && !d.exercises.some((e) => e.id === ex.id)
          ? { ...d, exercises: [...d.exercises, ex] }
          : d,
      ),
    )
    setPickingDay(null)
  }
  function removeExercise(dayIdx: number, exId: number) {
    setDays((ds) =>
      ds.map((d, j) =>
        j === dayIdx ? { ...d, exercises: d.exercises.filter((e) => e.id !== exId) } : d,
      ),
    )
  }

  function save() {
    const payload = {
      name: name.trim(),
      days: days.map((d, i) => ({
        name: d.name.trim() || `День ${i + 1}`,
        exercises: d.exercises.map((e) => ({ exercise_id: e.id })),
      })),
    }
    const done = { onSuccess: (prog: { id: number }) => navigate(`/program/${prog.id}`) }
    if (editing) update.mutate(payload, done)
    else create.mutate(payload, done)
  }

  return (
    <>
      <PageHeader
        title={editing ? 'Редактировать программу' : 'Новая программа'}
        right={
          <Link to={editing ? `/program/${editId}` : '/'} className="text-sm text-slate-400">
            ‹ Назад
          </Link>
        }
      />
      <div className="mx-auto max-w-3xl px-5 py-4">
        <label className="mb-1 block text-xs font-bold text-slate-500">Название</label>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Например: Фул бади или 5-дневный сплит"
          aria-label="Название программы"
          className="mb-5 w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-slate-100"
        />

        <div className="flex flex-col gap-4">
          {days.map((day, di) => (
            <section key={di} className="rounded-2xl border border-slate-800 bg-slate-900 p-3">
              <div className="mb-3 flex items-center gap-2">
                <input
                  value={day.name}
                  onChange={(e) => updateDay(di, { name: e.target.value })}
                  aria-label={`Название дня ${di + 1}`}
                  className="flex-1 rounded-lg border border-slate-800 bg-slate-950 px-2 py-1.5 font-bold text-slate-100"
                />
                {days.length > 1 && (
                  <button
                    type="button"
                    onClick={() => setDays((ds) => ds.filter((_, j) => j !== di))}
                    aria-label={`Удалить день ${di + 1}`}
                    className="rounded-lg border border-slate-700 px-2 py-1.5 text-sm text-slate-400"
                  >
                    Удалить день
                  </button>
                )}
              </div>

              {day.exercises.length === 0 && (
                <p className="mb-2 text-sm text-slate-500">Упражнений пока нет.</p>
              )}
              <ul className="mb-2 flex flex-col gap-1.5">
                {day.exercises.map((ex) => (
                  <li
                    key={ex.id}
                    className="flex items-center justify-between rounded-lg border border-slate-800 bg-slate-950 px-3 py-2"
                  >
                    <span className="text-sm text-slate-200">{ex.name}</span>
                    <button
                      type="button"
                      onClick={() => removeExercise(di, ex.id)}
                      aria-label={`Убрать ${ex.name}`}
                      className="text-slate-500"
                    >
                      ✕
                    </button>
                  </li>
                ))}
              </ul>

              {pickingDay === di ? (
                <ExercisePicker
                  onPick={(ex) => addExercise(di, { id: ex.id, name: ex.name })}
                  onClose={() => setPickingDay(null)}
                />
              ) : (
                <button
                  type="button"
                  onClick={() => setPickingDay(di)}
                  className="w-full rounded-xl border border-indigo-500/30 bg-indigo-500/10 px-3 py-2 text-sm font-bold text-indigo-300"
                >
                  + Добавить упражнение
                </button>
              )}
            </section>
          ))}
        </div>

        <button
          type="button"
          onClick={() => setDays((ds) => [...ds, { name: `День ${ds.length + 1}`, exercises: [] }])}
          className="mt-3 w-full rounded-2xl border border-slate-700 px-4 py-2.5 text-sm font-semibold text-slate-300"
        >
          + Добавить день
        </button>

        <button
          type="button"
          disabled={!canSave || pending}
          onClick={save}
          className="mt-5 w-full rounded-2xl bg-indigo-500 px-4 py-3.5 text-[15px] font-extrabold text-white disabled:opacity-50"
        >
          Сохранить программу
        </button>
        {(create.isError || update.isError) && (
          <p className="mt-2 text-sm text-rose-400">Не удалось сохранить. Проверьте поля.</p>
        )}
      </div>
    </>
  )
}
