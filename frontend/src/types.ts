export type UploadStatus =
  | "queued"
  | "uploading"
  | "indexing"
  | "ready"
  | "duplicate"
  | "error";

export interface DocumentItem {
  localId: string;
  serverId?: string;
  name: string;
  size: number;
  uploadedAt: string;
  status: UploadStatus;
  progress: number;
  pages?: number;
  error?: string;
}

export interface UploadResponse {
  id: string;
  file_name: string;
  size: number;
  status: string;
  message?: string;
  pages_count?: number;
  extracted_chars?: number;
  duplicate?: boolean;
}

export interface SearchResult {
  chunkId: string;
  fileName: string;
  page: number;
  text: string;
  score: number;
}

export interface SearchResponse {
  items: SearchResult[];
  total: number;
  page: number;
  totalPages: number;
}

export type UploadProgress = {
  phase: "uploading" | "indexing";
  value: number;
};
