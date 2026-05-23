import { Radar, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis, ResponsiveContainer } from "recharts";
import { AlertTriangle, CheckCircle } from "lucide-react";

export default function EvaluationPage() {
  const data = [
    { subject: 'Domain 1: Threats, Attacks', A: 85, fullMark: 100 },
    { subject: 'Domain 2: Architecture', A: 65, fullMark: 100 },
    { subject: 'Domain 3: Implementation', A: 90, fullMark: 100 },
    { subject: 'Domain 4: Operations & IR', A: 45, fullMark: 100 },
    { subject: 'Domain 5: Governance', A: 75, fullMark: 100 },
  ];

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold">Exam Evaluation</h1>
        <p className="text-slate-400">Pre-test Assessment Results</p>
      </div>

      <div className="grid md:grid-cols-2 gap-8">
        <div className="bg-slate-900/50 border border-slate-800 rounded-xl p-6 flex flex-col items-center justify-center">
          <div className="text-5xl font-black text-emerald-400 mb-2">720</div>
          <div className="text-slate-400">Total Score (Scale 100-900)</div>
          <div className="mt-4 px-4 py-1 rounded-full bg-amber-500/10 text-amber-500 border border-amber-500/20 text-sm font-medium">
            Threshold: 750 (Not Passed)
          </div>
        </div>

        <div className="bg-slate-900/50 border border-slate-800 rounded-xl p-6 h-80">
          <ResponsiveContainer width="100%" height="100%">
            <RadarChart cx="50%" cy="50%" outerRadius="80%" data={data}>
              <PolarGrid stroke="#334155" />
              <PolarAngleAxis dataKey="subject" tick={{ fill: '#94a3b8', fontSize: 12 }} />
              <PolarRadiusAxis angle={30} domain={[0, 100]} tick={false} axisLine={false} />
              <Radar name="Score" dataKey="A" stroke="#10b981" fill="#10b981" fillOpacity={0.4} />
            </RadarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="bg-slate-900/50 border border-slate-800 rounded-xl p-6 space-y-4">
        <h3 className="text-xl font-bold flex items-center gap-2">
          <AlertTriangle className="text-amber-500" />
          Skill Gap Analysis
        </h3>
        <ul className="space-y-3 text-slate-300">
          <li className="flex gap-3">
            <CheckCircle className="text-emerald-500 shrink-0 w-5 h-5" />
            <span><strong>Strength:</strong> You have an excellent understanding of Security Implementation and Architecture.</span>
          </li>
          <li className="flex gap-3">
            <AlertTriangle className="text-amber-500 shrink-0 w-5 h-5" />
            <span><strong>Weakness:</strong> Domain 4 (Operations & IR) is significantly below the threshold (45%). We recommend focusing on Incident Response procedures.</span>
          </li>
          <li className="flex gap-3">
            <AlertTriangle className="text-amber-500 shrink-0 w-5 h-5" />
            <span><strong>Recommendation:</strong> Complete the "Failed Login Anomaly" and "Ransomware Containment" SOC Scenarios to build practical IR skills.</span>
          </li>
        </ul>
        <div className="pt-4 mt-4 border-t border-slate-800 flex justify-end">
          <button className="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-md transition-colors">
            Start Recommended Scenarios
          </button>
        </div>
      </div>
    </div>
  );
}
