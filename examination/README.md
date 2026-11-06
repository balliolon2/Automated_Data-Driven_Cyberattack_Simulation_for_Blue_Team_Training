# examination/ — Source Exam Markdown

## 1. What This Directory Contains

This directory contains the source Markdown files for the Pre-test/Post-test question bank used by the learner assessment system: **540 questions** across **5 source files, one per Domain** (`domain*.md`), together with this `README.md`, which documents provenance, mapping, and usage policy.

Files in this directory are the **source of truth** for generating artifacts with the parser. The backend does not read them directly.

## 2. Source and Citation

All questions in this directory come from the following book:

Gorgotsias, N. (2024). *CompTIA Security+ SY0-701 practice tests & PBQs: Exam SY0-701* (3rd ed.). ExamsDigest LLC; LabsDigest LLC.

- Project reference key: **[R1]**
- In-text citation: **(Gorgotsias, 2024)**
- Reference registry, copyright restrictions, and open risks: [`../docs/REFERENCES.md`](../docs/REFERENCES.md)

## 3. Chapter → Domain → File Mapping

| Chapter | Book chapter title | Project Domain | Question range | File | Count |
|---|---|---|---|---|---|
| Ch1 | General Security Concepts | domain1 — Security Concepts | Q1–Q110 | [`domain1_security_concepts.md`](domain1_security_concepts.md) | 110 |
| Ch2 | Threats, Vulnerabilities, and Mitigations | domain2 — Threats, Vulnerabilities, and Mitigations | Q111–Q220 | [`domain2_threats_vulnerabilities.md`](domain2_threats_vulnerabilities.md) | 110 |
| Ch3 | Implementation | domain3 — Security Architecture | Q221–Q310 | [`domain3_security_architecture.md`](domain3_security_architecture.md) | 90 |
| Ch4 | Security Operations | domain4 — Security Operations | Q311–Q460 | [`domain4_security_operations.md`](domain4_security_operations.md) | 150 |
| Ch5 | Security Program Management and Oversight | domain5 — Security Management | Q461–Q540 | [`domain5_security_management.md`](domain5_security_management.md) | 80 |
| **Total** | | | **Q1–Q540** | | **540** |

Note: The source book names Chapter 3 **"Implementation"**, while the project maps it to CompTIA Domain 3, **"Security Architecture"**. This is a naming difference; the mapping follows chapter order and question ranges.

## 4. Normalization — Question 444

The file [`domain4_security_operations.md`](domain4_security_operations.md) previously had a formatting defect in the **Question 444** heading: there was no period after the question number, so the parser did not match the heading and skipped the question.

- Fixed on **2026-10-20** by adding **only a period after the question number**.
- **No question text, options, answers, or explanations were changed.**
- Result: the number of ingested questions increased from **539** to **540**, the expected size of the question bank.

## 5. Generated Artifact Policy

- Run the parser from the **repository root**:

  ```bash
  python scripts/parse_exams.py
  ```

- The official artifact is **`parsed_questions.json` at the repository root only**.
- The parser reads only the `domain*.md` source files in this directory and does not process other files such as `README.md`.
- `examination/questions.json` was once a duplicate file not referenced by any script and **has been removed**, leaving a single artifact.
- Never edit `parsed_questions.json` by hand; always regenerate it with the parser.

## 6. Metadata in JSON

Every question in `parsed_questions.json` has the following flat provenance fields:

| Field | Type | Meaning |
|---|---|---|
| `source_ref` | string | Reference key in the registry; the current value is `"R1"`. |
| `source_chapter` | integer | Source book chapter number (1–5). |
| `source_question_no` | integer | Source book question number (1–540). |

`source_chapter` comes from the Domain → Chapter mapping in the parser script, and `source_question_no` is read from the question headings in the source Markdown files.

## 7. Copyright Restrictions

- The question text, options, answers, and explanations are the copyright of the **author and publisher**.
- Use is limited to **internal academic use and training within the project**.
- **Do not publish or redistribute the question text** outside the internal use scope.
- Cite `[R1]` / (Gorgotsias, 2024) in reports or documents that reference the question bank.
- Details and open risks: [`../docs/REFERENCES.md`](../docs/REFERENCES.md)

## 8. Source Editing Rules

Do not edit the exam content in this directory's Markdown files without doing all of the following:

1. Announce and record the reason for the edit.
2. Update the provenance in [`../docs/REFERENCES.md`](../docs/REFERENCES.md) if the edit affects question counts, question ranges, or Domain mapping.
3. Update the reference registry (`R-ID`, Revision History) when the change affects data provenance.
4. Re-run `python scripts/parse_exams.py` and verify that the resulting question count matches the expected count.

Formatting edits (for example, adding a period in a question heading) must not change the question content and must be recorded in Section 4 of this document.
