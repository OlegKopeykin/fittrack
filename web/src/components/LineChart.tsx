type Pt = { date: string; kg: number }

const W = 600
const H = 200
const PAD = { l: 40, r: 14, t: 14, b: 26 }

const shortDate = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'short' })
function fmtDate(iso: string): string {
  const [y, m, d] = iso.split('-').map(Number)
  return shortDate.format(new Date(y, m - 1, d))
}

// LineChart — простой SVG-график тренда (слейт/индиго), без внешних зависимостей.
export default function LineChart({ points }: { points: Pt[] }) {
  if (points.length === 0) {
    return <p className="text-sm text-slate-500">Нет данных для графика.</p>
  }
  const values = points.map((p) => p.kg)
  const min = Math.min(...values)
  const max = Math.max(...values)
  const yMin = Math.floor(min - 1)
  const yMax = Math.ceil(max + 1)
  const span = yMax - yMin || 1

  const innerW = W - PAD.l - PAD.r
  const innerH = H - PAD.t - PAD.b
  const x = (i: number) => PAD.l + (points.length === 1 ? innerW / 2 : (i * innerW) / (points.length - 1))
  const y = (kg: number) => PAD.t + (1 - (kg - yMin) / span) * innerH

  const line = points.map((p, i) => `${x(i)},${y(p.kg)}`).join(' ')
  const area = `M${x(0)},${y(points[0].kg)} L${line.replaceAll(' ', ' L')} L${x(points.length - 1)},${H - PAD.b} L${x(0)},${H - PAD.b} Z`
  const last = points[points.length - 1]

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="w-full" role="img" aria-label="График веса тела">
      <defs>
        <linearGradient id="bw-grad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#6366f1" stopOpacity="0.3" />
          <stop offset="1" stopColor="#6366f1" stopOpacity="0" />
        </linearGradient>
      </defs>
      {/* сетка: верх/низ */}
      <line x1={PAD.l} y1={PAD.t} x2={W - PAD.r} y2={PAD.t} stroke="#1e293b" />
      <line x1={PAD.l} y1={H - PAD.b} x2={W - PAD.r} y2={H - PAD.b} stroke="#1e293b" />
      <text x={PAD.l - 6} y={PAD.t + 4} textAnchor="end" fontSize="11" fill="#64748b">{yMax}</text>
      <text x={PAD.l - 6} y={H - PAD.b + 4} textAnchor="end" fontSize="11" fill="#64748b">{yMin}</text>
      {/* даты по краям */}
      <text x={PAD.l} y={H - 6} textAnchor="start" fontSize="11" fill="#64748b">{fmtDate(points[0].date)}</text>
      {points.length > 1 && (
        <text x={W - PAD.r} y={H - 6} textAnchor="end" fontSize="11" fill="#64748b">{fmtDate(last.date)}</text>
      )}
      <path d={area} fill="url(#bw-grad)" />
      <polyline points={line} fill="none" stroke="#818cf8" strokeWidth="2.5" strokeLinejoin="round" />
      {points.map((p, i) => (
        <circle key={i} cx={x(i)} cy={y(p.kg)} r={i === points.length - 1 ? 4 : 2.5} fill="#818cf8" />
      ))}
    </svg>
  )
}
