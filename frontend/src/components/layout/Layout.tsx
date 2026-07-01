import { Outlet } from 'react-router-dom'
import { Navbar } from './Navbar'

export function Layout() {
  return (
    <div className="min-h-screen bg-surface">
      <Navbar />
      <main className="max-w-container mx-auto px-gutter py-6">
        <Outlet />
      </main>
    </div>
  )
}
