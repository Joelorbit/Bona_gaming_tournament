import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { SearchInput } from '@/components/ui/Input'
import { EmptyState } from '@/components/ui/EmptyState'
import { Grid3X3, List, Trophy, Users, MapPin, X } from 'lucide-react'
import { api, ApiError, type Tournament } from '@/lib/api'

type PaidFilter = '' | 'free' | 'paid'
type SortOption = 'newest' | 'starting_soon' | 'prize_high' | 'prize_low'

const STATUS_FILTERS: { value: string; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'open', label: 'Open' },
  { value: 'registration_closed', label: 'Closed' },
  { value: 'in_progress', label: 'In progress' },
  { value: 'completed', label: 'Completed' },
]

const PAID_FILTERS: { value: PaidFilter; label: string }[] = [
  { value: '', label: 'Any' },
  { value: 'free', label: 'Free' },
  { value: 'paid', label: 'Paid' },
]

const SORTS: { value: SortOption; label: string }[] = [
  { value: 'newest', label: 'Newest' },
  { value: 'starting_soon', label: 'Starting soon' },
  { value: 'prize_high', label: 'Prize: high to low' },
  { value: 'prize_low', label: 'Prize: low to high' },
]

export function BrowseTournaments() {
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [q, setQ] = useState('')
  const [status, setStatus] = useState('')
  const [paid, setPaid] = useState<PaidFilter>('')
  const [game, setGame] = useState('')
  const [sort, setSort] = useState<SortOption>('newest')

  const queryString = useMemo(() => {
    const sp = new URLSearchParams()
    if (q) sp.set('q', q)
    if (status) sp.set('status', status)
    if (paid) sp.set('paid', paid)
    if (game) sp.set('game', game)
    if (sort) sp.set('sort', sort)
    sp.set('limit', '50')
    return sp.toString()
  }, [q, status, paid, game, sort])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    const t = setTimeout(() => {
      api.get<Tournament[]>(`/api/v1/tournaments?${queryString}`)
        .then(data => { if (!cancelled) { setTournaments(data ?? []); setError(null) } })
        .catch(err => {
          if (cancelled) return
          setError(err instanceof ApiError ? err.message : 'Failed to load tournaments')
        })
        .finally(() => { if (!cancelled) setLoading(false) })
    }, 200)
    return () => { cancelled = true; clearTimeout(t) }
  }, [queryString])

  const games = useMemo(() => Array.from(new Set(tournaments.map(t => t.game))).sort(), [tournaments])
  const activeFilters = [status, paid, game].filter(Boolean).length

  function clearAll() {
    setStatus(''); setPaid(''); setGame(''); setQ('')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-headline-md text-on-surface">Tournaments</h1>
          <p className="text-body-md text-text-secondary">Find and join upcoming tournaments</p>
        </div>
        <Link to="/create">
          <Button icon={<Trophy className="w-4 h-4" />}>
            Create Tournament
          </Button>
        </Link>
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-3 flex-wrap">
          <div className="flex-1 min-w-[200px] max-w-md">
            <SearchInput
              placeholder="Search by name or game..."
              value={q}
              onChange={e => setQ(e.target.value)}
            />
          </div>
          <select
            className="h-10 px-3 border border-border rounded-lg bg-white text-body-sm"
            value={sort}
            onChange={e => setSort(e.target.value as SortOption)}
          >
            {SORTS.map(s => <option key={s.value} value={s.value}>{s.label}</option>)}
          </select>
          <div className="flex border border-border rounded-lg overflow-hidden">
            <button
              onClick={() => setViewMode('grid')}
              className={`p-2 ${viewMode === 'grid' ? 'bg-surface-container text-on-surface' : 'bg-white text-text-secondary hover:text-on-surface'}`}
            >
              <Grid3X3 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`p-2 ${viewMode === 'list' ? 'bg-surface-container text-on-surface' : 'bg-white text-text-secondary hover:text-on-surface'}`}
            >
              <List className="w-4 h-4" />
            </button>
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          {STATUS_FILTERS.map(s => (
            <FilterChip key={s.value || 'all-status'} active={status === s.value} onClick={() => setStatus(s.value)}>
              {s.label}
            </FilterChip>
          ))}
          <span className="w-px bg-border mx-1" />
          {PAID_FILTERS.map(p => (
            <FilterChip key={p.value || 'all-paid'} active={paid === p.value} onClick={() => setPaid(p.value)}>
              {p.label}
            </FilterChip>
          ))}
          {games.length > 0 && (
            <>
              <span className="w-px bg-border mx-1" />
              <select
                className="h-8 px-2 text-label-md border border-border rounded-full bg-white"
                value={game}
                onChange={e => setGame(e.target.value)}
              >
                <option value="">All games</option>
                {games.map(g => <option key={g} value={g}>{g}</option>)}
              </select>
            </>
          )}
          {activeFilters > 0 && (
            <button onClick={clearAll} className="flex items-center gap-1 text-label-md text-primary hover:text-primary-600 px-2 py-1">
              <X className="w-3 h-3" /> Clear filters
            </button>
          )}
        </div>
      </div>

      {loading ? (
        <div className="py-16 text-center text-text-secondary">Loading...</div>
      ) : error ? (
        <EmptyState
          title="Couldn't load tournaments"
          description={error}
          action={{ label: 'Try again', onClick: () => window.location.reload() }}
        />
      ) : tournaments.length === 0 ? (
        <EmptyState
          title="No tournaments match"
          description={q || activeFilters ? 'Try adjusting your filters' : 'Be the first to create one'}
          action={{ label: 'Create tournament', onClick: () => { window.location.href = '/create' } }}
        />
      ) : viewMode === 'grid' ? (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {tournaments.map(t => (
            <Link key={t.id} to={`/tournaments/${t.id}`}>
              <Card hover padding="lg" className="h-full group">
                <div className="flex items-center justify-between mb-3">
                  <Badge status={t.status} />
                  <span className="text-label-sm text-text-secondary">{t.format.replace(/_/g, ' ')}</span>
                </div>
                <h3 className="text-title-lg text-on-surface mb-1 group-hover:text-primary transition-colors">{t.title}</h3>
                <p className="text-body-sm text-text-secondary mb-4">{t.game}</p>
                <div className="space-y-2 text-body-sm text-text-secondary">
                  <div className="flex items-center gap-2">
                    <Users className="w-4 h-4" />
                    <span>up to {t.max_participants} players</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <MapPin className="w-4 h-4" />
                    <span>{t.location}</span>
                  </div>
                </div>
                <div className="flex items-center justify-between pt-4 mt-4 border-t border-border">
                  <div>
                    <p className="text-label-sm text-text-secondary">Prize Pool</p>
                    <p className="text-title-lg font-bold text-on-surface">{t.prize_pool.toLocaleString()} {t.currency}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-label-sm text-text-secondary">{t.entry_fee > 0 ? 'Entry' : 'Free'}</p>
                    <p className="text-body-sm font-medium text-on-surface">
                      {t.entry_fee > 0 ? `${t.entry_fee.toLocaleString()} ${t.currency}` : '—'}
                    </p>
                  </div>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      ) : (
        <div className="space-y-2">
          {tournaments.map(t => (
            <Link key={t.id} to={`/tournaments/${t.id}`}>
              <div className="flex items-center gap-4 p-4 bg-white rounded-xl border border-border hover:shadow-card-hover hover:border-border/80 transition-all duration-200 group">
                <div className="w-12 h-12 rounded-lg bg-primary-50 flex items-center justify-center flex-shrink-0">
                  <Trophy className="w-6 h-6 text-primary" />
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="text-title-lg text-on-surface group-hover:text-primary transition-colors truncate">{t.title}</h3>
                  <p className="text-body-sm text-text-secondary">{t.game} · {t.format.replace(/_/g, ' ')} · {t.location}</p>
                </div>
                <div className="hidden sm:flex items-center gap-6">
                  <div className="text-center">
                    <p className="text-label-sm text-text-secondary">Max</p>
                    <p className="text-body-sm font-medium">{t.max_participants}</p>
                  </div>
                  <div className="text-center">
                    <p className="text-label-sm text-text-secondary">Prize</p>
                    <p className="text-body-sm font-medium">{t.prize_pool.toLocaleString()} {t.currency}</p>
                  </div>
                </div>
                <Badge status={t.status} />
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

function FilterChip({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={`px-3 py-1 rounded-full text-label-md transition-colors ${
        active
          ? 'bg-primary text-white'
          : 'bg-surface-container-low text-text-secondary hover:text-on-surface hover:bg-surface-container'
      }`}
    >
      {children}
    </button>
  )
}
