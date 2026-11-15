import json
import re


# Weighted term rules per CompTIA Security+ (SY0-701) objective (1.1-5.6).
# Stronger, more specific multi-word signals carry higher weights so multi-topic
# scenarios classify into their primary tested objective cleanly.
SUBTOPIC_RULES = {
    "domain1": {
        "1.1": [
            (r"\bpreventive\b|\bdeterrent\b|\bdetective\b|\bcorrective\b|\bcompensating\b|\bdirective\b", 5),
            (r"technical control|managerial control|operational control|physical control|security control", 4),
            (r"bollard|mantrap|access control vestibule|warning banner|guard|cctv|badge|lock", 3),
        ],
        "1.2": [
            (r"non-?repudiation|cia triad|confidentiality|integrity|availability", 5),
            (r"zero trust|control plane|data plane|policy enforcement|policy decision|adaptive identity", 5),
            (r"honeypot|honeynet|honeyfile|deception|air-?gap", 4),
            (r"authentication, authorization, and accounting|\baaa\b|defense in depth|least privilege", 3),
        ],
        "1.3": [
            (r"change management|change control|change advisory|\bcab\b|change request", 6),
            (r"rollback|backout plan|maintenance window|impact analysis|configuration drift", 5),
            (r"requires a restart|service restart|routine update|scheduled maintenance|upgrade its|set for an os upgrade|standard operating procedure", 4),
        ],
        "1.4": [
            (r"\bpki\b|public key infrastructure|certificate authority|\bcrl\b|\bocsp\b|x\.509|digital certificate|key escrow", 6),
            (r"hardware security module|\bhsm\b|trusted platform module|\btpm\b|secure enclave|kms|key management", 5),
            (r"symmetric|asymmetric|\baes\b|\brsa\b|\becc\b|diffie-?hellman|elliptic curve", 5),
            (r"cryptograph|digital signature|\bhash|\bsha-?|\bmd5\b|salting|salt|key stretching|pbkdf2|bcrypt|steganograph|tokeniz|obfuscat|encrypt", 4),
        ],
    },
    "domain2": {
        "2.1": [
            (r"threat actor|nation-?state|\bapt\b|hacktivist|script kiddie|unskilled attacker|organized crime|shadow it", 6),
            (r"insider threat|competitor|motivation|espionage|philosophical|revenge", 4),
            (r"threat intelligence|\bosint\b|\bstix\b|\btaxii\b|\bisac\b|threat feed", 5),
        ],
        "2.2": [
            (r"\bphishing\b|\bspear phishing\b|\bwhaling\b|\bvishing\b|\bsmishing\b|business email compromise|\bbec\b", 6),
            (r"social engineering|pretexting|typosquatting|watering hole|impersonat|elicitation", 5),
            (r"attack surface|threat vector|removable media|\busb\b|supply chain vector|unsecured service|default credential", 4),
        ],
        "2.3": [
            (r"sql injection|\bxss\b|cross-site scripting|buffer overflow|integer overflow|race condition|toctou", 6),
            (r"insecure deserialization|\bcsrf\b|\bssrf\b|directory traversal|command injection|improper input", 6),
            (r"zero-?day|end-of-life|\beol\b|vm escape|firmware|hardware vulnerab|memory leak|outdated software|misconfigur", 4),
        ],
        "2.4": [
            (r"ransomware|trojan|worm|rootkit|keylogger|spyware|botnet|backdoor|logic bomb|remote access trojan", 5),
            (r"arp poisoning|dns poisoning|on-path|man-in-the-middle|mitm|replay attack|pass-the-hash|downgrade|dns tunneling", 5),
            (r"ddos|denial of service|brute force|dictionary attack|password spraying|credential stuffing|rainbow table", 4),
            (r"indicator of compromise|\bioc\b|impossible travel|anomalous|locked-?out account|missing logs|unauthorized scheduled", 4),
        ],
        "2.5": [
            (r"mitigation technique|network segmentation|application allowlist|whitelisting|blocklisting|blacklisting", 5),
            (r"hardening|disable unnecessary|patch management|remediat|quarantine|sandboxing|isolat|decommission", 4),
            (r"host-based firewall|endpoint protection|antivirus|least privilege|configuration baseline|acl", 3),
        ],
    },
    "domain3": {
        "3.1": [
            (r"\biaas\b|\bpaas\b|\bsaas\b|shared responsibility|public cloud|private cloud|hybrid cloud|multi-cloud", 5),
            (r"serverless|microservice|container|kubernetes|docker|infrastructure as code|\biac\b|software-defined|\bsdn\b", 5),
            (r"\biot\b|\bscada\b|\bics\b|\brtos\b|embedded system|operational technology|centralized|decentralized|virtualization|hypervisor", 4),
        ],
        "3.2": [
            (r"screened subnet|\bdmz\b|microsegmentation|bastion|jump server|air gap", 5),
            (r"stateful firewall|stateless|web application firewall|\bwaf\b|\bids\b|\bips\b|forward proxy|reverse proxy", 5),
            (r"\bvpn\b|\bipsec\b|load balancer|network access control|\bnac\b|802\.1x|\bradius\b|\btacacs|sd-wan|\bsase\b|\bztna\b|port security", 4),
        ],
        "3.3": [
            (r"data at rest|data in transit|data in use|data classification|data sovereignty|geographical", 6),
            (r"data loss prevention|\bdlp\b|tokenization|data masking|redaction|anonymization|pseudonymization|rights management|\birm\b|\bdrm\b", 5),
            (r"\bpii\b|\bphi\b|intellectual property|data retention|sensitive data", 4),
        ],
        "3.4": [
            (r"hot site|warm site|cold site|geographic dispersal|site resilience|disaster recovery", 6),
            (r"full backup|incremental backup|differential backup|immutable backup|snapshot|replication|3-2-1", 6),
            (r"high availability|failover|clustering|load balancing|redundancy|\braid\b|\bups\b|generator|multipath|\brpo\b|\brto\b|\bmtbf\b|\bmttr\b", 4),
        ],
    },
    "domain4": {
        "4.1": [
            (r"secure boot|\buefi\b|\bbios\b|selinux|apparmor|group policy|\bgpo\b|registry", 5),
            (r"\bmdm\b|mobile device management|\bmam\b|\bbyod\b|\bcope\b|\bcyod\b|containerization", 6),
            (r"\bwpa3\b|\bwpa2\b|\beap\b|captive portal|wireless security|site survey|endpoint baseline|server hardening|sandbox", 4),
        ],
        "4.2": [
            (r"media sanitization|sanitiz|degauss|shredding|pulping|incinerat|purging|clearing|cryptographic erase|data remanence", 6),
            (r"asset management|asset inventory|\bcmdb\b|asset tracking|asset ownership|procurement|acquisition|decommission|disposal|lifecycle", 5),
        ],
        "4.3": [
            (r"credentialed scan|non-credentialed scan|vulnerability scan|nessus|openvas|\bcvss\b|\bcve\b|bug bounty", 6),
            (r"false positive|false negative|prioritiz|remediat|validation|rescan|static analysis|\bsast\b|\bdast\b|\bscap\b", 4),
        ],
        "4.4": [
            (r"\bsiem\b|syslog|netflow|ipfix|\bsnmp|packet capture|\bpcap\b|tcpdump|wireshark", 6),
            (r"file integrity monitoring|\bfim\b|\bueba\b|alert tuning|correlation|log aggregation|ntp|collector|baseline monitoring", 5),
        ],
        "4.5": [
            (r"\bspf\b|\bdkim\b|\bdmarc\b|email security gateway|dns filtering|sinkholing|url filtering|content filter", 6),
            (r"firewall rule|access control list|\bacl\b|snmpv3|ssh|tls upgrade|ldaps|quarantine|attack surface reduction|port closure", 4),
        ],
        "4.6": [
            (r"identity and access management|\biam\b|provisioning|deprovisioning|privileged access|\bpam\b|just-in-time|\bjit\b", 6),
            (r"\bsaml\b|\boauth\b|\boidc\b|openid connect|single sign-on|\bsso\b|federation|kerberos|ldap|active directory", 5),
            (r"\brbac\b|\babac\b|\bdac\b|\bmac\b|multi-factor|\bmfa\b|fido2|biometric|totp|passwordless|token", 4),
        ],
        "4.7": [
            (r"\bsoar\b|orchestration|automated playbook|automation workflow|automation system|api integration|automated containment|continuous integration|ci/cd", 7),
            (r"integrating automation|automatically\b|manually deploying|automated remediation|guardrail|ticketing automation|scripted response|reduce human error", 6),
        ],
        "4.8": [
            (r"incident response|csirt|lessons learned|post-incident|tabletop exercise|attack simulation", 6),
            (r"containment|eradication|recovery phase|detection and analysis|preparation phase|root cause analysis|breach notification", 5),
        ],
        "4.9": [
            (r"digital forensic|chain of custody|legal hold|e-discovery|order of volatility|volatile memory|ram dump|memory capture", 6),
            (r"write blocker|bit-stream|disk image|integrity hash|forensic acquisition|evidence preservation|time offset", 5),
        ],
    },
    "domain5": {
        "5.1": [
            (r"acceptable use policy|\baup\b|security governance|steering committee|board of directors", 6),
            (r"data owner|data controller|data processor|data custodian|data steward|policy vs\.? standard|guideline|procedure", 6),
            (r"governance structure|centralized|decentralized|data sovereignty|onboarding policy|offboarding policy|sdlc policy", 4),
        ],
        "5.2": [
            (r"\bsle\b|\bale\b|\baro\b|exposure factor|single loss expectancy|annualized loss expectancy|asset value", 6),
            (r"risk register|\bkri\b|risk appetite|risk tolerance|risk threshold|risk acceptance|risk transfer|risk avoidance|risk mitigation", 6),
            (r"business impact analysis|\bbia\b|qualitative|quantitative|\brto\b|\brpo\b|\bmtbf\b|\bmttr\b|risk assessment", 4),
        ],
        "5.3": [
            (r"third-party|vendor assessment|supply chain risk|right-to-audit|right to audit|due diligence|conflict of interest", 6),
            (r"\bnda\b|\bsla\b|\bmou\b|\bmoa\b|\bmsa\b|\bbpa\b|\bsow\b|vendor questionnaire|independent assessment|vendor monitor", 6),
        ],
        "5.4": [
            (r"regulatory compliance|non-compliance|fines|sanctions|loss of license|attestation|acknowledgment", 6),
            (r"\bgdpr\b|\bhipaa\b|\bpci-?dss\b|\bsox\b|\bferpa\b|data subject|right to be forgotten|due care|due diligence|privacy compliance", 5),
        ],
        "5.5": [
            (r"internal audit|external audit|third-party audit|independent audit|self-assessment|soc 2|iso 27001 audit", 6),
            (r"penetration test|known environment|unknown environment|partially known|black box|white box|gray box|passive reconnaissance|active reconnaissance", 6),
        ],
        "5.6": [
            (r"security awareness|awareness training|simulated phishing|phishing simulation|phishing campaign|annual training", 6),
            (r"role-based training|security culture|user guidance|social engineering awareness|insider threat training|report suspicious", 5),
        ],
    },
}

