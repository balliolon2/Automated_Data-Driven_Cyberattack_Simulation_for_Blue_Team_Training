import { Outlet, Link, useNavigate } from "react-router-dom";
import { ShieldCheck, LogOut } from "lucide-react";

export default function Layout() {
  const navigate = useNavigate();
  const token = localStorage.getItem("token");

  const handleLogout = () => {
    localStorage.removeItem("token");
    navigate("/login");
  };

  return (
    <div className="min-h-screen flex flex-col bg-slate-950 text-slate-50">
      <header className="border-b border-slate-800 bg-slate-900/50 backdrop-blur-md sticky top-0 z-50">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 text-xl font-bold text-emerald-400">
            <ShieldCheck className="h-6 w-6" />
            CyberSim
          </Link>
          <nav className="flex gap-4 items-center">
            <Link to="/exam" className="hover:text-emerald-400 transition-colors">Exam</Link>
            <Link to="/evaluation" className="hover:text-emerald-400 transition-colors">Evaluation</Link>
            
            {token ? (
              <button 
                onClick={handleLogout}
                className="flex items-center gap-2 px-4 py-2 rounded-md bg-slate-800 text-slate-200 font-medium hover:bg-slate-700 transition-colors ml-4"
              >
                <LogOut className="w-4 h-4" />
                Logout
              </button>
            ) : (
              <Link to="/login" className="px-4 py-2 rounded-md bg-emerald-500 text-slate-950 font-medium hover:bg-emerald-400 transition-colors ml-4">
                Login
              </Link>
            )}
          </nav>
        </div>
      </header>
      <main className="flex-1 container mx-auto px-4 py-8">
        <Outlet />
      </main>
      <footer className="border-t border-slate-800 py-6 text-center text-slate-400 text-sm">
        &copy; 2026 CyberSim. All rights reserved.
      </footer>
    </div>
  );
}
