import type { ReactNode } from "react";
import type { SearchResult } from "../types";
import { formatScore } from "../utils/format";
import { Icon } from "./Icon";
import { Pagination } from "./Pagination";

interface SearchResultsProps {
  query: string;
  items: SearchResult[];
  total: number;
  page: number;
  totalPages: number;
  loading: boolean;
  hasSearched: boolean;
  error: string | null;
  onPageChange: (page: number) => void;
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^$(){}|[\]\\]/g, "\\$&");
}

function HighlightedText({ text, query }: { text: string; query: string }) {
  const words = Array.from(new Set(query.trim().split(/\s+/).filter((word) => word.length > 1)));
  if (!words.length) return <>{text}</>;
  const expression = new RegExp("(" + words.map(escapeRegExp).join("|") + ")", "giu");
  const parts = text.split(expression);
  const exactWords = new Set(words.map((word) => word.toLocaleLowerCase("ru")));
  return <>{parts.map((part, index) => exactWords.has(part.toLocaleLowerCase("ru"))
    ? <mark key={index}>{part}</mark>
    : <span key={index}>{part}</span>)}</>;
}

function LoadingCards() {
  return (
    <div className="result-list" aria-label="Поиск выполняется" aria-busy="true">
      {[1, 2, 3].map((item) => (
        <div className="result-card result-card--skeleton" key={item}>
          <span /><span /><span />
        </div>
      ))}
    </div>
  );
}

function ResultsState({ icon, title, children }: { icon: "search" | "alert" | "spark"; title: string; children: ReactNode }) {
  return (
    <div className="results-state">
      <span className="results-state__icon"><Icon name={icon} size={29} /></span>
      <h3>{title}</h3>
      <p>{children}</p>
    </div>
  );
}

/** Displays search states, highlighted result cards and pagination. */
export function SearchResults(props: SearchResultsProps) {
  const { query, items, total, page, totalPages, loading, hasSearched, error, onPageChange } = props;
  if (loading) return <LoadingCards />;
  if (error) return <ResultsState icon="alert" title="Не удалось выполнить поиск">{error}</ResultsState>;
  if (!hasSearched) {
    return <ResultsState icon="spark" title="Знания — ближе, чем кажется">Введите запрос, и система найдёт подходящие фрагменты во всех загруженных документах.</ResultsState>;
  }
  if (!items.length) {
    return <ResultsState icon="search" title="Ничего не найдено">По вашему запросу ничего не найдено. Попробуйте изменить формулировку</ResultsState>;
  }

  return (
    <section aria-labelledby="results-title">
      <div className="results-summary">
        <h2 id="results-title">Результаты</h2>
        <span>{total} {total === 1 ? "фрагмент" : "фрагментов"}</span>
      </div>
      <div className="result-list">
        {items.map((item, index) => (
          <article className="result-card" key={item.chunkId + "-" + index}>
            <div className="result-card__meta">
              <div className="result-file">
                <span><Icon name="file" size={18} /></span>
                <strong>{item.fileName}</strong>
              </div>
              <span className="page-chip">Страница {item.page}</span>
            </div>
            <p className="result-card__text"><HighlightedText text={item.text} query={query} /></p>
            <div className="relevance">
              <span>Релевантность</span>
              <div className="relevance__track"><span style={{ width: formatScore(item.score) }} /></div>
              <strong>{formatScore(item.score)}</strong>
            </div>
          </article>
        ))}
      </div>
      <Pagination page={page} totalPages={totalPages} onChange={onPageChange} />
    </section>
  );
}