# Explicit manual overrides by source_question_no where verified review shows a
# question specifically tests an edge objective (e.g., 5.6 awareness scenarios).
MANUAL_SUBTOPIC_OVERRIDES = {
    422: "4.7",  # Automating manual deployment & configuration tasks
    435: "4.7",  # Deploying enterprise automation system with guardrails
    442: "4.7",  # Integrating automation into healthcare operations
    458: "4.7",  # Automated cloud inventory/provisioning
    490: "5.6",  # Onboarding user awareness/AUP acknowledgment
    518: "5.6",  # Social engineering awareness / vishing reporting
    536: "5.6",  # Simulated phishing annual security training
    539: "5.6",  # Employee recognizing phishing email and reporting
}


def classify_question_subtopic(question):
    q_no = question.get("source_question_no")
    domain_id = question["domain_id"]

    if q_no in MANUAL_SUBTOPIC_OVERRIDES:
        override = MANUAL_SUBTOPIC_OVERRIDES[q_no]
        if override[0] == domain_id[-1]:
            return override

    rules = SUBTOPIC_RULES[domain_id]
    text = (
        f"{question.get('question_text', '')} "
        f"{' '.join((question.get('options') or {}).values())} "
        f"{question.get('explanation', '')}"
    ).lower()

    best_sub = sorted(rules.keys())[0]
    best_score = -1
    for sub_id in sorted(rules.keys()):
        score = 0
        for pattern, weight in rules[sub_id]:
            matches = len(re.findall(pattern, text))
            if matches:
                score += weight * min(matches, 3)
        if score > best_score:
            best_score = score
            best_sub = sub_id

    return best_sub


def classify_questions_file(path="parsed_questions.json"):
    with open(path, "r", encoding="utf-8") as f:
        questions = json.load(f)

    for q in questions:
        q["sub_topic_id"] = classify_question_subtopic(q)

    with open(path, "w", encoding="utf-8") as f:
        json.dump(questions, f, ensure_ascii=False, indent=2)

    return questions


if __name__ == "__main__":
    import collections

    qs = classify_questions_file()
    counts = collections.Counter(q["sub_topic_id"] for q in qs)
    for k in sorted(counts):
        print(f"{k}: {counts[k]}")
