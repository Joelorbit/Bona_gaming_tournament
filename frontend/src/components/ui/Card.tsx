import { cn } from '@/lib/utils'

interface CardProps {
  className?: string
  children: React.ReactNode
  hover?: boolean
  padding?: 'sm' | 'md' | 'lg' | 'none'
}

const paddings = {
  none: '',
  sm: 'p-4',
  md: 'p-6',
  lg: 'p-8',
}

export function Card({ className, children, hover = false, padding = 'md' }: CardProps) {
  return (
    <div
      className={cn(
        'bg-white rounded-xl border border-border shadow-card',
        hover && 'hover:shadow-card-hover hover:border-border/80 transition-all duration-200',
        paddings[padding],
        className
      )}
    >
      {children}
    </div>
  )
}

export function CardHeader({ className, children }: { className?: string; children: React.ReactNode }) {
  return <div className={cn('flex items-center justify-between mb-4', className)}>{children}</div>
}

export function CardTitle({ className, children }: { className?: string; children: React.ReactNode }) {
  return <h3 className={cn('text-headline-sm text-on-surface', className)}>{children}</h3>
}

export function CardContent({ className, children }: { className?: string; children: React.ReactNode }) {
  return <div className={cn('', className)}>{children}</div>
}
