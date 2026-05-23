import { Link } from "react-router-dom";
import { ShieldAlert, Crosshair, BrainCircuit } from "lucide-react";

export default function LandingPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[70vh] text-center space-y-12">
      <div className="space-y-6 max-w-3xl">
        <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-cyan-400">
          Master Cyber Defense
        </h1>
        <p className="text-xl text-slate-300">
          Train like a real SOC Analyst. Master CompTIA Security+ 701 through data-driven scenarios and AI-powered feedback.
        </p>
        <div className="pt-4 flex justify-center gap-4">
          <Link to="/login" className="px-8 py-3 rounded-md bg-emerald-500 text-slate-950 font-bold hover:bg-emerald-400 transition-colors">
            Start Free Assessment
          </Link>
          <Link to="/exam" className="px-8 py-3 rounded-md border border-slate-700 hover:border-slate-500 transition-colors">
            View Sample Exam
          </Link>
        </div>
      </div>

      <div className="grid md:grid-cols-3 gap-8 pt-12 border-t border-slate-800/50">
        <div className="p-6 rounded-xl bg-slate-900/50 border border-slate-800 flex flex-col items-center text-center space-y-4">
          <div className="p-3 bg-emerald-500/10 text-emerald-400 rounded-full">
            <ShieldAlert className="w-8 h-8" />
          </div>
          <h3 className="text-xl font-bold">CompTIA Sec+ Ready</h3>
          <p className="text-slate-400">100-question randomized pre/post tests covering all 5 core domains.</p>
        </div>
        <div className="p-6 rounded-xl bg-slate-900/50 border border-slate-800 flex flex-col items-center text-center space-y-4">
          <div className="p-3 bg-cyan-500/10 text-cyan-400 rounded-full">
            <Crosshair className="w-8 h-8" />
          </div>
          <h3 className="text-xl font-bold">SOC Analyst Skills</h3>
          <p className="text-slate-400">Hands-on scenarios based on real-world MITRE ATT&CK techniques.</p>
        </div>
        <div className="p-6 rounded-xl bg-slate-900/50 border border-slate-800 flex flex-col items-center text-center space-y-4">
          <div className="p-3 bg-purple-500/10 text-purple-400 rounded-full">
            <BrainCircuit className="w-8 h-8" />
          </div>
          <h3 className="text-xl font-bold">AI RAG Feedback</h3>
          <p className="text-slate-400">Intelligent post-action analysis and personalized skill gap tracking.</p>
        </div>
      </div>
    </div>
  );
}
