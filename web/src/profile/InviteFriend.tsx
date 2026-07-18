import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { invitesApi } from '../api/invites'
import { humanError } from '../lib/errors'

export default function InviteFriend() {
  const [link, setLink] = useState('')
  const [copied, setCopied] = useState(false)

  const create = useMutation({
    mutationFn: () => invitesApi.create(14),
    onSuccess: (inv) => {
      setLink(`${window.location.origin}/register?code=${inv.code}`)
      setCopied(false)
    },
  })

  async function copy() {
    try {
      await navigator.clipboard.writeText(link)
      setCopied(true)
    } catch {
      setCopied(false)
    }
  }

  return (
    <section className="mt-6 rounded-2xl border border-slate-800 bg-slate-900 p-4">
      <h2 className="mb-1 font-bold text-slate-100">Пригласить</h2>
      <p className="mb-3 text-sm text-slate-400">
        Создайте ссылку регистрации для нового человека. Одноразовая, действует 14 дней. У него будет
        свой отдельный аккаунт.
      </p>

      {!link ? (
        <button
          type="button"
          disabled={create.isPending}
          onClick={() => create.mutate()}
          className="w-full rounded-xl bg-indigo-500 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
        >
          Создать ссылку-приглашение
        </button>
      ) : (
        <>
          <input
            readOnly
            value={link}
            aria-label="Ссылка приглашения"
            onFocus={(e) => e.target.select()}
            className="w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2.5 text-sm text-slate-200"
          />
          <div className="mt-2 flex gap-2">
            <button
              type="button"
              onClick={copy}
              className="flex-1 rounded-xl bg-indigo-500 px-4 py-2.5 text-sm font-bold text-white"
            >
              {copied ? 'Скопировано ✓' : 'Скопировать'}
            </button>
            <button
              type="button"
              onClick={() => setLink('')}
              className="rounded-xl border border-slate-700 px-4 py-2.5 text-sm font-semibold text-slate-300"
            >
              Ещё
            </button>
          </div>
        </>
      )}
      {create.isError && <p className="mt-2 text-sm text-rose-400">{humanError(create.error)}</p>}
    </section>
  )
}
