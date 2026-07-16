// Тонкий HTTP-клиент над fetch: JSON, куки-сессия, единый конверт ошибок.

export type ApiErrorBody = {
  code: string
  message: string
  fields?: Record<string, string>
}

export class ApiError extends Error {
  code: string
  status: number
  fields?: Record<string, string>

  constructor(status: number, body: ApiErrorBody) {
    super(body.message || body.code)
    this.name = 'ApiError'
    this.status = status
    this.code = body.code
    this.fields = body.fields
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    credentials: 'same-origin',
    headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  })

  if (res.status === 204) {
    return undefined as T
  }

  const text = await res.text()
  const data = text ? JSON.parse(text) : undefined

  if (!res.ok) {
    const err: ApiErrorBody = data?.error ?? { code: 'internal', message: 'Ошибка сервера' }
    throw new ApiError(res.status, err)
  }
  return data as T
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, body),
  del: <T>(path: string) => request<T>('DELETE', path),
}
