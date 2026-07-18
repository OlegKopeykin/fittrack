import { useRef, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { backupApi } from '../api/backup'
import { humanError } from '../lib/errors'

export default function BackupRestore() {
  const qc = useQueryClient()
  const fileRef = useRef<HTMLInputElement>(null)
  const [msg, setMsg] = useState('')

  const imp = useMutation({
    mutationFn: (data: unknown) => backupApi.import(data),
    onSuccess: (r) => {
      setMsg(`Восстановлено тренировок: ${r.imported}${r.skipped ? `, пропущено: ${r.skipped}` : ''}`)
      qc.invalidateQueries({ queryKey: ['workouts'] })
    },
    onError: (e) => setMsg(humanError(e)),
  })

  async function onFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    setMsg('')
    let data: unknown
    try {
      data = JSON.parse(await file.text())
    } catch {
      setMsg('Не удалось прочитать файл: это не JSON.')
      return
    }
    imp.mutate(data)
  }

  return (
    <section className="mt-6 rounded-2xl border border-slate-800 bg-slate-900 p-4">
      <h2 className="mb-1 font-bold text-slate-100">Резервная копия</h2>
      <p className="mb-3 text-sm text-slate-400">
        Скачайте бэкап своего лога или восстановите тренировки из файла. Экспорт — только по вашим
        тренировкам.
      </p>
      <a
        href={backupApi.exportUrl}
        download
        className="block w-full rounded-xl border border-slate-700 px-4 py-2.5 text-center text-sm font-semibold text-slate-200"
      >
        Скачать бэкап (JSON)
      </a>
      <button
        type="button"
        disabled={imp.isPending}
        onClick={() => fileRef.current?.click()}
        className="mt-2 w-full rounded-xl border border-slate-700 px-4 py-2.5 text-sm font-semibold text-slate-200 disabled:opacity-50"
      >
        Восстановить из файла
      </button>
      <input
        ref={fileRef}
        type="file"
        accept="application/json,.json"
        onChange={onFile}
        aria-label="Файл бэкапа"
        className="hidden"
      />
      {msg && <p className="mt-2 text-sm text-slate-300">{msg}</p>}
    </section>
  )
}
