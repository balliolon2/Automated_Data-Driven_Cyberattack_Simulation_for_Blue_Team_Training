import { useState, useEffect, useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import axios from "axios";
import { Bell, CheckCheck, Clock } from "lucide-react";
import { cn } from "../lib/utils";
import { useI18n } from "../i18n";
import { resolveNotificationLocale } from "../i18n/format";

interface NotificationItem {
  notification_id: string;
  title: string;
  message: string;
  template_key?: string;
  template_params?: Record<string, unknown>;
  link_url: string;
  is_read: boolean;
  created_at: string;
}

export default function NotificationBell() {
  const { locale, t, formatDate } = useI18n();
  const navigate = useNavigate();
  const location = useLocation();
  const token = localStorage.getItem("token");

  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [unreadCount, setUnreadCount] = useState<number>(0);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!token) return;
    fetchNotifications();

    // Re-fetch notifications every 45s
    const interval = setInterval(fetchNotifications, 45000);
    return () => clearInterval(interval);
  }, [token]);

  // Re-fetch on route navigation
  useEffect(() => {
    if (token) {
      fetchNotifications();
    }
  }, [location.pathname]);

  // Close popover when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (popoverRef.current && !popoverRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen]);

  const fetchNotifications = async () => {
    if (!token) return;
    try {
      const res = await axios.get("/api/notifications?page_size=10", {
        headers: { Authorization: `Bearer ${token}` },
      });
      setNotifications(res.data.notifications || []);
      setUnreadCount(res.data.unread_count || 0);
    } catch (err) {
      console.error("Failed to fetch notifications:", err);
    }
  };

  const markAsRead = async (notif: NotificationItem) => {
    if (!token) return;
    try {
      if (!notif.is_read) {
        await axios.put(
          `/api/notifications/${notif.notification_id}/read`,
          {},
          { headers: { Authorization: `Bearer ${token}` } }
        );
        setNotifications((prev) =>
          prev.map((n) =>
            n.notification_id === notif.notification_id ? { ...n, is_read: true } : n
          )
        );
        setUnreadCount((prev) => Math.max(0, prev - 1));
      }
      setIsOpen(false);
      if (notif.link_url) {
        navigate(notif.link_url);
      }
    } catch (err) {
      console.error("Failed to mark notification as read:", err);
    }
  };

  const markAllAsRead = async () => {
    if (!token || unreadCount === 0) return;
    setLoading(true);
    try {
      await axios.put(
        "/api/notifications/read-all",
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
      setUnreadCount(0);
    } catch (err) {
      console.error("Failed to mark all as read:", err);
    } finally {
      setLoading(false);
    }
  };

  if (!token) return null;

  return (
    <div className="relative" ref={popoverRef}>
      {/* Bell Trigger Button */}
      <button
        onClick={() => {
          setIsOpen(!isOpen);
          if (!isOpen) fetchNotifications();
        }}
        aria-label={t("notifications.ariaLabel")}
        className="relative p-1.5 rounded-md text-graphite-400 hover:text-white hover:bg-graphite-900 transition-colors cursor-pointer"
      >
        <Bell className="w-4 h-4" />
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 flex h-5 min-w-[20px] px-1 items-center justify-center rounded-full bg-red-500 text-xs font-mono font-bold text-white shadow-sm animate-pulse leading-none">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </button>

      {/* Popover Dropdown */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 sm:w-96 rounded-xl bg-graphite-950 border border-graphite-800 shadow-2xl z-50 overflow-hidden animate-fade-in text-sm">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-graphite-800 bg-graphite-900/70">
            <div className="flex items-center gap-2">
              <span className="font-bold text-white text-sm">{t("notifications.title")}</span>
              {unreadCount > 0 && (
                <span className="px-2 py-0.5 rounded bg-indigo-500/20 text-indigo-300 font-mono text-xs font-semibold">
                  {t("notifications.newBadge", { count: unreadCount })}
                </span>
              )}
            </div>

            {unreadCount > 0 && (
              <button
                onClick={markAllAsRead}
                disabled={loading}
                className="flex items-center gap-1 text-xs text-graphite-400 hover:text-white transition-colors cursor-pointer disabled:opacity-50"
              >
                <CheckCheck className="w-4 h-4" />
                <span>{t("notifications.markAllRead")}</span>
              </button>
            )}
          </div>

          {/* List */}
          <div className="max-h-96 overflow-y-auto divide-y divide-graphite-800/60">
            {notifications.length === 0 ? (
              <div className="py-10 text-center text-graphite-400 font-mono text-sm">
                {t("notifications.empty")}
              </div>
            ) : (
              notifications.map((notif) => {
                const localized = resolveNotificationLocale(notif, locale);
                return (
                  <div
                    key={notif.notification_id}
                    onClick={() => markAsRead(notif)}
                    className={cn(
                      "p-4 hover:bg-graphite-900/60 transition-colors cursor-pointer space-y-1.5 relative",
                      !notif.is_read && "bg-indigo-950/15"
                    )}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex items-center gap-1.5 min-w-0">
                        {!notif.is_read && (
                          <span className="w-2 h-2 rounded-full bg-indigo-400 shrink-0" />
                        )}
                        <span
                          className={cn(
                            "text-sm font-medium leading-snug",
                            notif.is_read ? "text-graphite-300" : "text-white font-semibold"
                          )}
                        >
                          {localized.title}
                        </span>
                      </div>

                      <div className="flex items-center gap-1 text-xs font-mono text-graphite-400 shrink-0">
                        <Clock className="w-3.5 h-3.5" />
                        <span>{formatDate(notif.created_at)}</span>
                      </div>
                    </div>

                    <p className="text-xs sm:text-sm text-graphite-300 leading-relaxed line-clamp-2">
                      {localized.message}
                    </p>
                  </div>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
}
