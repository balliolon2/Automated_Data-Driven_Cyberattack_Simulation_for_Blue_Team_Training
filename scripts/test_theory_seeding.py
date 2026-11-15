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
            self.assertIn("REFERENCES.md", mod["source_url"])

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


if __name__ == "__main__":
    unittest.main()
