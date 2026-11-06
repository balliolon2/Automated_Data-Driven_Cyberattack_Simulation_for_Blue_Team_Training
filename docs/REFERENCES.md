# References & Source Provenance

This document is the canonical registry of all references used in the **Automated Data-Driven Cyberattack Simulation for Blue Team Training** project. It records where each dataset comes from, which Domain it is mapped to, what usage restrictions apply, and how it must be cited in reports or publications.

This document is documentation only. It does not change the database schema or the data formats used by the system.

## 1. Purpose

- Define a stable reference key for every data source used by the project.
- Record the provenance of the Pre-test/Post-test question bank, from chapter and question range down to the source file.
- Define a consistent citation format across the project, in project documents, reports, and source code.
- Record copyright restrictions and open risks so that stakeholders can make informed decisions.
- Serve as a template for adding new sources in the future.

## 2. Status and Dates

| Item | Value |
|---|---|
| Status | Active — used as the project's primary registry |
| Date created | 2026-10-20 |
| Last updated | 2026-10-20 |
| Document owner | Project Owner |
| Scope | Content sources and data ingested into the database |

## 3. Citation Convention

The project uses a two-tier citation convention, with **APA 7th** as the primary citation style:

1. **Project key `[R1]`** — used for internal traceability, such as in design documents, the Decision Log, issues, and comments in code or scripts. This key is bound to the registry in this document only.
2. **In-text citation `(Gorgotsias, 2024)`** — used in narrative sentences in reports, theses, and publications, following APA 7th.
3. **BibTeX key `gorgotsias2024securityplus`** — used when producing documents with LaTeX or tools that read `docs/references.bib`.

In project text and reports, use both `[R1]` and `(Gorgotsias, 2024)` together: `[R1]` identifies the source at the registry level, while `(Gorgotsias, 2024)` provides the academic citation under APA 7th.

Example:

> The project's Pre-test and Post-test question banks are reproduced from the Security+ SY0-701 practice-test book [R1] (Gorgotsias, 2024).

## 4. Reference Registry

| R-ID | Short citation | Source type | Used in project | Status |
|---|---|---|---|---|
| R1 | (Gorgotsias, 2024) | Book — practice tests & PBQs | Pre-test/Post-test question bank, 540 questions | Active |

`R-ID` codes are permanent. They must never be reused or redefined, even after the reference is retired.

## 5. R1 — Full Reference (APA 7th)

Gorgotsias, N. (2024). *CompTIA Security+ SY0-701 practice tests & PBQs: Exam SY0-701* (3rd ed.). ExamsDigest LLC; LabsDigest LLC.

| Field | Value |
|---|---|
| Author | Gorgotsias, Nikolaos |
| Year | 2024 |
| Title | CompTIA Security+ SY0-701 Practice Tests & PBQs: Exam SY0-701 |
| Edition | 3rd |
| Publisher | ExamsDigest LLC and LabsDigest LLC |
| ISBN | 9781953041241 |
| Pages | 1620 |
| Language | English |

Notes:

- The edition information (3rd ed.) is recorded as confirmed by the Project Owner on 2026-10-20.
- No place of publication is recorded in this registry, and none should be invented.
- The BibTeX entry is in [`docs/references.bib`](references.bib) under the key `gorgotsias2024securityplus`.

## 6. Provenance and Chapter → Domain Mapping

Chapters 1–5 of the source book are used as the source of the question bank, mapped one chapter per project Domain as follows:

| Chapter | Book TOC chapter title | domain_id | Project domain name | Question range | Source file | Count in original | Count ingested |
|---|---|---|---|---|---|---|---|
| Ch1 | General Security Concepts | domain1 | Security Concepts | Q1–Q110 | `examination/domain1_security_concepts.md` | 110 | 110 |
| Ch2 | Threats, Vulnerabilities, and Mitigations | domain2 | Threats, Vulnerabilities, and Mitigations | Q111–Q220 | `examination/domain2_threats_vulnerabilities.md` | 110 | 110 |
| Ch3 | Implementation | domain3 | Security Architecture | Q221–Q310 | `examination/domain3_security_architecture.md` | 90 | 90 |
| Ch4 | Security Operations | domain4 | Security Operations | Q311–Q460 | `examination/domain4_security_operations.md` | 150 | 150 |
| Ch5 | Security Program Management and Oversight | domain5 | Security Management | Q461–Q540 | `examination/domain5_security_management.md` | 80 | 80 |
| **Total** | | | | **Q1–Q540** | | **540** | **540** |

Note: The ingested counts in the table reflect the correct state after normalization (see Section 10), and `parsed_questions.json` at the repository root records 540 entries. If the counts do not match, run `python scripts/parse_exams.py` from the repository root to regenerate.

## 7. Use of the Question Bank

