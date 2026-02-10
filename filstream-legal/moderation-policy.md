# FilStream Content Moderation Policy

*Last updated: February 10, 2026*

FilStream is a decentralized video streaming platform built on Filecoin. We're committed to keeping the platform safe, legal, and useful for creators and viewers while respecting the decentralized nature of the network.

---

## 1. Content Guidelines

### Allowed Content
- Original video content you created or have rights to distribute
- Licensed or public domain content
- Fair use content (commentary, criticism, education, parody)
- Creative Commons and openly licensed works

### Prohibited Content
- **Copyright infringement** — Content you don't have rights to distribute
- **Child sexual abuse material (CSAM)** — Absolute zero tolerance (see Section 7)
- **Non-consensual intimate imagery** — Revenge porn or deepfake pornography
- **Illegal content** — Content that violates applicable law (e.g., terrorism recruitment, instructions for weapons of mass destruction)
- **Abuse and harassment** — Targeted harassment, doxxing, credible threats of violence
- **Fraud and scams** — Phishing, impersonation for fraud, deceptive schemes

---

## 2. Flagging Process

Any FilStream user can flag content they believe violates this policy.

### Flag Categories
| Category | Description |
|----------|-------------|
| **Copyright** | Content that infringes on your or someone else's copyright |
| **Illegal / CSAM** | Child exploitation material or other illegal content |
| **Abuse / Harassment** | Targeted harassment, threats, doxxing |

### How to Flag
1. Click the **Flag** button on any video or content page
2. Select the appropriate category
3. Provide a brief description (optional but helpful)
4. Submit — you'll receive a confirmation with a reference ID

Flags are confidential. The content creator will not see who flagged their content.

---

## 3. Review Process

1. **All flags are reviewed** by the FilStream moderation team
2. **Auto-escalation**: Content that receives **3 or more flags** from unique users is automatically escalated for priority review
3. **CSAM flags** are escalated immediately with zero delay (see Section 7)
4. **Review timeline**: Most flags are reviewed within 48 hours; escalated content within 24 hours
5. **Outcomes**:
   - **No action** — Content doesn't violate policy
   - **Content removed** — Content is added to the denylist and removed from the platform
   - **Account warning** — Creator receives a strike (see Section 6)
   - **Account terminated** — For severe or repeat violations

---

## 4. DMCA Notice-and-Takedown

FilStream complies with the Digital Millennium Copyright Act (17 U.S.C. § 512). If you believe your copyrighted work has been posted on FilStream without authorization, you can submit a DMCA takedown notice.

### How to Submit a DMCA Notice

Send a written notice to our designated DMCA agent (see [DMCA page](/dmca)) containing:

1. Your physical or electronic signature
2. Identification of the copyrighted work you claim is infringed
3. Identification of the infringing material on FilStream, with enough detail to locate it (URL or content hash)
4. Your contact information (name, address, phone, email)
5. A statement that you have a good faith belief the use is not authorized by the copyright owner, its agent, or the law
6. A statement, under penalty of perjury, that the information in your notice is accurate and that you are the copyright owner or authorized to act on their behalf

### What Happens After We Receive a Valid Notice

1. We promptly remove or disable access to the identified content
2. The content hash is added to the seeder denylist (see Section 8)
3. We notify the uploader that their content was removed and provide a copy of the notice (with your personal contact info redacted where possible)
4. The uploader may file a counter-notice (see Section 5)

---

## 5. DMCA Counter-Notice

If your content was removed and you believe it was removed in error or that you have authorization to use the material, you may submit a counter-notice.

### Required Elements

Your counter-notice must include:

1. Your physical or electronic signature
2. Identification of the material that was removed and where it appeared before removal
3. A statement under penalty of perjury that you have a good faith belief the material was removed by mistake or misidentification
4. Your name, address, and phone number
5. A statement that you consent to the jurisdiction of the federal court in your district (or any judicial district if outside the US) and that you will accept service of process from the person who filed the original notice

### Timeline

1. Upon receiving a valid counter-notice, we forward it to the original complainant
2. The complainant has **10 business days** to notify us they've filed a court action seeking to restrain the infringing activity
3. If we don't receive such notification within **10–14 business days**, we will restore the removed content
4. If a court action is filed, the content remains down pending resolution

---

## 6. Repeat Infringer Policy

FilStream enforces a **three-strike policy** for repeat violators:

| Strike | Consequence |
|--------|-------------|
| **1st strike** | Content removed + warning |
| **2nd strike** | Content removed + 7-day upload restriction |
| **3rd strike** | **Account permanently terminated** |

Strikes are issued for confirmed policy violations including valid DMCA takedowns. Strikes expire after **12 months** if no further violations occur.

Successful counter-notices or appeal reversals will remove the associated strike.

---

## 7. CSAM — Zero Tolerance

FilStream has **absolute zero tolerance** for child sexual abuse material.

- Any content identified as or reasonably suspected to be CSAM is **immediately removed**
- The account is **immediately and permanently terminated**
- All known information is reported to the **National Center for Missing & Exploited Children (NCMEC) CyberTipline** as required by federal law (18 U.S.C. § 2258A)
- Reports are also forwarded to relevant law enforcement agencies
- There is **no appeal** for CSAM violations

If you encounter CSAM on FilStream, flag it immediately or email our abuse team. Do not download, screenshot, or redistribute the material.

---

## 8. Seeder Compliance

FilStream operates on a decentralized network of seeders. All seeders participating in the FilStream network must comply with moderation decisions.

### Requirements

- Seeders must sync and honor the **content denylist** (distributed via Bloom filter)
- Denied content must stop being served **within 10 minutes** of a denylist update
- Seeders must run `MayContain()` checks before serving any content segment

### Non-Compliance

Seeders that fail to honor the denylist will be **delisted from the network** and will no longer receive routing or incentive rewards.

---

## 9. Appeal Process

If you believe a moderation decision was made in error (except CSAM — see Section 7), you can appeal.

### How to Appeal

1. Email the moderation team with your reference ID and explanation
2. Appeals are reviewed by a different moderator than the one who made the original decision
3. **Timeline**: Appeals are processed within 5 business days
4. **Outcomes**: The original decision may be upheld, modified, or reversed
5. If reversed, any associated strikes are removed and content is restored

You may only appeal each decision once.

---

## 10. Changes to This Policy

We may update this policy as the platform evolves. Significant changes will be announced on the platform. Continued use of FilStream after changes constitutes acceptance of the updated policy.

---

## Contact

For moderation inquiries, DMCA notices, or abuse reports, see our [DMCA page](/dmca) for contact details.
