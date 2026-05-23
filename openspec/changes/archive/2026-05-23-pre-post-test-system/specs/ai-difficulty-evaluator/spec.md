## ADDED Requirements

### Requirement: Authentication with Google Service Account
The script SHALL authenticate with Google Cloud using Application Default Credentials or a Service Account JSON file to access the Gemini API.

#### Scenario: Successful Authentication
- **WHEN** the script initializes the Gemini API client
- **THEN** it successfully connects using the provided Google Auth credentials without error

### Requirement: Evaluate Question Difficulty
The script SHALL pass the parsed question and choices to Gemini Pro and receive a difficulty rating (e.g., Easy, Medium, Hard, or 1-3).

#### Scenario: Difficulty Assignment
- **WHEN** a parsed question is sent to the Gemini API
- **THEN** the API returns a structured response indicating the difficulty level which is appended to the JSON object
