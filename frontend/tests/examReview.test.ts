import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  countReviewOutcomes,
  createInitialExpandedIds,
  filterReviewItems,
  formatDurationSeconds,
  formatSignedDelta,
  getReviewItemOutcome,
  getSortedOptionEntries,
  groupReviewItems,
  orderDistractorsForLearner,
  parseExamExplanation,
  resolveJumpToQuestion,
  type ExamDomainScore,
  type ExamReviewItem,
} from "../src/lib/examReview.ts";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "../..");

function makeSampleItem(overrides: Partial<ExamReviewItem>): ExamReviewItem {
  return {
    question_id: "q-1",
    domain_id: "domain1",
    domain_name: "General Security Concepts",
    sub_topic_id: "1.1",
    sub_topic_title: "Compare and contrast various types of security controls",
    type: "multiple_choice",
    question_text: "Sample question text?",
    options: { D: "Opt D", B: "Opt B", A: "Opt A", C: "Opt C" },
    correct_answer: "A",
    explanation:
      "Correct Answer: A. Opt A is right. Option B is incorrect. B fails. Option C is incorrect. C fails. Option D is incorrect. D fails.",
    order_index: 0,
    user_answer: "A",
    is_correct: true,
    is_unanswered: false,
    ...overrides,
  };
}

test("parseExamExplanation parses all 540 real questions in parsed_questions.json cleanly", () => {
  const parsedPath = path.join(repoRoot, "parsed_questions.json");
  const raw = fs.readFileSync(parsedPath, "utf8");
  const questions = JSON.parse(raw) as Array<{
    correct_answer: string;
    explanation: string;
  }>;

  assert.equal(questions.length, 540);
  for (let i = 0; i < questions.length; i++) {
    const q = questions[i];
    const res = parseExamExplanation(q.explanation, q.correct_answer);
    assert.equal(res.structured, true, `Question index ${i} failed structured parsing`);
    if (res.structured) {
      assert.equal(
        res.correctOption,
        q.correct_answer.toUpperCase(),
        `Question index ${i} parsed wrong correctOption`,
      );
      assert.ok(res.correctRationale.length > 0, `Question ${i} empty correctRationale`);
      assert.ok(res.distractors.length >= 1, `Question ${i} expected distractors`);
      for (const d of res.distractors) {
        assert.ok(d.rationale.length > 0, `Question ${i} empty distractor ${d.option}`);
      }
    }
  }
});

test("parseExamExplanation falls back gracefully on unstructured or empty text", () => {
  const emptyRes = parseExamExplanation("   ", "A");
  assert.deepEqual(emptyRes, { structured: false, raw: "" });

  const unstructured = "Digital signatures provide non-repudiation across public-key workflows.";
  const rawRes = parseExamExplanation(unstructured, "A");
  assert.deepEqual(rawRes, { structured: false, raw: unstructured });
});

test("orderDistractorsForLearner pins wrong user pick first and sorts remaining alphabetically", () => {
  const distractors = [
    { option: "D", rationale: "D wrong" },
    { option: "B", rationale: "B wrong" },
    { option: "C", rationale: "C wrong" },
  ];

  const pickedC = orderDistractorsForLearner(distractors, "c");
  assert.deepEqual(
    pickedC.map((d) => d.option),
    ["C", "B", "D"],
  );

  const unanswered = orderDistractorsForLearner(distractors, "");
  assert.deepEqual(
    unanswered.map((d) => d.option),
    ["B", "C", "D"],
  );
});

