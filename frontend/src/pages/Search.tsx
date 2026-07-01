import { useEffect, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Search as SearchIcon, Trophy, Users } from 'lucide-react'
import { Avatar } from '@/components/ui/Avatar'
import { Badge } from '@/components/ui/Badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { EmptyState } from '@/components/ui/EmptyState'
import { SearchInput } from '@/components/ui/Input'
import { api, ApiError, type Tournament } from '@/lib/api'

interface UserResult {
  id: string
  username: string
  display_name?: string | null
  avatar_url?: string | null
  role: string
}

export function SearchPage() {
  const [params, setParams] = useSearchParams()
  const initialQuery = params.get('q') || ''
  const [q, setQ] = useState(initialQuery)
  const [users, setUsers] = useState<UserResult[]>([])
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => setQ(initialQuery), [initialQuery])

  useEffect(() => {
    let cancelled = false
    const query = initialQuery.trim()
    if (!query) {
      setUsers([])
      setTournaments([])
      setLoading(false)
      setError(null)
      return
    }

    setLoading(true)
    setError(null)
    const id = window.setTimeout(() => {
      Promise.all([
        api.get<UserResult[]>(`/api/v1/users/search?q=${encodeURIComponent(query)}&limit=20`).catch(err => {
          throw err
        }),
        api.get<Tournament[]>(`/api/v1/tournaments?q=${encodeURIComponent(query)}&limit=20`).catch(err => {
          throw err
        }),
      ])
        .then(([userData, tournamentData]) => {
          if (cancelled) return
          setUsers(userData || [])
          setTournaments(tournamentData || [])
        })
        .catch(err => {
          if (cancelled) return
          setError(err instanceof ApiError ? err.message : 'Search failed')
        })
        .finally(() => { if (!cancelled) setLoading(false) })
    }, 200)

    return () => {
      cancelled = true
      window.clearTimeout(id)
    }
  }, [initialQuery])

  const hasResults = users.length > 0 || tournaments.length > 0
  const resultCount = useMemo(() => users.length + tournaments.length, [users.length, tournaments.length])

  function submit(e: React.FormEvent) {
    e.preventDefault()
    const next = q.trim()
    if (next) {
      setParams({ q: next })
    } else {
      setParams({})
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3">
        <div>
          <h1 className="text-headline-md text-on-surface">Search</h1>
          <p className="text-body-md text-text-secondary">Find real players by username and tournaments by title or game.</p>
        </div>
        <form onSubmit={submit} className="max-w-xl">
          <SearchInput value={q} onChange={e => setQ(e.target.value)} placeholder="Search usernames, players, tournaments, games..." />
        </form>
      </div>

      {loading ? (
        <div className="py-16 text-center text-text-secondary">Searching...</div>
      ) : error ? (
        <EmptyState title="Search failed" description={error} />
      ) : !initialQuery.trim() ? (
        <EmptyState icon={<SearchIcon className="w-8 h-8" />} title="Start a search" description="Type a username, player name, tournament title, or game." />
      ) : !hasResults ? (
        <EmptyState title="No real records found" description="Try another username, player name, tournament, or game." />
      ) : (
        <>
          <p className="text-body-sm text-text-secondary">{resultCount} result{resultCount === 1 ? '' : 's'} for "{initialQuery}"</p>

          <div className="grid lg:grid-cols-2 gap-6">
            <Card padding="lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2"><Users className="w-5 h-5" /> Players</CardTitle>
              </CardHeader>
              <CardContent>
                {users.length === 0 ? (
                  <EmptyState title="No players found" className="py-8" />
                ) : (
                  <div className="space-y-2">
                    {users.map(u => (
                      <Link key={u.id} to={`/u/${u.username}`} className="flex items-center gap-3 p-3 rounded-lg border border-border hover:border-primary/50 hover:bg-surface-container-low transition-all">
                        <Avatar src={u.avatar_url || undefined} name={u.display_name || u.username} size="sm" />
                        <div className="flex-1 min-w-0">
                          <p className="text-body-sm font-medium text-on-surface truncate">{u.display_name || u.username}</p>
                          <p className="text-label-sm text-text-secondary">@{u.username}</p>
                        </div>
                        <Badge>{u.role}</Badge>
                      </Link>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            <Card padding="lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2"><Trophy className="w-5 h-5" /> Tournaments</CardTitle>
              </CardHeader>
              <CardContent>
                {tournaments.length === 0 ? (
                  <EmptyState title="No tournaments found" className="py-8" />
                ) : (
                  <div className="space-y-2">
                    {tournaments.map(t => (
                      <Link key={t.id} to={`/tournaments/${t.id}`} className="block p-3 rounded-lg border border-border hover:border-primary/50 hover:bg-surface-container-low transition-all">
                        <div className="flex items-center justify-between gap-3">
                          <div className="min-w-0">
                            <p className="text-body-sm font-medium text-on-surface truncate">{t.title}</p>
                            <p className="text-label-sm text-text-secondary">{t.game} · {t.location}</p>
                          </div>
                          <Badge status={t.status}>{t.status.replace(/_/g, ' ')}</Badge>
                        </div>
                      </Link>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </div>
  )
}
