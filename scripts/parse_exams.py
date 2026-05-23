import os
import re
import json
import glob
import sys

def calculate_complexity(question_text, options, explanation):
    # A simple heuristic: longer questions and explanations tend to be harder.
    # We will use this to sort and distribute difficulties 1-5.
    length = len(question_text) + len(explanation)
    for k, v in options.items():
        length += len(v)
    return length

def parse_markdown_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    questions = []
    
    # Split content by "**Question" to get individual question blocks
    blocks = re.split(r'\*\*Question \d+\.\*\*', content)
    
    filename = os.path.basename(filepath)
    match = re.search(r'domain(\d+)', filename)
    domain_id = match.group(1) if match else "1"

    for block in blocks[1:]:
        block = block.strip()
        if not block:
            continue
            
        parts = re.split(r'\*?\*?\([A-D]\)\*?\*?', block)
        if len(parts) < 5:
            print(f"Skipping malformed block in {filename}...")
            continue
            
        question_text = parts[0].strip()
        
        option_a = parts[1].strip()
        option_b = parts[2].strip()
        option_c = parts[3].strip()
        option_d_parts = re.split(r'\*?\*?Explanation \d+[\.\:]', parts[4], maxsplit=1, flags=re.IGNORECASE)
        option_d = option_d_parts[0].strip()
        
        options_dict = {
            "A": option_a,
            "B": option_b,
            "C": option_c,
            "D": option_d
        }
        
        explanation_block = option_d_parts[1].strip() if len(option_d_parts) > 1 else ""
        
        correct_answer = "A"
        correct_match = re.search(r'Correct Answer:\s*([A-D])\.', explanation_block, re.IGNORECASE)
        if correct_match:
            correct_answer = correct_match.group(1).upper()
            
        # Calculate heuristic complexity instead of calling API
        score = calculate_complexity(question_text, options_dict, explanation_block)
        
        question_data = {
            "domain_id": f"domain{domain_id}",
            "type": "multiple_choice",
            "question_text": question_text,
            "options": options_dict,
            "correct_answer": correct_answer,
            "explanation": explanation_block,
            "complexity_score": score
        }
        
        questions.append(question_data)
        
    return questions

def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    target_dir = os.path.join(script_dir, "../examination")
    output_file = "parsed_questions.json"
    all_questions = []
    
    print("Starting deterministic parser (1-5 difficulty distribution)...")
    for filepath in glob.glob(os.path.join(target_dir, "*.md")):
        print(f"Parsing {filepath}...")
        questions = parse_markdown_file(filepath)
        all_questions.extend(questions)
        
    # Sort all questions by complexity score
    all_questions.sort(key=lambda x: x["complexity_score"])
    
    total = len(all_questions)
    # Distribute 1 to 5 evenly
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
        # Remove the temporary score key
        del q["complexity_score"]
        
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(all_questions, f, ensure_ascii=False, indent=2)
        
    print(f"Successfully parsed {len(all_questions)} questions and saved to {output_file}.")

if __name__ == "__main__":
    main()
