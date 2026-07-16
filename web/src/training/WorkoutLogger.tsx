import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Workout, WorkoutSet, Prescription, ExerciseSession } from '../api/training'
import type { Exercise } from '../api/exercises'
import {
  useProgramDay,
  useExerciseMap,
  useExerciseHistory,
  useAddSet,
  useUpdateSet,
  useDeleteSet,
  useFinishWorkout,
} from './useTraining'
import { exercisesApi } from '../api/exercises'
import { useQuery } from '@tanstack/react-query'
import { PageHeader } from '../components/AppShell'
import { formatSet } from '../lib/format'

function uid(): string {
  const c = globalThis.crypto
  if (c && typeof c.randomUUID === 'function') return c.randomUUID()
  return `cid-${Date.now()}-${Math.floor(Math.random() * 1e9)}`
}

function elapsed(fromISO?: string): string {
  if (!fromISO) return '0:00'
  const start = new Date(fromISO).getTime()
  const sec = Math.max(0, Math.floor((Date.now() - start) / 1000))
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  const mm = h ? String(m).padStart(2, '0') : String(m)
  const base = `${mm}:${String(s).padStart(2, '0')}`
  return h ? `${h}:${base}` : base
}

function useTick(active: boolean) {
  const [, setN] = useState(0)
  useEffect(() => {
    if (!active) return
    const id = setInterval(() => setN((n) => n + 1), 1000)
    return () => clearInterval(id)
  }, [active])
}

export default function WorkoutLogger({ workout }: { workout: Workout }) {
  const navigate = useNavigate()
  const programDay = useProgramDay(workout.program_day_id ?? undefined)
  const exMap = useExerciseMap()
  const finish = useFinishWorkout(workout.id)
  const [extra, setExtra] = useState<number[]>([])
  const [picking, setPicking] = useState(false)
  const [finishing, setFinishing] = useState(false)
  const running = !workout.finished_at
  useTick(running)

  const sets = useMemo(() => workout.sets ?? [], [workout.sets])

  // Порядок упражнений: предписания дня → упражнения с записанными подходами → добавленные вручную.
  const exerciseIds = useMemo(() => {
    const order: number[] = []
    const seen = new Set<number>()
    const push = (id: number) => {
      if (!seen.has(id)) {
        seen.add(id)
        order.push(id)
      }
    }
    programDay.data?.exercises.forEach((p) => push(p.exercise_id))
    sets.forEach((s) => push(s.exercise_id))
    extra.forEach(push)
    return order
  }, [programDay.data, sets, extra])

  const rxByExercise = useMemo(() => {
    const m = new Map<number, Prescription>()
    programDay.data?.exercises.forEach((p) => m.set(p.exercise_id, p))
    return m
  }, [programDay.data])

  return (
    <>
      <PageHeader
        title={workout.title || 'Тренировка'}
        right={
          <div className="flex items-center gap-2">
            {running && (
              <span className="rounded-lg border border-teal-500/30 bg-teal-500/10 px-2 py-1 text-sm font-bold tabular-nums text-teal-300">
                {elapsed(workout.started_at)}
              </span>
            )}
            {running ? (
              <button
                type="button"
                onClick={() => setFinishing(true)}
                className="rounded-lg bg-emerald-500 px-3 py-1.5 text-sm font-bold text-emerald-950"
              >
                Завершить
              </button>
            ) : (
              <button
                type="button"
                onClick={() => navigate('/workouts')}
                className="rounded-lg border border-slate-700 px-3 py-1.5 text-sm font-bold text-slate-300"
              >
                Готово
              </button>
            )}
          </div>
        }
      />
      <div className="mx-auto max-w-3xl px-4 py-4">
        {exerciseIds.length === 0 && (
          <p className="mb-4 text-sm text-slate-500">
            Пустая тренировка. Добавьте упражнение, чтобы начать.
          </p>
        )}
        <div className="flex flex-col gap-3">
          {exerciseIds.map((exId) => (
            <ExerciseBlock
              key={exId}
              workoutId={workout.id}
              exercise={exMap.data?.get(exId)}
              exerciseId={exId}
              prescription={rxByExercise.get(exId)}
              loggedSets={sets.filter((s) => s.exercise_id === exId)}
            />
          ))}
        </div>

        {picking ? (
          <ExercisePicker
            onPick={(id) => {
              setExtra((e) => (e.includes(id) ? e : [...e, id]))
              setPicking(false)
            }}
            onClose={() => setPicking(false)}
          />
        ) : (
          <button
            type="button"
            onClick={() => setPicking(true)}
            className="mt-3 w-full rounded-2xl border border-indigo-500/30 bg-indigo-500/10 px-4 py-3 text-sm font-bold text-indigo-300"
          >
            + Добавить упражнение
          </button>
        )}
      </div>

      {finishing && (
        <FinishSheet
          workout={workout}
          setCount={sets.length}
          exerciseCount={exerciseIds.length}
          onClose={() => setFinishing(false)}
          onFinish={(body) =>
            finish.mutate(
              { finished_at: new Date().toISOString(), ...body },
              { onSuccess: () => navigate('/workouts') },
            )
          }
          pending={finish.isPending}
        />
      )}
    </>
  )
}

