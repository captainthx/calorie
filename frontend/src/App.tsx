import { Routes, Route, Navigate } from "react-router-dom";
import { getToken, getRole } from "./lib/auth";
import TokenPage from "./pages/TokenPage";
import UserPage from "./pages/UserPage";
import AdminPage from "./pages/AdminPage";

interface RouteGuardProps {
  children: React.ReactNode;
  requiredRole: "user" | "admin";
}

function RouteGuard({ children, requiredRole }: RouteGuardProps) {
  const token = getToken();
  const role = getRole(token);

  if (!token || !role) return <TokenPage />;
  if (requiredRole === "admin" && role !== "admin")
    return <Navigate to="/user" replace />;
  if (requiredRole === "user" && role === "admin")
    return <Navigate to="/admin" replace />;
  return <>{children}</>;
}

export default function App() {
  const token = getToken();
  const role = getRole(token);

  if (!token || !role) return <TokenPage />;

  return (
    <Routes>
      <Route
        path="/"
        element={
          <Navigate to={role === "admin" ? "/admin" : "/user"} replace />
        }
      />
      <Route
        path="/user"
        element={
          <RouteGuard requiredRole="user">
            <UserPage />
          </RouteGuard>
        }
      />
      <Route
        path="/admin"
        element={
          <RouteGuard requiredRole="admin">
            <AdminPage />
          </RouteGuard>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
