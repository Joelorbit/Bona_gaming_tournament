import { cn } from '@/lib/utils'
import { Search } from 'lucide-react'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  icon?: React.ReactNode
}

export function Input({ className, label, error, icon, ...props }: InputProps) {
  return (
    <div className="space-y-1.5">
      {label && (
        <label className="block text-body-sm font-medium text-on-surface">
          {label}
        </label>
      )}
      <div className="relative">
        {icon && (
          <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-text-secondary">
            {icon}
          </div>
        )}
        <input
          className={cn(
            'w-full h-10 rounded-lg border border-border bg-white px-3 text-body-md text-on-surface placeholder:text-text-tertiary',
            'focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent',
            'transition-all duration-150',
            icon ? 'pl-10' : '',
            error ? 'border-danger focus:ring-danger' : '',
            className
          )} 
          {...props}
        />
      </div>
      {error && <p className="text-body-sm text-danger">{error}</p>}
    </div>
  )
}

export function SearchInput({ className, ...props }: Omit<InputProps, 'icon'>) {
  return (
    <Input
      icon={<Search className="h-4 w-4" />}
      placeholder="Search..."
      className={cn('pl-10', className)}
      {...props}
    />
  )
}
