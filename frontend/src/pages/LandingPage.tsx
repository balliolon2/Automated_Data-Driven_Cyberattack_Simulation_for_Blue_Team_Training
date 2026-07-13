import { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { 
  ShieldAlert, 
  Crosshair, 
  BrainCircuit, 
  ArrowRight, 
  Activity, 
  Terminal, 
  Loader2, 
  CheckCircle, 
  XCircle, 
  RefreshCw 
} from "lucide-react";

interface Question {
  question_id: string;
  domain_id: string;
  question_text: string;
  options: Record<string, string>;
  correct_answer: string;
  explanation: string;
  difficulty: number;
}

export default function LandingPage() {
  const navigate = useNavigate();
  const [redirectLoading, setRedirectLoading] = useState(false);
  
  // Sample Question State
  const [showSample, setShowSample] = useState(false);
  const [sampleLoading, setSampleLoading] = useState(false);
  const [question, setQuestion] = useState<Question | null>(null);
  const [selectedAnswer, setSelectedAnswer] = useState<string | null>(null);
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleStartAssessment = async () => {
    const token = localStorage.getItem("token");
    if (!token) {
      navigate("/login");
      return;
    }

    try {
      setRedirectLoading(true);

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
      // Fallback
      navigate("/pre-test");
    } finally {
      setRedirectLoading(false);
    }
  };

  const fetchSampleQuestion = async () => {
    try {
      setSampleLoading(true);
      setSelectedAnswer(null);
      setIsSubmitted(false);
      const res = await axios.get("/api/exams/sample");
      setQuestion(res.data);
    } catch (err) {
      console.error("Failed to load sample question", err);
    } finally {
      setSampleLoading(false);
    }
  };

  const handleViewSampleExam = () => {
    setShowSample(true);
    if (!question) {
      fetchSampleQuestion();
    }
    // Scroll smoothly to section
    setTimeout(() => {
      document.getElementById("sample-question-section")?.scrollIntoView({ behavior: "smooth" });
    }, 100);
  };

  const handleAnswerSelect = (optionKey: string) => {
    if (isSubmitted) return;
    setSelectedAnswer(optionKey);
  };

  const handleSubmitAnswer = () => {
    if (!selectedAnswer) return;
    setIsSubmitted(true);
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[75vh] w-full max-w-5xl mx-auto animate-slide-up pb-24">
      {/* Hero Section */}
      <div className="flex flex-col items-center text-center space-y-8 mt-12 mb-24 relative w-full">
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
          <button 
            onClick={handleStartAssessment}
            disabled={redirectLoading}
            className="group flex items-center justify-center gap-2 px-8 py-4 rounded-full bg-emerald-500 text-emerald-950 font-bold text-lg shadow-[0_0_20px_rgba(52,211,153,0.25)] hover:shadow-[0_0_35px_rgba(52,211,153,0.4)] hover:bg-emerald-400 disabled:bg-emerald-800 disabled:text-graphite-500 transition-all duration-300 hover:-translate-y-1"
          >
            {redirectLoading ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                Evaluating Progress...
              </>
            ) : (
              <>
                Start Assessment
                <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
              </>
            )}
          </button>
          
          <button 
            onClick={handleViewSampleExam}
            className="flex items-center justify-center gap-2 px-8 py-4 rounded-full glass-panel text-graphite-100 font-medium text-lg hover:bg-graphite-800/80 hover:text-white transition-all duration-300 hover:-translate-y-1"
          >
            <Terminal className="w-5 h-5 text-graphite-400" />
            View Sample Exam
          </button>
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

      {/* Interactive Sample Question Panel */}
      {showSample && (
        <div 
          id="sample-question-section" 
          className="w-full max-w-3xl mt-24 scroll-mt-20 glass-panel rounded-2xl p-8 border border-graphite-800 space-y-6"
        >
          <div className="flex justify-between items-center border-b border-graphite-800 pb-4">
            <div className="flex items-center gap-2">
              <Terminal className="w-5 h-5 text-emerald-400" />
              <h3 className="text-xl font-bold text-white tracking-wide">Sample Security+ Question</h3>
            </div>
            {question && (
              <span className="text-xs font-mono px-2.5 py-1 rounded bg-graphite-800 border border-graphite-700 text-graphite-300">
                Difficulty: {question.difficulty}/5
              </span>
            )}
          </div>

          {sampleLoading ? (
            <div className="py-12 flex flex-col items-center justify-center gap-3">
              <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
              <p className="text-sm font-mono text-graphite-400">Fetching sample question...</p>
            </div>
          ) : question ? (
            <div className="space-y-6">
              <div className="text-lg text-graphite-100 font-light leading-relaxed">
                {question.question_text}
              </div>

              <div className="space-y-3">
                {Object.entries(question.options).map(([key, value]) => {
                  const isSelected = selectedAnswer === key;
                  const isCorrectAnswer = question.correct_answer === key;
                  
                  let optionStyles = "border-graphite-800 hover:bg-graphite-800/40 text-graphite-300";
                  if (isSelected && !isSubmitted) {
                    optionStyles = "border-emerald-500/50 bg-emerald-500/5 text-emerald-300";
                  } else if (isSubmitted) {
                    if (isCorrectAnswer) {
                      optionStyles = "border-emerald-500 bg-emerald-950/20 text-emerald-400";
                    } else if (isSelected) {
                      optionStyles = "border-red-500 bg-red-950/20 text-red-400";
                    } else {
                      optionStyles = "border-graphite-900 opacity-60 text-graphite-400";
                    }
                  }

                  return (
                    <button
                      key={key}
                      onClick={() => handleAnswerSelect(key)}
                      disabled={isSubmitted}
                      className={`w-full text-left p-4 rounded-xl border font-light transition-all flex items-start gap-4 ${optionStyles}`}
                    >
                      <span className="font-bold text-emerald-400 shrink-0">{key})</span>
                      <span>{value}</span>
                    </button>
                  );
                })}
              </div>

              <div className="pt-4 flex justify-between items-center border-t border-graphite-800">
                <button
                  onClick={fetchSampleQuestion}
                  className="flex items-center gap-2 px-5 py-2.5 rounded-full border border-graphite-700 bg-graphite-800/40 text-graphite-300 text-sm font-medium hover:bg-graphite-800 hover:text-white transition-colors"
                >
                  <RefreshCw className="w-4 h-4" />
                  Try Another Question
                </button>

                {!isSubmitted && (
                  <button
                    onClick={handleSubmitAnswer}
                    disabled={!selectedAnswer}
                    className="px-6 py-2.5 rounded-full bg-emerald-500 text-emerald-950 text-sm font-bold shadow-[0_0_15px_rgba(52,211,153,0.2)] hover:bg-emerald-400 hover:shadow-[0_0_20px_rgba(52,211,153,0.3)] disabled:bg-graphite-800 disabled:text-graphite-500 disabled:shadow-none transition-all duration-300"
                  >
                    Submit Answer
                  </button>
                )}
              </div>

              {isSubmitted && (
                <div className="rounded-xl border p-5 space-y-3 animate-fade-in bg-graphite-950/40 border-graphite-800">
                  {selectedAnswer === question.correct_answer ? (
                    <div className="flex items-center gap-2 text-emerald-400 font-bold">
                      <CheckCircle className="w-5 h-5 shrink-0" />
                      <span>Correct! Nice job.</span>
                    </div>
                  ) : (
                    <div className="flex items-center gap-2 text-red-400 font-bold">
                      <XCircle className="w-5 h-5 shrink-0" />
                      <span>Incorrect. The correct answer was {question.correct_answer}.</span>
                    </div>
                  )}
                  <p className="text-sm text-graphite-300 font-light leading-relaxed">
                    <strong className="text-white block mb-1">Explanation:</strong>
                    {question.explanation}
                  </p>
                </div>
              )}
            </div>
          ) : (
            <div className="text-center py-6 text-graphite-400">
              No question loaded. Please try again.
            </div>
          )}
        </div>
      )}
    </div>
  );
}
