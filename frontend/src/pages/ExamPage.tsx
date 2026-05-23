import { useState, useEffect } from "react";
import { Clock, Flag, ChevronLeft, ChevronRight, CheckCircle2 } from "lucide-react";

export default function ExamPage() {
  const [currentQuestion, setCurrentQuestion] = useState(1);
  const totalQuestions = 100;
  
  // Timer state (90 mins)
  const [timeLeft, setTimeLeft] = useState(90 * 60);

  useEffect(() => {
    const timer = setInterval(() => {
      setTimeLeft((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  const formatTime = (seconds: number) => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 min-h-[75vh]">
      {/* Sidebar: Navigator */}
      <div className="lg:col-span-1 bg-slate-900/50 border border-slate-800 rounded-xl p-4 flex flex-col">
        <h3 className="font-bold text-lg mb-4">Question Navigator</h3>
        <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar">
          <div className="grid grid-cols-5 gap-2">
            {Array.from({ length: totalQuestions }).map((_, i) => (
              <button
                key={i}
                onClick={() => setCurrentQuestion(i + 1)}
                className={`w-10 h-10 rounded flex items-center justify-center text-sm font-medium border
                  ${currentQuestion === i + 1 
                    ? 'bg-emerald-500/20 border-emerald-500 text-emerald-400' 
                    : 'bg-slate-950 border-slate-700 text-slate-400 hover:border-slate-500'}`}
              >
                {i + 1}
              </button>
            ))}
          </div>
        </div>
        <div className="mt-4 pt-4 border-t border-slate-800">
          <button className="w-full py-2 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-md transition-colors flex items-center justify-center gap-2">
            <CheckCircle2 className="w-5 h-5" />
            Submit Exam
          </button>
        </div>
      </div>

      {/* Main Content: Question */}
      <div className="lg:col-span-3 bg-slate-900/50 border border-slate-800 rounded-xl p-6 flex flex-col">
        <div className="flex justify-between items-center border-b border-slate-800 pb-4 mb-6">
          <h2 className="text-xl font-bold">Question {currentQuestion} of {totalQuestions}</h2>
          <div className="flex items-center gap-2 text-emerald-400 font-mono text-xl bg-slate-950 px-4 py-2 rounded-lg border border-slate-800">
            <Clock className="w-5 h-5" />
            {formatTime(timeLeft)}
          </div>
        </div>

        <div className="flex-1 space-y-8">
          <div className="text-lg">
            A security analyst is reviewing logs and notices repeated failed login attempts from a single IP address followed by a successful login. Which of the following attacks is most likely occurring?
          </div>
          <div className="space-y-3">
            {['Brute force', 'SQL injection', 'Cross-site scripting', 'Pass the hash'].map((opt, i) => (
              <label key={i} className="flex items-center gap-3 p-4 border border-slate-700 rounded-lg cursor-pointer hover:bg-slate-800 transition-colors">
                <input type="radio" name={`q${currentQuestion}`} className="w-5 h-5 accent-emerald-500" />
                <span>{String.fromCharCode(65 + i)}) {opt}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="flex justify-between items-center mt-8 pt-6 border-t border-slate-800">
          <button className="px-4 py-2 flex items-center gap-2 text-slate-400 hover:text-white transition-colors">
            <Flag className="w-4 h-4" /> Flag for review
          </button>
          <div className="flex gap-4">
            <button 
              onClick={() => setCurrentQuestion(p => Math.max(1, p - 1))}
              className="px-6 py-2 rounded border border-slate-700 hover:bg-slate-800 flex items-center gap-2 disabled:opacity-50"
              disabled={currentQuestion === 1}
            >
              <ChevronLeft className="w-4 h-4" /> Previous
            </button>
            <button 
              onClick={() => setCurrentQuestion(p => Math.min(totalQuestions, p + 1))}
              className="px-6 py-2 rounded bg-slate-800 border border-slate-700 hover:bg-slate-700 flex items-center gap-2 disabled:opacity-50"
              disabled={currentQuestion === totalQuestions}
            >
              Next <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
