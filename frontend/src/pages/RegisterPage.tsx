import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Shield, ArrowRight, Loader2, Award, BookOpen, Upload, CheckCircle2 } from "lucide-react";
import axios from "axios";

export default function RegisterPage() {
  const navigate = useNavigate();
  const [accountType, setAccountType] = useState<"learner" | "specialist">("learner");
  const [email, setEmail] = useState("");
  const [nickname, setNickname] = useState("");
  const [password, setPassword] = useState("");
  
  // Specialist specific fields
  const [bio, setBio] = useState("");
  const [resumeFile, setResumeFile] = useState<File | null>(null);
  const [certFile, setCertFile] = useState<File | null>(null);
  const [linkedinUrl, setLinkedinUrl] = useState("");
  const [portfolioUrl, setPortfolioUrl] = useState("");

  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [successMessage, setSuccessMessage] = useState("");

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccessMessage("");
    setLoading(true);

    if (accountType === "specialist" && !resumeFile) {
      setError("Please attach your Resume (PDF or image).");
      setLoading(false);
      return;
    }

    if (accountType === "specialist" && !bio.trim()) {
      setError("Please provide a brief bio outlining your cybersecurity experience.");
      setLoading(false);
      return;
    }

    try {
      // 1. Register account
      await axios.post("/api/register", {
        email,
        password,
        nickname: nickname.trim(),
      });

      // 2. If Specialist, log in and upload credentials
      if (accountType === "specialist") {
        const loginRes = await axios.post("/api/login", { email, password });
        const token = loginRes.data.token;
        localStorage.setItem("token", token);
        localStorage.setItem("userEmail", loginRes.data.email);
        localStorage.setItem("userNickname", loginRes.data.nickname);
        localStorage.setItem("userRole", loginRes.data.role);

        const formData = new FormData();
        formData.append("bio", bio);
        if (resumeFile) formData.append("resume", resumeFile);
        if (certFile) formData.append("certificate", certFile);
        if (linkedinUrl) formData.append("linkedin_url", linkedinUrl);
        if (portfolioUrl) formData.append("portfolio_url", portfolioUrl);

        await axios.post("/api/specialist/apply", formData, {
          headers: {
            "Content-Type": "multipart/form-data",
            Authorization: `Bearer ${token}`,
          },
        });

        setSuccessMessage("Account created and Specialist application submitted! Redirecting to profile...");
        setTimeout(() => navigate("/profile"), 1500);
      } else {
        setSuccessMessage("Account created successfully! Redirecting to login...");
        setTimeout(() => navigate("/login"), 1000);
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Registration failed. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[80vh] flex items-center justify-center py-12">
      <div className="w-full max-w-lg bg-graphite-900 border border-graphite-800 rounded-lg p-8 shadow-xl">
        <div className="flex flex-col items-center mb-6">
          <div className="w-12 h-12 bg-graphite-950 text-white rounded border border-graphite-800 flex items-center justify-center mb-4">
            <Shield size={20} />
          </div>
          <h2 className="text-xl font-bold text-white tracking-tight">Create an Account</h2>
          <p className="text-graphite-400 text-xs mt-1 text-center font-light">
            Select your account type to join the CyberSim training environment.
          </p>
        </div>

        {/* Role Selector Tabs */}
        <div className="grid grid-cols-2 gap-2 p-1 bg-graphite-950 border border-graphite-800 rounded-md mb-6">
          <button
            type="button"
            onClick={() => { setAccountType("learner"); setError(""); }}
            className={`flex items-center justify-center gap-2 py-2 px-3 rounded text-xs font-medium transition-all ${
              accountType === "learner"
                ? "bg-white text-black font-semibold shadow-sm"
                : "text-graphite-400 hover:text-white"
            }`}
          >
            <BookOpen size={14} />
            <span>Learner</span>
          </button>
          <button
            type="button"
            onClick={() => { setAccountType("specialist"); setError(""); }}
            className={`flex items-center justify-center gap-2 py-2 px-3 rounded text-xs font-medium transition-all ${
              accountType === "specialist"
                ? "bg-white text-black font-semibold shadow-sm"
                : "text-graphite-400 hover:text-white"
            }`}
          >
            <Award size={14} />
            <span>Specialist Candidate</span>
          </button>
        </div>

        {error && (
          <div className="mb-4 p-3 rounded bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-mono">
            {error}
          </div>
        )}

        {successMessage && (
          <div className="mb-4 p-3 rounded bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-mono flex items-center gap-2">
            <CheckCircle2 size={16} />
            <span>{successMessage}</span>
          </div>
        )}

        <form onSubmit={handleRegister} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-graphite-300 mb-1">
              Nickname (ชื่อเล่น / Public Display Name) <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              required
              minLength={2}
              maxLength={30}
              className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
              placeholder="e.g. shadow_analyst"
            />
            <p className="text-[10px] text-graphite-500 mt-1">This name is visible on discussions and reviews to protect your email.</p>
          </div>

          <div>
            <label className="block text-xs font-medium text-graphite-300 mb-1">
              Email Address <span className="text-red-400">*</span>
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
            <label className="block text-xs font-medium text-graphite-300 mb-1">
              Password <span className="text-red-400">*</span>
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
              className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
              placeholder="••••••••"
            />
          </div>

          {/* Specialist Fields */}
          {accountType === "specialist" && (
            <div className="pt-3 border-t border-graphite-800 space-y-4 animate-fade-in">
              <div className="bg-graphite-950/60 p-3 rounded border border-graphite-800">
                <p className="text-xs text-graphite-300 font-medium mb-1">Specialist Application</p>
                <p className="text-[11px] text-graphite-400 leading-relaxed">
                  Specialist accounts require administrator review. You can still practice scenarios as a Learner while awaiting approval.
                </p>
              </div>

              <div>
                <label className="block text-xs font-medium text-graphite-300 mb-1">
                  Professional Bio / Experience Summary <span className="text-red-400">*</span>
                </label>
                <textarea
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  required={accountType === "specialist"}
                  rows={3}
                  className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
                  placeholder="Describe your cybersecurity background, certifications, and areas of expertise..."
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-graphite-300 mb-1">
                    Resume (PDF/Image) <span className="text-red-400">*</span>
                  </label>
                  <label className="flex flex-col items-center justify-center p-3 border border-dashed border-graphite-700 hover:border-white rounded bg-graphite-950 cursor-pointer transition-colors text-center">
                    <Upload size={16} className="text-graphite-400 mb-1" />
                    <span className="text-[11px] text-graphite-300 truncate max-w-full">
                      {resumeFile ? resumeFile.name : "Choose Resume (<10MB)"}
                    </span>
                    <input
                      type="file"
                      accept=".pdf,.png,.jpg,.jpeg"
                      onChange={(e) => e.target.files && setResumeFile(e.target.files[0])}
                      className="hidden"
                    />
                  </label>
                </div>

                <div>
                  <label className="block text-xs font-medium text-graphite-300 mb-1">
                    Certificate (Optional)
                  </label>
                  <label className="flex flex-col items-center justify-center p-3 border border-dashed border-graphite-700 hover:border-white rounded bg-graphite-950 cursor-pointer transition-colors text-center">
                    <Upload size={16} className="text-graphite-400 mb-1" />
                    <span className="text-[11px] text-graphite-300 truncate max-w-full">
                      {certFile ? certFile.name : "Choose Certificate"}
                    </span>
                    <input
                      type="file"
                      accept=".pdf,.png,.jpg,.jpeg"
                      onChange={(e) => e.target.files && setCertFile(e.target.files[0])}
                      className="hidden"
                    />
                  </label>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-graphite-300 mb-1">
                    LinkedIn URL
                  </label>
                  <input
                    type="url"
                    value={linkedinUrl}
                    onChange={(e) => setLinkedinUrl(e.target.value)}
                    className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-xs text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
                    placeholder="https://linkedin.com/in/..."
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-graphite-300 mb-1">
                    Portfolio / Credly Link
                  </label>
                  <input
                    type="url"
                    value={portfolioUrl}
                    onChange={(e) => setPortfolioUrl(e.target.value)}
                    className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-xs text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
                    placeholder="https://credly.com/..."
                  />
                </div>
              </div>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-white text-black font-semibold py-2.5 rounded text-xs hover:bg-white/90 transition-all flex justify-center items-center gap-2 mt-4 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
          >
            {loading ? <Loader2 className="animate-spin w-4 h-4" /> : accountType === "specialist" ? "Submit Application" : "Register"}
            {!loading && <ArrowRight size={14} />}
          </button>
        </form>

        <div className="mt-6 text-center text-xs text-graphite-400">
          Already have an account?{" "}
          <Link to="/login" className="text-white hover:underline">
            Sign in
          </Link>
        </div>
      </div>
    </div>
  );
}