- The system generates Pre-test and Post-test sets of **30 questions each** per learner.
- The per-Domain quota in each set is **Domain 1 = 4, Domain 2 = 7, Domain 3 = 5, Domain 4 = 8, Domain 5 = 6** (30 total).
- The system enforces **Zero Overlap**: questions that appear in a learner's Pre-test are not used in the same learner's Post-test.
- **Question text, options, answers, and explanations are content-verbatim**: no rewording or editing; the substantive content remains unchanged.
- The parser performs Markdown/whitespace normalization during extraction (for example, stripping `**` and collapsing consecutive spaces to one). This is existing extraction behavior and does not change the substantive content.
- The only edit to the source text is the single added period in the Question 444 heading, recorded in Section 10.

## 8. Note: Chapter 3 Title Does Not Match the Domain Name

The source book names Chapter 3 **"Implementation"** in its table of contents, while the project maps this chapter to **CompTIA Domain 3 "Security Architecture"**.

This difference is a naming difference, not a data discrepancy:

- The mapping follows the **chapter order and question ranges** (Q221–Q310) as the primary basis.
- Project Domain names follow the CompTIA Security+ SY0-701 Domain structure used by the system.
- When referring to Chapter 3 in reports, cite both the book's chapter title and the project's Domain name to avoid confusion.

## 9. Data Traceability

The parser-generated file (`parsed_questions.json` at the repository root) records flat provenance fields on every question:

| Field | Type | Meaning |
|---|---|---|
| `source_ref` | string | Reference key in this registry; the current value is `"R1"`. |
| `source_chapter` | integer | Source book chapter number (1–5). |
| `source_question_no` | integer | Source book question number (1–540). |

How it works:

- The parser sets `source_ref` from the source reference constant and derives `source_chapter` from the Domain → Chapter mapping table (`domain1` → Chapter 1, …, `domain5` → Chapter 5).
- `source_question_no` is read directly from the question headings in the source Markdown files.
- **No database schema changes** are made. These fields are artifact metadata used for traceability, not new PostgreSQL columns.
- `examination/questions.json` was once a duplicate file not referenced by any script and has been removed. The only official artifact is `parsed_questions.json` at the repository root.

## 10. Normalization Record — Question 444

Date: **2026-10-20**

The source file `examination/domain4_security_operations.md` had a formatting defect in the **Question 444** heading: **there was no period after the question number**, so the parser's heading pattern did not match and the question was skipped.

Fix:

- Add **only a period after the question number** (`**Question 444**` → `**Question 444.**`).
- **No question text, options, answers, or explanations were changed.**

Result:

| State | Questions ingested |
|---|---|
| Before normalization | 539 |
| After normalization | **540** |

The expected total for the question bank is **540 questions**, and this number must match the total in the table in Section 6.

## 11. Usage Restrictions and Copyright

- The source book and question bank are **copyrighted** works. The question text, options, answers, and explanations remain the copyright of the **author and publisher**.
- This project uses this material for **internal academic use and training only**.
- **Do not publish, resell, or redistribute the question text or explanations** in any form outside the project's internal use scope.
- Reports, presentations, and publications that reference the question bank **must cite the source** using the formats in Sections 3 and 5.
- This document records facts and usage restrictions only. It is **not legal advice** and does not confirm or grant any rights.

## 12. Open Risk

| Item | Detail |
|---|---|
| Risk | The project's current public GitHub repository includes a verbatim copy of the question bank, which may conflict with the usage restrictions of the source book. |
| Status | **Open** as of 2026-10-20 |
| Owner | Project Owner |
| Decision | Under discussion with the academic advisor; no conclusion yet. |
| Permission document | **No document confirming permission from the publisher exists**, and this document is not a claim that permission has been granted. |
| Impact | May affect publication of the repository, submission of the work, and downstream use of the work. |

Options considered (not confirmed): restrict repository access, separate the question bank from the public area, or replace some questions with self-authored items. A decision is pending from the Project Owner and the academic advisor.

## 13. Adding a New Reference

When the project adopts a new content source, the following steps must be completed **before** parsing or ingesting the data:

1. Assign the next `R-ID` in sequence (`R2`, `R3`, …); never reuse a code.
2. Add the full APA 7th reference entry in the registry section of this document.
3. Add a BibTeX entry to [`docs/references.bib`](references.bib) with a key consistent with the author name and year.
4. Record provenance: chapter/section, data range, source file, item count, and mapping to project Domains or entities.
5. Record usage and copyright restrictions, including any open risks.
6. Update the last-updated date and add an entry to the Revision History.

Sources that do not complete these steps must not enter the parsing or ingestion pipeline.

## 14. Revision History

| Date | Item |
|---|---|
| 2026-10-20 | Created the document and registered reference R1 (Gorgotsias, 2024), with the provenance table, Question 444 normalization record, copyright restrictions, and open risk. |
