import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useRegister } from '../auth/useAuth'
import { AuthCard, Field, FormError, SubmitButton } from '../components/ui'
import { humanError } from '../lib/errors'

export default function RegisterPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const register = useRegister()

  const [code, setCode] = useState(params.get('code') ?? '')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [localError, setLocalError] = useState<string | null>(null)

  function submit(e: FormEvent) {
    e.preventDefault()
    if (password.length < 8) {
      setLocalError('Пароль должен быть минимум 8 символов')
      return
    }
    setLocalError(null)
    register.mutate(
      { invite_code: code, username, password },
      { onSuccess: () => navigate('/', { replace: true }) },
    )
  }

  const error = localError ?? (register.isError ? humanError(register.error) : null)

  return (
    <AuthCard title="Регистрация">
      <form onSubmit={submit} className="flex flex-col gap-4">
        {error && <FormError>{error}</FormError>}
        <Field
          label="Код инвайта"
          value={code}
          onChange={(e) => setCode(e.target.value)}
        />
        <Field
          label="Логин"
          autoComplete="username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
        <Field
          label="Пароль"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <SubmitButton disabled={register.isPending}>
          {register.isPending ? 'Создаём…' : 'Зарегистрироваться'}
        </SubmitButton>
      </form>
      <p className="mt-6 text-center text-sm text-slate-400">
        Уже есть аккаунт?{' '}
        <Link to="/login" className="text-sky-400">
          Вход
        </Link>
      </p>
    </AuthCard>
  )
}
