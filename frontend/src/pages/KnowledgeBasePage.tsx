import { useCallback, useEffect, useRef, useState } from "react";
import { DocumentList } from "../components/DocumentList";
import { Icon } from "../components/Icon";
import { SearchBar } from "../components/SearchBar";
import { SearchHistory } from "../components/SearchHistory";
import { SearchResults } from "../components/SearchResults";
import { UploadZone } from "../components/UploadZone";
import { ApiError, searchDocuments, uploadDocument } from "../services/api";
import {
  addSearchHistory,
  clearSearchHistory,
  loadDocuments,
  loadSearchHistory,
  saveDocuments
} from "../services/storage";
import type { DocumentItem, SearchResponse } from "../types";
import { validateFile } from "../utils/files";

const PAGE_SIZE = 10;
const EMPTY_RESULTS: SearchResponse = { items: [], total: 0, page: 1, totalPages: 1 };

function localId(): string {
  return typeof crypto.randomUUID === "function"
    ? crypto.randomUUID()
    : Date.now().toString(36) + Math.random().toString(36).slice(2);
}

/** Coordinates document uploads, local history, API search and pagination. */
export function KnowledgeBasePage() {
  const [documents, setDocuments] = useState<DocumentItem[]>(loadDocuments);
  const [query, setQuery] = useState("");
  const [activeQuery, setActiveQuery] = useState("");
  const [history, setHistory] = useState<string[]>(loadSearchHistory);
  const [results, setResults] = useState<SearchResponse>(EMPTY_RESULTS);
  const [searching, setSearching] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);
  const searchAbort = useRef<AbortController | null>(null);

  useEffect(() => saveDocuments(documents), [documents]);
  useEffect(() => () => searchAbort.current?.abort(), []);

  const updateDocument = useCallback((id: string, patch: Partial<DocumentItem>) => {
    setDocuments((current) => current.map((item) =>
      item.localId === id ? { ...item, ...patch } : item
    ));
  }, []);

  const processFile = useCallback(async (file: File, id: string) => {
    updateDocument(id, { status: "uploading", progress: 2 });
    try {
      const response = await uploadDocument(file, ({ phase, value }) => {
        updateDocument(id, { status: phase, progress: value });
      });
      updateDocument(id, {
        serverId: response.id,
        name: response.file_name || file.name,
        size: response.size || file.size,
        pages: response.pages_count,
        progress: 100,
        status: response.duplicate || response.status === "duplicate" ? "duplicate" : "ready",
        error: undefined
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : "Неизвестная ошибка загрузки";
      updateDocument(id, { status: "error", progress: 0, error: message });
    }
  }, [updateDocument]);

  const handleFiles = useCallback((files: File[]) => {
    const now = new Date().toISOString();
    const incoming = files.map((file) => {
      const error = validateFile(file);
      return {
        file,
        item: {
          localId: localId(),
          name: file.name,
          size: file.size,
          uploadedAt: now,
          status: error ? "error" as const : "queued" as const,
          progress: 0,
          error: error ?? undefined
        }
      };
    });

    setDocuments((current) => [...incoming.map(({ item }) => item), ...current]);
    incoming.forEach(({ file, item }) => {
      if (!item.error) void processFile(file, item.localId);
    });
  }, [processFile]);

  const runSearch = useCallback(async (value: string, targetPage = 1) => {
    const normalized = value.trim();
    if (!normalized) return;

    searchAbort.current?.abort();
    const controller = new AbortController();
    searchAbort.current = controller;
    setSearching(true);
    setSearchError(null);
    setHasSearched(true);
    setActiveQuery(normalized);

    if (targetPage === 1) setHistory(addSearchHistory(normalized));

    try {
      const response = await searchDocuments(normalized, targetPage, PAGE_SIZE, controller.signal);
      setResults(response);
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") return;
      const message = error instanceof ApiError ? error.message : "Не удалось выполнить поиск";
      setResults({ ...EMPTY_RESULTS, page: targetPage });
      setSearchError(message);
    } finally {
      if (searchAbort.current === controller) setSearching(false);
    }
  }, []);

  const selectHistory = (value: string) => {
    setQuery(value);
    void runSearch(value, 1);
  };

  const readyCount = documents.filter((item) => item.status === "ready" || item.status === "duplicate").length;
  const activeUploads = documents.filter((item) => ["queued", "uploading", "indexing"].includes(item.status)).length;

  return (
    <div className="app-shell">
      <header className="topbar">
        <a className="brand" href="#top" aria-label="Контекст — на главную">
          <span className="brand__mark">К</span>
          <span><strong>КОНТЕКСТ</strong><small>база знаний университета</small></span>
        </a>
        <div className="system-status"><span />Система готова</div>
      </header>

      <main id="top">
        <section className="hero">
          <div className="hero__glow" aria-hidden="true" />
          <div className="hero__content">
            <span className="hero__eyebrow"><Icon name="spark" size={16} />Интеллектуальный поиск</span>
            <h1>Найдите точный ответ.<br /><em>Не весь документ.</em></h1>
            <p>Загружайте учебные материалы и находите нужный фрагмент за несколько секунд.</p>
            <SearchBar value={query} loading={searching} onChange={setQuery} onSubmit={() => void runSearch(query, 1)} />
            <SearchHistory
              items={history}
              onSelect={selectHistory}
              onClear={() => { clearSearchHistory(); setHistory([]); }}
            />
          </div>
        </section>

        <div className="workspace">
          <aside className="workspace__sidebar">
            <section className="panel upload-panel" aria-labelledby="upload-title">
              <div className="section-heading">
                <div><span className="eyebrow">Пополнение базы</span><h2 id="upload-title">Загрузить</h2></div>
              </div>
              <UploadZone onFiles={handleFiles} />
              {activeUploads > 0 && <p className="upload-note" role="status">Обрабатывается файлов: {activeUploads}</p>}
            </section>
            <DocumentList
              documents={documents}
              onRemove={(id) => setDocuments((current) => current.filter((item) => item.localId !== id))}
            />
          </aside>

          <div className="workspace__main">
            <div className="library-stat">
              <span><Icon name="file" size={18} /></span>
              <p><strong>{readyCount}</strong> документов доступно для поиска</p>
            </div>
            <SearchResults
              query={activeQuery}
              items={results.items}
              total={results.total}
              page={results.page}
              totalPages={results.totalPages}
              loading={searching}
              hasSearched={hasSearched}
              error={searchError}
              onPageChange={(page) => void runSearch(activeQuery, page)}
            />
          </div>
        </div>
      </main>
      <footer><span>КОНТЕКСТ</span><p>Учебный проект · 2026</p></footer>
    </div>
  );
}