// --- блок одного упражнения ---

function ExerciseBlock({
  workoutId,
  exerciseId,
  exercise,
  prescription,
  loggedSets,
}: {
  workoutId: number
  exerciseId: number
  exercise?: Exercise
  prescription?: Prescription
  loggedSets: WorkoutSet[]
}) {
  const history = useExerciseHistory(exerciseId)
  const addSet = useAddSet(workoutId)
  const del = useDeleteSet(workoutId)
  const kind = exercise?.kind
  const cardio = kind === 'cardio'
  const iso = kind === 'isometric'

  const prev = history.data?.[0]?.sets ?? []
  const working = loggedSets.filter((s) => s.role === 'working').length

  return (
    <div className="rounded-2xl border border-slate-800 bg-slate-900 p-3">
      <div className="mb-2 font-bold text-indigo-300">
        {exercise?.name ?? `Упражнение #${exerciseId}`}
        {prescription && (
          <span className="ml-2 text-xs font-semibold text-slate-500">
            цель {prescription.sets}
            {prescription.rep_max ? ` × ${prescription.rep_min ?? ''}${prescription.rep_min ? '–' : ''}${prescription.rep_max}` : ''}
          </span>
        )}
      </div>

      <table className="w-full tabular-nums">
        <thead>
          <tr className="text-[10px] uppercase tracking-wide text-slate-500">
            <th className="w-8 pb-1 font-bold">№</th>
            <th className="w-20 pb-1 font-bold">Прошлый</th>
            <th className="pb-1 font-bold">{cardio ? 'Км' : iso ? '' : 'Кг'}</th>
            <th className="pb-1 font-bold">{cardio || iso ? 'Сек' : 'Повт'}</th>
            <th className="w-10 pb-1 font-bold">✓</th>
          </tr>
        </thead>
        <tbody>
          {loggedSets.map((s, i) => (
            <LoggedRow
              key={s.id}
              set={s}
              index={s.role === 'working' ? workingIndex(loggedSets, i) : 0}
              prev={prev[i]}
              cardio={cardio}
              iso={iso}
              workoutId={workoutId}
              onDelete={() => del.mutate(s.id)}
            />
          ))}
          <ComposeRow
            nextNo={working + 1}
            prev={prev[loggedSets.length]}
            cardio={cardio}
            iso={iso}
            prescription={prescription}
            pending={addSet.isPending}
            onLog={(body) => addSet.mutate({ exercise_id: exerciseId, ...body })}
          />
        </tbody>
      </table>
    </div>
  )
}

function workingIndex(sets: WorkoutSet[], upto: number): number {
  let n = 0
  for (let i = 0; i <= upto; i++) if (sets[i].role === 'working') n++
  return n
}

// --- залогированный подход (редактируемый) ---

function LoggedRow({
  set,
  index,
  prev,
  cardio,
  iso,
  workoutId,
  onDelete,
}: {
  set: WorkoutSet
  index: number
  prev?: ExerciseSession['sets'][number]
  cardio: boolean
  iso: boolean
  workoutId: number
  onDelete: () => void
}) {
  const update = useUpdateSet(workoutId)
  const a = cardio ? set.distance_km : set.weight_kg
  const b = cardio || iso ? set.duration_sec : set.reps
  const [av, setAv] = useState(a?.toString() ?? '')
  const [bv, setBv] = useState(b?.toString() ?? '')

  function commit(field: 'a' | 'b', raw: string) {
    const num = raw === '' ? undefined : Number(raw)
    if (num !== undefined && Number.isNaN(num)) return
    const key = cardio
      ? field === 'a'
        ? 'distance_km'
        : 'duration_sec'
      : iso
        ? 'duration_sec'
        : field === 'a'
          ? 'weight_kg'
          : 'reps'
    update.mutate({ setId: set.id, body: { [key]: num } })
  }

  return (
    <tr className="[&>td]:py-1">
      <td className="text-center">
        <span
          className={
            'mx-auto flex h-6 w-6 items-center justify-center rounded-lg text-xs font-extrabold ' +
            (set.role === 'warmup'
              ? 'bg-amber-400/15 text-amber-300'
              : 'bg-emerald-400/15 text-emerald-300')
          }
        >
          {set.role === 'warmup' ? 'Р' : index}
        </span>
      </td>
      <td className="text-center text-[13px] font-semibold text-slate-600">
        {prev ? formatSet(prev as WorkoutSet) : '—'}
      </td>
      {!iso && (
        <td className="px-1">
          <input
            inputMode="decimal"
            value={av}
            onChange={(e) => setAv(e.target.value)}
            onBlur={(e) => commit('a', e.target.value)}
            className="w-full rounded-lg border border-emerald-500/30 bg-emerald-500/5 px-1 py-1.5 text-center text-[15px] font-bold text-emerald-100"
          />
        </td>
      )}
      <td className="px-1" colSpan={iso ? 2 : 1}>
        <input
          inputMode="decimal"
          value={bv}
          onChange={(e) => setBv(e.target.value)}
          onBlur={(e) => commit('b', e.target.value)}
          className="w-full rounded-lg border border-emerald-500/30 bg-emerald-500/5 px-1 py-1.5 text-center text-[15px] font-bold text-emerald-100"
        />
      </td>
      <td className="text-center">
        <button
          type="button"
          aria-label="Снять отметку"
          onClick={onDelete}
          className="mx-auto flex h-7 w-7 items-center justify-center rounded-lg border border-emerald-500 bg-emerald-500 text-sm font-extrabold text-emerald-950"
        >
          ✓
        </button>
      </td>
    </tr>
  )
}

