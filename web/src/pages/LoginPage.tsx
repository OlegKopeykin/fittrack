import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useLogin } from '../auth/useAuth'
import { AuthCard, Field, FormError, SubmitButton } from '../components/ui'
import { humanError } from '../lib/errors'

export default function LoginPage() {
  const navigate = useNavigate()
  const login = useLogin()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  function submit(e: FormEvent) {
    e.preventDefault()
    login.mutate(
      { username, password },
      { onSuccess: () => navigate('/', { replace: true }) },
    )
  }

  return (
    <AuthCard title="Вход">
      <form onSubmit={submit} className="flex flex-col gap-4">
        {login.isError && <FormError>{humanError(login.error)}</FormError>}
        <Field
          label="Логин"
          autoComplete="username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
        <Field
          label="Пароль"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <SubmitButton disabled={login.isPending}>
          {login.isPending ? 'Входим…' : 'Войти'}
        </SubmitButton>
      </form>
      <p className="mt-6 text-center text-sm text-slate-400">
        Нет аккаунта?{' '}
        <Link to="/register" className="text-sky-400">
          Регистрация
        </Link>
      </p>
    </AuthCard>
  )
}
