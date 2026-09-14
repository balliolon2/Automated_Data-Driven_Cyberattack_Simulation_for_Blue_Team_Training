import { BrowserRouter as Router, Routes, Route } from "react-router-dom"
import Layout from "./components/Layout"
import LandingPage from "./pages/LandingPage"
import LoginPage from "./pages/LoginPage"
import RegisterPage from "./pages/RegisterPage"
import ProfilePage from "./pages/ProfilePage"
import AdminSpecialistsPage from "./pages/AdminSpecialistsPage"
import ExamPage from "./pages/ExamPage"
import ExamRedirectPage from "./pages/ExamRedirectPage"
import EvaluationPage from "./pages/EvaluationPage"
import SimulationPage from "./pages/SimulationPage"
import ScenarioResultPage from "./pages/ScenarioResultPage"
import SimulationCompletePage from "./pages/SimulationCompletePage"
import SpecialistReviewPage from "./pages/SpecialistReviewPage"
import DiscussionsFeedPage from "./pages/DiscussionsFeedPage"
import ThreadEditorPage from "./pages/ThreadEditorPage"
import ThreadDetailPage from "./pages/ThreadDetailPage"

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<LandingPage />} />
          <Route path="login" element={<LoginPage />} />
          <Route path="register" element={<RegisterPage />} />
          <Route path="profile" element={<ProfilePage />} />
          <Route path="admin/specialists" element={<AdminSpecialistsPage />} />
          <Route path="specialist/reviews" element={<SpecialistReviewPage />} />
          <Route path="discussions" element={<DiscussionsFeedPage />} />
          <Route path="discussions/new" element={<ThreadEditorPage />} />
          <Route path="discussions/:id" element={<ThreadDetailPage />} />
          <Route path="discussions/:id/edit" element={<ThreadEditorPage />} />
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
