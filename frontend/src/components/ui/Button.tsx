import { ButtonHTMLAttributes, forwardRef } from 'react'
import { cn } from '@/lib/utils'
import { Loader2 } from 'lucide-react'

const variants = {
  primary: 'bg-primary text-white hover:bg-primary-600 active:bg-primary-700 shadow-sm',
  secondary: 'bg-white text-on-surface border border-border hover:bg-surface-container-low active:bg-surface-container',
  outline: 'bg-transparent text-primary border border-primary hover:bg-primary-50 active:bg-primary-100',
  danger: 'bg-danger text-white hover:bg-danger-600 active:bg-danger-700 shadow-sm',
  ghost: 'bg-transparent text-on-surface hover:bg-surface-container-low active:bg-surface-container',
}

const sizes = {
  sm: 'h-8 px-3 text-body-sm',
  md: 'h-10 px-4 text-body-sm',
  lg: 'h-12 px-6 text-body-md',
  xl: 'h-14 px-8 text-body-lg',
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof variants
  size?: keyof typeof sizes
  loading?: boolean
  icon?: React.ReactNode
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', loading, icon, children, disabled, ...props }, ref) => {
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={cn(
          'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none',
          variants[variant],
          sizes[size],
          className
        )}
        {...props}
      >
        {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : icon}
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'

export { Button, type ButtonProps }
