import { ApiError } from '../api/client'

// humanError переводит ошибку API/сети в понятное русское сообщение.
export function humanError(e: unknown): string {
  if (e instanceof ApiError) {
    switch (e.code) {
      case 'invalid_credentials':
        return 'Неверный логин или пароль'
      case 'invalid_invite':
        return 'Инвайт недействителен или уже использован'
      case 'rate_limited':
        return 'Слишком много попыток. Попробуйте позже'
      case 'conflict':
        return 'Такой логин уже занят'
      case 'invalid_input':
        return 'Проверьте правильность полей'
      default:
        return e.message || 'Ошибка сервера'
    }
  }
  return 'Нет связи с сервером'
}
