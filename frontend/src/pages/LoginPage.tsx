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
      // Save token to localStorage
      localStorage.setItem("token", response.data.token);
      // Navigate to the exam page or dashboard
      navigate("/exam");
    } catch (err: any) {
      setError(err.response?.data?.error || "Login failed. Please check your credentials.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[80vh] flex items-center justify-center">
      <div className="w-full max-w-md bg-slate-900/50 border border-slate-800 rounded-2xl p-8 shadow-xl">
        <div className="flex flex-col items-center mb-8">
          <div className="w-16 h-16 bg-blue-600/20 text-blue-500 rounded-full flex items-center justify-center mb-4">
            <Shield size={32} />
          </div>
          <h2 className="text-2xl font-bold text-white">System Access</h2>
          <p className="text-slate-400 text-sm mt-2 text-center">
            Enter your credentials to access the simulation environment.
          </p>
        </div>

        {error && (
          <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1">
              Email Address
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-white transition-all"
              placeholder="operator@soc.local"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-1 flex justify-between">
              <span>Password</span>
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full px-4 py-2 bg-slate-950 border border-slate-800 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-white transition-all"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-2.5 rounded-lg transition-colors flex justify-center items-center gap-2 mt-2 disabled:opacity-70 disabled:cursor-not-allowed"
          >
            {loading ? <Loader2 className="animate-spin w-5 h-5" /> : "Authenticate"}
            {!loading && <ArrowRight size={18} />}
          </button>
        </form>

        <div className="mt-6 text-center text-sm text-slate-400">
          Don't have an account?{" "}
          <Link to="/register" className="text-blue-400 hover:text-blue-300 hover:underline">
            Register here
          </Link>
        </div>
      </div>
    </div>
  );
}
