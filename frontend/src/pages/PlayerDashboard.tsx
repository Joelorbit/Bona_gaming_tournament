import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { EmptyState } from '@/components/ui/EmptyState'
import { Trophy, Calendar, Award, Receipt } from 'lucide-react'
import { api, type Tournament } from '@/lib/api'
import { formatDate } from '@/lib/utils'

interface RegistrationRow {
  id: string
  tournament_id: string
  payment_status: string
  registered_at: string
  title: string
  game: string
  status: string
  start_date: string
  prize_pool: number
}

interface MatchRow {
  id: string
  tournament_id: string
  round: number
  position: number
  status: string
  player_a_id?: string | null
  player_b_id?: string | null
}

interface PayoutRow {
  id: string
  tournament_id: string
  winner_id: string
  amount: number
  currency: string
  status: string
  paid_at?: string | null
  created_at: string
}

export function PlayerDashboard() {
  const [registrations, setRegistrations] = useState<RegistrationRow[]>([])
  const [matches, setMatches] = useState<MatchRow[]>([])
  const [payouts, setPayouts] = useState<PayoutRow[]>([])
  const [tournamentTitles, setTournamentTitles] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    Promise.all([
      api.get<RegistrationRow[]>('/api/v1/me/registrations').catch(() => []),
      api.get<MatchRow[]>('/api/v1/me/matches').catch(() => []),
      api.get<PayoutRow[]>('/api/v1/me/payouts').catch(() => []),
    ])
      .then(([r, m, p]) => {
        if (cancelled) return
        setRegistrations(r || [])
        setMatches(m || [])
        setPayouts(p || [])
        const knownTitles = new Map<string, string>()
        ;(r || []).forEach(row => knownTitles.set(row.tournament_id, row.title))
        const missingIDs = Array.from(new Set([...(m || []), ...(p || [])].map(row => row.tournament_id).filter(id => !knownTitles.has(id))))
        Promise.all(missingIDs.map(id => api.get<Tournament>(`/api/v1/tournaments/${id}`).then(t => [id, t.title] as const).catch(() => [id, 'Tournament'] as const)))
          .then(entries => {
            if (cancelled) return
            entries.forEach(([id, title]) => knownTitles.set(id, title))
            setTournamentTitles(Object.fromEntries(knownTitles))
          })
      })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading…</div>

  const upcoming = registrations.filter(r => r.status === 'open' || r.status === 'registration_closed' || r.status === 'in_progress')
  const completed = registrations.filter(r => r.status === 'completed')
  const totalWon = payouts.filter(p => p.status === 'paid').reduce((s, p) => s + p.amount, 0)
  const pendingPayouts = payouts.filter(p => p.status === 'pending')

  return (
    <div className="space-y-6">
      <h1 className="text-headline-md text-on-surface">My dashboard</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard icon={<Trophy />} label="Joined" value={registrations.length} />
        <StatCard icon={<Calendar />} label="Upcoming" value={upcoming.length} />
        <StatCard icon={<Award />} label="Wins" value={payouts.length} />
        <StatCard icon={<Receipt />} label="Total won" value={`${totalWon.toLocaleString()} ETB`} />
      </div>

      <Card padding="lg">
        <CardHeader><CardTitle>Open matches</CardTitle></CardHeader>
        <CardContent>
          {matches.length === 0 ? (
            <EmptyState title="No active matches" description="When a bracket is generated and you're seeded in, your matches appear here." />
          ) : (
            <div className="space-y-2">
              {matches.map(m => (
                <Link key={m.id} to={`/matches/${m.id}`} className="block p-3 border border-border rounded-lg hover:border-primary/50 hover:bg-surface-container-low transition-all">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-body-sm font-medium text-on-surface">Round {m.round} · Match {m.position + 1}</p>
                      <p className="text-label-md text-text-secondary">{tournamentTitles[m.tournament_id] || 'Tournament'}</p>
                    </div>
                    <Badge>{m.status.replace(/_/g, ' ')}</Badge>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid lg:grid-cols-2 gap-6">
        <Card padding="lg">
          <CardHeader><CardTitle>Upcoming tournaments</CardTitle></CardHeader>
          <CardContent>
            {upcoming.length === 0 ? (
              <EmptyState title="Nothing upcoming" description="Browse to find your next tournament." />
            ) : (
              <div className="space-y-2">
                {upcoming.map(r => <RegistrationRowView key={r.id} r={r} />)}
              </div>
            )}
          </CardContent>
        </Card>

        <Card padding="lg">
          <CardHeader><CardTitle>Completed</CardTitle></CardHeader>
          <CardContent>
            {completed.length === 0 ? (
              <EmptyState title="No completed tournaments" description="Wins will appear here once your tournaments finish." />
            ) : (
              <div className="space-y-2">
                {completed.map(r => <RegistrationRowView key={r.id} r={r} />)}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card padding="lg">
        <CardHeader><CardTitle>Prize history</CardTitle></CardHeader>
        <CardContent>
          {payouts.length === 0 ? (
            <EmptyState title="No prizes yet" description="Win a tournament and your prize history shows here." />
          ) : (
            <div className="space-y-2">
              {payouts.map(p => (
                <div key={p.id} className="flex items-center justify-between p-3 border border-border rounded-lg">
                  <div>
                    <p className="text-body-sm font-medium text-on-surface">
                      <Link to={`/tournaments/${p.tournament_id}`} className="hover:text-primary">{tournamentTitles[p.tournament_id] || 'Tournament'}</Link>
                    </p>
                    <p className="text-label-md text-text-secondary">{formatDate(p.created_at)}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-body-md font-bold text-on-surface">{p.amount.toLocaleString()} {p.currency}</p>
                    <Badge status={p.status === 'paid' ? 'completed' : 'pending'}>{p.status}</Badge>
                  </div>
                </div>
              ))}
            </div>
          )}
          {pendingPayouts.length > 0 && (
            <p className="text-body-sm text-text-secondary mt-3">
              {pendingPayouts.length} payout{pendingPayouts.length === 1 ? '' : 's'} pending from organizer.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: number | string }) {
  return (
    <Card padding="md">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">{icon}</div>
        <div>
          <p className="text-label-md text-text-secondary">{label}</p>
          <p className="text-title-lg font-bold text-on-surface">{value}</p>
        </div>
      </div>
    </Card>
  )
}

function RegistrationRowView({ r }: { r: RegistrationRow }) {
  return (
    <Link to={`/tournaments/${r.tournament_id}`} className="block p-3 border border-border rounded-lg hover:border-primary/50 hover:bg-surface-container-low transition-all">
      <div className="flex items-center justify-between gap-3">
        <div className="flex-1 min-w-0">
          <p className="text-body-sm font-medium text-on-surface truncate">{r.title}</p>
          <p className="text-label-md text-text-secondary">{r.game} · starts {formatDate(r.start_date)}</p>
        </div>
        <div className="flex items-center gap-2">
          <Badge status={r.payment_status === 'paid' ? 'active' : 'pending'}>{r.payment_status}</Badge>
          <Badge status={r.status}>{r.status.replace(/_/g, ' ')}</Badge>
        </div>
      </div>
    </Link>
  )
}
