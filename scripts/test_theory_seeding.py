import os
import sys
import tempfile
import unittest

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
if SCRIPT_DIR not in sys.path:
    sys.path.insert(0, SCRIPT_DIR)

import load_theory_to_db


class TheorySeedingValidationTests(unittest.TestCase):
    def test_repo_theory_modules_pass_validation(self):
        repo_theory_dir = os.path.join(SCRIPT_DIR, "..", "theory")
        modules = load_theory_to_db.load_and_validate_modules(repo_theory_dir)
        self.assertEqual(len(modules), 5)
        self.assertEqual(
            [m["domain_id"] for m in modules],
            ["domain1", "domain2", "domain3", "domain4", "domain5"],
        )
        for mod in modules:
            self.assertTrue(len(mod["content"]) > 500)
            self.assertTrue(mod["source_url"].startswith("https://"))
            self.assertIn("](https://", mod["content"])
            self.assertIn("https://en.wikipedia.org/", mod["content"])

    def test_rejects_missing_required_section(self):
        broken = """# Domain 1: Broken
## 1. Core Security+ Theory & Exam Concepts
Text only.
"""
        errors = load_theory_to_db.validate_theory_content("domain1", "Broken", broken)
        self.assertTrue(any("Section 2" in err for err in errors))
        self.assertTrue(any("Section 3" in err for err in errors))
        self.assertTrue(any("Section 4" in err for err in errors))

    def test_rejects_missing_file_in_empty_dir(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            with self.assertRaises(ValueError):
                load_theory_to_db.load_and_validate_modules(tmpdir)

    def test_repo_theory_subtopics_pass_validation(self):
        repo_theory_dir = os.path.join(SCRIPT_DIR, "..", "theory")
        subtopics = load_theory_to_db.load_and_validate_subtopics(repo_theory_dir)
        expected_ids = [
            "1.1", "1.2", "1.3", "1.4",
            "2.1", "2.2", "2.3", "2.4", "2.5",
            "3.1", "3.2", "3.3", "3.4",
            "4.1", "4.2", "4.3", "4.4", "4.5", "4.6", "4.7", "4.8", "4.9",
            "5.1", "5.2", "5.3", "5.4", "5.5", "5.6",
        ]
        self.assertEqual([s["sub_topic_id"] for s in subtopics], expected_ids)
        for sub in subtopics:
            self.assertTrue(sub["domain_id"].startswith("domain"))
            self.assertEqual(sub["domain_id"][-1], sub["sub_topic_id"][0])
            self.assertTrue(len(sub["content"]) > 300)
            self.assertTrue(sub["source_url"].startswith("https://"))
            self.assertIn("](https://", sub["content"])

    def test_subtopic_rejects_missing_required_sections_or_links(self):
        broken = """# 1.1 Broken Subtopic
## 1. Core Security+ Theory
Concepts only without practice or web references.
"""
        errors = load_theory_to_db.validate_subtopic_content("1.1", "Broken", broken)
        self.assertTrue(any("Section 2" in err for err in errors))
        self.assertTrue(any("Section 3" in err for err in errors))
        self.assertTrue(any("https://" in err for err in errors))

    def test_parsed_questions_classified_into_28_subtopics(self):
        import json

        parsed_path = os.path.join(SCRIPT_DIR, "..", "parsed_questions.json")
        with open(parsed_path, "r", encoding="utf-8") as f:
            questions = json.load(f)

        self.assertEqual(len(questions), 540)
        valid_ids = {s[0] for s in load_theory_to_db.SUBTOPIC_SPECS}
        domain_counts = {}
        subtopic_counts = {}
        for q in questions:
            sub_id = q.get("sub_topic_id")
            self.assertIn(sub_id, valid_ids, f"Question #{q.get('source_question_no')} missing/invalid sub_topic_id: {sub_id}")
            self.assertEqual(
                sub_id[0],
                q["domain_id"][-1],
                f"Question #{q.get('source_question_no')} domain_id={q['domain_id']} mismatched sub_topic_id={sub_id}",
            )
            domain_counts[q["domain_id"]] = domain_counts.get(q["domain_id"], 0) + 1
            subtopic_counts[sub_id] = subtopic_counts.get(sub_id, 0) + 1

        self.assertEqual(
            domain_counts,
            {"domain1": 110, "domain2": 110, "domain3": 90, "domain4": 150, "domain5": 80},
        )
        # Every one of the 28 SY0-701 sub-topics has at least 2 dedicated questions
        for sub_id in sorted(valid_ids):
            self.assertGreaterEqual(subtopic_counts.get(sub_id, 0), 2, f"Sub-topic {sub_id} has < 2 questions")

    def test_domain1_thai_pilot_translations_pass_validation(self):
        repo_theory_dir = os.path.join(SCRIPT_DIR, "..", "theory")
        th_modules, th_subtopics = load_theory_to_db.load_and_validate_thai_translations(
            repo_theory_dir, locale="th", require_all=False, require_pilot=True
        )
        self.assertGreaterEqual(len(th_modules), 1)
        self.assertIn("domain1", [m["entity_id"] for m in th_modules])
        self.assertGreaterEqual(len(th_subtopics), 4)
        sub_ids = {s["entity_id"] for s in th_subtopics}
        for pilot_id in ("1.1", "1.2", "1.3", "1.4"):
            self.assertIn(pilot_id, sub_ids)
        for item in th_modules + th_subtopics:
            self.assertEqual(item["locale"], "th")
            self.assertRegex(item["title"], r"[\u0E00-\u0E7F]")
            self.assertRegex(item["content"], r"[\u0E00-\u0E7F]")

    def test_thai_validator_rejects_untranslated_stub_mismatched_urls_and_thai_in_kql(self):
        en_sample = """# 1.1 Example
## 1. Core Security+ Theory
See [NIST](https://csrc.nist.gov/pubs/sp/800/53/r5/final).
## 2. SOC & Blue-Team Operational Practice
```kql
SigninLogs | where ResultType != "0"
```
## 3. References (Web Sources)
- [NIST](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
"""
        no_thai = en_sample
        errs_no_thai = load_theory_to_db.validate_thai_subtopic_content("1.1", "Example", en_sample, no_thai)
        self.assertTrue(any("Thai" in e for e in errs_no_thai))

        bad_url = """# 1.1 ตัวอย่าง
## 1. ทฤษฎีหลัก
ดู [Unregistered](https://evil.example.com/nope)
## 2. ปฏิบัติการ SOC
```kql
SigninLogs | where ResultType != "0"
```
## 3. แหล่งอ้างอิง
- [Unregistered](https://evil.example.com/nope)
"""
        errs_url = load_theory_to_db.validate_thai_subtopic_content("1.1", "Example", en_sample, bad_url)
        self.assertTrue(any("citation" in e.lower() or "https://" in e.lower() for e in errs_url))

        thai_in_kql = """# 1.1 ตัวอย่าง
## 1. ทฤษฎีหลัก
ดู [NIST](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
## 2. ปฏิบัติการ SOC
```kql
SigninLogs | where Message == "ภาษาไทย"
```
## 3. แหล่งอ้างอิง
- [NIST](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
"""
        errs_kql = load_theory_to_db.validate_thai_subtopic_content("1.1", "Example", en_sample, thai_in_kql)
        self.assertTrue(any("code block" in e.lower() for e in errs_kql))


if __name__ == "__main__":
    unittest.main()
