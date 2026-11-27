import json
import os
import sys
import unittest

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
if SCRIPT_DIR not in sys.path:
    sys.path.insert(0, SCRIPT_DIR)

import translate_questions_th  # noqa: E402


def make_sample_question():
    return {
        "domain_id": "domain5",
        "sub_topic_id": "5.2",
        "source_question_no": 506,
        "type": "multiple_choice",
        "question_text": "What is the expected Annualized Loss Expectancy (ALE) when ARO = 2 and SLE = $50,000?",
        "options": {"A": "$10,000", "B": "$100,000", "C": "$25,000", "D": "$1,000,000"},
        "correct_answer": "B",
        "explanation": (
            "Correct Answer: B. $100,000. ALE equals ARO multiplied by SLE ($100,000). "
            "Option A is incorrect. Incorrect arithmetic. "
            "Option C is incorrect. Half of SLE. "
            "Option D is incorrect. Overstated."
        ),
    }


def make_valid_entry(q):
    return {
        "domain_id": q["domain_id"],
        "source_question_no": q["source_question_no"],
        "en_sha256": translate_questions_th.compute_question_sha256(q),
        "th": {
            "question_text": "ค่า Annualized Loss Expectancy (ALE) ที่คาดการณ์เมื่อ ARO = 2 และ SLE = $50,000 คือข้อใด?",
            "options": {"A": "$10,000", "B": "$100,000", "C": "$25,000", "D": "$1,000,000"},
            "explanation": (
                "คำตอบที่ถูกต้อง: B. $100,000 ค่า ALE คำนวณจาก ARO × SLE ได้ $100,000 "
                "ตัวเลือก A ไม่ถูกต้อง เพราะคำนวณไม่ตรงสูตร "
                "ตัวเลือก C ไม่ถูกต้อง เพราะเป็นเพียงครึ่งหนึ่งของ SLE "
                "ตัวเลือก D ไม่ถูกต้อง เพราะสูงเกินค่าที่คำนวณได้"
            ),
        },
    }


class QuestionTranslationValidationTests(unittest.TestCase):
    def test_valid_hybrid_bilingual_entry_passes(self):
        q = make_sample_question()
        entry = make_valid_entry(q)
        errs = translate_questions_th.validate_translation_entry(q, entry)
        self.assertEqual(errs, [])

    def test_rejects_hash_drift_missing_thai_option_keys_and_answer_letter_mismatch(self):
        q = make_sample_question()

        bad_hash = make_valid_entry(q)
        bad_hash["en_sha256"] = "0" * 64
        self.assertTrue(any("en_sha256" in e for e in translate_questions_th.validate_translation_entry(q, bad_hash)))

        no_thai = make_valid_entry(q)
        no_thai["th"]["question_text"] = "English only question?"
        self.assertTrue(any("Thai" in e for e in translate_questions_th.validate_translation_entry(q, no_thai)))

        bad_opts = make_valid_entry(q)
        del bad_opts["th"]["options"]["D"]
        self.assertTrue(any("options" in e.lower() for e in translate_questions_th.validate_translation_entry(q, bad_opts)))

        wrong_ans = make_valid_entry(q)
        wrong_ans["th"]["explanation"] = wrong_ans["th"]["explanation"].replace("คำตอบที่ถูกต้อง: B", "คำตอบที่ถูกต้อง: A")
        self.assertTrue(any("คำตอบที่ถูกต้อง: B" in e for e in translate_questions_th.validate_translation_entry(q, wrong_ans)))

        missing_distractor = make_valid_entry(q)
        missing_distractor["th"]["explanation"] = "คำตอบที่ถูกต้อง: B. เหตุผล ตัวเลือก A ไม่ถูกต้อง เพราะผิด"
        self.assertTrue(any("ตัวเลือก" in e for e in translate_questions_th.validate_translation_entry(q, missing_distractor)))

    def test_canonical_parsed_questions_zero_thai_and_full_translations_valid(self):
        parsed_path = os.path.join(SCRIPT_DIR, "..", "parsed_questions.json")
        with open(parsed_path, "r", encoding="utf-8") as f:
            questions = json.load(f)
        self.assertEqual(len(questions), 540)
        self.assertEqual(translate_questions_th.validate_canonical_english_pool(questions), [])

        translations_path = os.path.join(SCRIPT_DIR, "question_translations_th.json")
        self.assertTrue(os.path.isfile(translations_path), "question_translations_th.json must exist")
        with open(translations_path, "r", encoding="utf-8") as f:
            translations = json.load(f)

        errs, per_domain = translate_questions_th.validate_translations_dataset(
            questions, translations, require_full=True
        )
        self.assertEqual(errs, [], f"Validation errors: {errs}")
        self.assertEqual(sum(per_domain.values()), 540)
        expected_counts = {"domain1": 110, "domain2": 110, "domain3": 90, "domain4": 150, "domain5": 80}
        self.assertEqual(per_domain, expected_counts)


if __name__ == "__main__":
    unittest.main()
