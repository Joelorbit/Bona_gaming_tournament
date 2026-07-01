import { cn } from '@/lib/utils'
import { Inbox } from 'lucide-react'
import { Button } from './Button'

interface EmptyStateProps {
  icon?: React.ReactNode
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
  className?: string
}

export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-16 px-4 text-center', className)}>
      <div className="w-16 h-16 rounded-full bg-surface-container flex items-center justify-center mb-4 text-text-secondary">
        {icon || <Inbox className="w-8 h-8" />}
      </div>
      <h3 className="text-headline-sm text-on-surface mb-2">{title}</h3>
      {description && (
        <p className="text-body-md text-text-secondary max-w-md mb-6">{description}</p>
      )}
      {action && (
        <Button onClick={action.onClick} size="lg">
          {action.label}
        </Button>
      )}
    </div>
  )
}
