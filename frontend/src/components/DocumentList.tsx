import type { DocumentItem, UploadStatus } from "../types";
import { formatBytes, formatDate } from "../utils/format";
import { Icon } from "./Icon";

interface DocumentListProps {
  documents: DocumentItem[];
  onRemove: (localId: string) => void;
}

const STATUS_LABELS: Record<UploadStatus, string> = {
  queued: "В очереди",
  uploading: "Загрузка...",
  indexing: "Индексация...",
  ready: "Готово",
  duplicate: "Уже загружен",
  error: "Ошибка"
};

function statusTone(status: UploadStatus): string {
  if (status === "ready" || status === "duplicate") return "success";
  if (status === "error") return "danger";
  return "progress";
}

/** Displays uploaded documents, processing progress and terminal statuses. */
export function DocumentList({ documents, onRemove }: DocumentListProps) {
  return (
    <section className="panel document-panel" aria-labelledby="documents-title">
      <div className="section-heading">
        <div>
          <span className="eyebrow">Библиотека</span>
          <h2 id="documents-title">Документы</h2>
        </div>
        <span className="count-badge" aria-label={"Документов: " + documents.length}>{documents.length}</span>
      </div>

      {documents.length === 0 ? (
        <div className="document-empty">
          <span><Icon name="file" size={25} /></span>
          <p>Загруженные документы появятся здесь</p>
        </div>
      ) : (
        <ul className="document-list">
          {documents.map((document) => {
            const isActive = ["queued", "uploading", "indexing"].includes(document.status);
            const extension = document.name.split(".").pop()?.toUpperCase() ?? "FILE";
            return (
              <li className="document-item" key={document.localId}>
                <div className={"file-mark file-mark--" + extension.toLowerCase()}>{extension}</div>
                <div className="document-item__body">
                  <div className="document-item__topline">
                    <strong title={document.name}>{document.name}</strong>
                    <span className={"status-chip status-chip--" + statusTone(document.status)}>
                      {document.status === "ready" && <Icon name="check" size={13} />}
                      {document.status === "error" && <Icon name="alert" size={13} />}
                      {STATUS_LABELS[document.status]}
                    </span>
                  </div>
                  <div className="document-meta">
                    <span>{formatBytes(document.size)}</span>
                    <span aria-hidden="true">•</span>
                    <span>{formatDate(document.uploadedAt)}</span>
                    {document.pages ? <><span aria-hidden="true">•</span><span>{document.pages} стр.</span></> : null}
                  </div>
                  {isActive && (
                    <div
                      className="progress"
                      role="progressbar"
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-valuenow={document.progress}
                      aria-label={STATUS_LABELS[document.status]}
                    >
                      <span style={{ width: document.progress + "%" }} />
                    </div>
                  )}
                  {document.error && <p className="document-error">{document.error}</p>}
                </div>
                {!isActive && (
                  <button
                    type="button"
                    className="icon-button"
                    onClick={() => onRemove(document.localId)}
                    aria-label={"Убрать " + document.name + " из списка"}
                    title="Убрать из списка"
                  >
                    <Icon name="trash" size={17} />
                  </button>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
