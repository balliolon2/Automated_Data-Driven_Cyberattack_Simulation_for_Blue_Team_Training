import argparse
from concurrent.futures import ThreadPoolExecutor, as_completed
import hashlib
import json
import os
import re
import sys
import time
import urllib.error
import urllib.request

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.abspath(os.path.join(SCRIPT_DIR, ".."))
DEFAULT_PARSED_PATH = os.path.join(REPO_ROOT, "parsed_questions.json")
DEFAULT_TRANSLATIONS_PATH = os.path.join(SCRIPT_DIR, "question_translations_th.json")
THAI_CHAR_RE = re.compile(r"[\u0E00-\u0E7F]")
ORDERED_DOMAINS = ("domain1", "domain2", "domain3", "domain4", "domain5")


def question_key(question):
    return f"{question['domain_id']}:{question['source_question_no']}"


def compute_question_sha256(question):
    text = (question.get("question_text") or "").encode("utf-8")
    return hashlib.sha256(text).hexdigest()


def contains_thai(text):
    return bool(THAI_CHAR_RE.search(text or ""))


def validate_canonical_english_pool(questions):
    errors = []
    for q in questions:
        key = question_key(q)
        if contains_thai(q.get("question_text", "")):
            errors.append(f"{key}: canonical question_text contains Thai characters")
        if contains_thai(q.get("explanation", "")):
            errors.append(f"{key}: canonical explanation contains Thai characters")
        opts = q.get("options") or {}
        if not isinstance(opts, dict) or set(opts.keys()) != {"A", "B", "C", "D"}:
            errors.append(f"{key}: canonical options must contain keys A-D")
            continue
        for opt_key, opt_val in opts.items():
            if contains_thai(str(opt_val)):
                errors.append(f"{key}: canonical option {opt_key} contains Thai characters")
    return errors


def validate_translation_entry(question, entry):
    key = question_key(question)
    errors = []
    if not isinstance(entry, dict):
        return [f"{key}: entry must be a JSON object"]

    if entry.get("domain_id") != question.get("domain_id") or entry.get("source_question_no") != question.get("source_question_no"):
        errors.append(f"{key}: domain_id/source_question_no mismatch")

    expected_sha = compute_question_sha256(question)
    if entry.get("en_sha256") != expected_sha:
        errors.append(f"{key}: en_sha256 mismatch (expected {expected_sha}, got {entry.get('en_sha256')})")

    th = entry.get("th")
    if not isinstance(th, dict):
        return [f"{key}: missing 'th' object"]

    q_text = th.get("question_text")
    if not isinstance(q_text, str) or not q_text.strip() or not contains_thai(q_text):
        errors.append(f"{key}: th.question_text must contain non-empty Thai text")

    en_opts = question.get("options") or {}
    th_opts = th.get("options")
    if not isinstance(th_opts, dict) or set(th_opts.keys()) != set(en_opts.keys()):
        errors.append(f"{key}: th.options keys must match canonical options {sorted(en_opts.keys())}")
    else:
        for opt_k, opt_v in th_opts.items():
            if not isinstance(opt_v, str) or not opt_v.strip():
                errors.append(f"{key}: th.options[{opt_k}] must be non-empty")

    correct_ans = (question.get("correct_answer") or "").strip().upper()
    if correct_ans not in ("A", "B", "C", "D"):
        errors.append(f"{key}: invalid canonical correct_answer '{correct_ans}'")
        return errors

    exp = th.get("explanation")
    if not isinstance(exp, str) or not exp.strip() or not contains_thai(exp):
        errors.append(f"{key}: th.explanation must contain non-empty Thai text")
    else:
        stripped = exp.strip()
        if not re.match(rf"^คำตอบที่ถูกต้อง:\s*{re.escape(correct_ans)}\.(?:\s|$)", stripped):
            errors.append(f"{key}: th.explanation must start with 'คำตอบที่ถูกต้อง: {correct_ans}.'")
        for opt_k in sorted(en_opts.keys()):
            if opt_k == correct_ans:
                continue
            if not re.search(rf"ตัวเลือก\s+{re.escape(opt_k)}\s+ไม่ถูกต้อง", stripped):
                errors.append(f"{key}: th.explanation missing distractor marker 'ตัวเลือก {opt_k} ไม่ถูกต้อง'")

    return errors


