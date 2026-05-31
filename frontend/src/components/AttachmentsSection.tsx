import { useRef, useState, type ChangeEvent, type DragEvent } from 'react';
import {
  Download,
  FileText,
  ImageIcon,
  Loader2,
  Paperclip,
  Trash2,
  Upload,
} from 'lucide-react';
import { useToast } from '@/contexts/ToastContext';
import {
  useAttachments,
  useDeleteAttachment,
  useIsViewer,
  useUploadAttachment,
} from '@/hooks';
import { attachmentService } from '@/services';
import { formatBytes } from '@/lib/format';
import { cn } from '@/lib/cn';
import { Skeleton } from '@/components/ui';
import {
  ATTACHMENT_ACCEPT,
  ATTACHMENT_ALLOWED_TYPES,
  ATTACHMENT_MAX_SIZE,
  type Attachment,
} from '@/types';

interface Props {
  transactionId: string;
}

const ALLOWED = ATTACHMENT_ALLOWED_TYPES as readonly string[];

function isImage(contentType: string): boolean {
  return contentType.startsWith('image/');
}

export function AttachmentsSection({ transactionId }: Props) {
  const readOnly = useIsViewer();
  const toast = useToast();
  const { data: attachments, isLoading, isError } = useAttachments(transactionId);
  const upload = useUploadAttachment(transactionId);
  const remove = useDeleteAttachment(transactionId);

  const inputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);
  const [downloadingId, setDownloadingId] = useState<string | null>(null);

  const handleFile = (file: File | undefined) => {
    if (!file) return;
    if (!ALLOWED.includes(file.type)) {
      toast.error('Formato não suportado. Use imagem (JPG, PNG, WEBP, GIF) ou PDF.');
      return;
    }
    if (file.size > ATTACHMENT_MAX_SIZE) {
      toast.error('Arquivo muito grande. O tamanho máximo é 10 MB.');
      return;
    }
    upload.mutate(file);
  };

  const onInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    handleFile(e.target.files?.[0]);
    // Reset so selecting the same file again re-triggers change.
    e.target.value = '';
  };

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setDragOver(false);
    if (readOnly || upload.isPending) return;
    handleFile(e.dataTransfer.files?.[0]);
  };

  const handleDownload = async (att: Attachment) => {
    setDownloadingId(att.id);
    try {
      await attachmentService.download(att);
    } catch {
      toast.error('Não foi possível baixar o comprovante.');
    } finally {
      setDownloadingId(null);
    }
  };

  const handleDelete = (att: Attachment) => {
    if (window.confirm(`Remover o comprovante "${att.file_name}"?`)) {
      remove.mutate(att.id);
    }
  };

  const items = attachments ?? [];

  return (
    <div className="space-y-3">
      {/* Upload dropzone (hidden for viewers) */}
      {!readOnly && (
        <div>
          <input
            ref={inputRef}
            type="file"
            accept={ATTACHMENT_ACCEPT}
            className="sr-only"
            onChange={onInputChange}
            disabled={upload.isPending}
            aria-label="Selecionar comprovante para anexar"
          />
          <div
            role="button"
            tabIndex={0}
            onClick={() => !upload.isPending && inputRef.current?.click()}
            onKeyDown={(e) => {
              if ((e.key === 'Enter' || e.key === ' ') && !upload.isPending) {
                e.preventDefault();
                inputRef.current?.click();
              }
            }}
            onDragOver={(e) => {
              e.preventDefault();
              setDragOver(true);
            }}
            onDragLeave={() => setDragOver(false)}
            onDrop={onDrop}
            aria-disabled={upload.isPending}
            className={cn(
              'flex cursor-pointer flex-col items-center justify-center gap-1.5 rounded-xl border border-dashed px-4 py-6 text-center transition',
              dragOver
                ? 'border-primary bg-primary/5'
                : 'border-gray-300 hover:border-primary/60 hover:bg-gray-50 dark:border-ink-border dark:hover:bg-ink-elevated',
              upload.isPending && 'pointer-events-none opacity-70',
            )}
          >
            {upload.isPending ? (
              <Loader2 className="h-5 w-5 animate-spin text-primary" />
            ) : (
              <Upload className="h-5 w-5 text-gray-400" />
            )}
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              {upload.isPending ? 'Enviando…' : 'Anexar comprovante'}
            </span>
            <span className="text-xs text-gray-400">
              Imagem ou PDF, até 10 MB. Arraste e solte ou clique para escolher.
            </span>
          </div>
        </div>
      )}

      {/* Attachment list / states */}
      {isLoading ? (
        <div className="space-y-2">
          <Skeleton className="h-12 rounded-lg" />
          <Skeleton className="h-12 rounded-lg" />
        </div>
      ) : isError ? (
        <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-600 dark:text-red-400">
          Não foi possível carregar os comprovantes.
        </p>
      ) : items.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-1.5 py-4 text-center">
          <Paperclip className="h-5 w-5 text-gray-300 dark:text-gray-600" />
          <p className="text-sm text-gray-400">Nenhum comprovante anexado</p>
        </div>
      ) : (
        <ul className="space-y-2">
          {items.map((att) => {
            const Icon = isImage(att.content_type) ? ImageIcon : FileText;
            const downloading = downloadingId === att.id;
            const deleting = remove.isPending && remove.variables === att.id;
            return (
              <li
                key={att.id}
                className="flex items-center gap-3 rounded-lg border border-gray-200 bg-white px-3 py-2 dark:border-ink-border dark:bg-ink-elevated"
              >
                <span
                  className={cn(
                    'flex h-9 w-9 shrink-0 items-center justify-center rounded-lg',
                    isImage(att.content_type)
                      ? 'bg-sky-500/10 text-sky-500'
                      : 'bg-red-500/10 text-red-500',
                  )}
                >
                  <Icon className="h-4 w-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
                    {att.file_name}
                  </p>
                  <p className="text-xs text-gray-400">{formatBytes(att.size)}</p>
                </div>
                <div className="flex shrink-0 items-center gap-0.5">
                  <button
                    type="button"
                    onClick={() => handleDownload(att)}
                    disabled={downloading}
                    className="rounded-md p-1.5 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700 disabled:cursor-not-allowed disabled:opacity-60 dark:hover:bg-ink-surface"
                    aria-label={`Baixar ${att.file_name}`}
                  >
                    {downloading ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Download className="h-4 w-4" />
                    )}
                  </button>
                  {!readOnly && (
                    <button
                      type="button"
                      onClick={() => handleDelete(att)}
                      disabled={deleting}
                      className="rounded-md p-1.5 text-gray-400 transition hover:bg-red-50 hover:text-red-600 disabled:cursor-not-allowed disabled:opacity-60 dark:hover:bg-red-500/10"
                      aria-label={`Remover ${att.file_name}`}
                    >
                      {deleting ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Trash2 className="h-4 w-4" />
                      )}
                    </button>
                  )}
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
