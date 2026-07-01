import { useEffect, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { api } from '@/lib/api'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/Button'
import { Avatar } from '@/components/ui/Avatar'
import { SearchInput } from '@/components/ui/Input'
import { useAuth } from '@/contexts/AuthContext'
import {
  Trophy,
  LayoutDashboard,
  Users,
  Plus,
  Bell,
  ChevronDown,
  Shield,
} from 'lucide-react'

const navLinks = [
  { href: '/tournaments', label: 'Tournaments', icon: Trophy },
  { href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { href: '/participants', label: 'Participants', icon: Users },
]

export function Navbar() {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, session, signOut } = useAuth()
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [unreadCount, setUnreadCount] = useState(0)
  const [search, setSearch] = useState('')

  useEffect(() => {
    if (!session) return
    let cancelled = false
    const load = () => {
      api.get<{ count: number }>('/api/v1/notifications/unread-count')
        .then(r => { if (!cancelled) setUnreadCount(r.count || 0) })
        .catch(() => {})
    }
    load()
    const id = window.setInterval(load, 30000)
    return () => { cancelled = true; window.clearInterval(id) }
  }, [session, location.pathname])

  const displayName = user?.user_metadata?.display_name || user?.email?.split('@')[0] || 'Guest'
  const email = user?.email ?? ''

  async function handleSignOut() {
    setShowUserMenu(false)
    try { await signOut() } catch {}
    navigate('/login')
  }

  function submitSearch(e: React.FormEvent) {
    e.preventDefault()
    const q = search.trim()
    if (!q) return
    navigate(`/search?q=${encodeURIComponent(q)}`)
    setSearch('')
  }

  return (
    <nav className="sticky top-0 z-50 h-nav bg-white/80 backdrop-blur-md border-b border-border">
      <div className="max-w-container mx-auto flex items-center justify-between h-full px-gutter">
        <div className="flex items-center gap-8">
          <Link to="/" className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
              <Trophy className="w-4 h-4 text-white" />
            </div>
            <span className="text-title-lg font-bold text-on-surface">Bona</span>
          </Link>

          {session && (
            <div className="hidden md:flex items-center gap-1">
              {navLinks.map(link => (
                <Link
                  key={link.href}
                  to={link.href}
                  className={cn(
                    'flex items-center gap-2 px-3 py-2 rounded-lg text-body-sm font-medium transition-colors duration-150',
                    location.pathname.startsWith(link.href)
                      ? 'bg-primary-50 text-primary'
                      : 'text-text-secondary hover:text-on-surface hover:bg-surface-container-low'
                  )}
                >
                  <link.icon className="w-4 h-4" />
                  {link.label}
                </Link>
              ))}
            </div>
          )}
        </div>

        <div className="flex items-center gap-3">
          {session ? (
            <>
              <form onSubmit={submitSearch} className="hidden sm:block w-64">
                <SearchInput value={search} onChange={e => setSearch(e.target.value)} placeholder="Search players or tournaments..." />
              </form>

              <Link to="/create">
                <Button size="sm" icon={<Plus className="w-4 h-4" />}>
                  <span className="hidden sm:inline">Create</span>
                </Button>
              </Link>

              <Link
                to="/notifications"
                className="relative p-2 rounded-lg text-text-secondary hover:text-on-surface hover:bg-surface-container-low transition-colors"
                aria-label="Notifications"
              >
                <Bell className="w-5 h-5" />
                {unreadCount > 0 && (
                  <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-danger text-white text-[10px] font-bold flex items-center justify-center">
                    {unreadCount > 9 ? '9+' : unreadCount}
                  </span>
                )}
              </Link>

              <div className="relative">
                <button
                  onClick={() => setShowUserMenu(!showUserMenu)}
                  className="flex items-center gap-2 p-1.5 rounded-lg hover:bg-surface-container-low transition-colors"
                >
                  <Avatar name={displayName} size="sm" />
                  <ChevronDown className="w-4 h-4 text-text-secondary hidden sm:block" />
                </button>

                {showUserMenu && (
                  <>
                    <div className="fixed inset-0 z-10" onClick={() => setShowUserMenu(false)} />
                    <div className="absolute right-0 top-full mt-1 w-56 bg-white rounded-xl border border-border shadow-dropdown z-20 py-1 animate-fade-in">
                      <div className="px-4 py-2 border-b border-border">
                        <p className="text-body-sm font-medium text-on-surface truncate">{displayName}</p>
                        <p className="text-body-sm text-text-secondary truncate">{email}</p>
                      </div>
                      <Link to="/profile" onClick={() => setShowUserMenu(false)} className="block px-4 py-2 text-body-sm text-text-secondary hover:bg-surface-container-low hover:text-on-surface">My profile</Link>
                      <Link to="/dashboard" onClick={() => setShowUserMenu(false)} className="block px-4 py-2 text-body-sm text-text-secondary hover:bg-surface-container-low hover:text-on-surface">Dashboard</Link>
                      <Link to="/me/dashboard" onClick={() => setShowUserMenu(false)} className="block px-4 py-2 text-body-sm text-text-secondary hover:bg-surface-container-low hover:text-on-surface">Player dashboard</Link>
                      <Link to="/me/payouts" onClick={() => setShowUserMenu(false)} className="block px-4 py-2 text-body-sm text-text-secondary hover:bg-surface-container-low hover:text-on-surface">My payouts</Link>
                      <Link to="/admin" onClick={() => setShowUserMenu(false)} className="flex items-center gap-2 px-4 py-2 text-body-sm text-text-secondary hover:bg-surface-container-low hover:text-on-surface">
                        <Shield className="h-4 w-4" />
                        Admin
                      </Link>
                      <div className="border-t border-border mt-1 pt-1">
                        <button onClick={handleSignOut} className="w-full text-left px-4 py-2 text-body-sm text-danger hover:bg-danger-50">Sign out</button>
                      </div>
                    </div>
                  </>
                )}
              </div>
            </>
          ) : (
            <Link to="/login">
              <Button size="sm" variant="primary">Sign in</Button>
            </Link>
          )}
        </div>
      </div>
    </nav>
  )
}