def validate_translations_dataset(questions, translations_map, require_full=True):
    errors = validate_canonical_english_pool(questions)
    if not isinstance(translations_map, dict):
        return ["translations dataset root must be a JSON object"], {}

    q_by_key = {question_key(q): q for q in questions}
    per_domain = {d: 0 for d in ORDERED_DOMAINS}

    for key in sorted(translations_map.keys()):
        if key not in q_by_key:
            errors.append(f"{key}: unknown key not found in parsed_questions.json")
            continue
        q = q_by_key[key]
        entry_errs = validate_translation_entry(q, translations_map[key])
        if entry_errs:
            errors.extend(entry_errs)
        else:
            per_domain[q["domain_id"]] = per_domain.get(q["domain_id"], 0) + 1

    if require_full:
        missing = sorted(set(q_by_key.keys()) - set(translations_map.keys()))
        if missing:
            errors.append(f"missing {len(missing)} translated questions (first 5: {', '.join(missing[:5])})")
    else:
        for dom in ORDERED_DOMAINS:
            if per_domain.get(dom, 0) < 2:
                errors.append(f"{dom}: partial validation requires at least 2 valid translated questions (got {per_domain.get(dom, 0)})")

    return errors, per_domain


def load_dotenv_fallback(env_path=None):
    cfg = {
        "LLM_API_KEY": os.environ.get("LLM_API_KEY", "").strip(),
        "LLM_BASE_URL": os.environ.get("LLM_BASE_URL", "").strip() or "https://api.openai.com/v1",
        "LLM_MODEL": os.environ.get("LLM_MODEL", "").strip() or "gpt-4o-mini",
    }
    target = env_path or os.path.join(REPO_ROOT, ".env")
    if os.path.isfile(target):
        with open(target, "r", encoding="utf-8") as f:
            for line in f:
                s = line.strip()
                if not s or s.startswith("#") or "=" not in s:
                    continue
                k, v = s.split("=", 1)
                k = k.strip()
                v = v.strip().strip('"').strip("'")
                if k in cfg and not cfg[k]:
                    cfg[k] = v
                elif k in ("LLM_BASE_URL", "LLM_MODEL") and k not in os.environ and v:
                    cfg[k] = v
    return cfg


def select_pilot_questions(questions, per_domain=2):
    picked = []
    for dom in ORDERED_DOMAINS:
        dom_qs = [q for q in questions if q.get("domain_id") == dom]
        picked.extend(dom_qs[:per_domain])
    return picked


def _extract_json_object(raw_text):
    cleaned = raw_text.strip()
    if cleaned.startswith("```"):
        cleaned = re.sub(r"^```(?:json)?\s*", "", cleaned)
        cleaned = re.sub(r"\s*```$", "", cleaned)
    start = cleaned.find("{")
    end = cleaned.rfind("}")
    if start >= 0 and end > start:
        cleaned = cleaned[start : end + 1]
    return json.loads(cleaned)


