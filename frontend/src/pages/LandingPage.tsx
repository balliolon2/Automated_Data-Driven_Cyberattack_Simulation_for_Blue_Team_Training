import { Link } from "react-router-dom";
import { ShieldAlert, Crosshair, BrainCircuit, ArrowRight, Activity, Terminal } from "lucide-react";

export default function LandingPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[75vh] w-full max-w-5xl mx-auto animate-slide-up">
      {/* Hero Section */}
      <div className="flex flex-col items-center text-center space-y-8 mt-12 mb-24 relative">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-sm font-medium mb-4 animate-pulse-slow">
          <Activity className="w-4 h-4" />
          <span>v2.0 Beta Environment Live</span>
        </div>
        
        <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight text-white max-w-4xl leading-[1.1]">
          Master Cyber Defense <br className="hidden md:block" />
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 via-cyan-400 to-emerald-300 drop-shadow-sm">
            Through Data-Driven Scenarios
          </span>
        </h1>
        
        <p className="text-xl md:text-2xl text-graphite-300 max-w-2xl font-light leading-relaxed">
          Train like a real SOC Analyst. Master CompTIA Security+ 701 with interactive simulations and AI-powered skill gap analysis.
        </p>
        
        <div className="pt-8 flex flex-col sm:flex-row justify-center gap-5 w-full sm:w-auto">
          <Link 
            to="/login" 
            className="group flex items-center justify-center gap-2 px-8 py-4 rounded-full bg-emerald-500 text-emerald-950 font-bold text-lg shadow-[0_0_20px_rgba(52,211,153,0.25)] hover:shadow-[0_0_35px_rgba(52,211,153,0.4)] hover:bg-emerald-400 transition-all duration-300 hover:-translate-y-1"
          >
            Start Assessment
            <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
          </Link>
          <Link 
            to="/exam" 
            className="flex items-center justify-center gap-2 px-8 py-4 rounded-full glass-panel text-graphite-100 font-medium text-lg hover:bg-graphite-800/80 hover:text-white transition-all duration-300 hover:-translate-y-1"
          >
            <Terminal className="w-5 h-5 text-graphite-400" />
            View Sample Exam
          </Link>
        </div>
      </div>

      {/* Feature Cards Grid */}
      <div className="grid md:grid-cols-3 gap-6 w-full mt-8">
        <div className="group glass-panel rounded-2xl p-8 hover:-translate-y-2 transition-all duration-500 relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
          <div className="relative z-10 flex flex-col items-start text-left space-y-4">
            <div className="p-3 bg-emerald-500/10 ring-1 ring-emerald-500/20 text-emerald-400 rounded-xl shadow-[0_0_15px_rgba(52,211,153,0.15)] group-hover:scale-110 transition-transform duration-500">
              <ShieldAlert className="w-7 h-7" />
            </div>
            <h3 className="text-xl font-bold text-white tracking-wide">CompTIA Sec+ Ready</h3>
            <p className="text-graphite-400 leading-relaxed font-light">
              100-question randomized pre/post tests covering all 5 core domains of the 701 exam architecture.
            </p>
          </div>
        </div>

        <div className="group glass-panel rounded-2xl p-8 hover:-translate-y-2 transition-all duration-500 relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-cyan-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
          <div className="relative z-10 flex flex-col items-start text-left space-y-4">
            <div className="p-3 bg-cyan-500/10 ring-1 ring-cyan-500/20 text-cyan-400 rounded-xl shadow-[0_0_15px_rgba(6,182,212,0.15)] group-hover:scale-110 transition-transform duration-500">
              <Crosshair className="w-7 h-7" />
            </div>
            <h3 className="text-xl font-bold text-white tracking-wide">SOC Analyst Skills</h3>
            <p className="text-graphite-400 leading-relaxed font-light">
              Hands-on scenarios and evaluations based on real-world MITRE ATT&CK adversarial techniques.
            </p>
          </div>
        </div>

        <div className="group glass-panel rounded-2xl p-8 hover:-translate-y-2 transition-all duration-500 relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-purple-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
          <div className="relative z-10 flex flex-col items-start text-left space-y-4">
            <div className="p-3 bg-purple-500/10 ring-1 ring-purple-500/20 text-purple-400 rounded-xl shadow-[0_0_15px_rgba(168,85,247,0.15)] group-hover:scale-110 transition-transform duration-500">
              <BrainCircuit className="w-7 h-7" />
            </div>
            <h3 className="text-xl font-bold text-white tracking-wide">AI RAG Feedback</h3>
            <p className="text-graphite-400 leading-relaxed font-light">
              Intelligent post-action analysis provides personalized insights and tracks your evolving skill gaps over time.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
