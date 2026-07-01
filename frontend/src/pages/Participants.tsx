import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Avatar } from '@/components/ui/Avatar'
import { SearchInput } from '@/components/ui/Input'
import { Tabs } from '@/components/ui/Tabs'
import { EmptyState } from '@/components/ui/EmptyState'
import { Trophy, Users } from 'lucide-react'
import { api, ApiError, type Tournament } from '@/lib/api'

interface Registration {
  id: string
  tournament_id: string
  user_id: string
  payment_status: string
  seed?: number | null
  registered_at: string
  username?: string | null
  display_name?: string | null
  avatar_url?: string | null
}

export function Participants() {
  const { id } = useParams<{ id?: string }>()
  const [tournament, setTournament] = useState<Tournament | null>(null)
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [participants, setParticipants] = useState<Registration[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState('all')
  const [search, setSearch] = useState('')

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    const load = id
      ? Promise.all([
          api.get<Tournament>(`/api/v1/tournaments/${id}`),
          api.get<Registration[]>(`/api/v1/tournaments/${id}/players`),
        ]).then(([t, p]) => {
          if (cancelled) return
          setTournament(t)
          setParticipants(p || [])
          setTournaments([])
        })
      : api.get<Tournament[]>('/api/v1/tournaments?limit=100').then(data => {
          if (cancelled) return
          setTournaments(data || [])
          setTournament(null)
          setParticipants([])
        })

    load
      .catch(err => {
        if (cancelled) return
        setError(err instanceof ApiError ? err.message : 'Could not load participants')
      })
      .finally(() => { if (!cancelled) setLoading(false) })

    return () => { cancelled = true }
  }, [id])

  const tabs = useMemo(() => [
    { id: 'all', label: 'All', count: participants.length },
    { id: 'paid', label: 'Paid', count: participants.filter(p => p.payment_status === 'paid').length },
    { id: 'pending', label: 'Pending', count: participants.filter(p => p.payment_status !== 'paid').length },
  ], [participants])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return participants.filter(p => {
      if (activeTab === 'paid' && p.payment_status !== 'paid') return false
      if (activeTab === 'pending' && p.payment_status === 'paid') return false
      if (!q) return true
      const label = `${p.display_name || ''} ${p.username || ''} ${p.user_id}`.toLowerCase()
      return label.includes(q)
    })
  }, [activeTab, participants, search])

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading...</div>

  if (error) {
    return <EmptyState title="Could not load participants" description={error} />
  }

  if (!id) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-headline-md text-on-surface">Participants</h1>
          <p className="text-body-md text-text-secondary">Choose a real tournament to view its registered players.</p>
        </div>

        {tournaments.length === 0 ? (
          <EmptyState title="No tournaments yet" description="Create or join a tournament, then participants will appear here." />
        ) : (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
            {tournaments.map(t => (
              <Link key={t.id} to={`/tournaments/${t.id}/participants`}>
                <Card hover padding="lg" className="h-full">
                  <div className="flex items-center justify-between mb-3">
                    <div className="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
                      <Trophy className="w-5 h-5" />
                    </div>
                    <Badge status={t.status}>{t.status.replace(/_/g, ' ')}</Badge>
                  </div>
                  <h2 className="text-title-lg text-on-surface">{t.title}</h2>
                  <p className="text-body-sm text-text-secondary mt-1">{t.game}</p>
                  <p className="text-label-md text-text-secondary mt-4">Max {t.max_participants} players</p>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-headline-md text-on-surface">Participants</h1>
          <p className="text-body-md text-text-secondary">{tournament?.title || 'Tournament'} · {participants.length} registered</p>
        </div>
        <Link to={`/tournaments/${id}`} className="text-body-sm text-primary hover:text-primary-600">
          Back to tournament
        </Link>
      </div>

      <div className="flex flex-col sm:flex-row gap-3 sm:items-center sm:justify-between">
        <div className="max-w-sm w-full">
          <SearchInput value={search} onChange={e => setSearch(e.target.value)} placeholder="Search username or player..." />
        </div>
        <Tabs tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />
      </div>

      <Card padding="none">
        <CardHeader className="px-6 pt-6 mb-0">
          <CardTitle className="flex items-center gap-2"><Users className="w-5 h-5" /> Registered players</CardTitle>
        </CardHeader>
        <CardContent>
          {filtered.length === 0 ? (
            <EmptyState title="No players found" description={participants.length === 0 ? 'No one has registered yet.' : 'Try another search or filter.'} />
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left pl-6 pr-4 py-3 text-label-sm text-text-secondary font-medium">Seed</th>
                    <th className="text-left px-4 py-3 text-label-sm text-text-secondary font-medium">Player</th>
                    <th className="text-center px-4 py-3 text-label-sm text-text-secondary font-medium">Payment</th>
                    <th className="text-right pr-6 pl-4 py-3 text-label-sm text-text-secondary font-medium">Profile</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((p, index) => (
                    <tr key={p.id} className="border-b border-border/50 last:border-0 hover:bg-surface-container-low/50 transition-colors">
                      <td className="pl-6 pr-4 py-4">
                        <span className="text-body-sm font-mono text-text-secondary">#{p.seed || index + 1}</span>
                      </td>
                      <td className="px-4 py-4">
                        <div className="flex items-center gap-3">
                          <Avatar src={p.avatar_url || undefined} name={p.display_name || p.username || p.user_id} size="sm" />
                          <div>
                            <p className="text-body-sm font-medium text-on-surface">{p.display_name || p.username || 'Player'}</p>
                            {p.username && <p className="text-label-sm text-text-secondary">@{p.username}</p>}
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-4 text-center">
                        <Badge status={p.payment_status === 'paid' ? 'active' : 'pending'}>{p.payment_status}</Badge>
                      </td>
                      <td className="pr-6 pl-4 py-4 text-right">
                        {p.username ? (
                          <Link to={`/u/${p.username}`} className="text-body-sm text-primary hover:text-primary-600">View</Link>
                        ) : (
                          <span className="text-body-sm text-text-secondary">Unavailable</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
