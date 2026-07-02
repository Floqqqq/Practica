import { Icon } from "./Icon";

interface SearchHistoryProps {
  items: string[];
  onSelect: (query: string) => void;
  onClear: () => void;
}

/** Displays recent queries and exposes selection and clearing actions. */
export function SearchHistory({ items, onSelect, onClear }: SearchHistoryProps) {
  if (!items.length) return null;
  return (
    <div className="history" aria-label="История поиска">
      <div className="history__label"><Icon name="clock" size={15} />Недавние запросы</div>
      <div className="history__items">
        {items.map((item) => (
          <button key={item} type="button" onClick={() => onSelect(item)}>{item}</button>
        ))}
      </div>
      <button type="button" className="history__clear" onClick={onClear}>Очистить</button>
    </div>
  );
}
