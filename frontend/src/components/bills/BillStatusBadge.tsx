import { Badge } from '@/components/ui';
import type { TransactionStatus } from '@/types';

const statusMeta: Record<TransactionStatus, { label: string; variant: 'success' | 'danger' | 'warning' }> = {
  paid: { label: 'Paga', variant: 'success' },
  overdue: { label: 'Vencida', variant: 'danger' },
  open: { label: 'Em aberto', variant: 'warning' },
};

export function BillStatusBadge({ status }: { status?: TransactionStatus }) {
  const meta = statusMeta[status ?? 'open'];
  return (
    <Badge variant={meta.variant} dot>
      {meta.label}
    </Badge>
  );
}
