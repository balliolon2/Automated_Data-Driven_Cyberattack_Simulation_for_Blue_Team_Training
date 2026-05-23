import { BrowserRouter as Router, Routes, Route } from "react-router-dom"
import Layout from "./components/Layout"
import LandingPage from "./pages/LandingPage"
import LoginPage from "./pages/LoginPage"
import RegisterPage from "./pages/RegisterPage"
import ExamPage from "./pages/ExamPage"
import EvaluationPage from "./pages/EvaluationPage"

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<LandingPage />} />
          <Route path="login" element={<LoginPage />} />
          <Route path="register" element={<RegisterPage />} />
          <Route path="exam" element={<ExamPage />} />
          <Route path="evaluation" element={<EvaluationPage />} />
        </Route>
      </Routes>
    </Router>
  )
}

export default App
