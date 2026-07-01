import { cn } from '@/lib/utils'

const avatarColors = [
  'bg-primary text-white',
  'bg-success-500 text-white',
  'bg-warning-500 text-white',
  'bg-danger text-white',
  'bg-primary-200 text-primary-800',
  'bg-success-200 text-success-800',
]

interface AvatarProps {
  src?: string
  name: string
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

export function Avatar({ src, name, size = 'md', className }: AvatarProps) {
  const initials = name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

  const colorIndex = name.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0) % avatarColors.length

  const sizeClasses = {
    sm: 'w-8 h-8 text-body-sm',
    md: 'w-10 h-10 text-body-md',
    lg: 'w-12 h-12 text-body-lg',
  }

  if (src) {
    return (
      <img
        src={src}
        alt={name}
        className={cn('rounded-full object-cover', sizeClasses[size], className)}
      />
    )
  }

  return (
    <div
      className={cn(
        'rounded-full flex items-center justify-center font-semibold',
        avatarColors[colorIndex],
        sizeClasses[size],
        className
      )}
    >
      {initials}
    </div>
  )
}
