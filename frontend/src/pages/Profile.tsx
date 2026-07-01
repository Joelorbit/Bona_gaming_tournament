import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Trophy, Users, Award, MapPin, Mail, Edit3, Calendar } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Avatar } from '@/components/ui/Avatar'
import { Badge } from '@/components/ui/Badge'
import { EmptyState } from '@/components/ui/EmptyState'
import { api, type ProfileWithStats } from '@/lib/api'
import { useAuth } from '@/contexts/AuthContext'
import { formatDate } from '@/lib/utils'

export function Profile() {
  const { username } = useParams<{ username?: string }>()
  const { user } = useAuth()
  const [profile, setProfile] = useState<ProfileWithStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    const path = username ? `/api/v1/users/${encodeURIComponent(username)}` : '/api/v1/users/me'
    api.get<ProfileWithStats>(path)
      .then(p => { if (!cancelled) setProfile(p) })
      .catch(err => { if (!cancelled) setError(err?.message || 'Failed to load profile') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [username])

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading profile…</div>
  if (error || !profile) {
    return (
      <EmptyState
        title="Profile not found"
        description={error || 'No profile data available'}
      />
    )
  }

  const isOwnProfile = !username || (user && profile.id === user.id)
  const displayName = profile.display_name || profile.username

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <Card padding="lg">
        <div className="flex flex-col sm:flex-row items-start gap-6">
          <Avatar src={profile.avatar_url || undefined} name={displayName} size="lg" />
          <div className="flex-1">
            <div className="flex items-start justify-between flex-wrap gap-3">
              <div>
                <h1 className="text-headline-md text-on-surface">{displayName}</h1>
                <p className="text-body-md text-text-secondary">@{profile.username}</p>
              </div>
              {isOwnProfile && (
                <Link to="/profile/edit">
                  <Button variant="outline" size="sm" icon={<Edit3 className="w-4 h-4" />}>
                    Edit profile
                  </Button>
                </Link>
              )}
            </div>
            {profile.bio && (
              <p className="text-body-md text-on-surface mt-4">{profile.bio}</p>
            )}
            <div className="flex flex-wrap items-center gap-4 mt-4 text-body-sm text-text-secondary">
              {profile.country && (
                <span className="flex items-center gap-1.5">
                  <MapPin className="w-4 h-4" />
                  {profile.country}
                </span>
              )}
              {profile.email && isOwnProfile && (
                <span className="flex items-center gap-1.5">
                  <Mail className="w-4 h-4" />
                  {profile.email}
                </span>
              )}
              <span className="flex items-center gap-1.5">
                <Calendar className="w-4 h-4" />
                Joined {formatDate(profile.created_at)}
              </span>
              <Badge>{profile.role}</Badge>
            </div>
          </div>
        </div>
      </Card>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard icon={<Users className="w-5 h-5" />} label="Tournaments played" value={profile.stats.tournaments_played} />
        <StatCard icon={<Trophy className="w-5 h-5" />} label="Tournaments hosted" value={profile.stats.tournaments_hosted} />
        <StatCard icon={<Award className="w-5 h-5" />} label="Wins" value={profile.stats.wins} />
      </div>
    </div>
  )
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: number }) {
  return (
    <Card padding="md">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
          {icon}
        </div>
        <div>
          <p className="text-body-sm text-text-secondary">{label}</p>
          <p className="text-headline-sm text-on-surface">{value}</p>
        </div>
      </div>
    </Card>
  )
}
