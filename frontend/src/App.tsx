import { BrowserRouter as Router, Routes, Route } from "react-router-dom"
import Layout from "./components/Layout"
import LandingPage from "./pages/LandingPage"
import LoginPage from "./pages/LoginPage"
import RegisterPage from "./pages/RegisterPage"
import ExamPage from "./pages/ExamPage"
import ExamRedirectPage from "./pages/ExamRedirectPage"
import EvaluationPage from "./pages/EvaluationPage"
import SimulationPage from "./pages/SimulationPage"
import ScenarioResultPage from "./pages/ScenarioResultPage"
import SimulationCompletePage from "./pages/SimulationCompletePage"

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<LandingPage />} />
          <Route path="login" element={<LoginPage />} />
          <Route path="register" element={<RegisterPage />} />
          <Route path="exam" element={<ExamRedirectPage />} />
          <Route path="pre-test" element={<ExamPage mode="pre" />} />
          <Route path="post-test" element={<ExamPage mode="post" />} />
          <Route path="evaluation" element={<EvaluationPage />} />
          <Route path="simulation" element={<SimulationPage />} />
          <Route path="simulation/result/:id" element={<ScenarioResultPage />} />
          <Route path="simulation/complete" element={<SimulationCompletePage />} />
        </Route>
      </Routes>
    </Router>
  )
}

export default App
