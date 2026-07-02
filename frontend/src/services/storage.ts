import type { DocumentItem } from "../types";

const DOCUMENTS_KEY = "knowledge-search:documents";
const HISTORY_KEY = "knowledge-search:history";

function parseArray<T>(value: string | null): T[] {
  if (!value) return [];
  try {
    const parsed = JSON.parse(value) as unknown;
    return Array.isArray(parsed) ? (parsed as T[]) : [];
  } catch {
    return [];
  }
}

/** Returns the persisted client-side document registry. */
export function loadDocuments(): DocumentItem[] {
  return parseArray<DocumentItem>(localStorage.getItem(DOCUMENTS_KEY));
}

/** Persists completed upload entries for the document list. */
export function saveDocuments(documents: DocumentItem[]): void {
  const completed = documents.filter((item) =>
    ["ready", "duplicate", "error"].includes(item.status)
  );
  localStorage.setItem(DOCUMENTS_KEY, JSON.stringify(completed.slice(0, 50)));
}

/** Returns up to eight recent search queries. */
export function loadSearchHistory(): string[] {
  return parseArray<string>(localStorage.getItem(HISTORY_KEY));
}

/** Adds a query to history, removes duplicates and returns the new list. */
export function addSearchHistory(query: string): string[] {
  const normalized = query.trim();
  const next = [
    normalized,
    ...loadSearchHistory().filter((item) => item.toLowerCase() !== normalized.toLowerCase())
  ].slice(0, 8);
  localStorage.setItem(HISTORY_KEY, JSON.stringify(next));
  return next;
}

/** Removes all locally persisted search queries. */
export function clearSearchHistory(): void {
  localStorage.removeItem(HISTORY_KEY);
}
