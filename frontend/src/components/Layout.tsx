import { Outlet, Link, useNavigate, useLocation } from "react-router-dom";
import { ShieldCheck, LogOut, User, ShieldAlert, Award } from "lucide-react";
import { clsx } from "clsx";
import NotificationBell from "./NotificationBell";
import { useI18n } from "../i18n/index.tsx";

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { locale, setLocale, t } = useI18n();
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
    { path: "/exam", label: t("nav.exam") },
    { path: "/learn", label: t("nav.learn") },
    { path: "/evaluation", label: t("nav.evaluation") },
    { path: "/discussions", label: t("nav.discussions") },
    ...(token && (userRole === "specialist" || userRole === "admin")
      ? [{ path: "/specialist/reviews", label: t("nav.reviews") }]
      : []),
    ...(token ? [{ path: "/profile", label: t("nav.profile") }] : []),
    ...(token && userRole === "admin" ? [{ path: "/admin/specialists", label: t("nav.adminPortal") }] : []),
  ];

  return (
    <div className="min-h-screen flex flex-col relative bg-graphite-950 text-graphite-50 selection:bg-white selection:text-black">
      <header className="sticky top-0 z-50 border-b border-graphite-800 bg-graphite-950/80 backdrop-blur-md">
        <div className="container mx-auto px-6 h-16 flex items-center justify-between gap-4">
          <Link to="/" className="flex items-center gap-2.5 group shrink-0">
            <ShieldCheck className="h-5 w-5 text-white transition-transform group-hover:scale-105" />
            <span className="text-lg font-semibold tracking-tight text-white">
              {t("common.brandPrefix")}<span className="text-graphite-400 font-normal">{t("common.brandSuffix")}</span>
            </span>
          </Link>
          
          <nav className="flex items-center gap-3 xl:gap-5">
            <div className="flex gap-3 xl:gap-5">
              {navLinks.map((link) => {
                const isActive = link.path === "/exam"
                  ? ["/exam", "/pre-test", "/post-test"].includes(location.pathname) ||
                    location.pathname.startsWith("/exams/")
                  : location.pathname.startsWith(link.path);

                return (
                  <Link
                    key={link.path}
                    to={link.path}
                    className={clsx(
                      "text-sm font-medium transition-colors py-1 relative whitespace-nowrap",
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
            
            <div className="h-4 w-px bg-graphite-800 hidden sm:block shrink-0" />

            <div
              role="group"
              aria-label={t("common.languageToggleAria")}
              className="inline-flex items-center rounded-md border border-graphite-800 bg-graphite-900 p-0.5 shrink-0"
            >
              <button
                type="button"
                onClick={() => setLocale("en")}
                aria-pressed={locale === "en"}
                className={clsx(
                  "px-2.5 py-1 rounded text-xs font-mono font-semibold transition-colors cursor-pointer",
                  locale === "en"
                    ? "bg-white text-black"
                    : "text-graphite-400 hover:text-white"
                )}
              >
                EN
              </button>
              <button
                type="button"
                onClick={() => setLocale("th")}
                aria-pressed={locale === "th"}
                className={clsx(
                  "px-2.5 py-1 rounded text-xs font-mono font-semibold transition-colors cursor-pointer",
                  locale === "th"
                    ? "bg-white text-black"
                    : "text-graphite-400 hover:text-white"
                )}
              >
                TH
              </button>
            </div>
            
            {token ? (
              <div className="flex items-center gap-3 shrink-0">
                <NotificationBell />

                <Link to="/profile" className="hidden sm:flex items-center gap-2 text-sm text-graphite-300 hover:text-white transition-colors">
                  <div className="w-7 h-7 rounded-full bg-graphite-900 flex items-center justify-center border border-graphite-800 shrink-0">
                    {userRole === "admin" ? (
                      <ShieldAlert className="w-4 h-4 text-amber-400" />
                    ) : userRole === "specialist" ? (
                      <Award className="w-4 h-4 text-indigo-400" />
                    ) : (
                      <User className="w-4 h-4 text-graphite-400" />
                    )}
                  </div>
                  <span className="font-medium text-sm font-mono text-graphite-200 whitespace-nowrap">
                    {userNickname || userEmail}
                  </span>

                  {/* Role Badge */}
                  {userRole === "admin" && (
                    <span className="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20 whitespace-nowrap shrink-0">
                      {t("common.roles.adminBadge")}
                    </span>
                  )}
                  {userRole === "specialist" && (
                    <span className="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 whitespace-nowrap shrink-0">
                      {t("common.roles.specialistBadge")}
                    </span>
                  )}
                </Link>

                <button 
                  onClick={handleLogout}
                  className="flex items-center gap-2 px-3.5 py-1.5 rounded-md border border-graphite-800 bg-transparent text-graphite-300 text-sm font-medium hover:bg-graphite-900 hover:text-white transition-all duration-200 cursor-pointer whitespace-nowrap"
                >
                  <LogOut className="w-4 h-4" />
                  <span className="hidden sm:inline">{t("nav.logout")}</span>
                </button>
              </div>
            ) : (
              <Link 
                to="/login" 
                className="px-4 py-1.5 rounded-md bg-graphite-100 hover:bg-white text-graphite-950 text-sm font-semibold transition-all duration-200 shrink-0 whitespace-nowrap"
              >
                {t("nav.login")}
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
          <div className="flex items-center gap-2 opacity-90">
            <ShieldCheck className="h-4 w-4 text-white" />
            <span className="font-semibold text-xs tracking-wider text-white uppercase font-mono">
              {t("common.footerBrand")}
            </span>
          </div>
          <p className="text-graphite-400 text-xs font-mono">
            {t("common.copyright", { year: new Date().getFullYear() })}
          </p>
        </div>
      </footer>
    </div>
  );
}
