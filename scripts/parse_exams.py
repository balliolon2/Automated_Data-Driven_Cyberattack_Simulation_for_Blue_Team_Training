import os
import re
import json
import glob
import sys


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

    blocks = re.split(r"\*\*Question \d+\.\*\*", content)

    filename = os.path.basename(filepath)
    match = re.search(r"domain(\d+)", filename)
    domain_id = match.group(1) if match else "1"

    for block in blocks[1:]:
        block = block.strip()
        if not block:
            continue

        parts = re.split(r"\*?\*?\([A-D]\)\*?\*?", block)
        if len(parts) < 5:
            print(f"Skipping malformed block in {filename}...")
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
            "domain_id": f"domain{domain_id}",
            "type": "multiple_choice",
            "question_text": question_text,
            "options": options_dict,
            "correct_answer": correct_answer,
            "explanation": explanation_block,
            "complexity_score": score,
        }

        questions.append(question_data)

    return questions


def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    target_dir = os.path.join(script_dir, "../examination")
    output_file = "parsed_questions.json"
    all_questions = []

    source_files = glob.glob(os.path.join(target_dir, "*.md"))
    if not source_files:
        print(f"No examination source files found in {target_dir}; refusing to seed an empty pool.", file=sys.stderr)
        sys.exit(1)

    print("Starting deterministic parser (1-5 difficulty distribution)...")
    for filepath in source_files:
        print(f"Parsing {filepath}...")
        questions = parse_markdown_file(filepath)
        all_questions.extend(questions)

    if not all_questions:
        print("Parsed zero questions from examination sources; refusing to seed an empty pool.", file=sys.stderr)
        sys.exit(1)

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