def translate_batch_via_llm(batch_questions, llm_cfg, retry_hint=""):
    payload = [
        {
            "key": question_key(q),
            "domain_id": q["domain_id"],
            "question_text": q["question_text"],
            "options": q["options"],
            "correct_answer": q["correct_answer"],
            "explanation": q["explanation"],
        }
        for q in batch_questions
    ]

    system_prompt = (
        "You translate CompTIA Security+ (SY0-701) exam questions from English into natural, accessible "
        "Hybrid Bilingual Thai for Thai high-school learners.\n"
        "STRICT RULES:\n"
        "1. Preserve canonical cybersecurity acronyms, technical terms, protocols, tools, file paths, commands, ports, and numbers in English inline inside natural Thai phrasing "
        "(e.g., 'Firewall', 'Phishing', 'Least Privilege', 'ALE', 'SIEM', 'TCP/443', 'AES-256').\n"
        "2. Translate question_text and ALL options (A, B, C, D) accurately. If an option is purely numeric/currency/port/acronym (like '$100,000', '443', 'AES-256'), keep that value verbatim.\n"
        "3. Write idiomatic, smooth Thai — avoid stiff word-for-word machine translations or awkward calques "
        "(never write awkward phrases like 'ถูกเรียกว่าอะไรกันแต่ทั่วไป' or literal 'มัน...' pronouns; use natural Thai such as 'เรียกว่าอะไรโดยทั่วไป?' and natural subject omission).\n"
        "4. Format explanation EXACTLY in this structured Thai pattern (preserving letter positions and exact correct_answer verbatim):\n"
        "   'คำตอบที่ถูกต้อง: <CORRECT_LETTER>. <อธิบายเหตุผลที่ถูกต้อง> ตัวเลือก <WRONG_1> ไม่ถูกต้อง <เหตุผล> ตัวเลือก <WRONG_2> ไม่ถูกต้อง <เหตุผล> ตัวเลือก <WRONG_3> ไม่ถูกต้อง <เหตุผล>'\n"
        "5. Every non-correct option letter in ['A','B','C','D'] MUST appear with 'ตัวเลือก <K> ไม่ถูกต้อง'.\n"
        "6. Output ONLY valid JSON matching: "
        '{"items":[{"key":"domainX:N","question_text":"...","options":{"A":"...","B":"...","C":"...","D":"..."},"explanation":"..."}]}'
    )
    user_content = f"Translate these {len(payload)} CompTIA Security+ questions cleanly:\n{json.dumps(payload, ensure_ascii=False)}"
    if retry_hint:
        user_content += f"\n\nFix these validation errors from the prior attempt:\n{retry_hint}"

    req_body = json.dumps(
        {
            "model": llm_cfg["LLM_MODEL"],
            "temperature": 0.2,
            "response_format": {"type": "json_object"},
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_content},
            ],
        }
    ).encode("utf-8")

    endpoint = llm_cfg["LLM_BASE_URL"].rstrip("/") + "/chat/completions"
    req = urllib.request.Request(
        endpoint,
        data=req_body,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {llm_cfg['LLM_API_KEY']}",
        },
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=120) as resp:
        data = json.loads(resp.read().decode("utf-8"))
    content = data["choices"][0]["message"]["content"]
    parsed = _extract_json_object(content)
    items = parsed.get("items") or []
    return {it["key"]: it for it in items if isinstance(it, dict) and "key" in it}


def translate_and_validate_group(group, llm_cfg):
    hint = ""
    by_key = {question_key(q): q for q in group}
    pending = list(group)
    accepted = {}

    for attempt in range(4):
        if not pending:
            break
        try:
            raw_map = translate_batch_via_llm(pending, llm_cfg, retry_hint=hint)
        except Exception as exc:
            hint = f"Prior call failed: {exc}"
            time.sleep(1.5 * (attempt + 1))
            continue

        next_pending = []
        errs_for_retry = []
        for q in pending:
            k = question_key(q)
            raw_item = raw_map.get(k)
            if not raw_item:
                next_pending.append(q)
                errs_for_retry.append(f"{k}: missing from output")
                continue
            candidate = {
                "domain_id": q["domain_id"],
                "source_question_no": q["source_question_no"],
                "en_sha256": compute_question_sha256(q),
                "th": {
                    "question_text": str(raw_item.get("question_text", "")).strip(),
                    "options": {
                        ok: str(ov).strip()
                        for ok, ov in (raw_item.get("options") or {}).items()
                    },
                    "explanation": str(raw_item.get("explanation", "")).strip(),
                },
            }
            item_errs = validate_translation_entry(q, candidate)
            if item_errs:
                next_pending.append(q)
                errs_for_retry.extend(item_errs)
            else:
                accepted[k] = candidate

        pending = next_pending
        hint = "\n".join(errs_for_retry[:10])
        if pending and len(pending) > 1 and attempt >= 1:
            # Split failing group down to individual retries on later attempts
            for sq in list(pending):
                sk = question_key(sq)
                try:
                    single_map = translate_batch_via_llm([sq], llm_cfg, retry_hint=hint)
                    ri = single_map.get(sk)
                    if ri:
                        cand = {
                            "domain_id": sq["domain_id"],
                            "source_question_no": sq["source_question_no"],
                            "en_sha256": compute_question_sha256(sq),
                            "th": {
                                "question_text": str(ri.get("question_text", "")).strip(),
                                "options": {ok: str(ov).strip() for ok, ov in (ri.get("options") or {}).items()},
                                "explanation": str(ri.get("explanation", "")).strip(),
                            },
                        }
                        if not validate_translation_entry(sq, cand):
                            accepted[sk] = cand
                            pending = [p for p in pending if question_key(p) != sk]
                except Exception:
                    pass

    if pending:
        missing_keys = ", ".join(question_key(q) for q in pending)
        raise RuntimeError(f"Failed translating questions after retries: {missing_keys}; last_err={hint}")
    return accepted