test("Smart Expand, filtering, domain grouping, and Jump Grid resolution preserve learning alignment", () => {
  const items: ExamReviewItem[] = [
    makeSampleItem({
      question_id: "q-d4-correct",
      domain_id: "domain4",
      domain_name: "Security Operations",
      order_index: 0,
      user_answer: "A",
      is_correct: true,
    }),
    makeSampleItem({
      question_id: "q-d1-wrong",
      domain_id: "domain1",
      domain_name: "General Security Concepts",
      order_index: 1,
      user_answer: "C",
      is_correct: false,
      is_unanswered: false,
    }),
    makeSampleItem({
      question_id: "q-d2-unanswered",
      domain_id: "domain2",
      domain_name: "Threats, Vulnerabilities, and Mitigations",
      order_index: 2,
      user_answer: "",
      is_correct: false,
      is_unanswered: true,
    }),
  ];

  // Smart Expand opens incorrect + unanswered only
  const initialExpanded = createInitialExpandedIds(items);
  assert.equal(initialExpanded.has("q-d4-correct"), false);
  assert.equal(initialExpanded.has("q-d1-wrong"), true);
  assert.equal(initialExpanded.has("q-d2-unanswered"), true);

  // Outcome counts sum to 3
  assert.deepEqual(countReviewOutcomes(items), {
    all: 3,
    correct: 1,
    incorrect: 1,
    unanswered: 1,
  });

  // Filtering
  assert.deepEqual(
    filterReviewItems(items, "incorrect", "all").map((q) => q.question_id),
    ["q-d1-wrong"],
  );
  assert.deepEqual(
    filterReviewItems(items, "unanswered", "all").map((q) => q.question_id),
    ["q-d2-unanswered"],
  );
  assert.deepEqual(
    filterReviewItems(items, "all", "domain4").map((q) => q.question_id),
    ["q-d4-correct"],
  );

  // Grouping: "domain" puts domain1 -> domain2 -> domain4 regardless of order_index
  const domainScores: ExamDomainScore[] = [
    { domain_id: "domain1", domain_name: "D1", total: 1, correct: 0, incorrect: 1, unanswered: 0, percentage: 0, threshold: 70, passed: false },
    { domain_id: "domain2", domain_name: "D2", total: 1, correct: 0, incorrect: 0, unanswered: 1, percentage: 0, threshold: 70, passed: false },
    { domain_id: "domain4", domain_name: "D4", total: 1, correct: 1, incorrect: 0, unanswered: 0, percentage: 100, threshold: 70, passed: true },
  ];
  const byDomain = groupReviewItems(items, "domain", domainScores);
  assert.deepEqual(
    byDomain.map((s) => s.domainId),
    ["domain1", "domain2", "domain4"],
  );
  assert.equal(byDomain[0].domainScore?.percentage, 0);

  // Grouping: "order" keeps 0 -> 1 -> 2 in single section
  const byOrder = groupReviewItems(items, "order", domainScores);
  assert.equal(byOrder.length, 1);
  assert.deepEqual(
    byOrder[0].items.map((q) => q.question_id),
    ["q-d4-correct", "q-d1-wrong", "q-d2-unanswered"],
  );

  // Jump Grid: jumping to collapsed correct item while filtered to "incorrect" resets filters and expands it
  const jumped = resolveJumpToQuestion(items[0], "incorrect", "domain1", initialExpanded);
  assert.equal(jumped.nextStatusFilter, "all");
  assert.equal(jumped.nextDomainFilter, "all");
  assert.equal(jumped.nextExpandedIds.has("q-d4-correct"), true);
  assert.equal(jumped.nextExpandedIds.has("q-d1-wrong"), true);
});

test("getSortedOptionEntries, formatDurationSeconds, formatSignedDelta, and getReviewItemOutcome produce deterministic output", () => {
  assert.deepEqual(
    getSortedOptionEntries({ D: "Four", B: "Two", A: "One", C: "Three" }).map(([k]) => k),
    ["A", "B", "C", "D"],
  );
  assert.equal(formatDurationSeconds(null), "—");
  assert.equal(formatDurationSeconds(65), "01:05");
  assert.equal(formatDurationSeconds(3661), "01:01:01");
  assert.equal(formatSignedDelta(12.34), "+12.3%");
  assert.equal(formatSignedDelta(-5), "-5.0%");
  assert.equal(formatSignedDelta(0), "0.0%");
  assert.equal(getReviewItemOutcome(makeSampleItem({ is_correct: true, user_answer: "A" })), "correct");
  assert.equal(getReviewItemOutcome(makeSampleItem({ is_correct: false, user_answer: "B", is_unanswered: false })), "incorrect");
  assert.equal(getReviewItemOutcome(makeSampleItem({ is_correct: false, user_answer: "", is_unanswered: true })), "unanswered");
});