// --- строка ввода нового подхода ---

function ComposeRow({
  nextNo,
  prev,
  cardio,
  iso,
  prescription,
  pending,
  onLog,
}: {
  nextNo: number
  prev?: ExerciseSession['sets'][number]
  cardio: boolean
  iso: boolean
  prescription?: Prescription
  pending: boolean
  onLog: (body: Partial<import('../api/training').NewSet>) => void
}) {
  const targetA = cardio ? undefined : prescription?.weight_max_kg
  const targetB = cardio || iso ? undefined : prescription?.rep_max
  const [av, setAv] = useState('')
  const [bv, setBv] = useState('')
  const clientId = useRef(uid())

  function fill() {
    if (cardio) {
      if (prev?.distance_km != null) setAv(String(prev.distance_km))
      if (prev?.duration_sec != null) setBv(String(prev.duration_sec))
    } else if (iso) {
      if (prev?.duration_sec != null) setBv(String(prev.duration_sec))
    } else {
      if (prev?.weight_kg != null) setAv(String(prev.weight_kg))
      if (prev?.reps != null) setBv(String(prev.reps))
    }
  }

  function bump(setter: (v: string) => void, cur: string, by: number) {
    const n = cur === '' ? 0 : Number(cur)
    if (Number.isNaN(n)) return
    setter(String(Math.max(0, Math.round((n + by) * 100) / 100)))
  }

  function log() {
    const a = av === '' ? undefined : Number(av)
    const b = bv === '' ? undefined : Number(bv)
    if (a === undefined && b === undefined) return
    const body: Partial<import('../api/training').NewSet> = {
      role: 'working',
      client_id: clientId.current,
    }
    if (cardio) {
      body.distance_km = a
      body.duration_sec = b
    } else if (iso) {
      body.duration_sec = b
    } else {
      body.weight_kg = a
      body.reps = b
    }
    onLog(body)
    setAv('')
    setBv('')
    clientId.current = uid()
  }

  const placeholderA = targetA != null ? String(targetA) : cardio ? 'км' : 'кг'
  const placeholderB = targetB != null ? String(targetB) : cardio || iso ? 'сек' : 'повт'

  return (
    <tr className="[&>td]:py-1">
      <td className="text-center">
        <span className="mx-auto flex h-6 w-6 items-center justify-center rounded-lg bg-slate-800 text-xs font-extrabold text-slate-400">
          {nextNo}
        </span>
      </td>
      <td className="text-center">
        {prev ? (
          <button
            type="button"
            onClick={fill}
            className="text-[13px] font-semibold text-slate-400 underline decoration-dotted"
          >
            {formatSet(prev as WorkoutSet)}
          </button>
        ) : (
          <span className="text-[13px] text-slate-600">—</span>
        )}
      </td>
      {!iso && (
        <td className="px-1">
          <div className="flex items-center gap-1">
            <button type="button" aria-label="минус" onClick={() => bump(setAv, av, cardio ? -0.5 : -2.5)} className="h-7 w-6 rounded-md bg-slate-800 text-slate-400">−</button>
            <input
              inputMode="decimal"
              value={av}
              placeholder={placeholderA}
              onChange={(e) => setAv(e.target.value)}
              aria-label="вес"
              className="w-full min-w-0 rounded-lg border border-slate-700 bg-slate-950 px-1 py-1.5 text-center text-[15px] font-bold text-slate-100"
            />
            <button type="button" aria-label="плюс" onClick={() => bump(setAv, av, cardio ? 0.5 : 2.5)} className="h-7 w-6 rounded-md bg-slate-800 text-slate-400">+</button>
          </div>
        </td>
      )}
      <td className="px-1" colSpan={iso ? 2 : 1}>
        <div className="flex items-center gap-1">
          <button type="button" aria-label="минус повт" onClick={() => bump(setBv, bv, -1)} className="h-7 w-6 rounded-md bg-slate-800 text-slate-400">−</button>
          <input
            inputMode="decimal"
            value={bv}
            placeholder={placeholderB}
            onChange={(e) => setBv(e.target.value)}
            aria-label="повторы"
            className="w-full min-w-0 rounded-lg border border-slate-700 bg-slate-950 px-1 py-1.5 text-center text-[15px] font-bold text-slate-100"
          />
          <button type="button" aria-label="плюс повт" onClick={() => bump(setBv, bv, 1)} className="h-7 w-6 rounded-md bg-slate-800 text-slate-400">+</button>
        </div>
      </td>
      <td className="text-center">
        <button
          type="button"
          aria-label="Записать подход"
          disabled={pending}
          onClick={log}
          className="mx-auto flex h-7 w-7 items-center justify-center rounded-lg border border-slate-700 bg-slate-950 text-sm font-extrabold text-slate-500"
        >
          ✓
        </button>
      </td>
    </tr>
  )
}

