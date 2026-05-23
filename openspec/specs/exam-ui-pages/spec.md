## ADDED Requirements

### Requirement: Application Routing
The system SHALL provide routing between the Landing, Login, Exam, and Evaluation pages.

#### Scenario: Navigation
- **WHEN** a user visits `/login`
- **THEN** the Login page is displayed

### Requirement: Exam Simulator UI
The system SHALL display a paginated exam interface for 100 questions.

#### Scenario: Question navigation
- **WHEN** a user clicks "Next" on the exam page
- **THEN** the next question is displayed without reloading the entire page

### Requirement: Exam Timer and Auto-submit
The exam page SHALL enforce a 90-minute timer and automatically submit when time expires.

#### Scenario: Time expiration
- **WHEN** the 90-minute timer reaches 00:00:00
- **THEN** the exam automatically triggers the submit action and navigates to the evaluation page
