import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  useMuscleGroups,
  useExercise,
  useCreateExercise,
  useUpdateExercise,
} from '../exercises/useExercises'
import { kindLabel, equipmentLabel, type ExerciseKind } from '../api/exercises'
import { PageHeader } from '../components/AppShell'

const KINDS: ExerciseKind[] = ['compound', 'isolation', 'isometric', 'bodyweight', 'cardio']
const EQUIPMENT = ['', 'barbell', 'dumbbell', 'machine', 'cable', 'bodyweight', 'band', 'kettlebell', 'other', 'none']

export default function ExerciseFormPage() {
  const navigate = useNavigate()
  const { id } = useParams()
  const editId = id ? Number(id) : undefined
  const editing = editId !== undefined

  const groups = useMuscleGroups()
  const exercise = useExercise(editId ?? 0, editing)
  const create = useCreateExercise()
  const update = useUpdateExercise(editId ?? 0)

  const [name, setName] = useState('')
  const [muscleGroup, setMuscleGroup] = useState('')
  const [kind, setKind] = useState<ExerciseKind>('compound')
  const [equipment, setEquipment] = useState('')
  const [perArm, setPerArm] = useState(false)
  const [notes, setNotes] = useState('')

  const slugById = useMemo(() => {
    const m = new Map<number, string>()
    groups.data?.forEach((g) => m.set(g.id, g.slug))
    return m
  }, [groups.data])

  const seeded = useRef(false)
  useEffect(() => {
    if (editing && !seeded.current && exercise.data && slugById.size > 0) {
      const ex = exercise.data
      setName(ex.name)
      setMuscleGroup(slugById.get(ex.muscle_group_id) ?? '')
      setKind(ex.kind)
      setEquipment(ex.equipment ?? '')
      setPerArm(ex.per_arm)
      setNotes(ex.technique_notes ?? '')
      seeded.current = true
    }
  }, [editing, exercise.data, slugById])

  const pending = create.isPending || update.isPending
  const canSave = name.trim() !== '' && muscleGroup !== ''
  const err = create.error || update.error
  const globalReadonly = editing && exercise.data?.global

  function save() {
    const body = {
      name: name.trim(),
      muscle_group: muscleGroup,
      kind,
      per_arm: perArm,
      equipment,
      technique_notes: notes.trim(),
    }
    const done = { onSuccess: () => navigate('/exercises') }
    if (editing) update.mutate(body, done)
    else create.mutate(body, done)
  }

  return (
    <>
      <PageHeader
        title={editing ? 'Изменить упражнение' : 'Новое упражнение'}
        right={
          <Link to="/exercises" className="text-sm text-slate-400">
            ‹ Назад
          </Link>
        }
      />
      <div className="mx-auto max-w-xl px-5 py-4">
        {globalReadonly && (
          <p className="mb-4 rounded-xl border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-300">
            Это упражнение из общего каталога — его нельзя менять. Создайте своё.
          </p>
        )}

        <label className="mb-1 block text-xs font-bold text-slate-500">Название</label>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          aria-label="Название упражнения"
          disabled={globalReadonly}
          className="mb-4 w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-slate-100 disabled:opacity-50"
        />

        <label className="mb-1 block text-xs font-bold text-slate-500">Группа мышц</label>
        <select
          value={muscleGroup}
          onChange={(e) => setMuscleGroup(e.target.value)}
          aria-label="Группа мышц"
          disabled={globalReadonly}
          className="mb-4 w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-slate-100 disabled:opacity-50"
        >
          <option value="">— выберите —</option>
          {groups.data?.map((g) => (
            <option key={g.slug} value={g.slug}>
              {g.name_ru}
            </option>
          ))}
        </select>

        <label className="mb-1 block text-xs font-bold text-slate-500">Вид</label>
        <select
          value={kind}
          onChange={(e) => setKind(e.target.value as ExerciseKind)}
          aria-label="Вид упражнения"
          disabled={globalReadonly}
          className="mb-4 w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-slate-100 disabled:opacity-50"
        >
          {KINDS.map((k) => (
            <option key={k} value={k}>
              {kindLabel[k]}
            </option>
          ))}
        </select>

        <label className="mb-1 block text-xs font-bold text-slate-500">Снаряд</label>
        <select
          value={equipment}
          onChange={(e) => setEquipment(e.target.value)}
          aria-label="Снаряд"
          disabled={globalReadonly}
          className="mb-4 w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-slate-100 disabled:opacity-50"
        >
          {EQUIPMENT.map((eq) => (
            <option key={eq} value={eq}>
              {equipmentLabel[eq] ?? eq}
            </option>
          ))}
        </select>

        <label className="mb-4 flex items-center gap-2 text-sm text-slate-300">
          <input
            type="checkbox"
            checked={perArm}
            onChange={(e) => setPerArm(e.target.checked)}
            disabled={globalReadonly}
            className="size-4"
          />
          Считать на одну руку/ногу
        </label>

        <label className="mb-1 block text-xs font-bold text-slate-500">Заметка по технике</label>
        <textarea
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          aria-label="Заметка по технике"
          disabled={globalReadonly}
          rows={3}
          className="mb-5 w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-slate-100 disabled:opacity-50"
        />

        {!globalReadonly && (
          <button
            type="button"
            disabled={!canSave || pending}
            onClick={save}
            className="w-full rounded-2xl bg-indigo-500 px-4 py-3.5 text-[15px] font-extrabold text-white disabled:opacity-50"
          >
            Сохранить упражнение
          </button>
        )}
        {err && (
          <p className="mt-2 text-sm text-rose-400">
            {err instanceof Error && err.message.includes('уже') ? 'Упражнение с таким именем уже есть.' : 'Не удалось сохранить.'}
          </p>
        )}
      </div>
    </>
  )
}