// --- выбор упражнения ---

function ExercisePicker({
  onPick,
  onClose,
}: {
  onPick: (id: number) => void
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
              onClick={() => onPick(e.id)}
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

// --- лист завершения ---

function FinishSheet({
  workout,
  setCount,
  exerciseCount,
  onClose,
  onFinish,
  pending,
}: {
  workout: Workout
  setCount: number
  exerciseCount: number
  onClose: () => void
  onFinish: (body: { bodyweight_kg?: number; feeling?: string }) => void
  pending: boolean
}) {
  const [bw, setBw] = useState(workout.bodyweight_kg?.toString() ?? '')
  const [feeling, setFeeling] = useState(workout.feeling ?? '')
  const feelings = ['тяжело', 'бодро', 'лёгкость']
  return (
    <div className="fixed inset-0 z-20 flex items-end bg-black/50" onClick={onClose}>
      <div
        className="w-full rounded-t-3xl border-t border-slate-800 bg-slate-900 p-5"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mx-auto mb-4 h-1 w-10 rounded-full bg-slate-700" />
        <h3 className="text-lg font-extrabold text-slate-100">Завершить тренировку</h3>
        <div className="mb-4 mt-3 flex gap-2">
          <div className="flex-1 rounded-xl border border-slate-800 bg-slate-950 py-2 text-center">
            <div className="text-lg font-extrabold tabular-nums">{exerciseCount}</div>
            <div className="text-[10px] font-bold uppercase tracking-wide text-slate-500">упражн.</div>
          </div>
          <div className="flex-1 rounded-xl border border-slate-800 bg-slate-950 py-2 text-center">
            <div className="text-lg font-extrabold tabular-nums">{setCount}</div>
            <div className="text-[10px] font-bold uppercase tracking-wide text-slate-500">подходов</div>
          </div>
        </div>
        <label className="mb-1 block text-xs font-bold text-slate-500">Вес тела</label>
        <div className="mb-3 flex items-center gap-2">
          <input
            inputMode="decimal"
            value={bw}
            onChange={(e) => setBw(e.target.value)}
            aria-label="Вес тела"
            className="flex-1 rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-center text-lg font-extrabold tabular-nums text-slate-100"
          />
          <span className="font-bold text-slate-500">кг</span>
        </div>
        <label className="mb-1 block text-xs font-bold text-slate-500">Самочувствие</label>
        <div className="mb-4 flex gap-2">
          {feelings.map((f) => (
            <button
              key={f}
              type="button"
              onClick={() => setFeeling(f)}
              className={
                'flex-1 rounded-xl border px-2 py-2 text-sm font-bold ' +
                (feeling === f
                  ? 'border-indigo-500/50 bg-indigo-500/15 text-indigo-300'
                  : 'border-slate-700 bg-slate-950 text-slate-400')
              }
            >
              {f}
            </button>
          ))}
        </div>
        <button
          type="button"
          disabled={pending}
          onClick={() =>
            onFinish({
              bodyweight_kg: bw === '' ? undefined : Number(bw),
              feeling: feeling || undefined,
            })
          }
          className="w-full rounded-xl bg-emerald-500 py-3.5 text-[15px] font-extrabold text-emerald-950 disabled:opacity-60"
        >
          Сохранить и завершить
        </button>
      </div>
    </div>
  )
}
