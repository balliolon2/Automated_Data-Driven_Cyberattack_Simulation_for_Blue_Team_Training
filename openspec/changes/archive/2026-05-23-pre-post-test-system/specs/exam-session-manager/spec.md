## ADDED Requirements

### Requirement: Randomize Question Set
The system SHALL randomize exactly 100 questions per exam session, distributed as: Domain 1 (12), Domain 2 (22), Domain 3 (18), Domain 4 (28), and Domain 5 (20).

#### Scenario: Pre-Test Question Generation
- **WHEN** a user starts a new pre-test
- **THEN** the system fetches a randomized set of 100 questions matching the domain distribution and links them to a new exam session

### Requirement: Strict Post-Test Randomization
For post-tests, the system SHALL select a completely different set of randomized questions that the user did not receive during their pre-test.

#### Scenario: Post-Test Generation
- **WHEN** a user starts a post-test
- **THEN** the system fetches 100 randomized questions matching the domain distribution, ensuring none of them were included in the user's pre-test session

### Requirement: State Persistence
The system SHALL persist the user's progress through the exam session.

#### Scenario: User Disconnection
- **WHEN** a user refreshes the page or reconnects after losing internet
- **THEN** the system retrieves their active exam session and resumes from the last unanswered question
