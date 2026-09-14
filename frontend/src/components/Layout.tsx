import { Outlet, Link, useNavigate, useLocation } from "react-router-dom";
import { ShieldCheck, LogOut, User, ShieldAlert, Award } from "lucide-react";
import { clsx } from "clsx";
import NotificationBell from "./NotificationBell";

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const token = localStorage.getItem("token");
  const userEmail = localStorage.getItem("userEmail");
  const userNickname = localStorage.getItem("userNickname") || (userEmail ? userEmail.split("@")[0] : "");
  const userRole = localStorage.getItem("userRole") || "learner";

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("userEmail");
    localStorage.removeItem("userNickname");
    localStorage.removeItem("userRole");
    localStorage.removeItem("userId");
    navigate("/login");
  };

  const navLinks = [
    { path: "/exam", label: "Exam" },
    { path: "/evaluation", label: "Evaluation" },
    { path: "/discussions", label: "Discussions" },
    ...(token && (userRole === "specialist" || userRole === "admin")
      ? [{ path: "/specialist/reviews", label: "Reviews" }]
      : []),
    ...(token ? [{ path: "/profile", label: "Profile" }] : []),
    ...(token && userRole === "admin" ? [{ path: "/admin/specialists", label: "Admin Portal" }] : []),
  ];

  return (
    <div className="min-h-screen flex flex-col relative bg-graphite-950 text-graphite-50 selection:bg-white selection:text-black">
      <header className="sticky top-0 z-50 border-b border-graphite-800 bg-graphite-950/80 backdrop-blur-md">
        <div className="container mx-auto px-6 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2.5 group">
            <ShieldCheck className="h-5 w-5 text-white transition-transform group-hover:scale-105" />
            <span className="text-base font-semibold tracking-tight text-white">
              CyberSim<span className="text-graphite-400 font-normal">/SOC</span>
            </span>
          </Link>
          
          <nav className="flex items-center gap-6">
            <div className="flex gap-4 sm:gap-6">
              {navLinks.map((link) => {
                const isActive = link.path === "/exam"
                  ? ["/exam", "/pre-test", "/post-test"].includes(location.pathname)
                  : location.pathname.startsWith(link.path);

                return (
                  <Link
                    key={link.path}
                    to={link.path}
                    className={clsx(
                      "text-sm font-medium transition-colors py-1 relative",
                      isActive
                        ? "text-white"
                        : "text-graphite-400 hover:text-white"
                    )}
                  >
                    {link.label}
                    {isActive && (
                      <span className="absolute bottom-0 left-0 right-0 h-[2px] bg-white rounded-full animate-fade-in" />
                    )}
                  </Link>
                );
              })}
            </div>
            
            <div className="h-4 w-px bg-graphite-800 hidden sm:block" />
            
            {token ? (
              <div className="flex items-center gap-3 sm:gap-4">
                <NotificationBell />

                <Link to="/profile" className="hidden sm:flex items-center gap-2 text-sm text-graphite-300 hover:text-white transition-colors">
                  <div className="w-6 h-6 rounded-full bg-graphite-900 flex items-center justify-center border border-graphite-800">
                    {userRole === "admin" ? (
                      <ShieldAlert className="w-3.5 h-3.5 text-amber-400" />
                    ) : userRole === "specialist" ? (
                      <Award className="w-3.5 h-3.5 text-indigo-400" />
                    ) : (
                      <User className="w-3.5 h-3.5 text-graphite-400" />
                    )}
                  </div>
                  <span className="font-medium text-xs font-mono text-graphite-200">
                    {userNickname || userEmail}
                  </span>

                  {/* Role Badge */}
                  {userRole === "admin" && (
                    <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                      ADMIN
                    </span>
                  )}
                  {userRole === "specialist" && (
                    <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                      SPECIALIST
                    </span>
                  )}
                </Link>

                <button 
                  onClick={handleLogout}
                  className="flex items-center gap-2 px-3 py-1.5 rounded-md border border-graphite-800 bg-transparent text-graphite-300 text-xs font-medium hover:bg-graphite-900 hover:text-white transition-all duration-200 cursor-pointer"
                >
                  <LogOut className="w-3.5 h-3.5" />
                  <span className="hidden sm:inline">Logout</span>
                </button>
              </div>
            ) : (
              <Link 
                to="/login" 
                className="px-4 py-1.5 rounded-md bg-graphite-100 hover:bg-white text-graphite-950 text-xs font-semibold transition-all duration-200"
              >
                Login
              </Link>
            )}
          </nav>
        </div>
      </header>

      <main className="flex-1 w-full max-w-7xl mx-auto px-4 sm:px-6 py-12 relative z-0">
        <Outlet />
      </main>

      <footer className="border-t border-graphite-800 py-8 mt-auto z-10 relative bg-graphite-950">
        <div className="container mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2 opacity-80">
            <ShieldCheck className="h-4 w-4 text-white" />
            <span className="font-semibold text-xs tracking-wider text-white uppercase font-mono">
              CyberSim SOC Trainer
            </span>
          </div>
          <p className="text-graphite-500 text-xs font-mono">
            &copy; {new Date().getFullYear()} CyberSim. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
