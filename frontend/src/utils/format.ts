/** Formats a byte count using Russian locale units. */
export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 Б";
  const units = ["Б", "КБ", "МБ", "ГБ"];
  const unit = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / 1024 ** unit;
  const formatted = new Intl.NumberFormat("ru-RU", {
    maximumFractionDigits: unit === 0 ? 0 : 1
  }).format(value);
  return formatted + " " + units[unit];
}

/** Formats an ISO date for the Russian interface. */
export function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}

/** Converts a normalized or percentage score to a bounded percentage label. */
export function formatScore(score: number): string {
  const normalized = score <= 1 ? score * 100 : score;
  return Math.max(0, Math.min(100, Math.round(normalized))) + "%";
}
