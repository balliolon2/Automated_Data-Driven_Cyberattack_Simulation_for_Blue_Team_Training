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

export default function ExamPage({ mode }: { mode: "pre" | "post" }) {
  const [currentQuestionIdx, setCurrentQuestionIdx] = useState(0);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [session, setSession] = useState<ExamSession | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [timeLeft, setTimeLeft] = useState(30 * 60);
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

        if (res.data && res.data.session) {
          const activeType = res.data.session.exam_type;
          if (activeType !== mode) {
            // Mismatch! Redirect to the correct page
            if (activeType === "pre") {
              navigate("/pre-test");
            } else {
              navigate("/post-test");
            }
            return;
          }
          setSession(res.data.session);
          setQuestions(res.data.questions);
          
          const initialAnswers: Record<string, string> = {};
          res.data.questions.forEach((q: Question) => {
            if (q.user_answer) {
              initialAnswers[q.question_id] = q.user_answer;
            }
          });
          setAnswers(initialAnswers);
        }
      } catch (err) {
        console.error("Failed to load exam session:", err);
      } finally {
        setLoading(false);
      }
    };
    setLoading(true);
    setQuestions([]);
    setSession(null);
    setAnswers({});
    fetchSession();
  }, [mode, navigate]);

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
    return (
      <div className="max-w-xl mx-auto mt-12 bg-graphite-900 p-8 rounded-lg border border-graphite-800 text-center space-y-6 animate-slide-up">
        {mode === "pre" ? (
          <>
            <h2 className="text-2xl font-bold text-white tracking-tight">Security+ Pre-Test Assessment</h2>
            <p className="text-graphite-400 font-light text-sm leading-relaxed">
              This assessment consists of <strong className="text-white">30 questions</strong> spanning 5 core CompTIA Security+ domains. 
              The test evaluates your current baseline proficiency and designs your targeted adaptive simulation roadmap.
            </p>
            <div className="pt-4">
              <button 
                onClick={() => startExam("pre")}
                className="w-full py-3 bg-white hover:bg-white/90 text-black font-semibold text-sm rounded-md transition-all duration-200"
              >
                Start Pre-Test Assessment
              </button>
            </div>
          </>
        ) : (
          <>
            <h2 className="text-2xl font-bold text-white tracking-tight">Security+ Post-Test Assessment</h2>
            <p className="text-graphite-400 font-light text-sm leading-relaxed">
              This final assessment consists of <strong className="text-white">30 questions</strong>. It evaluates your overall proficiency after 
              completing the required hands-on SOC simulation training scenarios.
            </p>
            <div className="pt-4">
              <button 
                onClick={() => startExam("post")}
                className="w-full py-3 bg-white hover:bg-white/90 text-black font-semibold text-sm rounded-md transition-all duration-200"
              >
                Start Post-Test Assessment
              </button>
            </div>
          </>
        )}
        <button 
          onClick={() => navigate("/")}
          className="mt-4 text-xs text-graphite-400 hover:text-white transition-colors underline decoration-dotted"
        >
          Go back to Dashboard
        </button>
      </div>
    );
  }

  const currentQ = questions[currentQuestionIdx];
  const totalQuestions = questions.length;
  const answeredCount = questions.filter(q => !!answers[q.question_id]).length;
  const progressPercent = totalQuestions > 0 ? Math.round((answeredCount / totalQuestions) * 100) : 0;
  const isComplete = totalQuestions > 0 && answeredCount === totalQuestions;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 min-h-[75vh]">
      {/* Sidebar: Navigator */}
      <div className="lg:col-span-1 bg-graphite-900 border border-graphite-800 rounded-lg p-4 flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-2">
            <h3 className="font-semibold text-sm text-white">Question Navigator</h3>
            <span
              className={`text-xs font-mono px-2 py-0.5 rounded border transition-colors ${
                isComplete
                  ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400 font-medium'
                  : 'bg-graphite-950 border-graphite-800 text-graphite-300'
              }`}
            >
              {answeredCount}/{totalQuestions} Answered
            </span>
          </div>

          {/* Micro Progress Bar */}
          <div className="mb-4">
            <div className="w-full bg-graphite-950 rounded-full h-1.5 border border-graphite-800/80 overflow-hidden">
              <div
                className={`h-full transition-all duration-300 rounded-full ${
                  isComplete ? 'bg-emerald-500' : 'bg-white'
                }`}
                style={{ width: `${progressPercent}%` }}
              />
            </div>
          </div>
          <div className="max-h-[350px] overflow-y-auto pr-2 custom-scrollbar">
            <div className="grid grid-cols-5 gap-2">
              {questions.map((q, i) => {
                const isAnswered = !!answers[q.question_id];
                const isCurrent = currentQuestionIdx === i;
                
                let btnStyles = "bg-graphite-950 border-graphite-800 text-graphite-400 hover:border-graphite-700";
                if (isCurrent) {
                  btnStyles = "bg-white/10 border-white text-white";
                } else if (isAnswered) {
                  btnStyles = "bg-graphite-800 border-graphite-600 text-white";
                }

                return (
                  <button
                    key={q.question_id}
                    onClick={() => setCurrentQuestionIdx(i)}
                    className={`w-10 h-10 rounded text-xs font-mono border transition-colors flex items-center justify-center ${btnStyles}`}
                  >
                    {(i + 1).toString().padStart(2, '0')}
                  </button>
                );
              })}
            </div>
          </div>
        </div>
        <div className="mt-6 pt-4 border-t border-graphite-800">
          <button 
            onClick={handleSubmitExam}
            disabled={submitting}
            className="w-full py-2 bg-white text-black hover:bg-white/90 disabled:bg-graphite-800 disabled:text-graphite-500 text-xs font-semibold rounded-md transition-all flex items-center justify-center gap-2"
          >
            <CheckCircle2 className="w-4 h-4" />
            {submitting ? "Submitting..." : "Submit Exam"}
          </button>
        </div>
      </div>

      {/* Main Content: Question */}
      <div className="lg:col-span-3 bg-graphite-900 border border-graphite-800 rounded-lg p-6 flex flex-col justify-between">
        <div>
          <div className="flex justify-between items-center border-b border-graphite-800 pb-4 mb-6">
            <h2 className="text-base font-semibold text-white">Question {currentQuestionIdx + 1} of {totalQuestions}</h2>
            <div className="flex items-center gap-2 text-white font-mono text-sm bg-graphite-950 px-3 py-1.5 rounded border border-graphite-800 select-none">
              <Clock className="w-4 h-4 text-graphite-400" />
              {formatTime(timeLeft)}
            </div>
          </div>

          {timeLeft === 0 && (
            <div className="mb-6 bg-red-500/10 border border-red-500/20 rounded p-4 text-red-400 text-xs">
              <p className="font-medium flex items-center gap-2">
                <span>⏰</span> Time's up! You can review your answers but cannot change them. Please submit.
              </p>
            </div>
          )}

          <div className="space-y-6">
            <div className="text-base text-graphite-100 font-light leading-relaxed">
              {currentQ.question_text}
            </div>
            <div className="space-y-3">
              {Object.entries(currentQ.options || {}).map(([key, optText]) => {
                const isChecked = answers[currentQ.question_id] === key;
                return (
                  <label 
                    key={key} 
                    className={`flex items-start gap-4 p-4 border rounded-md cursor-pointer transition-colors text-sm font-light ${
                      isChecked 
                        ? 'border-white bg-graphite-950 text-white' 
                        : 'border-graphite-800 bg-graphite-950/40 hover:bg-graphite-950 text-graphite-300'
                    }`}
                  >
                    <input 
                      type="radio" 
                      name={`q${currentQ.question_id}`} 
                      className={`w-4 h-4 mt-0.5 accent-white ${timeLeft === 0 ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'}`} 
                      checked={isChecked}
                      onChange={() => {
                        if (timeLeft > 0) {
                          handleAnswerSelect(currentQ.question_id, key);
                        }
                      }}
                      disabled={timeLeft === 0}
                    />
                    <span className="font-semibold font-mono text-white shrink-0">{key})</span>
                    <span>{optText as string}</span>
                  </label>
                );
              })}
            </div>
          </div>
        </div>

        <div className="flex justify-between items-center mt-12 pt-4 border-t border-graphite-800">
          <button className="px-3 py-1.5 flex items-center gap-2 text-xs text-graphite-400 hover:text-white transition-colors">
            <Flag className="w-3.5 h-3.5" /> Flag for review
          </button>
          <div className="flex gap-3">
            <button 
              onClick={() => setCurrentQuestionIdx(p => Math.max(0, p - 1))}
              className="px-4 py-2 rounded-md border border-graphite-800 text-xs text-graphite-300 hover:bg-graphite-900 hover:text-white flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={currentQuestionIdx === 0}
            >
              <ChevronLeft className="w-3.5 h-3.5" /> Previous
            </button>
            <button 
              onClick={() => setCurrentQuestionIdx(p => Math.min(totalQuestions - 1, p + 1))}
              className="px-4 py-2 rounded-md bg-white text-black text-xs font-semibold hover:bg-white/90 flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={currentQuestionIdx === totalQuestions - 1}
            >
              Next <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
