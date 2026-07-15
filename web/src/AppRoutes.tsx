import { Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import TodayPage from './pages/TodayPage'
import WorkoutHistoryPage from './pages/WorkoutHistoryPage'
import WorkoutDetailPage from './pages/WorkoutDetailPage'
import ProgramDetailPage from './pages/ProgramDetailPage'
import ExercisesPage from './pages/ExercisesPage'
import ProgressPage from './pages/ProgressPage'
import ProfilePage from './pages/ProfilePage'
import RequireAuth from './components/RequireAuth'
import AppShell from './components/AppShell'

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        element={
          <RequireAuth>
            <AppShell />
          </RequireAuth>
        }
      >
        <Route path="/" element={<TodayPage />} />
        <Route path="/workouts" element={<WorkoutHistoryPage />} />
        <Route path="/workout/:id" element={<WorkoutDetailPage />} />
        <Route path="/program/:id" element={<ProgramDetailPage />} />
        <Route path="/exercises" element={<ExercisesPage />} />
        <Route path="/progress" element={<ProgressPage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
