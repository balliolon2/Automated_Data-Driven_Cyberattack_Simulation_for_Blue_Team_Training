## ADDED Requirements

### Requirement: Extract Question Components
The parser SHALL extract the question text, options (A, B, C, D), correct answer, and explanation from the markdown file.

#### Scenario: Successful Extraction
- **WHEN** the script processes a correctly formatted markdown question block
- **THEN** it outputs a structured JSON object containing all required fields

### Requirement: Identify Domain
The parser SHALL identify the domain of the question based on the markdown file being processed (e.g. domain1 to domain5).

#### Scenario: Domain Association
- **WHEN** parsing `domain1_security_concepts.md`
- **THEN** all parsed questions are tagged with domain = 1
