## Why

To support the "Automated Data-Driven Cyberattack Simulation for Continuous Blue Team Workforce Training" (CyberSim) project, we need a robust, modern frontend application. This application will serve as the primary interface for users to take their CompTIA Security+ 701 assessments (Pre-test and Post-test) and engage with the SOC Analyst simulation platform.

## What Changes

- Initialize a React + Vite frontend application in a `frontend` directory.
- Configure TailwindCSS and integrate Shadcn/UI for consistent, accessible, and fast UI component development.
- Scaffold the key user flows and pages:
  - Landing Page
  - Login/Register Page
  - Pre-test and Post-test Pages (100 questions, 90-minute timer, paginated exam simulator view, auto-submit)
  - Evaluation Page (Radar chart for 5 domains, skill gap analysis)

## Capabilities

### New Capabilities
- `frontend-core-setup`: The base React + Vite application with TailwindCSS and Shadcn/UI.
- `exam-ui-pages`: The core UI pages and components (Landing, Login, Test Simulator, Evaluation Dashboard).

### Modified Capabilities
- 

## Impact

- Establishes the frontend architecture for the entire project.
- Provides a mockable UI structure that the backend can later wire up to real data (e.g., fetching the 100 questions from the 5 domain markdown files).
