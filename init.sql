-- 1.2 Write init.sql with enums and vector / pgcrypto extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE role AS ENUM ('learner', 'expert', 'admin');
CREATE TYPE test_type AS ENUM ('pre', 'post');
CREATE TYPE question_type AS ENUM ('multiple_choice', 'true_false', 'scenario_based', 'drag_drop');
CREATE TYPE scenario_status AS ENUM ('draft', 'active', 'archived');
CREATE TYPE session_status AS ENUM ('in_progress', 'completed', 'abandoned');
CREATE TYPE action_type AS ENUM ('view_log', 'triage_alert', 'query_rag', 'escalate', 'respond', 'help_click', 'submit_decision');
CREATE TYPE feedback_source AS ENUM ('rag', 'rule_based', 'expert');

-- 1.3 Add core tables
CREATE TABLE users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR UNIQUE NOT NULL,
  password_hash VARCHAR NOT NULL,
  role role DEFAULT 'learner',
  current_tier INT DEFAULT 1,
  created_at TIMESTAMP DEFAULT now(),
  last_login TIMESTAMP,
  is_active BOOLEAN DEFAULT true
);

CREATE TABLE security_domains (
  domain_id VARCHAR PRIMARY KEY,
  name VARCHAR NOT NULL,
  description TEXT,
  weight_score FLOAT DEFAULT 1.0,
  compTIA_reference VARCHAR
);

CREATE TABLE mitre_techniques (
  technique_id VARCHAR PRIMARY KEY,
  tactic VARCHAR NOT NULL,
  name VARCHAR NOT NULL,
  description TEXT,
  url VARCHAR
);

-- 1.4 Add assessment tables
CREATE TABLE questions (
  question_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  domain_id VARCHAR REFERENCES security_domains(domain_id),
  type question_type NOT NULL,
  question_text TEXT NOT NULL,
  options JSONB,
  correct_answer VARCHAR NOT NULL,
  explanation TEXT,
  difficulty INT DEFAULT 3,
  mitre_ref VARCHAR REFERENCES mitre_techniques(technique_id),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE test_attempts (
  attempt_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(user_id),
  test_type test_type NOT NULL,
  score FLOAT NOT NULL,
  total_questions INT NOT NULL,
  answers JSONB,
  started_at TIMESTAMP DEFAULT now(),
  completed_at TIMESTAMP,
  time_spent_seconds INT
);

-- 1.5 Add simulation tables
CREATE TABLE scenarios (
  scenario_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR NOT NULL,
  description TEXT,
  mitre_technique VARCHAR REFERENCES mitre_techniques(technique_id),
  domain_id VARCHAR REFERENCES security_domains(domain_id),
  difficulty INT DEFAULT 3,
  initial_logs JSONB NOT NULL,
  playbook_steps JSONB,
  expected_outcomes JSONB,
  status scenario_status DEFAULT 'active',
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP
);

CREATE TABLE simulation_sessions (
  session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(user_id),
  scenario_id UUID REFERENCES scenarios(scenario_id),
  status session_status DEFAULT 'in_progress',
  final_score FLOAT,
  skill_gap JSONB,
  started_at TIMESTAMP DEFAULT now(),
  completed_at TIMESTAMP,
  total_actions INT DEFAULT 0
);

CREATE TABLE session_actions (
  action_id BIGSERIAL PRIMARY KEY,
  session_id UUID REFERENCES simulation_sessions(session_id),
  step_order INT NOT NULL,
  action_type action_type NOT NULL,
  payload JSONB,
  is_correct BOOLEAN,
  points INT DEFAULT 0,
  timestamp TIMESTAMP DEFAULT now(),
  UNIQUE (session_id, step_order)
);
CREATE INDEX idx_session_actions_type ON session_actions(action_type);
CREATE INDEX idx_session_actions_time ON session_actions(timestamp);

-- 1.6 Add AI & Feedback tables
CREATE TABLE rag_documents (
  doc_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR NOT NULL,
  content TEXT NOT NULL,
  domain_id VARCHAR REFERENCES security_domains(domain_id),
  embedding vector(768),
  source_url VARCHAR,
  mitre_ref VARCHAR REFERENCES mitre_techniques(technique_id),
  updated_at TIMESTAMP DEFAULT now()
);
CREATE INDEX idx_rag_docs_embedding ON rag_documents USING hnsw (embedding vector_cosine_ops);

CREATE TABLE feedback_logs (
  feedback_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID REFERENCES simulation_sessions(session_id),
  trigger_action_id BIGINT REFERENCES session_actions(action_id),
  feedback_text TEXT NOT NULL,
  reference_theory VARCHAR,
  source feedback_source NOT NULL,
  relevance_score FLOAT,
  timestamp TIMESTAMP DEFAULT now()
);

-- 1.7 Add analytics tables
CREATE TABLE user_skill_profiles (
  profile_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(user_id),
  domain_id VARCHAR REFERENCES security_domains(domain_id),
  proficiency_score FLOAT DEFAULT 0,
  last_practiced TIMESTAMP,
  scenarios_completed INT DEFAULT 0,
  UNIQUE (user_id, domain_id)
);

CREATE TABLE evaluation_metrics (
  eval_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(user_id),
  pre_score FLOAT NOT NULL,
  post_score FLOAT NOT NULL,
  improvement_pct FLOAT,
  csuq_scores JSONB,
  nasa_tlx_scores JSONB,
  skill_gap_summary JSONB,
  eval_date TIMESTAMP DEFAULT now(),
  notes TEXT
);

CREATE TABLE help_click_analytics (
  click_id BIGSERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(user_id),
  session_id UUID REFERENCES simulation_sessions(session_id),
  button_type VARCHAR,
  context_step INT,
  scenario_id UUID REFERENCES scenarios(scenario_id),
  timestamp TIMESTAMP DEFAULT now()
);
CREATE INDEX idx_help_clicks_user ON help_click_analytics(user_id);
CREATE INDEX idx_help_clicks_scenario ON help_click_analytics(scenario_id);
CREATE INDEX idx_help_clicks_time ON help_click_analytics(timestamp);
