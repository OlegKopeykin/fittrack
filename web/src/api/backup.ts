import { api } from './client'

export type ImportResult = { imported: number; skipped: number }

export const backupApi = {
  // Прямая ссылка для скачивания JSON-бэкапа лога.
  exportUrl: '/api/v1/profile/export',
  import: (data: unknown) => api.post<ImportResult>('/api/v1/profile/import', data),
}
