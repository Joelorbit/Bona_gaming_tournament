import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { LandingPage } from '@/pages/Landing'
import { Login } from '@/pages/Login'
import { ForgotPassword } from '@/pages/ForgotPassword'
import { ResetPassword } from '@/pages/ResetPassword'
import { BrowseTournaments } from '@/pages/BrowseTournaments'
import { TournamentDetails } from '@/pages/TournamentDetails'
import { OrganizerDashboard } from '@/pages/OrganizerDashboard'
import { CreateTournament } from '@/pages/CreateTournament'
import { BracketView } from '@/pages/BracketView'
import { Participants } from '@/pages/Participants'
import { Profile } from '@/pages/Profile'
import { EditProfile } from '@/pages/EditProfile'
import { Notifications } from '@/pages/Notifications'
import { MatchDetail } from '@/pages/MatchDetail'
import { PlayerDashboard } from '@/pages/PlayerDashboard'
import { MePayouts } from '@/pages/MePayouts'
import { Admin } from '@/pages/Admin'
import { PaymentReturn } from '@/pages/PaymentReturn'
import { SearchPage } from '@/pages/Search'
import { useAuth } from '@/contexts/AuthContext'
import type { ReactNode } from 'react'

function RequireAuth({ children }: { children: ReactNode }) {
  const { session, loading } = useAuth()
  const location = useLocation()
  if (loading) return <div className="p-12 text-center text-text-secondary">Loading...</div>
  if (!session) {
    return <Navigate to="/login" replace state={{ from: location }} />
  }
  return <>{children}</>
}

function RedirectIfSignedIn({ children }: { children: ReactNode }) {
  const { session, loading } = useAuth()
  const location = useLocation()
  if (loading) return null
  if (session) {
    const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname
    return <Navigate to={from && from !== '/login' ? from : '/tournaments'} replace />
  }
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<LandingPage />} />
          <Route path="/login" element={<RedirectIfSignedIn><Login /></RedirectIfSignedIn>} />
          <Route path="/forgot-password" element={<RedirectIfSignedIn><ForgotPassword /></RedirectIfSignedIn>} />
          <Route path="/reset-password" element={<ResetPassword />} />
          <Route path="/tournaments" element={<RequireAuth><BrowseTournaments /></RequireAuth>} />
          <Route path="/tournaments/:id" element={<RequireAuth><TournamentDetails /></RequireAuth>} />
          <Route path="/dashboard" element={<RequireAuth><OrganizerDashboard /></RequireAuth>} />
          <Route path="/create" element={<RequireAuth><CreateTournament /></RequireAuth>} />
          <Route path="/tournaments/:id/bracket" element={<RequireAuth><BracketView /></RequireAuth>} />
          <Route path="/tournaments/:id/participants" element={<RequireAuth><Participants /></RequireAuth>} />
          <Route path="/participants" element={<RequireAuth><Participants /></RequireAuth>} />
          <Route path="/search" element={<RequireAuth><SearchPage /></RequireAuth>} />
          <Route path="/profile" element={<RequireAuth><Profile /></RequireAuth>} />
          <Route path="/profile/edit" element={<RequireAuth><EditProfile /></RequireAuth>} />
          <Route path="/u/:username" element={<RequireAuth><Profile /></RequireAuth>} />
          <Route path="/notifications" element={<RequireAuth><Notifications /></RequireAuth>} />
          <Route path="/matches/:id" element={<RequireAuth><MatchDetail /></RequireAuth>} />
          <Route path="/payments/return" element={<RequireAuth><PaymentReturn /></RequireAuth>} />
          <Route path="/me/dashboard" element={<RequireAuth><PlayerDashboard /></RequireAuth>} />
          <Route path="/me/payouts" element={<RequireAuth><MePayouts /></RequireAuth>} />
          <Route path="/admin" element={<RequireAuth><Admin /></RequireAuth>} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
