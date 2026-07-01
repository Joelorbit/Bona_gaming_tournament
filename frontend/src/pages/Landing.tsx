import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { EmptyState } from '@/components/ui/EmptyState'
import { Trophy, Users, Award, ArrowRight, Search, Plus, TrendingUp } from 'lucide-react'
import { api, type Tournament } from '@/lib/api'

export function LandingPage() {
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    api.get<Tournament[]>('/api/v1/tournaments?limit=6&sort=starting_soon')
      .then(data => { if (!cancelled) setTournaments(data || []) })
      .catch(() => { if (!cancelled) setTournaments([]) })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  const activeCount = tournaments.filter(t => ['open', 'registration_closed', 'in_progress'].includes(t.status)).length
  const totalCapacity = useMemo(() => tournaments.reduce((sum, t) => sum + t.max_participants, 0), [tournaments])
  const prizePool = useMemo(() => tournaments.reduce((sum, t) => sum + (t.prize_pool || 0), 0), [tournaments])
  const games = useMemo(() => new Set(tournaments.map(t => t.game)).size, [tournaments])

  const stats = [
    { label: 'Active tournaments', value: activeCount.toLocaleString(), icon: Trophy, detail: 'Open or in progress' },
    { label: 'Player capacity', value: totalCapacity.toLocaleString(), icon: Users, detail: 'Across listed tournaments' },
    { label: 'Prize pool', value: `${prizePool.toLocaleString()} ETB`, icon: Award, detail: 'From current tournaments' },
    { label: 'Games', value: games.toLocaleString(), icon: TrendingUp, detail: 'Represented now' },
  ]

  return (
    <div className="space-y-12">
      <section className="pt-8 pb-4">
        <div className="max-w-3xl">
          <h1 className="text-display-hero text-on-surface mb-4 leading-[1]">
            Bona tournaments
          </h1>
          <p className="text-body-lg text-text-secondary mb-8">
            Create real tournaments, take registrations, collect entry payments, generate brackets, and track match results.
          </p>
          <div className="flex items-center gap-3 flex-wrap">
            <Link to="/tournaments">
              <Button size="xl" icon={<Search className="w-5 h-5" />}>
                Browse tournaments
              </Button>
            </Link>
            <Link to="/create">
              <Button size="xl" variant="secondary" icon={<Plus className="w-5 h-5" />}>
                Create tournament
              </Button>
            </Link>
          </div>
        </div>
      </section>

      <section className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map(stat => (
          <Card key={stat.label} padding="lg">
            <div className="w-10 h-10 rounded-lg bg-primary-50 flex items-center justify-center mb-3">
              <stat.icon className="w-5 h-5 text-primary" />
            </div>
            <p className="text-3xl font-bold text-on-surface mb-1">{stat.value}</p>
            <p className="text-body-sm text-text-secondary">{stat.label}</p>
            <p className="text-label-sm text-text-secondary mt-2">{stat.detail}</p>
          </Card>
        ))}
      </section>

      <section>
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-headline-md text-on-surface">Current tournaments</h2>
            <p className="text-body-md text-text-secondary">Live data from your database</p>
          </div>
          <Link to="/tournaments">
            <Button variant="ghost" size="sm" icon={<ArrowRight className="w-4 h-4" />}>
              View all
            </Button>
          </Link>
        </div>

        {loading ? (
          <div className="py-12 text-center text-text-secondary">Loading tournaments...</div>
        ) : tournaments.length === 0 ? (
          <EmptyState title="No tournaments yet" description="Create the first tournament to populate this page with real data." action={{ label: 'Create tournament', onClick: () => { window.location.href = '/create' } }} />
        ) : (
          <div className="grid md:grid-cols-3 gap-4">
            {tournaments.slice(0, 3).map(t => (
              <Link key={t.id} to={`/tournaments/${t.id}`}>
                <Card hover padding="lg" className="h-full group">
                  <div className="flex items-center justify-between mb-4">
                    <Badge status={t.status}>{t.status.replace(/_/g, ' ')}</Badge>
                    <span className="text-label-sm text-text-secondary">{t.max_participants} max</span>
                  </div>
                  <h3 className="text-title-lg text-on-surface mb-1 group-hover:text-primary transition-colors">{t.title}</h3>
                  <p className="text-body-sm text-text-secondary mb-4">{t.game}</p>
                  <div className="flex items-center justify-between pt-4 border-t border-border">
                    <div>
                      <p className="text-label-sm text-text-secondary">Prize pool</p>
                      <p className="text-title-lg font-bold text-on-surface">{t.prize_pool.toLocaleString()} {t.currency}</p>
                    </div>
                    <div className="text-right">
                      <p className="text-label-sm text-text-secondary">Entry</p>
                      <p className="text-body-sm font-medium text-on-surface">{t.entry_fee > 0 ? `${t.entry_fee.toLocaleString()} ${t.currency}` : 'Free'}</p>
                    </div>
                  </div>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
