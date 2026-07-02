import type {
  SearchResponse,
  SearchResult,
  UploadProgress,
  UploadResponse
} from "../types";

const API_BASE = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");

/** Represents an HTTP or network failure returned by the application API. */
export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status = 0) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

function errorMessage(body: unknown, fallback: string): string {
  if (body && typeof body === "object" && "message" in body) {
    const message = (body as { message?: unknown }).message;
    if (typeof message === "string" && message.trim()) return message;
  }
  return fallback;
}

/** Uploads one document and reports its transfer and indexing phases. */
export function uploadDocument(
  file: File,
  onProgress: (progress: UploadProgress) => void
): Promise<UploadResponse> {
  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest();
    const data = new FormData();
    data.append("file", file);

    request.open("POST", API_BASE + "/api/v1/documents/upload");
    request.responseType = "json";
    request.timeout = 120_000;

    request.upload.addEventListener("progress", (event) => {
      if (!event.lengthComputable) return;
      const percent = Math.min(68, Math.round((event.loaded / event.total) * 68));
      onProgress({ phase: "uploading", value: percent });
    });
    request.upload.addEventListener("load", () => {
      onProgress({ phase: "indexing", value: 76 });
    });

    request.addEventListener("load", () => {
      const response = request.response as unknown;
      if (request.status >= 200 && request.status < 300) {
        onProgress({ phase: "indexing", value: 96 });
        resolve(response as UploadResponse);
        return;
      }
      reject(new ApiError(errorMessage(response, "Не удалось загрузить документ"), request.status));
    });
    request.addEventListener("error", () => {
      reject(new ApiError("Сервер недоступен. Проверьте подключение к API."));
    });
    request.addEventListener("timeout", () => {
      reject(new ApiError("Обработка заняла слишком много времени. Повторите попытку."));
    });
    request.addEventListener("abort", () => reject(new ApiError("Загрузка отменена.")));
    request.send(data);
  });
}

function toSearchResult(raw: unknown, index: number): SearchResult {
  const item = (raw && typeof raw === "object" ? raw : {}) as Record<string, unknown>;
  const page = Number(item.page ?? item.page_number ?? 1);
  const score = Number(item.score ?? item.relevance ?? 0);
  return {
    chunkId: String(item.chunk_id ?? item.id ?? index),
    fileName: String(item.file_name ?? item.fileName ?? "Документ"),
    page: Number.isFinite(page) && page > 0 ? page : 1,
    text: String(item.text ?? item.fragment ?? ""),
    score: Number.isFinite(score) ? score : 0
  };
}

/** Requests one page of ranked search results and normalizes supported API shapes. */
export async function searchDocuments(
  query: string,
  page: number,
  pageSize: number,
  signal?: AbortSignal
): Promise<SearchResponse> {
  const params = new URLSearchParams({ q: query, page: String(page), limit: String(pageSize) });
  let response: Response;
  try {
    response = await fetch(API_BASE + "/api/v1/search?" + params.toString(), {
      headers: { Accept: "application/json" },
      signal
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") throw error;
    throw new ApiError("Сервер поиска недоступен. Проверьте подключение к API.");
  }

  let body: unknown = null;
  try {
    body = await response.json();
  } catch {
    // The status-specific message below is more useful than a JSON parse error.
  }
  if (!response.ok) {
    throw new ApiError(errorMessage(body, "Ошибка при выполнении поиска"), response.status);
  }

  if (Array.isArray(body)) {
    const allItems = body.map(toSearchResult);
    const start = (page - 1) * pageSize;
    return {
      items: allItems.slice(start, start + pageSize),
      total: allItems.length,
      page,
      totalPages: Math.max(1, Math.ceil(allItems.length / pageSize))
    };
  }

  const data = (body && typeof body === "object" ? body : {}) as Record<string, unknown>;
  const rawItems = Array.isArray(data.results)
    ? data.results
    : Array.isArray(data.items)
      ? data.items
      : [];
  const total = Number(data.total ?? data.count ?? rawItems.length);
  const currentPage = Number(data.page ?? page);
  const declaredPages = Number(data.total_pages ?? data.totalPages);
  const totalPages = Number.isFinite(declaredPages) && declaredPages > 0
    ? declaredPages
    : Math.max(1, Math.ceil(total / pageSize));

  return {
    items: rawItems.map(toSearchResult),
    total: Number.isFinite(total) ? total : rawItems.length,
    page: Number.isFinite(currentPage) ? currentPage : page,
    totalPages
  };
}
