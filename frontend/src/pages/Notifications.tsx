import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Bell, Check, Trash2 } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { api } from '@/lib/api'
import { formatTimeAgo } from '@/lib/utils'

export interface Notification {
  id: string
  user_id: string
  type: string
  title: string
  message: string
  link?: string | null
  read_at?: string | null
  created_at: string
}

export function Notifications() {
  const [items, setItems] = useState<Notification[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const list = await api.get<Notification[]>('/api/v1/notifications')
      setItems(list || [])
    } catch (err: any) {
      setError(err?.message || 'Failed to load notifications')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  async function markRead(id: string) {
    try {
      await api.post(`/api/v1/notifications/${id}/read`)
      setItems(prev => prev.map(n => n.id === id ? { ...n, read_at: new Date().toISOString() } : n))
    } catch {}
  }

  async function markAllRead() {
    try {
      await api.post('/api/v1/notifications/read-all')
      const now = new Date().toISOString()
      setItems(prev => prev.map(n => n.read_at ? n : { ...n, read_at: now }))
    } catch {}
  }

  async function remove(id: string) {
    try {
      await api.delete(`/api/v1/notifications/${id}`)
      setItems(prev => prev.filter(n => n.id !== id))
    } catch {}
  }

  const unread = items.filter(n => !n.read_at).length

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading…</div>

  return (
    <div className="max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-headline-md text-on-surface">Notifications</h1>
        {unread > 0 && (
          <Button variant="outline" size="sm" icon={<Check className="w-4 h-4" />} onClick={markAllRead}>
            Mark all read
          </Button>
        )}
      </div>

      {error && (
        <Card padding="md" className="mb-4">
          <p className="text-body-sm text-danger">{error}</p>
        </Card>
      )}

      {items.length === 0 ? (
        <EmptyState
          icon={<Bell className="w-6 h-6" />}
          title="No notifications yet"
          description="When something happens — a payment confirms, a match is up — you'll see it here."
        />
      ) : (
        <div className="space-y-2">
          {items.map(n => (
            <Card key={n.id} padding="md" className={n.read_at ? '' : 'border-primary/40 bg-primary/5'}>
              <div className="flex items-start justify-between gap-3">
                <div className="flex-1 min-w-0">
                  {n.link ? (
                    <Link to={n.link} onClick={() => !n.read_at && markRead(n.id)} className="block">
                      <NotificationBody n={n} />
                    </Link>
                  ) : (
                    <NotificationBody n={n} />
                  )}
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  {!n.read_at && (
                    <button
                      onClick={() => markRead(n.id)}
                      title="Mark read"
                      className="p-2 rounded-lg text-text-secondary hover:text-on-surface hover:bg-surface-container-low"
                    >
                      <Check className="w-4 h-4" />
                    </button>
                  )}
                  <button
                    onClick={() => remove(n.id)}
                    title="Delete"
                    className="p-2 rounded-lg text-text-secondary hover:text-danger hover:bg-danger-50"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}

function NotificationBody({ n }: { n: Notification }) {
  return (
    <>
      <p className="text-body-md font-medium text-on-surface">{n.title}</p>
      <p className="text-body-sm text-text-secondary mt-0.5">{n.message}</p>
      <p className="text-label-sm text-text-secondary mt-1">{formatTimeAgo(n.created_at)}</p>
    </>
  )
}
