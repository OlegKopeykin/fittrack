import { type InputHTMLAttributes, type ReactNode } from 'react'

// Мелкие переиспользуемые примитивы форм в мобильном стиле.

export function Field({
  label,
  error,
  ...props
}: { label: string; error?: string } & InputHTMLAttributes<HTMLInputElement>) {
  return (
    <label className="flex flex-col gap-1 text-sm text-slate-300">
      {label}
      <input
        {...props}
        className="rounded-lg border border-slate-700 bg-slate-900 px-3 py-3 text-base text-slate-100 outline-none focus:border-sky-500"
      />
      {error && <span className="text-sm text-red-400">{error}</span>}
    </label>
  )
}

export function AuthCard({ title, children }: { title: string; children: ReactNode }) {
  return (
    <main className="flex min-h-dvh flex-col items-center justify-center bg-slate-950 px-6">
      <div className="w-full max-w-sm">
        <h1 className="mb-1 text-3xl font-bold tracking-tight text-slate-100">FitTrack</h1>
        <h2 className="mb-6 text-lg text-slate-400">{title}</h2>
        {children}
      </div>
    </main>
  )
}

export function SubmitButton({
  children,
  disabled,
}: {
  children: ReactNode
  disabled?: boolean
}) {
  return (
    <button
      type="submit"
      disabled={disabled}
      className="mt-2 rounded-lg bg-sky-600 px-4 py-3 text-base font-semibold text-white disabled:opacity-50"
    >
      {children}
    </button>
  )
}

export function FormError({ children }: { children: ReactNode }) {
  if (!children) return null
  return (
    <p role="alert" className="rounded-lg bg-red-950/60 px-3 py-2 text-sm text-red-300">
      {children}
    </p>
  )
}
