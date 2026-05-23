import os
import re
import json
import glob
import google.auth
from google.auth.transport.requests import Request
import vertexai
from vertexai.generative_models import GenerativeModel

# Google Cloud Vertex AI Authentication
# This uses Application Default Credentials (ADC).
# To authenticate, run: gcloud auth application-default login
# Or set the GOOGLE_APPLICATION_CREDENTIALS environment variable pointing to your Service Account JSON.
try:
    credentials, project_id = google.auth.default()
    # If project_id is not automatically detected, fall back to environment variable
    if not project_id:
        project_id = os.environ.get("GCP_PROJECT_ID")
        
    if not project_id:
        raise ValueError("GCP Project ID could not be detected. Please set GCP_PROJECT_ID environment variable.")
        
    vertexai.init(project=project_id, location=os.environ.get("GCP_LOCATION", "us-central1"))
    # Use gemini-1.5-pro for Gemini Pro
    model = GenerativeModel("gemini-1.5-pro")
    print(f"Authenticated successfully with Google Cloud. Project ID: {project_id}")
except Exception as e:
    print(f"Failed to authenticate with Google Cloud: {e}")
    print("Please ensure Google Cloud SDK is installed and you have run 'gcloud auth application-default login' or set GOOGLE_APPLICATION_CREDENTIALS.")
    sys.exit(1)

def evaluate_difficulty(question_text, options, explanation):
    prompt = f"""
    You are an expert cybersecurity examiner. Read the following question, options, and explanation.
    Evaluate the difficulty of this question on a scale of 1 to 3 (1 = Easy, 2 = Medium, 3 = Hard).
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

    # Regex patterns (adjust as needed based on exact markdown formatting)
    questions = []
    
    # Split content by "**Question" to get individual question blocks
    blocks = re.split(r'\*\*Question \d+\.\*\*', content)
    
    # Identify domain from filename (e.g., domain1_security_concepts.md)
    filename = os.path.basename(filepath)
    match = re.search(r'domain(\d+)', filename)
    domain_id = match.group(1) if match else "1"

    for block in blocks[1:]: # Skip the first split which is before Question 1
        block = block.strip()
        if not block:
            continue
            
        # Extract Question text
        parts = re.split(r'\*\*\([A-D]\)\*\*|\([A-D]\)', block)
        if len(parts) < 5:
            print(f"Skipping malformed block in {filename}...")
            continue
            
        question_text = parts[0].strip()
        
        # Options
        option_a = parts[1].strip()
        option_b = parts[2].strip()
        option_c = parts[3].strip()
        option_d_parts = re.split(r'\*\*Explanation \d+\.', parts[4])
        option_d = option_d_parts[0].strip()
        
        options_dict = {
            "A": option_a,
            "B": option_b,
            "C": option_c,
            "D": option_d
        }
        
        # Explanation and Correct Answer
        explanation_block = option_d_parts[1].strip() if len(option_d_parts) > 1 else ""
        
        correct_answer = "A" # Default
        correct_match = re.search(r'Correct Answer:\s*([A-D])\.', explanation_block, re.IGNORECASE)
        if correct_match:
            correct_answer = correct_match.group(1).upper()
            
        # Evaluate difficulty using Gemini
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
    target_dir = "../examination"
    output_file = "parsed_questions.json"
    all_questions = []
    
    # Process all markdown files in the examination folder
    for filepath in glob.glob(os.path.join(target_dir, "*.md")):
        print(f"Parsing {filepath}...")
        questions = parse_markdown_file(filepath)
        all_questions.extend(questions)
        
    # Save to JSON
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(all_questions, f, ensure_ascii=False, indent=2)
        
    print(f"Successfully parsed {len(all_questions)} questions and saved to {output_file}.")

if __name__ == "__main__":
    import sys
    main()
