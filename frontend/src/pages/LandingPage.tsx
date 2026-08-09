import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { 
  ArrowRight, 
  Terminal, 
  Loader2, 
  CheckCircle, 
  XCircle, 
  RefreshCw,
  Activity
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
  const [activeWorkbenchTab, setActiveWorkbenchTab] = useState<"quiz" | "logs">("quiz");
  const [sampleLoading, setSampleLoading] = useState(false);
  const [question, setQuestion] = useState<Question | null>(null);
  const [selectedAnswer, setSelectedAnswer] = useState<string | null>(null);
  const [isSubmitted, setIsSubmitted] = useState(false);

  useEffect(() => {
    fetchSampleQuestion();
  }, []);

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
    setActiveWorkbenchTab("quiz");
    if (!question) {
      fetchSampleQuestion();
    }
    // Scroll smoothly to section on mobile
    if (window.innerWidth < 1024) {
      setTimeout(() => {
        document.getElementById("analyst-workbench")?.scrollIntoView({ behavior: "smooth" });
      }, 100);
    }
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
    <div className="flex flex-col w-full max-w-6xl mx-auto pb-24 space-y-24 animate-slide-up">
      {/* Hero Section: Left-Heavy Asymmetric Split Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center mt-12 md:mt-20">
        
        {/* Left Column: Heading and CTAs */}
        <div className="lg:col-span-5 space-y-6 text-left">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-graphite-800 text-graphite-400 text-[10px] font-mono bg-graphite-900/20">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
            <span>v2.0 Beta Environment Live</span>
          </div>
          
          <h1 className="text-4xl md:text-5xl font-extrabold tracking-tighter text-white leading-[1.05] text-wrap">
            Master Cyber Defense <br />
            Through Data-Driven Scenarios
          </h1>
          
          <p className="text-xs md:text-sm text-graphite-400 font-light leading-relaxed max-w-[45ch]">
            Train like a real SOC Analyst. Master CompTIA Security+ 701 with interactive simulations and AI-powered skill gap analysis.
          </p>
          
          <div className="pt-4 flex flex-col sm:flex-row gap-4 w-full sm:w-auto">
            <button 
              onClick={handleStartAssessment}
              disabled={redirectLoading}
              className="group flex items-center justify-center gap-2 px-6 py-2.5 rounded-md bg-white text-black font-semibold text-xs hover:bg-white/90 disabled:bg-graphite-800 disabled:text-graphite-500 transition-all duration-200"
            >
              {redirectLoading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Evaluating...
                </>
              ) : (
                <>
                  Start Assessment
                  <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-0.5" />
                </>
              )}
            </button>
            
            <button 
              onClick={handleViewSampleExam}
              className="flex items-center justify-center gap-2 px-6 py-2.5 rounded-md border border-graphite-800 bg-transparent text-graphite-300 font-semibold text-xs hover:bg-graphite-900 hover:text-white transition-all duration-200"
            >
              <Terminal className="w-4 h-4" />
              View Sample Exam
            </button>
          </div>
        </div>

        {/* Right Column: Tabbed Analyst Dashboard (SOC Workbench) */}
        <div id="analyst-workbench" className="lg:col-span-7 bg-graphite-900 border border-graphite-800 rounded-lg overflow-hidden shadow-2xl min-h-[420px] flex flex-col justify-between">
          
          {/* Header Tab Bar */}
          <div className="flex bg-graphite-950 border-b border-graphite-800 px-4 h-11 items-center justify-between select-none">
            <div className="flex gap-2">
              <button
                onClick={() => setActiveWorkbenchTab("quiz")}
                className={`flex items-center gap-2 px-3 py-2 text-xs font-mono transition-all border-b-2 h-11 ${
                  activeWorkbenchTab === "quiz"
                    ? "border-white text-white font-medium"
                    : "border-transparent text-graphite-500 hover:text-graphite-300"
                }`}
              >
                <Terminal className="w-3.5 h-3.5" />
                triage-quiz.md
              </button>
              <button
                onClick={() => setActiveWorkbenchTab("logs")}
                className={`flex items-center gap-2 px-3 py-2 text-xs font-mono transition-all border-b-2 h-11 ${
                  activeWorkbenchTab === "logs"
                    ? "border-white text-white font-medium"
                    : "border-transparent text-graphite-500 hover:text-graphite-300"
                }`}
              >
                <Activity className="w-3.5 h-3.5" />
                live-siem-logs.sh
              </button>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-graphite-800 border border-graphite-700" />
              <span className="w-2 h-2 rounded-full bg-graphite-800 border border-graphite-700" />
              <span className="w-2 h-2 rounded-full bg-graphite-800 border border-graphite-700" />
            </div>
          </div>

          {/* Workbench Body */}
          <div className="flex-1 p-6 overflow-y-auto">
            {activeWorkbenchTab === "quiz" ? (
              // Quiz content
              sampleLoading ? (
                <div className="h-full min-h-[260px] flex flex-col items-center justify-center gap-3">
                  <Loader2 className="w-6 h-6 text-white animate-spin" />
                  <p className="text-[10px] font-mono text-graphite-550">Fetching sample question...</p>
                </div>
              ) : question ? (
                <div className="space-y-4">
                  <div className="flex justify-between items-center text-[10px] font-mono text-graphite-500 border-b border-graphite-850 pb-2">
                    <span>CompTIA Security+ 701 Question</span>
                    <span>Difficulty: {question.difficulty}/5</span>
                  </div>
                  <div className="text-xs text-graphite-200 font-light leading-relaxed">
                    {question.question_text}
                  </div>
                  <div className="space-y-2 pt-2">
                    {Object.entries(question.options).map(([key, value]) => {
                      const isSelected = selectedAnswer === key;
                      const isCorrectAnswer = question.correct_answer === key;
                      
                      let optionStyles = "border-graphite-800 bg-graphite-950/40 hover:bg-graphite-950 text-graphite-300";
                      if (isSelected && !isSubmitted) {
                        optionStyles = "border-white bg-graphite-950 text-white";
                      } else if (isSubmitted) {
                        if (isCorrectAnswer) {
                          optionStyles = "border-emerald-500/30 bg-emerald-950/20 text-emerald-400";
                        } else if (isSelected) {
                          optionStyles = "border-red-500/30 bg-red-950/20 text-red-400";
                        } else {
                          optionStyles = "border-graphite-950 opacity-40 text-graphite-500";
                        }
                      }

                      return (
                        <button
                          key={key}
                          onClick={() => handleAnswerSelect(key)}
                          disabled={isSubmitted}
                          className={`w-full text-left p-3 rounded border text-xs font-light transition-all flex items-start gap-3 ${optionStyles}`}
                        >
                          <span className="font-semibold text-white shrink-0 font-mono">{key})</span>
                          <span>{value}</span>
                        </button>
                      );
                    })}
                  </div>
                  
                  {isSubmitted && (
                    <div className="rounded border p-4 space-y-2 animate-fade-in bg-graphite-950 border-graphite-800 text-[11px] font-light leading-relaxed">
                      {selectedAnswer === question.correct_answer ? (
                        <div className="flex items-center gap-1.5 text-emerald-400 font-semibold">
                          <CheckCircle className="w-3.5 h-3.5" />
                          <span>Correct answer!</span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1.5 text-red-400 font-semibold">
                          <XCircle className="w-3.5 h-3.5" />
                          <span>Incorrect. Correct was {question.correct_answer}.</span>
                        </div>
                      )}
                      <p className="text-graphite-405">
                        <strong className="text-white font-medium">Explanation:</strong> {question.explanation}
                      </p>
                    </div>
                  )}
                </div>
              ) : (
                <div className="h-full flex items-center justify-center text-graphite-500 font-mono text-xs">
                  No question loaded.
                </div>
              )
            ) : (
              // Live logs content
              <div className="h-full min-h-[260px] flex flex-col justify-between font-mono text-[10px] text-graphite-300">
                <div className="space-y-2.5 select-none">
                  <div className="text-graphite-500 border-b border-graphite-850 pb-2 mb-2 flex justify-between">
                    <span>active-sim-triage.log</span>
                    <span className="text-emerald-500 flex items-center gap-1">
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" /> LIVE STREAM
                    </span>
                  </div>
                  <div className="text-red-400 font-semibold">[ALERT] Port Scan T1046 detected from 192.0.2.14</div>
                  <div className="text-graphite-500">[INFO] Executing KQL filter: SecurityAlert | where ProviderName == "MDATP"</div>
                  <div className="text-graphite-500">SrcIP: 198.51.100.45 -{"->"} DstIP: 10.0.2.15</div>
                  <div className="text-graphite-400">&gt; T1110.001 Brute Force target identified</div>
                  <div className="text-emerald-400">&gt; Playbook Containment: Isolation rule dispatched to firewall.</div>
                  <div className="text-graphite-500">[STATUS] Host isolation status: COMPLETED</div>
                </div>
                <div className="border-t border-graphite-850 pt-4 flex justify-between items-center">
                  <span className="text-[9px] text-graphite-500">Threat Mitigation Loop Active</span>
                  <span className="text-white bg-graphite-800 px-2 py-0.5 rounded border border-graphite-700 text-[8px]">MITRE T1046</span>
                </div>
              </div>
            )}
          </div>

          {/* Workbench Footer Actions */}
          {activeWorkbenchTab === "quiz" && question && (
            <div className="bg-graphite-950 border-t border-graphite-800 px-6 py-3 flex justify-between items-center select-none">
              <button
                onClick={fetchSampleQuestion}
                className="flex items-center gap-2 px-3 py-1.5 rounded border border-graphite-800 bg-transparent text-graphite-400 text-[10px] font-medium hover:bg-graphite-900 hover:text-white transition-colors"
              >
                <RefreshCw className="w-3 h-3" />
                Try Another
              </button>

              {!isSubmitted ? (
                <button
                  onClick={handleSubmitAnswer}
                  disabled={!selectedAnswer}
                  className="px-4 py-1.5 rounded bg-white text-black text-[10px] font-semibold hover:bg-white/90 disabled:bg-graphite-800 disabled:text-graphite-500 transition-all duration-200"
                >
                  Submit Answer
                </button>
              ) : (
                <span className="text-[10px] text-graphite-500 font-mono">Answers submitted</span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* High-Visual Curriculum Bento Grid */}
      <div className="space-y-6 pt-12">
        <div className="text-left space-y-1">
          <h2 className="text-lg font-bold tracking-tight text-white">Platform Curriculum & Design</h2>
          <p className="text-[10px] text-graphite-400 font-mono uppercase tracking-wider">Structured Incident Response & Assessment Pipeline</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-6 gap-6 w-full">
          
          {/* Cell 1: Adaptive Training Loop (2 columns) */}
          <div className="md:col-span-2 bg-graphite-900 border border-graphite-800/80 rounded-lg p-6 flex flex-col justify-between hover:border-graphite-700 transition-colors duration-300">
            <div>
              <span className="text-[9px] uppercase tracking-wider font-mono text-graphite-500">Adaptive Loop</span>
              <h3 className="text-sm font-semibold text-white mt-1">Adaptive Learning Flow</h3>
              <p className="text-graphite-400 text-xs mt-2 font-light leading-relaxed">
                Our RAG-based analysis generates custom roadmaps based on your pre-test baseline skills.
              </p>
            </div>
            
            {/* Visual Loop diagram */}
            <div className="mt-8 pt-4 border-t border-graphite-850 flex flex-col gap-3 font-mono text-[9px] text-graphite-400 select-none">
              <div className="flex items-center gap-2">
                <span className="w-5 h-5 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium">01</span>
                <span>Pre-Test Diagnostic</span>
              </div>
              <div className="w-px h-3 bg-graphite-800 ml-2.5" />
              <div className="flex items-center gap-2">
                <span className="w-5 h-5 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium">02</span>
                <span className="text-white font-medium">AI Gap Roadmap</span>
              </div>
              <div className="w-px h-3 bg-graphite-800 ml-2.5" />
              <div className="flex items-center gap-2">
                <span className="w-5 h-5 rounded border border-graphite-800 bg-graphite-950 flex items-center justify-center text-white font-medium">03</span>
                <span>Practical Scenarios</span>
              </div>
            </div>
          </div>

          {/* Cell 2: Incident Response Playbooks (4 columns) */}
          <div className="md:col-span-4 bg-graphite-900 border border-graphite-800/80 rounded-lg p-6 hover:border-graphite-700 transition-colors duration-300 flex flex-col justify-between">
            <div className="space-y-2">
              <span className="text-[9px] uppercase tracking-wider font-mono text-graphite-500">Playbook Engineering</span>
              <h3 className="text-sm font-semibold text-white">Interactive IR Playbooks</h3>
              <p className="text-graphite-400 text-xs font-light leading-relaxed">
                Triage alerts and apply incident response techniques mapped directly to containment, eradication, and recovery procedures.
              </p>
            </div>
            
            {/* Playbook interactive step visual */}
            <div className="grid grid-cols-3 gap-3 mt-8 pt-4 border-t border-graphite-850 font-mono text-[10px] select-none">
              <div className="border border-graphite-800 rounded bg-graphite-950/60 p-3 text-center space-y-1.5">
                <div className="text-white font-bold">1. Containment</div>
                <div className="text-[8px] text-graphite-500">Isolate host network logs</div>
              </div>
              <div className="border border-graphite-800 rounded bg-graphite-950/60 p-3 text-center space-y-1.5">
                <div className="text-white font-bold">2. Eradication</div>
                <div className="text-[8px] text-graphite-500">Purge SSH key access</div>
              </div>
              <div className="border border-graphite-800 rounded bg-graphite-950/60 p-3 text-center space-y-1.5">
                <div className="text-white font-bold">3. Recovery</div>
                <div className="text-[8px] text-graphite-500">Restore sanitized images</div>
              </div>
            </div>
          </div>

          {/* Cell 3: Continuous AI Diagnostics (6 columns) */}
          <div className="md:col-span-6 bg-graphite-900 border border-graphite-800/80 rounded-lg p-6 hover:border-graphite-700 transition-colors duration-300">
            <div className="grid md:grid-cols-12 gap-6 items-center">
              <div className="md:col-span-7 space-y-2 text-left">
                <span className="text-[9px] uppercase tracking-wider font-mono text-graphite-500">Sec+ Proficiency</span>
                <h3 className="text-sm font-semibold text-white">Continuous AI Diagnostics</h3>
                <p className="text-graphite-400 text-xs font-light leading-relaxed">
                  Real-time telemetry gauges calculate your current proficiency scores across all 5 Security+ domains as you work.
                </p>
              </div>
              
              <div className="md:col-span-5 space-y-3 font-mono text-[9px] text-graphite-400 select-none">
                <div>
                  <div className="flex justify-between text-graphite-300 mb-1">
                    <span>Domain 1: Security Concepts</span>
                    <span className="text-white font-bold">85%</span>
                  </div>
                  <div className="w-full bg-graphite-950 h-1.5 rounded overflow-hidden border border-graphite-800">
                    <div className="bg-white h-full" style={{ width: "85%" }} />
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-graphite-300 mb-1">
                    <span>Domain 2: Threats & Mitigations</span>
                    <span className="text-white font-bold">72%</span>
                  </div>
                  <div className="w-full bg-graphite-950 h-1.5 rounded overflow-hidden border border-graphite-800">
                    <div className="bg-white h-full" style={{ width: "72%" }} />
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-graphite-300 mb-1">
                    <span>Domain 3: Security Architecture</span>
                    <span className="text-white font-bold">90%</span>
                  </div>
                  <div className="w-full bg-graphite-950 h-1.5 rounded overflow-hidden border border-graphite-800">
                    <div className="bg-white h-full" style={{ width: "90%" }} />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
