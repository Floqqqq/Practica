import { useRef, useState } from "react";
import type { ChangeEvent, DragEvent } from "react";
import { Icon } from "./Icon";

interface UploadZoneProps {
  disabled?: boolean;
  onFiles: (files: File[]) => void;
}

/** Accepts multiple PDF/DOCX files through drag-and-drop or the system picker. */
export function UploadZone({ disabled = false, onFiles }: UploadZoneProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const dragDepth = useRef(0);
  const [isDragging, setIsDragging] = useState(false);

  const openPicker = () => {
    if (!disabled) inputRef.current?.click();
  };

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files ?? []);
    if (files.length) onFiles(files);
    event.target.value = "";
  };

  const handleDragEnter = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    if (disabled) return;
    dragDepth.current += 1;
    setIsDragging(true);
  };

  const handleDragLeave = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    dragDepth.current -= 1;
    if (dragDepth.current <= 0) {
      dragDepth.current = 0;
      setIsDragging(false);
    }
  };

  const handleDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    dragDepth.current = 0;
    setIsDragging(false);
    if (disabled) return;
    const files = Array.from(event.dataTransfer.files);
    if (files.length) onFiles(files);
  };


  return (
    <div
      className={"upload-zone" + (isDragging ? " is-dragging" : "") + (disabled ? " is-disabled" : "")}
      onDragEnter={handleDragEnter}
      onDragOver={(event) => event.preventDefault()}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      role="group"
      aria-label="Загрузить документы PDF или DOCX"
      data-testid="upload-zone"
    >
      <input
        ref={inputRef}
        type="file"
        accept=".pdf,.docx,application/pdf,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        multiple
        hidden
        onChange={handleChange}
      />
      <span className="upload-zone__icon"><Icon name="upload" size={26} /></span>
      <div>
        <strong>{isDragging ? "Отпустите файлы здесь" : "Перетащите документы"}</strong>
        <p>или <button type="button" className="link-button" onClick={openPicker}>выберите на устройстве</button></p>
      </div>
      <span className="upload-zone__hint">PDF, DOCX · до 20 МБ · можно несколько</span>
    </div>
  );
}
