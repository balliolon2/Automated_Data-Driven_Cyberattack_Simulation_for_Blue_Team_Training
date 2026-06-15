import { Outlet, Link, useNavigate, useLocation } from "react-router-dom";
import { ShieldCheck, LogOut, User } from "lucide-react";
import { clsx } from "clsx";

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const token = localStorage.getItem("token");
  const userEmail = localStorage.getItem("userEmail");

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("userEmail");
    navigate("/login");
  };

  const navLinks = [
    { path: "/exam", label: "Exam" },
    { path: "/evaluation", label: "Evaluation" }
  ];

  return (
    <div className="min-h-screen flex flex-col relative">
      {/* Background ambient glow */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[300px] bg-emerald-500/10 blur-[120px] rounded-full pointer-events-none -z-10" />

      <header className="sticky top-0 z-50 glass-panel border-x-0 border-t-0 rounded-none bg-graphite-900/40">
        <div className="container mx-auto px-6 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 group">
            <div className="relative flex items-center justify-center w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/20 group-hover:bg-emerald-500/20 transition-colors">
              <ShieldCheck className="h-5 w-5 text-emerald-400 drop-shadow-[0_0_8px_rgba(52,211,153,0.8)]" />
            </div>
            <span className="text-xl font-bold tracking-tight text-graphite-50 group-hover:text-white transition-colors">
              Cyber<span className="text-emerald-400">Sim</span>
            </span>
          </Link>
          
          <nav className="flex items-center gap-6">
            <div className="flex gap-1 bg-graphite-900/50 p-1 rounded-full border border-graphite-800">
              {navLinks.map((link) => (
                <Link
                  key={link.path}
                  to={link.path}
                  className={clsx(
                    "px-4 py-1.5 rounded-full text-sm font-medium transition-all duration-300",
                    location.pathname === link.path
                      ? "bg-graphite-800 text-white shadow-sm"
                      : "text-graphite-400 hover:text-graphite-100 hover:bg-graphite-800/50"
                  )}
                >
                  {link.label}
                </Link>
              ))}
            </div>
            
            <div className="h-6 w-px bg-graphite-800 hidden sm:block" />
            
            {token ? (
              <div className="flex items-center gap-4">
                {userEmail && (
                  <div className="hidden sm:flex items-center gap-2 text-sm text-graphite-300">
                    <div className="w-7 h-7 rounded-full bg-graphite-800 flex items-center justify-center border border-graphite-700">
                      <User className="w-4 h-4 text-emerald-400" />
                    </div>
                    <span className="font-medium">{userEmail}</span>
                  </div>
                )}
                <button 
                  onClick={handleLogout}
                  className="flex items-center gap-2 px-4 py-2 rounded-full border border-graphite-700 bg-graphite-800/50 text-graphite-300 text-sm font-medium hover:bg-red-500/10 hover:text-red-400 hover:border-red-500/30 transition-all duration-300"
                >
                  <LogOut className="w-4 h-4" />
                  <span className="hidden sm:inline">Logout</span>
                </button>
              </div>
            ) : (
              <Link 
                to="/login" 
                className="px-5 py-2 rounded-full bg-emerald-500 text-emerald-950 text-sm font-bold shadow-[0_0_15px_rgba(52,211,153,0.3)] hover:shadow-[0_0_25px_rgba(52,211,153,0.5)] hover:bg-emerald-400 transition-all duration-300 hover:-translate-y-0.5"
              >
                Login
              </Link>
            )}
          </nav>
        </div>
      </header>

      <main className="flex-1 w-full max-w-7xl mx-auto px-4 sm:px-6 py-8 animate-fade-in relative z-0">
        <Outlet />
      </main>

      <footer className="border-t border-graphite-800 py-8 mt-auto z-10 relative bg-graphite-950/80">
        <div className="container mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2 opacity-60">
            <ShieldCheck className="h-5 w-5 text-emerald-400" />
            <span className="font-medium text-sm tracking-wide text-graphite-300">
              CyberSim SOC Trainer
            </span>
          </div>
          <p className="text-graphite-500 text-sm">
            &copy; {new Date().getFullYear()} CyberSim. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
