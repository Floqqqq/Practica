export const MAX_FILE_SIZE = 20 * 1024 * 1024;

/** Validates a file extension, non-empty content and the 20 MB size limit. */
export function validateFile(file: File): string | null {
  const extension = file.name.split(".").pop()?.toLowerCase();
  if (extension !== "pdf" && extension !== "docx") {
    return "Поддерживаются только файлы PDF и DOCX";
  }
  if (file.size <= 0) return "Файл пуст";
  if (file.size > MAX_FILE_SIZE) return "Размер файла превышает 20 МБ";
  return null;
}
