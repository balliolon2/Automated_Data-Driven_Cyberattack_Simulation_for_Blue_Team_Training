import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Shield, ArrowRight, Loader2 } from "lucide-react";
import axios from "axios";

export default function LoginPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const response = await axios.post("/api/login", {
        email,
        password,
      });
      // Save user details to localStorage
      localStorage.setItem("token", response.data.token);
      localStorage.setItem("userEmail", response.data.email);
      localStorage.setItem("userNickname", response.data.nickname || response.data.email.split('@')[0]);
      localStorage.setItem("userRole", response.data.role || "learner");
      if (response.data.user_id) {
        localStorage.setItem("userId", response.data.user_id);
      }
      
      if (response.data.role === "admin") {
        navigate("/admin/specialists");
      } else {
        navigate("/exam");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Login failed. Please check your credentials.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[80vh] flex items-center justify-center">
      <div className="w-full max-w-md bg-graphite-900 border border-graphite-800 rounded-lg p-8">
        <div className="flex flex-col items-center mb-8">
          <div className="w-12 h-12 bg-graphite-950 text-white rounded border border-graphite-800 flex items-center justify-center mb-4">
            <Shield size={20} />
          </div>
          <h2 className="text-xl font-bold text-white tracking-tight">System Access</h2>
          <p className="text-graphite-400 text-xs mt-2 text-center font-light leading-relaxed">
            Enter your credentials to access the simulation environment.
          </p>
        </div>

        {error && (
          <div className="mb-4 p-3 rounded bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-mono">
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-graphite-300 mb-1">
              Email Address
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
              placeholder="operator@soc.local"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-graphite-300 mb-1 flex justify-between">
              <span>Password</span>
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-white text-black font-semibold py-2 rounded text-xs hover:bg-white/90 transition-all flex justify-center items-center gap-2 mt-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? <Loader2 className="animate-spin w-4 h-4" /> : "Authenticate"}
            {!loading && <ArrowRight size={14} />}
          </button>
        </form>

        <div className="mt-6 text-center text-xs text-graphite-400">
          Don't have an account?{" "}
          <Link to="/register" className="text-white hover:underline">
            Register here
          </Link>
        </div>
      </div>
    </div>
  );
}
