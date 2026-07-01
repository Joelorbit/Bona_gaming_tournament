import { cn } from '@/lib/utils'
import type { ReactNode } from 'react'

const statusStyles: Record<string, string> = {
  open: 'bg-success-50 text-success-700 border-success-200',
  draft: 'bg-surface-container text-on-surface-variant border-border',
  registration_closed: 'bg-warning-50 text-warning-700 border-warning-200',
  in_progress: 'bg-primary-50 text-primary-700 border-primary-200',
  completed: 'bg-surface-variant text-on-surface-variant border-outline-variant',
  cancelled: 'bg-danger-50 text-danger-700 border-danger-200',
  active: 'bg-success-50 text-success-700 border-success-200',
  eliminated: 'bg-surface-variant text-on-surface-variant border-outline-variant',
  winner: 'bg-warning-50 text-warning-700 border-warning-200',
  pending: 'bg-surface-container text-on-surface-variant border-border',
}

const statusLabels: Record<string, string> = {
  open: 'Open',
  draft: 'Draft',
  registration_closed: 'Registration Closed',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
  active: 'Active',
  eliminated: 'Eliminated',
  winner: 'Winner',
  pending: 'Pending',
}

interface BadgeProps {
  status?: string
  className?: string
  children?: ReactNode
}

export function Badge({ status, className, children }: BadgeProps) {
  const badgeStatus = status || 'pending'

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-label-sm font-medium',
        statusStyles[badgeStatus] || 'bg-surface-container text-on-surface-variant border-border',
        className
      )}
    >
      <span className={cn(
        'w-1.5 h-1.5 rounded-full',
        badgeStatus === 'open' && 'bg-success-500',
        badgeStatus === 'draft' && 'bg-outline',
        badgeStatus === 'registration_closed' && 'bg-warning-500',
        badgeStatus === 'in_progress' && 'bg-primary',
        badgeStatus === 'completed' && 'bg-outline',
        badgeStatus === 'cancelled' && 'bg-danger',
        badgeStatus === 'winner' && 'bg-warning-500',
        badgeStatus === 'active' && 'bg-success-500',
        badgeStatus === 'eliminated' && 'bg-outline',
      )} />
      {children ?? statusLabels[badgeStatus] ?? badgeStatus}
    </span>
  )
}
