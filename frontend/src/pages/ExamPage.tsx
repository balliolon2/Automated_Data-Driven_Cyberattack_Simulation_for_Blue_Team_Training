import { useState, useEffect } from "react";
import { Clock, Flag, ChevronLeft, ChevronRight, CheckCircle2 } from "lucide-react";
import axios from "axios";
import { useNavigate } from "react-router-dom";

interface Question {
  question_id: string;
  domain_id: string;
  type: string;
  question_text: string;
  options: Record<string, string>;
  order_index: number;
  user_answer?: string;
}

interface ExamSession {
  session_id: string;
  user_id: string;
  exam_type: string;
  status: string;
  total_questions: number;
  time_spent_seconds: number | null;
}

export default function ExamPage() {
  const [currentQuestionIdx, setCurrentQuestionIdx] = useState(0);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [session, setSession] = useState<ExamSession | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [timeLeft, setTimeLeft] = useState(90 * 60);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchSession = async () => {
      try {
        const token = localStorage.getItem("token");
        const res = await axios.get("/api/exams/session", {
          headers: { Authorization: `Bearer ${token}` }
        });
        setSession(res.data.session);
        setQuestions(res.data.questions);
        
        const initialAnswers: Record<string, string> = {};
        res.data.questions.forEach((q: Question) => {
          if (q.user_answer) {
            initialAnswers[q.question_id] = q.user_answer;
          }
        });
        setAnswers(initialAnswers);
      } catch (err) {
        console.error("Failed to load exam session:", err);
      } finally {
        setLoading(false);
      }
    };
    fetchSession();
  }, []);

  useEffect(() => {
    if (loading) return;
    const timer = setInterval(() => {
      setTimeLeft((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(timer);
  }, [loading]);

  const formatTime = (seconds: number) => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  const handleAnswerSelect = (qId: string, answer: string) => {
    setAnswers(prev => ({ ...prev, [qId]: answer }));
  };

  const handleSubmitExam = async () => {
    if (!session) return;
    
    const unanswered = questions.filter(q => !answers[q.question_id]).length;
    if (unanswered > 0) {
      if (!window.confirm(`You have ${unanswered} unanswered questions. Are you sure you want to submit? Your score won't be calculated until all questions are answered.`)) {
        return;
      }
    }

    setSubmitting(true);
    const token = localStorage.getItem("token");
    
    try {
      for (const q of questions) {
        const userAns = answers[q.question_id];
        if (userAns && userAns !== q.user_answer) {
          await axios.post("/api/exams/submit-answer", {
            session_id: session.session_id,
            question_id: q.question_id,
            user_answer: userAns,
            time_spent_seconds: 0
          }, {
            headers: { Authorization: `Bearer ${token}` }
          });
        }
      }
      navigate("/evaluation");
    } catch (err) {
      console.error(err);
      alert("Failed to submit exam.");
    } finally {
      setSubmitting(false);
    }
  };

  const startExam = async (type: "pre" | "post") => {
    try {
      setLoading(true);
      const token = localStorage.getItem("token");
      const endpoint = type === "pre" ? "/api/exams/pre-test" : "/api/exams/post-test";
      const res = await axios.post(endpoint, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setSession(res.data.session);
      setQuestions(res.data.questions);
      setAnswers({});
      setCurrentQuestionIdx(0);
    } catch (err: any) {
      console.error("Failed to start exam:", err);
      alert(err.response?.data?.error || "Failed to start exam. Check if the database has questions loaded.");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="p-8 text-white font-mono flex items-center justify-center min-h-[50vh]">Loading active exam session...</div>;
  }

  if (!questions.length) {
    return <div className="p-8 text-white font-mono flex flex-col items-center justify-center min-h-[50vh]">
      <div className="mb-8 text-xl">No active exam session found.</div>
      <div className="flex gap-4">
        <button 
          onClick={() => startExam("pre")}
          className="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 rounded text-white font-bold transition-colors"
        >
          Start Pre-Test
        </button>
        <button 
          onClick={() => startExam("post")}
          className="px-6 py-2 bg-blue-600 hover:bg-blue-500 rounded text-white font-bold transition-colors"
        >
          Start Post-Test
        </button>
      </div>
      <button 
        onClick={() => navigate("/")}
        className="mt-6 px-6 py-2 border border-slate-700 hover:border-slate-500 rounded text-slate-300 font-medium transition-colors"
      >
        Go back to Dashboard
      </button>
    </div>;
  }

  const currentQ = questions[currentQuestionIdx];
  const totalQuestions = questions.length;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 min-h-[75vh]">
      {/* Sidebar: Navigator */}
      <div className="lg:col-span-1 bg-slate-900/50 border border-slate-800 rounded-xl p-4 flex flex-col">
        <h3 className="font-bold text-lg mb-4 text-slate-100">Question Navigator</h3>
        <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar">
          <div className="grid grid-cols-5 gap-2">
            {questions.map((q, i) => {
              const isAnswered = !!answers[q.question_id];
              return (
                <button
                  key={q.question_id}
                  onClick={() => setCurrentQuestionIdx(i)}
                  className={`w-10 h-10 rounded flex items-center justify-center text-sm font-medium border transition-colors
                    ${currentQuestionIdx === i 
                      ? 'bg-emerald-500/20 border-emerald-500 text-emerald-400' 
                      : isAnswered
                        ? 'bg-slate-800 border-emerald-500/50 text-slate-300 hover:border-emerald-400'
                        : 'bg-slate-950 border-slate-700 text-slate-400 hover:border-slate-500'}`}
                >
                  {i + 1}
                </button>
              );
            })}
          </div>
        </div>
        <div className="mt-4 pt-4 border-t border-slate-800">
          <button 
            onClick={handleSubmitExam}
            disabled={submitting}
            className="w-full py-2 bg-emerald-600 hover:bg-emerald-500 disabled:bg-emerald-800 disabled:text-slate-400 text-white font-bold rounded-md transition-colors flex items-center justify-center gap-2"
          >
            <CheckCircle2 className="w-5 h-5" />
            {submitting ? "Submitting..." : "Submit Exam"}
          </button>
        </div>
      </div>

      {/* Main Content: Question */}
      <div className="lg:col-span-3 bg-slate-900/50 border border-slate-800 rounded-xl p-6 flex flex-col">
        <div className="flex justify-between items-center border-b border-slate-800 pb-4 mb-6">
          <h2 className="text-xl font-bold text-slate-100">Question {currentQuestionIdx + 1} of {totalQuestions}</h2>
          <div className="flex items-center gap-2 text-emerald-400 font-mono text-xl bg-slate-950 px-4 py-2 rounded-lg border border-slate-800">
            <Clock className="w-5 h-5" />
            {formatTime(timeLeft)}
          </div>
        </div>

        <div className="flex-1 space-y-8">
          <div className="text-lg text-slate-200">
            {currentQ.question_text}
          </div>
          <div className="space-y-3">
            {Object.entries(currentQ.options || {}).map(([key, optText]) => (
              <label key={key} className="flex items-center gap-3 p-4 border border-slate-700 rounded-lg cursor-pointer hover:bg-slate-800 transition-colors">
                <input 
                  type="radio" 
                  name={`q${currentQ.question_id}`} 
                  className="w-5 h-5 accent-emerald-500 cursor-pointer" 
                  checked={answers[currentQ.question_id] === key}
                  onChange={() => handleAnswerSelect(currentQ.question_id, key)}
                />
                <span className="text-slate-300 font-medium">{key})</span>
                <span className="text-slate-200">{optText as string}</span>
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
              onClick={() => setCurrentQuestionIdx(p => Math.max(0, p - 1))}
              className="px-6 py-2 rounded border border-slate-700 text-slate-300 hover:bg-slate-800 flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={currentQuestionIdx === 0}
            >
              <ChevronLeft className="w-4 h-4" /> Previous
            </button>
            <button 
              onClick={() => setCurrentQuestionIdx(p => Math.min(totalQuestions - 1, p + 1))}
              className="px-6 py-2 rounded bg-slate-800 border border-slate-700 text-slate-100 hover:bg-slate-700 flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={currentQuestionIdx === totalQuestions - 1}
            >
              Next <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
