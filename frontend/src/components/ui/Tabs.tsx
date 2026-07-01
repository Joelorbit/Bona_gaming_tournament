import { cn } from '@/lib/utils'

interface TabsProps {
  tabs: { id: string; label: string; count?: number }[]
  activeTab: string
  onTabChange: (id: string) => void
  className?: string
}

export function Tabs({ tabs, activeTab, onTabChange, className }: TabsProps) {
  return (
    <div className={cn('flex gap-1 border-b border-border', className)}>
      {tabs.map(tab => (
        <button
          key={tab.id}
          onClick={() => onTabChange(tab.id)}
          className={cn(
            'relative px-4 py-3 text-body-sm font-medium transition-colors duration-150',
            'focus:outline-none',
            activeTab === tab.id
              ? 'text-primary'
              : 'text-text-secondary hover:text-on-surface'
          )}
        >
          <span className="flex items-center gap-2">
            {tab.label}
            {tab.count !== undefined && (
              <span className={cn(
                'inline-flex items-center justify-center min-w-[20px] h-5 px-1.5 rounded-full text-label-sm',
                activeTab === tab.id
                  ? 'bg-primary-50 text-primary'
                  : 'bg-surface-container text-text-secondary'
              )}>
                {tab.count}
              </span>
            )}
          </span>
          {activeTab === tab.id && (
            <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full" />
          )}
        </button>
      ))}
    </div>
  )
}
