import { Icon } from "./Icon";

interface PaginationProps {
  page: number;
  totalPages: number;
  onChange: (page: number) => void;
}

function visiblePages(current: number, total: number): Array<number | "gap-left" | "gap-right"> {
  if (total <= 7) return Array.from({ length: total }, (_, index) => index + 1);
  const pages: Array<number | "gap-left" | "gap-right"> = [1];
  if (current > 3) pages.push("gap-left");
  for (let value = Math.max(2, current - 1); value <= Math.min(total - 1, current + 1); value += 1) {
    pages.push(value);
  }
  if (current < total - 2) pages.push("gap-right");
  pages.push(total);
  return pages;
}

/** Renders compact accessible pagination for search result pages. */
export function Pagination({ page, totalPages, onChange }: PaginationProps) {
  if (totalPages <= 1) return null;
  return (
    <nav className="pagination" aria-label="Страницы результатов">
      <button type="button" onClick={() => onChange(page - 1)} disabled={page <= 1} aria-label="Предыдущая страница">
        <Icon name="chevron-left" size={18} />
      </button>
      {visiblePages(page, totalPages).map((item) =>
        typeof item === "number" ? (
          <button
            type="button"
            key={item}
            className={item === page ? "is-current" : ""}
            onClick={() => onChange(item)}
            aria-current={item === page ? "page" : undefined}
            aria-label={"Страница " + item}
          >{item}</button>
        ) : <span key={item} aria-hidden="true">…</span>
      )}
      <button type="button" onClick={() => onChange(page + 1)} disabled={page >= totalPages} aria-label="Следующая страница">
        <Icon name="chevron-right" size={18} />
      </button>
    </nav>
  );
}
