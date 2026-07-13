import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { Loader2 } from "lucide-react";

export default function ExamRedirectPage() {
  const navigate = useNavigate();

  useEffect(() => {
    const checkProgressAndRedirect = async () => {
      const token = localStorage.getItem("token");
      if (!token) {
        navigate("/login");
        return;
      }

      try {
        // 1. Check for active exam session first
        const sessionRes = await axios.get("/api/exams/session", {
          headers: { Authorization: `Bearer ${token}` }
        });
        
        if (sessionRes.data && sessionRes.data.session) {
          const type = sessionRes.data.session.exam_type;
          if (type === "pre") {
            navigate("/pre-test");
          } else {
            navigate("/post-test");
          }
          return;
        }
      } catch (err) {
        // No active exam session, proceed to check simulation status
      }

      try {
        // 2. Check simulation status
        const statusRes = await axios.get("/api/simulation/status", {
          headers: { Authorization: `Bearer ${token}` }
        });

        if (statusRes.data.error || statusRes.data.error?.includes("pre-test")) {
          navigate("/pre-test");
        } else if (statusRes.data.needs_training) {
          navigate("/simulation");
        } else {
          navigate("/post-test");
        }
      } catch (err) {
        // Fallback to pre-test if status check fails
        navigate("/pre-test");
      }
    };

    checkProgressAndRedirect();
  }, [navigate]);

  return (
    <div className="min-h-[60vh] flex flex-col items-center justify-center gap-4">
      <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
      <p className="text-sm font-mono text-graphite-400">Redirecting to active assessment stage...</p>
    </div>
  );
}
