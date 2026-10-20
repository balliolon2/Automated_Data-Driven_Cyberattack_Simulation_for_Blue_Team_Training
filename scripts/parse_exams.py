import os
import re
import json
import glob
import sys
from collections import Counter


SOURCE_REF = "R1"

DOMAIN_CHAPTER = {f"domain{i}": i for i in range(1, 6)}

EXPECTED_R1_COUNTS = {
    "domain1": 110,
    "domain2": 110,
    "domain3": 90,
    "domain4": 150,
    "domain5": 80,
}


def normalize_text(text):
    text = re.sub(r"\*\*", "", text)
    text = re.sub(r"\s+", " ", text)
    return text.strip()


def calculate_complexity(question_text, options, explanation):
    length = len(question_text) + len(explanation)
    for v in options.values():
        length += len(v)
    return length


def parse_markdown_file(filepath):
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    questions = []

    blocks = re.split(r"\*\*Question (\d+)\.\*\*", content)

    filename = os.path.basename(filepath)
    match = re.search(r"domain(\d+)", filename)
    domain_id = match.group(1) if match else "1"
    domain_key = f"domain{domain_id}"

    for i in range(1, len(blocks) - 1, 2):
        question_no = int(blocks[i])
        block = blocks[i + 1].strip()
        if not block:
            continue

        parts = re.split(r"\*?\*?\([A-D]\)\*?\*?", block)
        if len(parts) < 5:
            print(f"Skipping malformed block in {filename} (question {question_no})...")
            continue

        question_text = normalize_text(parts[0])

        option_a = normalize_text(parts[1])
        option_b = normalize_text(parts[2])
        option_c = normalize_text(parts[3])
        option_d_parts = re.split(r"\*?\*?Explanation \d+[\.\:]", parts[4], maxsplit=1, flags=re.IGNORECASE)
        option_d = normalize_text(option_d_parts[0])

        options_dict = {
            "A": option_a,
            "B": option_b,
            "C": option_c,
            "D": option_d,
        }

        raw_explanation = option_d_parts[1].strip() if len(option_d_parts) > 1 else ""

        correct_answer = "A"
        correct_match = re.search(r"Correct Answer:\s*([A-D])\.", raw_explanation, re.IGNORECASE)
        if correct_match:
            correct_answer = correct_match.group(1).upper()

        explanation_block = normalize_text(raw_explanation)

        score = calculate_complexity(question_text, options_dict, explanation_block)

        question_data = {
            "domain_id": domain_key,
            "type": "multiple_choice",
            "question_text": question_text,
            "options": options_dict,
            "correct_answer": correct_answer,
            "explanation": explanation_block,
            "complexity_score": score,
            "source_ref": SOURCE_REF,
            "source_chapter": DOMAIN_CHAPTER[domain_key],
            "source_question_no": question_no,
        }

        questions.append(question_data)

    return questions


def validate_questions(questions):
    print("=" * 60)
    print(f"Post-parse validation (source {SOURCE_REF}): {len(questions)} questions")

    domain_stats = {}
    for q in questions:
        domain = q["domain_id"]
        number = q["source_question_no"]
        stats = domain_stats.setdefault(domain, {"count": 0, "min": number, "max": number})
        stats["count"] += 1
        stats["min"] = min(stats["min"], number)
        stats["max"] = max(stats["max"], number)

    for domain in sorted(domain_stats):
        stats = domain_stats[domain]
        expected = EXPECTED_R1_COUNTS.get(domain)
        expected_label = f"expected {expected}" if expected is not None else "no expected count"
        warning = ""
        if expected is not None and stats["count"] != expected:
            warning = "  WARNING: unexpected count"
        print(
            f"  {domain}: count={stats['count']} ({expected_label}), "
            f"source_question_no={stats['min']}-{stats['max']}{warning}"
        )

    for domain in sorted(set(EXPECTED_R1_COUNTS) - set(domain_stats)):
        print(f"  WARNING: {domain} has no parsed questions (expected {EXPECTED_R1_COUNTS[domain]})")

    number_counts = Counter(q["source_question_no"] for q in questions)
    duplicates = sorted(number for number, count in number_counts.items() if count > 1)
    if duplicates:
        print(f"  WARNING: duplicate source_question_no values: {duplicates}")
    else:
        print("  duplicate source_question_no values: none")

    out_of_range = sorted(number for number in number_counts if not 1 <= number <= 540)
    if out_of_range:
        print(f"  WARNING: source_question_no outside 1-540: {out_of_range}")
    else:
        print("  source_question_no outside 1-540: none")
    print("=" * 60)


def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    target_dir = os.path.join(script_dir, "../examination")
    output_file = "parsed_questions.json"
    all_questions = []

    source_files = glob.glob(os.path.join(target_dir, "domain*.md"))
    source_files.sort()
    if not source_files:
        print(f"No examination source files found in {target_dir} matching domain*.md; refusing to seed an empty pool.", file=sys.stderr)
        sys.exit(1)

    print("Starting deterministic parser (1-5 difficulty distribution)...")
    for filepath in source_files:
        print(f"Parsing {filepath}...")
        questions = parse_markdown_file(filepath)
        all_questions.extend(questions)

    if not all_questions:
        print("Parsed zero questions from examination sources; refusing to seed an empty pool.", file=sys.stderr)
        sys.exit(1)

    validate_questions(all_questions)

    all_questions.sort(key=lambda x: x["complexity_score"])

    total = len(all_questions)
    for i, q in enumerate(all_questions):
        percentile = i / total
        if percentile < 0.2:
            q["difficulty"] = 1
        elif percentile < 0.4:
            q["difficulty"] = 2
        elif percentile < 0.6:
            q["difficulty"] = 3
        elif percentile < 0.8:
            q["difficulty"] = 4
        else:
            q["difficulty"] = 5
        del q["complexity_score"]

    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(all_questions, f, ensure_ascii=False, indent=2)

    print(f"Successfully parsed {len(all_questions)} questions and saved to {output_file}.")


if __name__ == "__main__":
    main()
