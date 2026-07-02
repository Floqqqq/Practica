import type { FormEvent } from "react";
import { Icon } from "./Icon";

interface SearchBarProps {
  value: string;
  loading: boolean;
  onChange: (value: string) => void;
  onSubmit: () => void;
}

/** Renders the controlled search form and submits it by button or Enter. */
export function SearchBar({ value, loading, onChange, onSubmit }: SearchBarProps) {
  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit();
  };

  return (
    <form className="search-form" role="search" onSubmit={submit}>
      <Icon name="search" size={22} className="search-form__icon" />
      <label className="sr-only" htmlFor="knowledge-query">Поисковый запрос</label>
      <input
        id="knowledge-query"
        type="search"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder="Например, жизненный цикл программного продукта"
        autoComplete="off"
        enterKeyHint="search"
      />
      <button type="submit" className="primary-button" disabled={loading || !value.trim()}>
        {loading ? <span className="button-spinner" aria-hidden="true" /> : <Icon name="search" size={18} />}
        {loading ? "Ищем" : "Найти"}
      </button>
    </form>
  );
}
