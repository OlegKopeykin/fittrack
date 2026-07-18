import { useState } from 'react'
import {
  useTelegram,
  useSetTelegram,
  useLinkTelegram,
  useTestTelegram,
  useRemoveTelegram,
} from './useTelegram'
import type { TelegramFrequency } from '../api/telegram'
import { humanError } from '../lib/errors'

const FREQ: { key: TelegramFrequency; label: string }[] = [
  { key: 'daily', label: 'Каждый день' },
  { key: 'weekly', label: 'Раз в неделю' },
  { key: 'monthly', label: 'Раз в месяц' },
]

export default function TelegramExport() {
  const tg = useTelegram()
  const set = useSetTelegram()
  const link = useLinkTelegram()
  const test = useTestTelegram()
  const remove = useRemoveTelegram()

  const [token, setToken] = useState('')
  const [testMsg, setTestMsg] = useState('')

  const st = tg.data

  return (
    <section className="mt-6 rounded-2xl border border-slate-800 bg-slate-900 p-4">
      <div className="mb-1 flex items-center gap-2">
        <span className="flex size-7 items-center justify-center rounded-lg bg-sky-500 text-sm font-bold text-white">
          ✈
        </span>
        <h2 className="font-bold text-slate-100">Экспорт в Telegram</h2>
      </div>
      <p className="mb-3 text-sm text-slate-400">
        Бот присылает бэкап тренировок, заметок и веса в ваш чат файлом JSON — по расписанию.
      </p>

      {tg.isLoading && <p className="text-sm text-slate-500">Загрузка…</p>}

      {/* Шаг 1 — токен */}
      {st && !st.configured && (
        <>
          <label className="mb-1 block text-xs font-bold uppercase tracking-wide text-slate-500">
            Токен бота
          </label>
          <input
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="123456789:AA…"
            aria-label="Токен бота"
            className="w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 font-mono text-sm text-slate-100"
          />
          <button
            type="button"
            disabled={token.trim() === '' || set.isPending}
            onClick={() => set.mutate({ bot_token: token.trim() })}
            className="mt-2 w-full rounded-xl bg-sky-500 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
          >
            Подключить
          </button>
          {set.isError && <p className="mt-2 text-sm text-rose-400">{humanError(set.error)}</p>}
          <ol className="mt-3 space-y-1.5 text-xs text-slate-400">
            <li>1. Создайте бота у <b className="text-slate-200">@BotFather</b> → команда <code className="text-slate-300">/newbot</code>.</li>
            <li>2. Скопируйте выданный токен и вставьте сюда.</li>
            <li>3. Откройте своего бота в Telegram и нажмите <b className="text-slate-200">Start</b>.</li>
          </ol>
        </>
      )}

      {/* Шаг 2 — связать чат */}
      {st && st.configured && !st.chat_linked && (
        <>
          <p className="mb-2 text-sm text-slate-300">
            Бот <b className="text-sky-300">@{st.bot_username}</b> подключён. Откройте его в Telegram и
            нажмите <b>Start</b>, затем проверьте связь.
          </p>
          <button
            type="button"
            disabled={link.isPending}
            onClick={() => link.mutate()}
            className="w-full rounded-xl bg-indigo-500 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
          >
            Проверить связь
          </button>
          {link.isError && (
            <p className="mt-2 text-sm text-rose-400">Сообщений боту нет — нажмите Start и повторите.</p>
          )}
          <button
            type="button"
            onClick={() => remove.mutate()}
            className="mt-2 w-full rounded-xl border border-slate-700 px-4 py-2 text-sm text-slate-400"
          >
            Сменить токен
          </button>
        </>
      )}

      {/* Шаг 3 — подключено */}
      {st && st.configured && st.chat_linked && (
        <>
          <div className="mb-3 flex items-center gap-2 rounded-xl border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm font-semibold text-emerald-300">
            ✓ Подключено · @{st.bot_username}
          </div>

          <div className="flex items-center justify-between py-2">
            <div>
              <div className="text-sm font-bold text-slate-100">Автовыгрузка</div>
              <div className="text-xs text-slate-500">Присылать бэкап по расписанию</div>
            </div>
            <button
              type="button"
              role="switch"
              aria-checked={st.enabled}
              aria-label="Автовыгрузка"
              onClick={() => set.mutate({ enabled: !st.enabled })}
              className={`relative h-6 w-11 rounded-full transition ${st.enabled ? 'bg-emerald-500' : 'bg-slate-700'}`}
            >
              <span
                className={`absolute top-0.5 size-5 rounded-full bg-white transition-all ${st.enabled ? 'right-0.5' : 'left-0.5'}`}
              />
            </button>
          </div>

          <div className="mb-1 mt-2 text-xs font-bold uppercase tracking-wide text-slate-500">
            Частота (ночью в 00:00)
          </div>
          <div className="flex gap-1 rounded-xl border border-slate-800 bg-slate-950 p-1">
            {FREQ.map((f) => (
              <button
                key={f.key}
                type="button"
                onClick={() => set.mutate({ frequency: f.key })}
                className={`flex-1 rounded-lg px-2 py-2 text-xs font-bold ${
                  st.frequency === f.key ? 'bg-indigo-500 text-white' : 'text-slate-400'
                }`}
              >
                {f.label}
              </button>
            ))}
          </div>

          <button
            type="button"
            disabled={test.isPending}
            onClick={() =>
              test.mutate(undefined, {
                onSuccess: () => setTestMsg('Отправлено ✓'),
                onError: (e) => setTestMsg(humanError(e)),
              })
            }
            className="mt-3 w-full rounded-xl bg-indigo-500 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
          >
            Отправить сейчас (проверка)
          </button>
          {testMsg && <p className="mt-2 text-sm text-slate-300">{testMsg}</p>}

          <button
            type="button"
            onClick={() => remove.mutate()}
            className="mt-2 w-full rounded-xl border border-slate-700 px-4 py-2 text-sm text-slate-400"
          >
            Отключить и удалить токен
          </button>
        </>
      )}
    </section>
  )
}
