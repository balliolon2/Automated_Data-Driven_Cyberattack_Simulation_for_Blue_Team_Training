## Context

The "Automated Data-Driven Cyberattack Simulation for Continuous Blue Team Workforce Training" (CyberSim) platform requires a frontend application. The platform involves taking 100-question CompTIA Security+ 701 exams (Pre-test and Post-test) with a 90-minute time limit, as well as accessing simulation dashboards and evaluation results. The backend is written in Go. The user prefers a modern React + Vite stack with TailwindCSS and Shadcn/UI.

## Goals / Non-Goals

**Goals:**
- Establish a React + Vite application structure.
- Integrate TailwindCSS and Shadcn/UI for rapid component development.
- Implement responsive routing using React Router.
- Design the scaffolding for the 5 core pages: Landing, Login, Exam Simulator (Pre/Post), and Evaluation.
- Implement the client-side 90-minute timer and auto-submit logic for the exam.

**Non-Goals:**
- Implementing the backend API or database connections.
- Parsing the markdown files to JSON directly in the frontend (the backend or a seed script should handle this).
- Finalizing the full Cyber Simulation scenario dashboard (focus is on the exam flow right now).

## Decisions

- **Framework:** React with Vite. Vite provides extremely fast HMR and build times compared to Create React App.
- **Styling:** TailwindCSS + Shadcn/UI. This combination offers a perfect balance of utility-first styling and pre-built, accessible, customizable components (like buttons, progress bars, radar charts via Recharts).
- **Routing:** React Router DOM (v6) for handling the client-side routing between Landing, Login, Exam, and Dashboard.
- **Exam State Management:** React Context or Zustand will be used to manage the 100 questions, user answers, and the 90-minute timer without prop-drilling.

## Risks / Trade-offs

- **[Risk] State loss on refresh:** If a user refreshes the page during the 90-minute exam, they might lose their answers and timer state.
  - *Mitigation:* The exam state (current answers, remaining time) should be periodically synced to `localStorage` or the backend API.
- **[Risk] Security of auto-submit:** The client-side timer can be manipulated by technical users.
  - *Mitigation:* The backend must also record the exam start time and enforce the 90-minute limit strictly, treating the client-side timer as a UX feature rather than a security measure.
