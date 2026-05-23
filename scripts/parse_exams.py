import os
import re
import json
import glob
import sys

try:
    import google.generativeai as genai
except ImportError:
    print("Please install google-generativeai first: pip install google-generativeai")
    sys.exit(1)

# Retrieve the API key from environment variables
api_key = os.environ.get("GOOGLE_API_KEY")

if not api_key:
    print("ERROR: GOOGLE_API_KEY environment variable not found.")
    print("To get a FREE API key, visit: https://aistudio.google.com/app/apikey")
    print("Then set it in your terminal:")
    print('  $env:GOOGLE_API_KEY="your-api-key" (for PowerShell)')
    sys.exit(1)

genai.configure(api_key=api_key)

# Use gemini-2.5-pro
model = genai.GenerativeModel('gemini-2.5-pro')

def evaluate_difficulty(question_text, options, explanation):
    prompt = f"""
    You are an expert cybersecurity examiner. Read the following question, options, and explanation.
    Evaluate the difficulty of this question on a scale of 1 to 3:
    1 = Easy (Basic definitions, common terms, straightforward recall)
    2 = Medium (Scenario-based application, standard operational concepts)
    3 = Hard (Complex multi-step analysis, niche technical knowledge, tricky distractors)
    
    Please provide a varied and realistic distribution across all your evaluations.
    Only return a single integer representing the difficulty level. No other text.
    
    Question: {question_text}
    Options: {json.dumps(options)}
    Explanation: {explanation}
    """
    try:
        response = model.generate_content(prompt)
        difficulty = int(response.text.strip())
        if difficulty not in [1, 2, 3]:
            return 2 # default fallback
        return difficulty
    except Exception as e:
        print(f"Error evaluating difficulty: {e}")
        return 2 # default fallback

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
            
        # Call Gemini API
        difficulty = evaluate_difficulty(question_text, options_dict, explanation_block)
        
        question_data = {
            "domain_id": f"domain{domain_id}",
            "type": "multiple_choice",
            "question_text": question_text,
            "options": options_dict,
            "correct_answer": correct_answer,
            "explanation": explanation_block,
            "difficulty": difficulty
        }
        
        questions.append(question_data)
        
    return questions

def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    target_dir = os.path.join(script_dir, "../examination")
    output_file = "parsed_questions.json"
    all_questions = []
    
    print("Starting parser using Google AI Studio API Key...")
    for filepath in glob.glob(os.path.join(target_dir, "*.md")):
        print(f"Parsing {filepath}...")
        questions = parse_markdown_file(filepath)
        all_questions.extend(questions)
        
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(all_questions, f, ensure_ascii=False, indent=2)
        
    print(f"Successfully parsed {len(all_questions)} questions and saved to {output_file}.")

if __name__ == "__main__":
    main()
