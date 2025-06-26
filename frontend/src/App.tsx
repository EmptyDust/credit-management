import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { Toaster } from "react-hot-toast";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import Students from "./pages/Students";
import Teachers from "./pages/Teachers";
import Affairs from "./pages/Affairs";
import AffairDetail from "./pages/AffairDetail";
import ActivityEdit from "./pages/ActivityEdit";
import Applications from "./pages/Applications";
import ProfilePage from "./pages/Profile";
import { AuthProvider } from "./contexts/AuthContext";
import { ThemeProvider } from "./contexts/ThemeContext";
import ProtectedRoute from "./components/ProtectedRoute";
import RoleBasedRoute from "./components/RoleBasedRoute";
import Layout from "./components/Layout";

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <AuthProvider>
        <Router>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route element={<ProtectedRoute />}>
              <Route element={<Layout />}>
                <Route path="/dashboard" element={<Dashboard />} />
                <Route
                  path="/students"
                  element={
                    <RoleBasedRoute allowedRoles={["teacher", "admin"]}>
                      <Students />
                    </RoleBasedRoute>
                  }
                />
                <Route
                  path="/teachers"
                  element={
                    <RoleBasedRoute allowedRoles={["admin"]}>
                      <Teachers />
                    </RoleBasedRoute>
                  }
                />
                <Route path="/affairs" element={<Affairs />} />
                <Route path="/affairs/:id" element={<AffairDetail />} />
                <Route path="/affairs/edit/:id" element={<ActivityEdit />} />
                <Route path="/applications" element={<Applications />} />
                <Route path="/profile" element={<ProfilePage />} />
              </Route>
            </Route>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </Router>
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              background: "hsl(var(--background))",
              color: "hsl(var(--foreground))",
              border: "1px solid hsl(var(--border))",
            },
          }}
        />
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