def save_translations_sorted(path, translations_map, questions):
    order_index = {question_key(q): idx for idx, q in enumerate(questions)}
    sorted_keys = sorted(translations_map.keys(), key=lambda k: order_index.get(k, 999999))
    ordered = {k: translations_map[k] for k in sorted_keys}
    tmp_path = f"{path}.tmp"
    with open(tmp_path, "w", encoding="utf-8") as f:
        json.dump(ordered, f, ensure_ascii=False, indent=2)
        f.write("\n")
    os.replace(tmp_path, path)


def run_translation(target_questions, all_questions, output_path, batch_size=5, workers=5):
    existing = {}
    if os.path.isfile(output_path):
        with open(output_path, "r", encoding="utf-8") as f:
            existing = json.load(f)

    todo = []
    for q in target_questions:
        k = question_key(q)
        if k in existing and not validate_translation_entry(q, existing[k]):
            continue
        todo.append(q)

    print(f"Target questions: {len(target_questions)} | Already valid: {len(target_questions) - len(todo)} | Remaining: {len(todo)}")
    if not todo:
        return existing

    llm_cfg = load_dotenv_fallback()
    if not llm_cfg["LLM_API_KEY"]:
        raise RuntimeError("LLM_API_KEY is required to run translation generation.")

    batches = [todo[i : i + batch_size] for i in range(0, len(todo), batch_size)]
    completed_batches = 0
    with ThreadPoolExecutor(max_workers=workers) as pool:
        futures = {pool.submit(translate_and_validate_group, b, llm_cfg): b for b in batches}
        for fut in as_completed(futures):
            res = fut.result()
            existing.update(res)
            completed_batches += 1
            save_translations_sorted(output_path, existing, all_questions)
            print(f"[{completed_batches}/{len(batches)}] Saved batch (+{len(res)} items; total valid={len(existing)})", flush=True)

    return existing


def main():
    parser = argparse.ArgumentParser(description="Translate and validate CompTIA Security+ questions in Hybrid Bilingual Thai.")
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--pilot", action="store_true", help="Translate 10 representative questions (2 per domain) and save to translations file.")
    mode.add_argument("--all", action="store_true", help="Resumable batch translation for all 540 questions.")
    mode.add_argument("--validate-only", action="store_true", help="Zero-network deterministic validation check.")
    parser.add_argument("--allow-partial", action="store_true", help="With --validate-only, allow partial pilot dataset (>=2/domain) instead of full 540/540.")
    parser.add_argument("--parsed", default=DEFAULT_PARSED_PATH, help="Path to parsed_questions.json")
    parser.add_argument("--output", default=DEFAULT_TRANSLATIONS_PATH, help="Path to question_translations_th.json")
    parser.add_argument("--batch-size", type=int, default=5)
    parser.add_argument("--workers", type=int, default=6)
    args = parser.parse_args()

    with open(args.parsed, "r", encoding="utf-8") as f:
        questions = json.load(f)

    if args.validate_only:
        if not os.path.isfile(args.output):
            print(f"Error: Translations file not found at {args.output}", file=sys.stderr)
            sys.exit(1)
        with open(args.output, "r", encoding="utf-8") as f:
            translations = json.load(f)
        errs, per_domain = validate_translations_dataset(
            questions, translations, require_full=not args.allow_partial
        )
        if errs:
            print("Question translation validation FAILED:", file=sys.stderr)
            for err in errs[:25]:
                print(f" - {err}", file=sys.stderr)
            if len(errs) > 25:
                print(f" ... and {len(errs) - 25} more errors", file=sys.stderr)
            sys.exit(1)
        print(f"Question translation validation PASSED ({sum(per_domain.values())}/{len(questions)} validated): {per_domain}")
        return

    targets = select_pilot_questions(questions, per_domain=2) if args.pilot else questions
    result_map = run_translation(
        targets, questions, args.output, batch_size=args.batch_size, workers=args.workers
    )
    if args.pilot:
        print("\n=== PILOT 10 TRANSLATION PREVIEW ===")
        for q in targets:
            k = question_key(q)
            th = result_map[k]["th"]
            print(f"\n[{k} | Sub-topic {q.get('sub_topic_id')} | Ans={q['correct_answer']}]")
            print(f"EN: {q['question_text']}")
            print(f"TH: {th['question_text']}")
            print(f"Options TH: {json.dumps(th['options'], ensure_ascii=False)}")
            print(f"Exp TH: {th['explanation']}")


if __name__ == "__main__":
    main()
