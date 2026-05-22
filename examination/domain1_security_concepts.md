**Question 1.** A client disputes having signed a digital contract.
The service provider needs to prove that the signature was
indeed from the client and hasn’t been tampered with. Which of
the following security concepts is the service provider relying
on?
**(A)** Authentication
**(B)** Confidentiality
**(C) Non-repudiation
(D)** Access Control

**Explanation 1. Correct Answer: C. Non-repudiation.** Non-
repudiation ensures that a party in a dispute cannot deny the
authenticity of their actions. In this scenario, it would provide
evidence that the client did sign the contract and that it hasn’t
been tampered with post-signature.

**Option A is incorrect.** Authentication confirms the identity of a
user or system. While it plays a part in ensuring that the right
person is accessing the system, it doesn’t directly provide
evidence about the actions post-authentication, like signing a
contract.

**Option B is incorrect.** Confidentiality ensures that information
is only accessible to those with the appropriate permissions. It
doesn’t provide evidence of an action being taken by a specific
entity.


**Option D is incorrect.** Access Control determines who or what
can view or use resources in a computing environment. It
doesn’t ensure the validity of actions taken within the system.

**Question 2.** Carlos, an IT consultant, advises a startup company
on cybersecurity best practices. The company plans to launch
several microsites under various subdomains. They want a
solution that is cost-effective but also ensures that the sites are
validated by a third-party. What type of certificate should Carlos
recommend?
**(A)** A separate self-signed certificate for each microsite
**(B)** An individual third-party certificate for each subdomain
**(C) A third-party wildcard certificate
(D)** An EV certificate issued by an internal CA

**Explanation 2. Correct Answer: C. A third-party wildcard
certificate.** A third-party wildcard certificate allows an
organization to secure multiple subdomains with a single
certificate. It’s cost-effective as the company doesn’t need to
purchase and manage separate certificates for each subdomain,
and because it’s issued by a third-party Certificate Authority, it
provides validation for external users.

**Option A is incorrect.** Self-signed certificates won’t provide
third-party validation, which could result in trust issues for
external users.

**Option B is incorrect.** While individual third-party certificates
for each subdomain will provide third-party validation, this
approach would not be as cost-effective as a wildcard
certificate.


**Option D is incorrect.** An EV certificate provides high
assurance, but one issued by an internal CA will not be
inherently trusted by external users.

**Question 3.** A company wants to ensure that security incidents
are detected and addressed as quickly as possible by on-duty
personnel. Which of the following operational security controls
would be BEST to implement for this purpose?
**(A)** Deploying a Network Intrusion Prevention System
(NIPS)
**(B) Establishing a 24/7 Security Operations Center
(SOC)
(C)** Creating a company-wide security policy
**(D)** Implementing end-to-end data encryption

**Explanation 3. Correct Answer: B. Establishing a 24/7
Security Operations Center (SOC).** A Security Operations
Center (SOC) is an operational control that provides real-time
monitoring, detection, and response to security incidents. With a
24/7 SOC, the company ensures that there is always personnel
available to handle security incidents as they occur.

**Option A is incorrect.** Deploying a Network Intrusion
Prevention System (NIPS) is a technical control. While it can
prevent unauthorized activities on the network, it does not
ensure that there is personnel available around the clock to
address incidents.

**Option C is incorrect.** Creating a company-wide security
policy is a managerial control. It sets the guidelines and


procedures for security but does not ensure continuous
monitoring and immediate response to incidents.

**Option D is incorrect.** Implementing end-to-end data
encryption is a technical control that ensures data
confidentiality. While it protects data, it does not ensure that
incidents are detected and addressed by on-duty personnel in
real-time.

**Question 4.** During a routine check, the IT department
discovered that several employees had left their computers on
and unattended during lunch break. Which operational security
control can help mitigate the risk associated with this behavior?
**(A)** Implementing biometric authentication
**(B)** Enforcing a strict password policy
**(C) Deploying an automatic screen lock after inactivity
(D)** Implementing a secure coding practice

**Explanation 4. Correct Answer: C. Deploying an automatic
screen lock after inactivity.** Deploying an automatic screen
lock after a certain period of inactivity is an operational control.
It ensures that unattended devices are protected from
unauthorized access, thereby mitigating risks associated with
employees leaving their computers on and unattended.

**Option A is incorrect.** Implementing biometric authentication
is a technical control. While it enhances security at the point of
access, it doesn’t ensure that active sessions on unattended
devices are secured against unauthorized access.


**Option B is incorrect.** Enforcing a strict password policy is a
managerial control that dictates the creation and use of strong
passwords. While it enhances access security, it doesn’t secure
active sessions on unattended devices.

**Option D is incorrect.** Implementing a secure coding practice
is a technical and sometimes managerial control. It ensures
software is written to prevent vulnerabilities but doesn’t directly
address the risk of unattended computers.

**Question 5.** An art gallery wants to deploy a security solution to
detect movement in an open courtyard that features several
sculptures. This space has varying temperature conditions,
which might cause false alarms in some motion detection
technologies. Which type of sensor would be MOST
appropriate to ensure consistent motion detection in such
conditions?
**(A)** Thermal imaging sensors
**(B)** Pressure-sensitive mats
**(C)** Ultrasonic detectors
**(D) Microwave motion detectors**

**Explanation 5. Correct Answer: D. Microwave motion
detectors.** Microwave motion detectors are suitable for open
areas and are less affected by temperature changes. They emit
microwave beams to create an invisible detection zone and can
consistently detect motion when an object interrupts this zone,
irrespective of the ambient temperature.

**Option A is incorrect.** Thermal imaging sensors detect
variations in heat. While they can be effective, the varying


temperature conditions in the courtyard may cause
inconsistencies in detection.

**Option B is incorrect.** Pressure-sensitive mats are designed to
detect weight or pressure changes when stepped on. They would
not be suitable for an open courtyard where movement needs to
be detected across a larger area.

**Option C is incorrect.** Ultrasonic detectors emit sound waves
to detect motion. However, they might also be affected by
external environmental factors and are not as suitable for open
courtyards as microwave motion detectors.

**Question 6.** A company’s primary security control for accessing
secure server rooms is a biometric fingerprint scanner.
However, the scanner occasionally malfunctions in high
humidity. The security team is considering an alternative
solution to grant access when the primary method fails. Which
of the following would be the MOST appropriate compensating
control?
**(A) Implementing a security token-based authentication
system
(B)** Employing security guards at the main entrance
**(C)** Installing security cameras inside the server room
**(D)** Conducting regular server room audits

**Explanation 6. Correct Answer: A. Implementing a security
token-based authentication system.** A security token-based
authentication system would act as an alternative method for
verifying the identity of individuals when the primary control


(biometric fingerprint scanner) fails. This serves as a direct
compensating control for access.

**Option B is incorrect.** While security guards at the main
entrance can provide an added layer of security, they aren’t a
direct compensating control for a malfunctioning biometric
system in a specific location like the server room.

**Option C is incorrect.** While security cameras provide
surveillance, they don’t act as an alternative method for granting
or denying access to the server room.

**Option D is incorrect.** Conducting regular server room audits
is a detective control. It won’t provide real-time access or
compensate for the malfunctioning fingerprint scanner.

**Question 7.** A financial institution wants to ensure that
customers are aware of the bank’s policies on information
sharing and how their personal data is used. Which of the
following security controls would BEST communicate this to
customers?
**(A)** Implementing end-to-end encryption for online
transactions
**(B) Publishing a privacy policy on the bank's website
(C)** Conducting annual cybersecurity awareness training for
employees
**(D)** Using multi-factor authentication for online banking

**Explanation 7. Correct Answer: B. Publishing a privacy
policy on the bank’s website.** A privacy policy serves as a
directive control as it informs customers about the bank’s


practices regarding the collection, use, and sharing of their
personal data. By reading the policy, customers understand their
rights and the bank’s responsibilities.

**Option A is incorrect.** While end-to-end encryption ensures the
confidentiality of online transactions, it doesn’t inform
customers about the bank’s policies on information sharing or
how their data is used.

**Option C is incorrect.** Annual cybersecurity awareness training
is aimed at employees, not customers. It wouldn’t directly
communicate the bank’s information-sharing policies to its
customers.

**Option D is incorrect.** Using multi-factor authentication
improves the security of online banking by requiring multiple
forms of verification. However, it doesn’t communicate to
customers how their personal data is used or the bank’s
information-sharing policies.

**Question 8.** A large financial organization wants to ensure that
all employees understand the importance of cybersecurity and
the role they play in safeguarding company assets. Which of the
following managerial security controls will be MOST effective
in achieving this?
**(A)** Installing a firewall at the network perimeter
**(B) Regular security awareness training for employees
(C)** Deploying an Intrusion Detection System (IDS)
**(D)** Encrypting all company data


**Explanation 8. Correct Answer: B. Regular security
awareness training for employees.** Security awareness training
is a managerial control aiming to educate employees about
security risks and the necessary precautions they need to take.
By regularly training employees, the organization ensures that
all staff are aware of potential threats and their roles in
cybersecurity.

**Option A is incorrect.** Installing a firewall is a technical control
focused on preventing unauthorized access to or from a private
network. While it protects the network, it doesn’t directly
educate employees about their roles in cybersecurity.

**Option C is incorrect.** Deploying an Intrusion Detection
System (IDS) is a technical control. It monitors network traffic
for suspicious activities but does not directly focus on educating
employees.

**Option D is incorrect.** Encrypting company data is a technical
control. While it ensures the confidentiality of data, it doesn’t
address the employees’ knowledge or awareness regarding
cybersecurity.

**Question 9.** A company has faced multiple instances of
unauthorized individuals gaining access to their office premises.
Which of the following preventive security controls would be
MOST effective in preventing unauthorized physical access?
**(A)** Implementing a log monitoring solution for network
traffic
**(B) Installing video surveillance cameras at all entry and
exit points**


**(C)** Conducting regular security awareness training for
employees
**(D)** Implementing a multi-factor authentication system for
network access.

**Explanation 9. Correct Answer: B. Installing video
surveillance cameras at all entry and exit points.** Installing
video surveillance cameras at all entry and exit points acts as a
preventive control by deterring unauthorized individuals from
attempting to gain access, given the increased risk of detection
and recording.

**Option A is incorrect.** Implementing a log monitoring solution
is a detective control that provides insights into network
activities but doesn’t prevent unauthorized physical access.

**Option C is incorrect.** Conducting regular security awareness
training is a preventive measure, but its main focus is on
making employees aware of security risks and best practices,
not directly preventing unauthorized physical access.

**Option D is incorrect.** Implementing a multi-factor
authentication system is a preventive control for unauthorized
digital access but doesn’t address the prevention of
unauthorized physical access.

**Question 10.** TechVault, a company specializing in secure
storage solutions, recently had an unauthorized intrusion where
a burglar managed to bypass their motion sensors. In a bid to
prevent future breaches, they are considering deploying a
system that can detect weight changes in a restricted floor area


to alert any unauthorized access. Which of the following would
be BEST for this requirement?
**(A)** Ultrasonic motion detectors
**(B) Pressure-sensitive floor mats
(C)** CCTV cameras with facial recognition
**(D)** Glass break sensors

**Explanation 10. Correct Answer: B. Pressure-sensitive floor
mats.** Pressure-sensitive floor mats are designed to detect
weight changes or pressure when stepped on. This makes them
an effective solution for monitoring restricted areas and alerting
unauthorized access based on weight detection.

**Option A is incorrect.** Ultrasonic motion detectors use sound
waves to detect motion in an area but do not measure weight or
pressure.

**Option C is incorrect.** CCTV cameras with facial recognition
provide visual surveillance and can identify individuals, but
they don’t detect weight changes on the floor.

**Option D is incorrect.** Glass break sensors detect the sound of
breaking glass and are primarily used for windows and glass
doors, not for detecting pressure or weight changes on a floor.

**Question 11.** A system administrator is setting up an
authentication system for a new web application. Which of the
following security controls falls under the technical category
and ensures that users prove their identity before gaining
access?
**(A)** Implementing a security awareness training program


**(B)** Conducting a background check for new employees
**(C) Using multi-factor authentication
(D)** Establishing a clean desk policy

**Explanation 11. Correct Answer: C. Using multi-factor
authentication.** Multi-factor authentication is a technical
control that requires users to present two or more pieces of
evidence (factors) before gaining access. It provides an
additional layer of security to ensure that users are who they say
they are.

**Option A is incorrect.** Implementing a security awareness
training program is an administrative control, as it involves
educating employees on security best practices rather than using
technical measures to enforce them.

**Option B is incorrect.** Conducting a background check is an
administrative control as it involves vetting potential employees
before they’re hired. This process doesn’t directly enforce
technical measures on systems or networks.

**Option D is incorrect.** Establishing a clean desk policy is an
administrative control. It sets a guideline for employees to keep
their workspaces tidy and free of sensitive information, rather
than enforcing technical measures.

**Question 12.** An e-commerce company has experienced a
Distributed Denial of Service (DDoS) attack, which caused its
website to become inaccessible for several hours. To mitigate
the impact of such attacks in the future, which of the following
would be the BEST corrective control to implement?


**(A)** Displaying a seal for third-party security certifications
on the website
**(B) Establishing a Web Application Firewall (WAF) with
DDoS protection
(C)** Conducting routine vulnerability assessments on the
website
**(D)** Implementing strong password policies for website
administrators

**Explanation 12. Correct Answer: B. Establishing a Web
Application Firewall (WAF) with DDoS protection.** A Web
Application Firewall (WAF) with DDoS protection can identify
and filter out malicious traffic associated with DDoS attacks. As
a corrective control, it can help in mitigating the impact and
restoring normal service during and after an attack.

**Option A is incorrect.** Displaying a seal for third-party security
certifications on the website acts as a deterrent by showing
visitors and potential attackers that the site adheres to security
standards. However, it does not mitigate or correct the effects of
a DDoS attack.

**Option C is incorrect.** Conducting routine vulnerability
assessments is a detective control that helps in identifying
weaknesses. While it’s essential for overall security, it doesn’t
directly correct or mitigate the effects of a DDoS attack.

**Option D is incorrect.** Implementing strong password policies
for website administrators is a preventive control. It ensures that
administrators’ accounts are secure, but it does not address or
correct the issues caused by a DDoS attack.


**Question 13.** GreenTech Industries has a manufacturing facility
located in a relatively secluded area. Recent incidents of theft
and trespassing have alarmed the management. Which of the
following would MOST effectively deter unauthorized
nighttime access to the perimeter of the facility?
**(A)** Installing infrared sensors
**(B) Using bright perimeter lighting
(C)** Deploying additional security guards inside the facility
**(D)** Increasing the height of the facility walls

**Explanation 13. Correct Answer: B. Using bright perimeter
lighting.** Bright perimeter lighting acts as a strong deterrent for
unauthorized individuals, as it reduces hiding spots, makes
surveillance cameras more effective, and can make it easier for
security personnel to spot potential threats. In secluded areas,
proper lighting is particularly essential to illuminate dark spots
and deter potential intruders.

**Option A is incorrect.** While infrared sensors can detect
movement, they do not act as a visible deterrent in the same
way bright lighting does.

**Option C is incorrect.** Deploying additional security guards
inside the facility does not address the immediate concern of
unauthorized nighttime access to the perimeter.

**Option D is incorrect.** Increasing the height of the walls can
act as a deterrent, but it doesn’t illuminate or expose potential
intruders like bright lighting does.


**Question 14.** While conducting a routine security review, Jake,
a security specialist, discovers an unexpected piece of data
placed in the organization’s financial system. Upon asking, he
learns that this piece of data is intentionally placed and
monitored to see if any unauthorized user or system interacts
with it. What is this deceptive piece of data known as?
**(A)** Honeystring
**(B) Honeytoken
(C)** Canary token
**(D)** Security marker

**Explanation 14. Correct Answer: B. Honeytoken.**
Honeytokens are strategically placed deceptive pieces of data
that have no actual value or real-world use but are closely
monitored. Their sole purpose is to detect unauthorized
interactions, as any access or use of a honeytoken is likely
malicious or unauthorized.

**Option A is incorrect.** There isn’t a commonly recognized
security term known as “Honeystring” in the context described.

**Option C is incorrect.** Canary tokens are a specific type of
honeytoken and can serve the same purpose. However, given
the choices provided and the context of the question,
“Honeytoken” is the most accurate answer.

**Option D is incorrect.** A security marker, in a general sense,
can be any mark or indicator used for security purposes, but it
isn’t specifically a deceptive piece of data placed to detect
unauthorized access.


**Question 15.** An organization is deploying new IoT devices in
its smart office. To ensure that only authorized devices can
connect to the corporate network, each device will be given a
unique key pair. Which of the following best describes the
system authentication approach the organization is using?
**(A)** Shared secret authentication
**(B) Public key infrastructure (PKI)
(C)** Token-based authentication
**(D)** Username and password authentication

**Explanation 15. Correct Answer: B. Public key
infrastructure (PKI).** Public key infrastructure (PKI) is a
combination of hardware, software, policies, and standards that
work together to provide a framework for secure
communications. One of the primary features of PKI is the use
of a pair of keys (public and private) to authenticate entities. In
the scenario, each IoT device is given a unique key pair,
indicating the use of PKI for system authentication.

**Option A is incorrect.** Shared secret authentication typically
involves two parties having a shared secret that they use to
authenticate one another. The scenario mentions a unique key
pair for each device, which doesn’t align with the concept of a
shared secret.

**Option C is incorrect.** Token-based authentication typically
involves using a hardware or software token that generates a
time-sensitive code. The scenario is describing the use of key
pairs, not tokens.


**Option D is incorrect.** Username and password authentication
is a method where entities provide a username and a secret
password to verify their identity. The scenario does not mention
the use of usernames or passwords.

**Question 16.** In the new branch of BankSecure, the
management has decided to install a security system at the main
entrance that forces visitors to go through two separate
authorization checks before entering the main premises. Which
physical security measure should they consider?
**(A)** Turnstiles
**(B)** Security Guards
**(C) Access Control Vestibule
(D)** Keycard Readers

**Explanation 16. Correct Answer: C. Access Control
Vestibule.** An access control vestibule, often referred to as a
mantrap, is a two-stage authentication system. It consists of two
doors: a person enters the first door, undergoes an authorization
check (like a badge reader or biometric scanner), and only after
being approved can they proceed to the second door, where they
undergo another authorization check before accessing the main
premises.

**Option A is incorrect.** Turnstiles control the flow of individuals
into a location and can prevent tailgating to some extent, but
they do not force a two-stage authorization check.

**Option B is incorrect.** While security guards can perform
authorization checks and control access, they alone do not


provide a two-stage authorization system like an access control
vestibule.

**Option D is incorrect.** Keycard readers are a form of access
control that checks the credentials of individuals, but on their
own, they don’t ensure two separate authorization checks.

**Question 17.** The IT department wants to monitor network
traffic in real time to detect any anomalies or malicious
activities. Which of the following security controls can
accomplish this?
**(A)** Security policy documentation
**(B) Intrusion Detection System (IDS)
(C)** Employee code of conduct
**(D)** Access Control Lists (ACL)

**Explanation 17. Correct Answer: B. Intrusion Detection
System (IDS).** An Intrusion Detection System (IDS) is a
technical control that monitors network traffic in real-time and
alerts administrators to any suspicious or malicious activities
based on predefined rules or heuristics.

**Option A is incorrect.** Security policy documentation is an
administrative control. It provides guidelines and procedures for
maintaining security but does not actively monitor network
traffic.

**Option C is incorrect.** Employee code of conduct is an
administrative control. It provides guidelines on how employees
should behave in a professional setting but does not actively
monitor network traffic.


**Option D is incorrect.** Access Control Lists (ACL) are
technical controls, but they are used to define permissions on
who can access specific resources. They do not actively monitor
network traffic in real-time for anomalies or malicious
activities.

**Question 18.** Jenna, a web administrator for a growing online
retail business, is in the process of obtaining SSL certificates for
the company’s domain. The company uses several subdomains
for different services, such as shop.example.com,
blog.example.com, and support.example.com. Instead of
obtaining individual certificates for each subdomain, Jenna
wants to use one certificate. What type of certificate should
Jenna pursue?
**(A)** Extended Validation Certificate
**(B) Wildcard Certificate
(C)** Certificate with Subject Alternative Names (SAN)
**(D)** Code Signing Certificate

**Explanation 18. Correct Answer: B. Wildcard Certificate.** A
Wildcard Certificate is designed to secure a domain and its
subdomains under the same top-level domain. For example, a
wildcard certificate for *.example.com would secure
shop.example.com, blog.example.com, and any other
subdomain of example.com.

**Option A is incorrect.** An Extended Validation Certificate
provides the highest level of validation but does not necessarily
cover multiple subdomains by default.


**Option C is incorrect.** While a Certificate with Subject
Alternative Names (SAN) can secure multiple domains and
subdomains, it is not specifically tailored for all subdomains
under a single domain as the Wildcard Certificate is.

**Option D is incorrect.** A Code Signing Certificate is used to
sign software code, ensuring its integrity and authenticity, not
for securing domains or subdomains.

**Question 19.** At a newly established museum, management
wants to install sensors in the exhibit rooms to detect any
unauthorized movement after hours. The rooms are often filled
with a mix of air conditioning and external noise from the city.
Which sensor would be BEST suited to detect movement in
such conditions without being affected by the noise?
**(A)** Acoustic sensors
**(B)** Glass break detectors
**(C) Ultrasonic sensors
(D)** Thermal imaging cameras

**Explanation 19. Correct Answer: C. Ultrasonic sensors.**
Ultrasonic sensors emit high-frequency sound waves to detect
motion. These sound waves are beyond the range of human
hearing and won’t be affected by ambient noise, making them
ideal for environments with varying noise conditions. When
motion is detected, as indicated by changes in the reflected
waves, an alarm is triggered.

**Option A is incorrect.** Acoustic sensors detect specific sounds.
The external noise from the city might cause false alarms or
interfere with their detection capabilities.


**Option B is incorrect.** Glass break detectors are designed to
detect the sound or vibration of breaking glass. They aren’t
designed primarily to detect movement.

**Option D is incorrect.** Thermal imaging cameras detect heat
signatures and would be more susceptible to variations in room
temperature due to air conditioning, potentially leading to false
detections.

**Question 20.** A company is setting up a secure communication
channel between its headquarters and a remote branch office. To
ensure that data transmitted over this channel originates from a
legitimate system at the branch office, the company is
considering using digital certificates. Which authentication
method for systems is the company contemplating?
**(A)** Kerberos authentication
**(B)** Password-based authentication
**(C) Certificate-based authentication
(D)** Biometric-based authentication

**Explanation 20. Correct Answer: C. Certificate-based
authentication.** Certificate-based authentication uses digital
certificates to verify the identity of systems or individuals. In
the given scenario, the company wants to verify that data
transmitted over the communication channel originates from a
legitimate system, making digital certificates an appropriate
choice.

**Option A is incorrect.** Kerberos authentication is a ticket-based
authentication protocol primarily used to authenticate users in a


network, not specifically for system-to-system authentication
using digital certificates.

**Option B is incorrect.** Password-based authentication requires
systems or users to provide a secret password to prove their
identity. It doesn’t involve the use of digital certificates.

**Option D is incorrect.** Biometric-based authentication involves
using unique physical or behavioral attributes of a person for
verification, such as fingerprints or facial patterns. It is not
applicable to system-to-system authentication.

**Question 21.** A financial institution has experienced an uptick
in unauthorized transactions. They want to implement a control
that will allow them to identify suspicious transactions in real-
time. Which of the following would be the BEST detective
control for this scenario?
**(A)** Implementing a multi-factor authentication system for
all users
**(B)** Establishing a Security Operations Center (SOC) to
monitor network traffic
**(C) Installing an Intrusion Detection System (IDS) on
their network
(D)** Restricting transaction capabilities to only a few trusted
IP addresses.

**Explanation 21. Correct Answer: C. Installing an Intrusion
Detection System (IDS) on their network.** An Intrusion
Detection System (IDS) serves as a detective control by
monitoring network traffic for suspicious activities and potential
threats. In this context, it can be configured to detect patterns


related to unauthorized transactions, thereby allowing timely
intervention.

**Option A is incorrect.** Implementing a multi-factor
authentication system is a preventive control that provides an
additional layer of security by requiring two or more
verification methods. While it reduces the risk of unauthorized
access, it does not detect suspicious transactions.

**Option B is incorrect.** Establishing a Security Operations
Center (SOC) is a broad approach to handle security events, and
while it can include detective controls, merely setting up a SOC
does not provide specific real-time detection of unauthorized
transactions.

**Option D is incorrect.** Restricting transaction capabilities to
only a few trusted IP addresses is a preventive control that
limits the sources of potential transactions. While it can reduce
the number of unauthorized transactions, it does not detect
them.

**Question 22.** TechHaus has recently experienced multiple
security breaches where unauthorized personnel have managed
to infiltrate their server rooms after hours. To enhance security
measures, the company decided to deploy a new system. Which
of the following options would BEST detect human intruders
based on their body heat even in complete darkness?
**(A)** Installing CCTV cameras with LED lights
**(B)** Using ultrasonic motion sensors
**(C) Deploying infrared (IR) sensors
(D)** Implementing RFID badge readers at the entrance


**Explanation 22. Correct Answer: C. Deploying infrared (IR)
sensors.** Infrared (IR) sensors detect infrared radiation, such as
the heat emitted by the human body. This makes them
particularly effective in detecting human intruders, even in
complete darkness, based on the body heat they emit.

**Option A is incorrect.** While CCTV cameras with LED lights
can provide visual surveillance, they rely on light to produce
images and may not detect intruders in complete darkness as
efficiently as infrared sensors.

**Option B is incorrect.** Ultrasonic motion sensors detect
movement through sound waves, not body heat, making them
less efficient in differentiating between a human intruder and
other moving objects.

**Option D is incorrect.** RFID badge readers control access at
entry points but do not detect human intruders based on their
body heat inside a facility.

**Question 23.** After detecting an unauthorized intrusion into
their network, a financial institution wants to implement a
control that will restore compromised systems to a known good
state. Which of the following would be the MOST appropriate
corrective control?
**(A)** Implementing Intrusion Detection Systems (IDS) across
the network
**(B)** Frequently updating firewall rules
**(C) Restoring systems from verified backups
(D)** Enabling multi-factor authentication for users


**Explanation 23. Correct Answer: C. Deploying infrared (IR)
sensors.** Infrared (IR) sensors detect infrared radiation, such as
the heat emitted by the human body. This makes them
particularly effective in detecting human intruders, even in
complete darkness, based on the body heat they emit.

**Option A is incorrect.** While CCTV cameras with LED lights
can provide visual surveillance, they rely on light to produce
images and may not detect intruders in complete darkness as
efficiently as infrared sensors.

**Option B is incorrect.** Ultrasonic motion sensors detect
movement through sound waves, not body heat, making them
less efficient in differentiating between a human intruder and
other moving objects.

**Option D is incorrect.** RFID badge readers control access at
entry points but do not detect human intruders based on their
body heat inside a facility.

**Question 24.** After a recent security breach, Sarah, a
cybersecurity analyst, is implementing additional measures to
detect unauthorized activities. She decides to embed specific
values in the database that serve no real purpose but are
monitored for any unauthorized access or usage. These values
are designed to raise alerts if they are ever accessed or used.
What are these specific values commonly referred to as?
**(A)** Security flags
**(B)** Honeypots
**(C) Honeytokens
(D)** Audit trails


**Explanation 24. Correct Answer: C. Restoring systems from
verified backups.** Restoring systems from verified backups is a
corrective control, as it can restore compromised systems to
their last known good state. This action corrects the adverse
effects of the intrusion and ensures that any malicious
alterations are removed.

**Option A is incorrect.** Implementing Intrusion Detection
Systems (IDS) is a detective control. It monitors and detects
malicious activities in the network but doesn’t correct the
adverse impacts of an intrusion.

**Option B is incorrect.** Frequently updating firewall rules is a
preventive measure, aiming to block malicious traffic and
prevent potential intrusions. While vital, it doesn’t correct the
impacts of an already occurred breach.

**Option D is incorrect.** Enabling multi-factor authentication is a
preventive control, aiming to provide additional layers of
verification. While it enhances security, it doesn’t correct the
adverse impacts of an intrusion.

**Question 25.** Bob receives an email prompting him to verify his
identity by clicking on a link. The link directs him to a webpage
where he has to provide his username, password, and answer a
personal security question. What type of authentication method
is being employed here?
**(A)** Biometric authentication
**(B)** Token-based authentication
**(C) Two-factor authentication
(D)** Single sign-on


**Explanation 25. Correct Answer: C. Two-factor
authentication.** Two-factor authentication (2FA) is a security
process in which users provide two different authentication
factors to verify their identity. In this scenario, Bob is providing
something he knows (username and password) and also
answering a personal security question, which is another form
of “something he knows.”

**Option A is incorrect.** Biometric authentication involves using
unique physical or behavioral attributes of a person for
verification, such as fingerprints or facial patterns. The scenario
doesn’t mention any biometric data.

**Option B is incorrect.** Token-based authentication typically
involves using a hardware or software token that generates a
time-sensitive code. This was not described in the scenario.

**Option D is incorrect.** Single sign-on (SSO) allows a user to
log in once and gain access to multiple systems without being
prompted to log in again for each system. The scenario
describes a two-factor authentication process, not SSO.

**Question 26.** In an effort to minimize data breaches from
malware, a company is deciding on a control to prevent
malicious software from being executed on company devices.
Which of the following would be the BEST preventive control?
**(A)** Deploying a Network Intrusion Detection System
(NIDS)
**(B)** Regularly backing up critical data
**(C) Installing an antivirus software with real-time**


**scanning
(D)** Performing a forensic analysis after a security incident

**Explanation 26. Correct Answer: C. Two-factor
authentication.** Two-factor authentication (2FA) is a security
process in which users provide two different authentication
factors to verify their identity. In this scenario, Bob is providing
something he knows (username and password) and also
answering a personal security question, which is another form
of “something he knows.”

**Option A is incorrect.** Biometric authentication involves using
unique physical or behavioral attributes of a person for
verification, such as fingerprints or facial patterns. The scenario
doesn’t mention any biometric data.

**Option B is incorrect.** Token-based authentication typically
involves using a hardware or software token that generates a
time-sensitive code. This was not described in the scenario.

**Option D is incorrect.** Single sign-on (SSO) allows a user to
log in once and gain access to multiple systems without being
prompted to log in again for each system. The scenario
describes a two-factor authentication process, not SSO.

**Question 27.** After undergoing a major infrastructure upgrade,
GlobalMed Corp experienced several unanticipated security
issues. In retrospect, the IT manager realized they skipped an
essential step in their change management process which could
have predicted and mitigated these issues. What step did they
most likely overlook?


**(A)** Procurement of new hardware
**(B)** Training of IT staff on the new systems
**(C) Impact analysis
(D)** Integration with legacy systems

**Explanation 27. Correct Answer: C. Impact analysis.** An
impact analysis is vital in the change management process as it
evaluates the potential ramifications of a proposed change. By
conducting this analysis, organizations can anticipate potential
security challenges and mitigate them before implementing the
change.

**Option A is incorrect.** While procurement is essential, merely
purchasing new hardware wouldn’t directly help in predicting or
mitigating potential security issues stemming from an
infrastructure upgrade.

**Option B is incorrect.** Training IT staff is essential for effective
implementation and operation, but it doesn’t directly address
predicting and understanding potential security consequences of
the upgrade.

**Option D is incorrect.** Integration with legacy systems is a
crucial consideration, especially for compatibility. However, the
focus of the scenario is on predicting and understanding
potential security issues, which is primarily addressed through
an impact analysis.

**Question 28.** MegaCorp recently introduced a new web
application for its customers. Before its release, the software
underwent rigorous testing in a controlled environment. When


the application was deployed in production, several security
vulnerabilities were reported. Which of the following reasons
can explain the mismatch between the test results and actual
vulnerabilities?
**(A)** The testing environment was an exact replica of the
production environment
**(B) Test results were not thoroughly reviewed
(C)** The software was not tested for zero-day vulnerabilities
**(D)** Penetration testing was done post-production

**Explanation 28. Correct Answer: B. Test results were not
thoroughly reviewed.** Even if an application is tested
rigorously, it is crucial to thoroughly review and interpret the
test results to identify any potential security vulnerabilities.
Failing to review or misinterpreting these results can lead to
vulnerabilities going unnoticed and unresolved.

**Option A is incorrect.** Having a testing environment that
mirrors the production environment is a best practice. This
ensures that the tests are representative of how the software will
behave in production.

**Option C is incorrect.** While zero-day vulnerabilities are a
concern, by definition, they are unknown vulnerabilities.
Testing specifically for them would be challenging. However,
thorough testing and review processes can mitigate potential
risks.

**Option D is incorrect.** Penetration testing is an essential aspect
of security testing, but doing it post-production doesn’t explain


the mismatch between the test results and actual vulnerabilities
if the initial test results were not reviewed correctly.

**Question 29.** An online banking website employs a system that
automatically logs out users after 10 minutes of inactivity to
ensure that if a user forgets to log out, no one else can alter the
user’s banking details. Which principle of the CIA triad is the
banking website MOST directly addressing?
**(A)** Confidentiality
**(B)** Availability
**(C)** Authentication
**(D) Integrity**

**Explanation 29. Correct Answer: D. Integrity.** The integrity
pillar of the CIA triad ensures the accuracy and reliability of
data. By logging out users after a period of inactivity, the
banking website aims to prevent unauthorized modifications
(potentially by someone else who might gain access to the
unattended session) to the user’s banking details, thereby
maintaining the integrity of the data.

**Option A is incorrect.** While logging out users does have a
confidentiality aspect, the primary aim in this scenario is to
prevent unauthorized changes rather than unauthorized viewing.

**Option B is incorrect.** Availability ensures that data and
systems are accessible to authorized users when they need it.
This scenario doesn’t discuss providing or restricting access
based on system uptime or accessibility.


**Option C is incorrect.** Authentication ensures that users are
who they claim to be. While the scenario does touch on security
measures, the primary concern here is preventing unauthorized
changes to data, which aligns with integrity, not authentication.

**Question 30.** A company is located in an area prone to natural
disasters such as earthquakes and floods. Which of the
following physical security controls would be MOST effective
in ensuring the safety of the company’s IT infrastructure?
**(A)** Using biometric authentication for server access
**(B)** Deploying a firewall to protect against cyber threats
**(C) Establishing a raised floor system in the data center
(D)** Conducting penetration testing on a regular basis

**Explanation 30. Correct Answer: C. Establishing a raised
floor system in the data center.** A raised floor system in a data
center serves as a physical control by elevating equipment off
the ground, helping to protect it from potential water damage in
the event of flooding and providing some protection from other
environmental risks.

**Option A is incorrect.** Using biometric authentication is a
technical control that enhances security by confirming users’
identities based on physical or behavioral attributes. While it
strengthens access security, it doesn’t provide protection against
natural disasters.

**Option B is incorrect.** Deploying a firewall is a technical
control that guards against unauthorized access to or from a
private network. While it protects against cyber threats, it


doesn’t offer protection against physical threats like natural
disasters.

**Option D is incorrect.** Conducting penetration testing is a
technical and sometimes operational control that identifies
vulnerabilities in an organization’s digital assets. While it
enhances cyber security, it doesn’t protect infrastructure against
physical threats.

**Question 31.** TechBank has just opened a new branch in the
city center. Due to its location, the management is concerned
about potential vehicular attacks on the facility. Which of the
following physical security measures can TechBank employ to
specifically deter such attacks?
**(A)** Surveillance Cameras
**(B) Bollards
(C)** Access Badges
**(D)** Security Guards

**Explanation 31. Correct Answer: B. Bollards.** Bollards are
short, sturdy vertical posts that are typically used to control road
traffic. In the context of physical security, they serve to prevent
vehicles from entering areas where they are not allowed,
thereby acting as a deterrent against potential vehicular attacks.

**Option A is incorrect.** While surveillance cameras can monitor
and record activities, they do not serve as a physical barrier
against vehicular attacks.

**Option C is incorrect.** Access badges control personnel access
to facilities but do not deter vehicular attacks.


**Option D is incorrect.** While security guards can respond to
threats and control access, they are not a specific measure to
deter vehicular attacks like bollards.

**Question 32.** During a security assessment, Maria, a security
consultant, identifies a self-signed certificate being used on a
client’s public-facing web server. What is the PRIMARY
security concern related to this finding?
**(A)** The web server might be vulnerable to Distributed
Denial of Service (DDoS) attacks
**(B)** The certificate could be expired
**(C) Users cannot validate the authenticity of the website
easily
(D)** The web server might not support modern encryption
algorithms

**Explanation 32. Correct Answer: C. Users cannot validate
the authenticity of the website easily.** Self-signed certificates
are not signed by a recognized Certificate Authority. As a result,
when users connect to a website using a self-signed certificate,
they typically receive a warning that the certificate is not
trusted. This poses a risk as users cannot easily validate the
authenticity of the website, making them more susceptible to
man-in-the-middle attacks.

**Option A is incorrect.** While DDoS attacks are a concern for
public-facing web servers, they aren’t directly related to the use
of self-signed certificates.

**Option B is incorrect.** Any certificate, whether self-signed or
CA-signed, can expire. However, expiration is not the primary


concern related to the use of self-signed certificates on public-
facing servers.

**Option D is incorrect.** The use of modern encryption
algorithms is independent of whether a certificate is self-signed
or not.

**Question 33.** TechFin Bank is considering implementing a new
software system for their transaction processing. Before rolling
it out, the cybersecurity team insists on carrying out a specific
type of analysis to understand how this change might affect the
organization’s security posture. What is the team referring to?
**(A)** Risk appetite assessment
**(B)** Performance benchmarking
**(C) Impact analysis
(D)** Penetration testing

**Explanation 33. Correct Answer: C. Impact analysis.** An
impact analysis assesses the potential consequences of a change
within an organization. In the context of TechFin Bank, the
cybersecurity team would use this analysis to understand how
the new software system might introduce new vulnerabilities,
affect existing security measures, or otherwise impact the
bank’s overall security.

**Option A is incorrect.** Risk appetite assessment refers to
determining the amount and type of risk an organization is
willing to accept in pursuit of its objectives. It doesn’t focus on
the consequences of a specific change.


**Option B is incorrect.** Performance benchmarking focuses on
comparing an organization’s performance metrics to industry
standards or best practices, not assessing the potential security
impact of a change.

**Option D is incorrect.** While penetration testing is crucial to
assess the vulnerabilities in a system, it doesn’t provide a
holistic view of the potential impacts a change might have on an
organization’s security posture.

**Question 34.** To discourage potential cybercriminals from
targeting their online storefront, an e-commerce company is
considering various security measures. Which of the following
would act MOST effectively as a deterrent control?
**(A) Displaying a seal for third-party security
certifications on the website
(B)** Using a Web Application Firewall (WAF)
**(C)** Conducting monthly vulnerability assessments
**(D)** Storing customer data in encrypted databases

**Explanation 34. Correct Answer: A. Displaying a seal for
third-party security certifications on the website.** Displaying
a seal for third-party security certifications on the website
serves as a deterrent control. It sends a message to potential
attackers that the site is recognized for its security measures,
which can discourage them from attempting an attack.

**Option B is incorrect.** Using a Web Application Firewall
(WAF) is a preventive control. It filters, monitors, and blocks
malicious web traffic. While it helps in defending against cyber
threats, it doesn’t act as a visible deterrent to attackers.


**Option C is incorrect.** Conducting monthly vulnerability
assessments is a detective control. It identifies vulnerabilities to
enhance the security posture but doesn’t act as a deterrent to
potential attackers.

**Option D is incorrect.** Storing customer data in encrypted
databases is a preventive control that ensures unauthorized
individuals cannot access or comprehend the stored data. It
protects data but doesn’t deter potential cybercriminals.

**Question 35.** The security team of a multinational company
deployed a network of honeypots globally, making it appear as
an interconnected and realistic environment. They aim to study
coordinated multi-stage attacks. This deceptive setup is known
as:
**(A)** Firewall Cluster
**(B)** Virtual LAN (VLAN)
**(C)** Distributed Denial of Service (DDoS) Prevention
**(D) Honeynet**

**Explanation 35. Correct Answer: D. Honeynet.** A honeynet is
essentially a network of honeypots. It is designed to be
attractive to attackers, making them believe they are attacking a
real network, while in reality, they are being observed, and their
tactics and techniques are being analyzed.

**Option A is incorrect.** A Firewall Cluster is a group of
firewalls operating together to increase reliability and
performance. It doesn’t serve as a decoy to attract attackers.


**Option B is incorrect.** A Virtual LAN (VLAN) is a network
protocol used to create logically segmented networks within a
physical network, improving performance and security. It’s not
a decoy system.

**Option C is incorrect.** Distributed Denial of Service (DDoS)
Prevention solutions focus on identifying and mitigating large-
scale attempts to disrupt network service availability. They do
not present a deceptive environment for attackers.

**Question 36.** ExamsDigest Corp, a technology company,
recently conducted a security assessment to align with industry
best practices. The company’s current security posture was
compared to its desired future state, revealing discrepancies.
Which of the following best describes the approach
ExamsDigest Corp employed?
**(A)** Vulnerability Assessment
**(B)** Penetration Testing
**(C) Gap Analysis
(D)** Threat Modeling

**Explanation 36. Correct Answer: C. Gap Analysis.** Gap
analysis is a method of comparing the current state of
something (such as security measures) with a future desired
state to identify the discrepancies or “gaps”. In the scenario,
ExamsDigest Corp compared their current security posture to a
desired future state, which is consistent with the process of gap
analysis.


**Option A is incorrect.** A vulnerability assessment focuses on
identifying, quantifying, and ranking vulnerabilities in a system,
not comparing the current state with a desired future state.

**Option B is incorrect.** Penetration testing is an authorized
simulated cyberattack on a system to evaluate its security, not to
compare its current state with a desired future state.

**Option D is incorrect.** Threat modeling is the process of
identifying potential threats to a system and determining the risk
they pose, not comparing the current state with a desired future
state.

**Question 37.** A pharmaceutical company is concerned about
competitors accessing their formula for a new drug. Which
pillar of the CIA triad is MOST directly addressed by their
concern?
**(A)** Availability
**(B) Confidentiality
(C)** Integrity
**(D)** Non-repudiation

**Explanation 37. Correct Answer: B. Confidentiality.** The
confidentiality pillar of the CIA triad ensures that information is
accessible only to those with authorized access. In this scenario,
the company wants to ensure that its drug formula remains
secret and is not accessible to unauthorized individuals,
particularly competitors.

**Option A is incorrect.** Availability ensures that information is
accessible to authorized users when needed. The concern here is


not about access to the data but rather about preventing
unauthorized access.

**Option C is incorrect.** Integrity ensures the accuracy and
reliability of data and systems. The scenario doesn’t mention
concerns about the formula being altered, only about
unauthorized access.

**Option D is incorrect.** Non-repudiation is a concept ensuring
that a party in a dispute cannot deny the authenticity of their
actions. It’s not directly related to the company’s concern about
the secrecy of their drug formula.

**Question 38.** FinCorp, a financial institution, has recently
adopted a new security framework. In this framework, every
device and user inside the organization’s network is treated as if
they were outside the perimeter, necessitating rigorous
verification processes even for internal requests. Which security
paradigm has FinCorp implemented?
**(A)** Demilitarized Zone (DMZ)
**(B)** Network Segmentation
**(C)** Intrusion Detection System (IDS)
**(D) Zero Trust**

**Explanation 38. Correct Answer: D. Zero Trust.** Zero Trust
is a security model that treats every access request with
skepticism, regardless of its origin, be it from within or outside
the organization’s traditional perimeter. It requires rigorous
verification processes for every interaction.


**Option A is incorrect.** A Demilitarized Zone (DMZ) is a
physical or logical subnetwork that exposes an organization’s
external-facing services to a larger, untrusted network, usually
the internet.

**Option B is incorrect.** Network Segmentation divides a
network into multiple segments or subnets, allowing
administrators to control the flow of traffic between them. It
does not inherently distrust all traffic like Zero Trust.

**Option C is incorrect.** Intrusion Detection System (IDS) is a
device or software application that monitors a network or
systems for malicious activity or policy violations. It doesn’t
define how trust is managed across interactions.

**Question 39.** GreenValley Mall, located in a busy urban area,
has recently faced security concerns due to the proximity of its
main entrance to a major road. Which physical security
enhancement can the mall management implement to create a
protective barrier between the road and the entrance, ensuring
pedestrian safety and preventing unauthorized vehicular access?
**(A)** Reinforced Walls
**(B)** Metal Detectors
**(C) Bollards
(D)** Perimeter Fencing

**Explanation 39. Correct Answer: C. Bollards. B** ollards are
robust vertical posts, usually made of steel or concrete, which
can be placed at specific intervals to form a protective barrier.
They can effectively prevent vehicles from accessing pedestrian


areas or building entrances while allowing pedestrian
movement.

**Option A is incorrect.** While reinforced walls can offer
protection against various threats, they would not be practical
for separating a mall entrance from a road as they would block
pedestrian access as well.

**Option B is incorrect.** Metal detectors are used for detecting
metal objects and weapons on individuals entering a facility, not
for stopping vehicular access.

**Option D is incorrect.** Perimeter fencing can deter
unauthorized access, but it might not specifically prevent fast-
moving vehicular threats like bollards do. Furthermore, a fence
might not be aesthetically pleasing or practical in front of a mall
entrance.

**Question 40.** A tech company, InnovateTech, has recently faced
multiple incidents of unauthorized personnel trying to access
their R&D labs. They wish to monitor and record all activities
near the entrance of this sensitive area. Which physical security
measure would be most effective for this requirement?
**(A)** RFID Badge Readers
**(B)** Biometric Scanners
**(C) Video Surveillance Cameras
(D)** Mantrap

**Explanation 40. Correct Answer: C. Video Surveillance
Cameras.** Video surveillance cameras provide a continuous
visual monitoring capability and can record activities near


specific areas. For the purpose of observing and recording
incidents near the entrance of the R&D labs, video surveillance
would be the most direct and effective solution.

**Option A is incorrect.** While RFID badge readers can control
access and log which badges are used at entrances, they do not
visually monitor or record activities.

**Option B is incorrect.** Biometric scanners are an authentication
mechanism, and while they offer a high level of security for
access control, they do not provide visual monitoring or
recording capabilities.

**Option D is incorrect.** A mantrap is a physical security access
control system that prevents tailgating into secure areas. While
it can enhance security at entrances, it does not visually record
activities.

**Question 41.** A cybersecurity analyst at XYZ Corp is looking to
deploy a system that appears to be vulnerable and enticing to
attackers. The main goal is to study the tactics, techniques, and
procedures (TTPs) of potential adversaries, without them
realizing that they’re interacting with a decoy. Which of the
following would BEST meet this requirement?
**(A)** Intrusion Detection System (IDS)
**(B)** Firewall
**(C) Honeypot
(D)** VPN Concentrator

**Explanation 41. Correct Answer: C. Honeypot.** A honeypot is
a security mechanism designed to lure attackers into interacting


with a seemingly vulnerable system. Its primary purpose is not
to block or prevent attacks but to log and study them. By
analyzing the activities on the honeypot, security professionals
can gain insights into the methods and motivations of the
attackers.

**Option A is incorrect.** An Intrusion Detection System (IDS) is
designed to detect malicious activities on a network and alert
administrators. While it can identify threats, it doesn’t actively
lure attackers.

**Option B is incorrect.** A firewall is designed to block or allow
traffic based on specific rules. It doesn’t present itself as a
vulnerable target to lure attackers.

**Option D is incorrect.** A VPN Concentrator is a device that
provides remote access to a network over a secure connection.
Its primary purpose is to enable secure remote access, not to act
as a decoy for attackers.

**Question 42.** A multinational organization recently experienced
a significant security breach. After investigating, it was
determined that a change to the network infrastructure was
made without undergoing the standard approval process. As a
result, there was a misconfiguration which allowed
unauthorized access. What security principle related to change
management did the organization neglect?
**(A)** Configuration baseline reviews
**(B) Least privilege enforcement
(C)** Approval process adherence
**(D)** Patch management


**Explanation 42. Correct Answer: C. Approval process
adherence.** The approval process is a critical aspect of change
management. Before any changes are made, especially to
critical systems like network infrastructure, they need to
undergo a rigorous approval process. This ensures that multiple
experts evaluate the change for potential vulnerabilities or
issues. In this scenario, skipping the approval process led to a
significant security breach.

**Option A is incorrect.** Configuration baseline reviews involve
regularly checking and ensuring that systems are configured as
per the organization’s approved baseline. While it’s important,
the issue in the scenario was directly related to bypassing the
approval process.

**Option B is incorrect.** Least privilege enforcement means
providing only the minimal necessary access to users to perform
their tasks. This scenario doesn’t deal with access rights or
privileges.

**Option D is incorrect.** Patch management concerns the process
of applying updates to software and systems. The breach in the
question wasn’t related to missing patches but was due to
bypassing the approval process.

**Question 43.** After a series of cyber-attacks on a company’s
infrastructure, the IT team decided to deploy a solution that
would seem like a legitimate part of their network but is
intentionally isolated and monitored. They intend to detect and
analyze malicious activities in this isolated environment. What
technology are they most likely implementing?


**(A)** Network segmentation
**(B) Honeypot
(C)** DMZ (Demilitarized Zone)
**(D)** Sandboxing

**Explanation 43. Correct Answer: B. Honeypot.** A honeypot is
intentionally set up to appear as a legitimate part of a network,
but it is isolated and closely monitored. Its purpose is to attract
attackers and observe their actions, thereby providing insights
into potential threats and the methods employed by adversaries.

**Option A is incorrect.** Network segmentation involves dividing
a network into smaller sub-networks. While this can enhance
security by limiting attackers’ access to specific segments, it
doesn’t act as a decoy to attract attackers.

**Option C is incorrect.** A DMZ (Demilitarized Zone) is a
subnet that acts as a buffer between the internet and an
organization’s internal network. While it can contain servers
accessible to external users, its primary purpose is not to act as a
decoy but to provide a layer of protection.

**Option D is incorrect.** Sandboxing is a security mechanism
that allows programs to run in a separate environment to
prevent them from affecting the broader system. It’s used for
testing and analyzing potentially malicious software, not for
luring attackers.

**Question 44.** Liam, the CTO of a medium-sized enterprise,
noticed that several software applications were not updated
regularly, leading to potential security vulnerabilities. Upon


investigation, he realized that no specific team or individual was
assigned as the owner of these applications. To enhance
security, what should Liam emphasize?
**(A)** Immediate decommissioning of all unowned
applications
**(B) Assignment of clear ownership to all business
applications
(C)** Conducting monthly vulnerability assessments on all
applications
**(D)** Outsourcing the management of these applications to
third-party vendors

**Explanation 44. Correct Answer: B. Assignment of clear
ownership to all business applications.** Assigning clear
ownership ensures that there’s a designated team or individual
responsible for the upkeep, updates, and security of an
application. When there’s clear ownership, the owner has the
accountability to maintain and secure the application, reducing
the risk of oversights like missing updates.

**Option A is incorrect.** Immediate decommissioning may not be
practical or feasible, especially if the applications are critical to
business operations.

**Option C is incorrect.** While monthly vulnerability
assessments can help identify security issues, they don’t address
the root cause of the problem highlighted in the scenario – the
lack of ownership and accountability.

**Option D is incorrect.** Outsourcing may shift the responsibility,
but it doesn’t ensure that the applications will be better managed


or more secure. Ownership clarity is paramount, whether the
management is internal or outsourced.

**Question 45.** TechSoft Corp, a mid-sized software development
firm, is relocating its main office to a new building. The
management is concerned about potential threats after hours,
particularly due to the increasing reports of cyber-espionage.
They are evaluating different security measures. Which option
would provide an immediate physical presence and deterrence
during non-business hours?
**(A)** CCTV with motion detection
**(B)** Retinal scan at all entrances
**(C) Security guard presence
(D)** Reinforced doors and windows

**Explanation 45. Correct Answer: C. Security guard
presence.** A security guard provides a visible deterrent and
immediate physical presence. This human element is invaluable
in situations where an immediate response to threats, deterrence
of potential intruders, or evaluation of suspicious activities is
required, especially during non-business hours.

**Option A is incorrect.** While CCTV with motion detection can
monitor and alert on movement, it doesn’t provide the
immediate human response and deterrence a security guard
does.

**Option B is incorrect.** A retinal scan is an authentication
mechanism for controlling access. While it offers high security,
it doesn’t offer the visible deterrence or immediate response of a
security guard.


**Option D is incorrect.** Reinforced doors and windows enhance
the physical security of a building, but they don’t provide an
active and visible human deterrence like a security guard.

**Question 46.** Alice, a system administrator for a startup, is
preparing to deploy a new website for her company. To ensure
secure communications between the users and the website, she
plans to obtain a digital certificate for the site. Before doing so,
which step must Alice first undertake to get a certificate from a
Certificate Authority (CA)?
**(A) Generate a public-private key pair
(B)** Submit her passport copy to the CA
**(C)** Download the latest CA root certificate
**(D)** Encrypt the website with symmetric encryption

**Explanation 46. Correct Answer: A. Generate a public-
private key pair.** Before Alice can request a digital certificate
from a CA, she must first generate a public-private key pair.
Once this is done, she creates a Certificate Signing Request
(CSR) containing her public key and some additional
information. The CSR is then submitted to the CA for signing.

**Option B is incorrect.** CAs do not typically require a passport
copy for standard SSL/TLS certificates. They might have
identity verification processes, but it’s usually for extended
validation certificates.

**Option C is incorrect.** While it may be necessary to trust a CA
by downloading its root certificate, this is not the step required
before requesting a digital certificate.


**Option D is incorrect.** Symmetric encryption is unrelated to the
process of obtaining a digital certificate.

**Question 47.** Julia, a security administrator, is concerned about
potential unauthorized access to confidential project files stored
on a company server. She decides to place a document within
the project folders that seems enticing but is actually monitored
for access. This strategy aims to detect if someone is accessing
files without authorization. What is this document commonly
known as?
**(A)** Salt file
**(B) Honeyfile
(C)** Log file
**(D)** Backup file

**Explanation 47. Correct Answer: B. Honeyfile.** A honeyfile is
a monitored file placed intentionally to act as a decoy. If
accessed, it can provide an alert that someone might be
accessing files without proper authorization, or it might be an
indication of a potential insider threat.

**Option A is incorrect.** A salt is random data that is used as an
additional input to a one-way function that hashes data or
passwords. It isn’t a decoy file.

**Option C is incorrect.** A log file records events in an operating
system or other software to aid in troubleshooting and activity
monitoring, but it isn’t used as a deceptive measure.


**Option D is incorrect.** A backup file is a copy of a file or
database that can be used for data recovery. It’s not a decoy to
detect unauthorized access.

**Question 48.** After a recent incident of vandalism, a corporate
building is considering implementing security controls that
would dissuade potential perpetrators. Which of the following
would serve BEST as a deterrent control?
**(A)** Encrypting all stored data
**(B)** Installing biometric access controls on all entrances
**(C)** Implementing regular data backups
**(D) Placing visible security signage indicating 24/7
surveillance**

**Explanation 48. Correct Answer: D. Placing visible security
signage indicating 24/7 surveillance.** Visible security signage
serves as a deterrent control as it discourages potential
perpetrators by signaling the risk of detection and
consequences, even if actual surveillance might not be active at
all times.

**Option A is incorrect.** Encrypting all stored data is a preventive
control that ensures unauthorized individuals cannot access or
understand the encrypted information. It doesn’t deter acts of
physical vandalism.

**Option B is incorrect.** Installing biometric access controls is a
preventive control that restricts physical access based on unique
biological attributes. While it prevents unauthorized access, it
doesn’t act as a visible deterrent to vandalism or other potential
threats.


**Option C is incorrect.** Implementing regular data backups is a
corrective control designed to restore data after a security
incident. It doesn’t deter potential threats.

**Question 49.** Alice wants to access a restricted online portal.
The portal asks her to enter a unique username and a secret
passphrase only she should know. This process helps the system
ensure that Alice is who she claims to be. What security concept
is the portal employing?
**(A)** Authorization
**(B)** Accounting
**(C)** Multifactor authentication
**(D) Authentication**

**Explanation 49. Correct Answer: D. Authentication.**
Authentication is the process of verifying the identity of a user,
system, or application. In the described scenario, Alice is
proving her identity to the system by providing a unique
username and a passphrase, which are credentials that
supposedly only she possesses.

**Option A is incorrect.** Authorization determines what actions,
resources, or services a verified identity is allowed to access or
perform. It does not deal with verifying the identity itself.

**Option B is incorrect.** Accounting involves tracking user
activities and recording them for audit purposes. It does not
directly verify a user’s identity.

**Option C is incorrect.** Multifactor authentication requires two
or more methods of verification from different categories of


credentials. The scenario only mentioned a username and
passphrase, which is a single-factor authentication method.

**Question 50.** Sophia, the cybersecurity lead at XYZ Corp, is in
the process of drafting a new security policy. During the
drafting process, she primarily consults with her security team.
However, upon implementation, several departments pushed
back due to the policy interfering with their operations. Which
best describes the misstep Sophia made during the policy
creation process?
**(A)** Not using a standardized security framework
**(B)** Over-reliance on automated security solutions
**(C) Not including key stakeholders in the policy drafting
process
(D)** Focusing too much on external threats rather than
internal ones

**Explanation 50. Correct Answer: C. Not including key
stakeholders in the policy drafting process.** Stakeholders
from different departments provide crucial insights into how
security measures can impact various operations and processes
within an organization. By including them in the policy drafting
process, Sophia would have received feedback that could have
helped shape a policy that not only maintains security but also
aligns with the needs of different departments.

**Option A is incorrect.** While using a standardized security
framework can provide guidance, it doesn’t necessarily account
for the unique operational needs of different departments within
an organization.


**Option B is incorrect.** The scenario doesn’t mention any
reliance, over or otherwise, on automated security solutions.

**Option D is incorrect.** While both external and internal threats
are crucial considerations, the primary issue here was the lack
of consultation with key stakeholders.

**Question 51.** BioGen Inc., a biotechnology company, has
implemented a layered security approach. They are considering
adding a human element to their security measures for their
research labs. Which of the following would best provide the
ability to evaluate and respond to various security situations
with human judgment?
**(A)** Installing biometric locks
**(B) Employing security guards
(C)** Implementing an access control vestibule
**(D)** Deploying AI-driven security cameras

**Explanation 51. Correct Answer: B. Employing security
guards.** Security guards provide the advantage of human
judgment and can evaluate, respond, and adapt to a wide variety
of security situations, making them ideal for adding a human
element to a layered security approach.

**Option A is incorrect.** While biometric locks can control access
based on unique human features, they don’t provide the
evaluation and response capabilities of a human guard.

**Option C is incorrect.** An access control vestibule controls
access into an area, often with two sets of doors, but it does not


provide the evaluation, judgment, or immediate response that a
security guard does.

**Option D is incorrect.** While AI-driven security cameras can
provide advanced monitoring and potentially detect suspicious
activities, they don’t replace the judgment and immediate
response capabilities of a human security guard.

**Question 52.** While analyzing server logs, Mike, an IT security
analyst, noticed that an unfamiliar document was frequently
accessed. Upon investigation, he realized that this document
was deliberately placed by the security team and had no real
data but was closely monitored. The purpose of this file is
MOST likely:
**(A)** To serve as a redundancy copy in case of data loss
**(B) To act as a decoy to attract and detect unauthorized
access
(C)** To maintain a record of all user activities for auditing
**(D)** To be encrypted and sent to clients as a sample

**Explanation 52. Correct Answer: B. To act as a decoy to
attract and detect unauthorized access.** Honeyfiles serve as
deceptive measures, attracting potential malicious actors or
unauthorized users. If such files are accessed, it can be an
indication of unauthorized or suspicious activities in the system.

**Option A is incorrect.** Redundancy copies or backups are
created to prevent data loss due to unforeseen issues, but they
are not monitored as decoys.


**Option C is incorrect.** User activity logs maintain records of
actions taken within a system or application, which is different
from a deceptive measure like a honeyfile.

**Option D is incorrect.** Files encrypted for client samples serve
a different purpose and are not typically used as decoys to
detect unauthorized access.

**Question 53.** DataCenter Inc. is located in a region prone to
protests and vandalism. They wish to enhance their perimeter
security to deter potential intruders and make it visibly clear
that unauthorized access is restricted. Which of the following
physical security measures would be the most effective first line
of defense for the company?
**(A)** Sliding Doors
**(B)** Security Cameras
**(C) High-security Fencing
(D)** Proximity Card Readers

**Explanation 53. Correct Answer: C. High-security Fencing.**
High-security fencing is a primary physical barrier that serves to
deter, delay, and detect intrusions. It provides a clear visual
indicator that delineates private property and can be equipped
with additional deterrents such as barbed wire or sensors.
Especially in areas prone to vandalism or protests, a robust
fence acts as an immediate barrier to unauthorized access.

**Option A is incorrect.** Sliding doors are more applicable to
internal security or entrance points and do not serve as a
primary external barrier.


**Option B is incorrect.** While security cameras monitor and
record activities, they do not serve as a physical barrier to
prevent or deter unauthorized access.

**Option D is incorrect.** Proximity card readers are a form of
access control that checks the credentials of individuals but do
not serve as a primary deterrent against vandalism or protests.

**Question 54.** SecureTech Corp, a company dealing with
sensitive client data, is redesigning its main office entrance to
enhance security. They want to ensure that only one person
gains access at a time, even if multiple people try to enter using
a single authorized access badge. Which of the following would
best serve this purpose?
**(A)** CCTV Cameras
**(B) Mantrap
(C)** Biometric Scanners
**(D)** Motion Detectors

**Explanation 54. Correct Answer: B. Mantrap.** A mantrap,
also known as an access control vestibule, is a physical security
access control system comprising a small space with two sets of
interlocking doors. The first door must close before the second
door opens, ensuring that only one person can pass through at a
time. This design prevents tailgating or piggybacking, where
unauthorized individuals attempt to enter a secure area by
following closely behind an authorized individual.

**Option A is incorrect.** While CCTV cameras monitor and
record activities, they do not physically prevent multiple people
from entering at once using a single access badge.


**Option C is incorrect.** Biometric scanners provide a means of
authenticating individuals based on unique physical or
behavioral characteristics, but they do not prevent tailgating on
their own.

**Option D is incorrect.** Motion detectors can detect movement
but do not restrict the entry of multiple individuals trying to use
a single authorized access badge.

**Question 55.** While setting up a new internal web application,
Laura, a system administrator, decides to use a digital certificate
for SSL/TLS encryption. Due to budget constraints, she can’t
procure a certificate from a commercial Certificate Authority
(CA). Which of the following would be a viable option for
Laura to secure the application?
**(A)** Rely on plaintext HTTP for the application
**(B)** Obtain a certificate from a free Certificate Authority
**(C) Generate a self-signed certificate
(D)** Use a shared certificate from another application

**Explanation 55. Correct Answer: C. Generate a self-signed
certificate.** A self-signed certificate can be generated by Laura
without the need for a Certificate Authority. While self-signed
certificates can cause trust issues in public-facing applications
(since they aren’t signed by a recognized CA), they can be
appropriate for internal applications where users can be
informed and trust can be established manually.

**Option A is incorrect.** Relying on plaintext HTTP doesn’t
provide any encryption or security for the application, leaving it
vulnerable to various attacks.


**Option B is incorrect.** While obtaining a certificate from a free
Certificate Authority is a valid option, it wasn’t the best choice
given the specific scenario which emphasizes not using a CA.

**Option D is incorrect.** Using a shared certificate from another
application can introduce security risks and is not a
recommended practice.

**Question 56.** A network administrator has received a new
security patch for a mission-critical application. Which of the
following is the BEST action to take before applying this patch
in the live environment?
**(A)** Apply the patch immediately to ensure system security
**(B)** Notify all users about the upcoming downtime due to
the patch
**(C) Test the patch in a separate testing environment
(D)** Take a backup of only the mission-critical application

**Explanation 56. Correct Answer: C. Test the patch in a
separate testing environment.** Testing any changes, including
patches, in a separate environment before deploying them to
production is essential to ensure there are no unintended
technical implications. This is a key aspect of change
management processes and helps prevent system outages or
vulnerabilities from being introduced.

**Option A is incorrect.** While applying patches is crucial for
maintaining security, doing so immediately without proper
testing can lead to unforeseen technical problems.


**Option B is incorrect.** Notifying users is important, but it’s
premature to notify them without first testing the patch.

**Option D is incorrect.** Taking a backup is a good practice, but
it is not a substitute for testing the patch first.

**Question 57.** After implementing a major security update to its
database system, TechCo experienced unexpected downtime
and system incompatibilities. The CISO wants to ensure that
such incidents can be quickly addressed in the future. Which of
the following should TechCo have had in place before
deploying the update to mitigate the impact of these kinds of
incidents?
**(A)** A comprehensive list of all updates
**(B)** An automated system recovery tool
**(C) A backout plan
(D)** A detailed user manual for the update

**Explanation 57. Correct Answer: C. A backout plan.** A
backout plan is a pre-arranged strategy or set of procedures to
reverse changes made to the system in case the changes have
adverse effects. In scenarios like this, where a significant update
causes unintended problems, a backout plan would allow the
organization to revert the system to its previous stable state
quickly.

**Option A is incorrect.** While having a comprehensive list of all
updates is good for documentation and auditing purposes, it
would not directly help in mitigating the effects of an adverse
update.


**Option B is incorrect.** An automated system recovery tool
might assist in reverting changes or recovering the system.
However, a backout plan is more specific to undoing changes
made during an update or change process, making it more
suitable in this context.

**Option D is incorrect.** A detailed user manual for the update is
beneficial for training and troubleshooting, but it wouldn’t serve
the direct purpose of reverting unintended adverse changes.

**Question 58.** A financial institution processes thousands of
credit card transactions daily. To ensure the security and
integrity of these transactions, the security officer wants to
employ a solution that will safely manage and store
cryptographic keys. Which of the following would be the
MOST suitable solution?
**(A)** Trusted Platform Module (TPM)
**(B)** Full Disk Encryption (FDE)
**(C) Hardware Security Module (HSM)
(D)** Software Key Repository

**Explanation 58. Correct Answer: C. Hardware Security
Module (HSM).** Hardware Security Modules (HSMs) are
physical devices specifically designed to manage, process, and
store cryptographic keys. They provide a high level of
protection against both physical and logical attacks and are
commonly used by financial institutions to ensure the security
of high-value transactions.

**Option A is incorrect.** While TPMs provide hardware-level
security for individual devices, they are not designed for the


high-capacity cryptographic needs of an institution processing
numerous transactions.

**Option B is incorrect.** Full Disk Encryption (FDE) secures data
at rest on a hard drive but doesn’t specifically manage
cryptographic keys used in transaction processing.

**Option D is incorrect.** While a software key repository can
store cryptographic keys, it lacks the same level of physical and
logical protection provided by an HSM.

**Question 59.** During the setup of a secure communication
channel, Alice and Bob need to agree upon a shared secret key
without sending the key directly to each other, as they fear
eavesdropping. Which protocol would best facilitate this
requirement?
**(A)** RSA
**(B)** HMAC
**(C) Diffie-Hellman
(D)** AES

**Explanation 59. Correct Answer: C. Diffie-Hellman.** The
Diffie-Hellman key exchange protocol allows two parties to
each generate public and private key pairs, exchange the public
keys, and then derive a shared secret key. This secret key can
then be used for symmetric encryption. The beauty of this
protocol is that the shared secret can be derived without directly
sending it over the communication channel, preventing
eavesdroppers from obtaining the secret key directly.


**Option A is incorrect.** RSA is an asymmetric encryption
method, not a key exchange protocol.

**Option B is incorrect.** HMAC is a specific type of message
authentication code that involves hashing and is not used for
key exchange.

**Option D is incorrect.** AES is a symmetric encryption
algorithm and does not offer a key exchange mechanism.

**Question 60.** A company is developing a new video
conferencing tool. They want to make sure that all video and
audio data transmitted between participants are encrypted and
protected from eavesdropping. Which type of encryption should
the developers implement to achieve this?
**(A)** Endpoint Encryption
**(B) Transport-layer Encryption
(C)** Volume-level Encryption
**(D)** Database-level EncryptionAccess Control

**Explanation 60. Correct Answer: B. Transport-layer
Encryption.** Transport-layer Encryption is tailored for securing
data while it’s in transit. By implementing this encryption, the
company ensures that all video and audio data during the video
conference are confidential, maintaining the privacy of the
participants.

**Option A is incorrect.** Endpoint Encryption is designed for data
on specific devices like laptops or mobile phones, not for data
being transmitted over networks.


**Option C is incorrect.** Volume-level Encryption pertains to
encrypting specific logical drives or volumes, not data in transit.

**Option D is incorrect.** Database-level Encryption secures data
within a database and is not specific to data transmission over
networks.

**Question 61.** After a significant cybersecurity incident, ABC
Tech revamped its incident response procedures. However, the
documentation was not updated to reflect these changes. During
a subsequent minor incident, there was confusion regarding the
steps to be followed. Which of the following is the MOST direct
implication of not updating the incident response
documentation?
**(A)** The company may have to invest in new cybersecurity
tools
**(B)** Stakeholders might lose trust in the company’s ability to
handle incidents
**(C) Incident response might be inconsistent and less
effective
(D)** ABC Tech may have to hire external consultants for
incident response

**Explanation 61. Correct Answer: C. Incident response
might be inconsistent and less effective.** Without up-to-date
documentation reflecting the most recent incident response
procedures, there’s a risk that the response will be inconsistent,
leading to inefficiencies and potential oversights.


**Option A is incorrect.** While new tools might be beneficial, the
direct concern with outdated documentation is the effectiveness
of the response.

**Option B is incorrect.** While stakeholder trust is important, the
immediate implication of outdated documentation is the quality
of the incident response.

**Option D is incorrect.** Hiring external consultants might be an
option, but the direct consequence of outdated documentation is
the potential ineffectiveness of the internal response process.

**Question 62.** A financial organization is considering
implementing a system that allows all users to view all
transactions, but once a transaction is recorded, it cannot be
altered or deleted. They want this transparency to foster trust
among their users. Which of the following would best meet this
requirement?
**(A)** Digital certificate
**(B) Open public ledger
(C)** Symmetric encryption
**(D)** Secure file transfer protocol

**Explanation 62. Correct Answer: B. Open public ledger.** An
open public ledger provides transparency by allowing all users
to view all transactions. Moreover, once a transaction is added
to the ledger, it becomes immutable, meaning it cannot be
altered or deleted, ensuring data integrity and fostering trust
among participants.


**Option A is incorrect.** A digital certificate is used to verify the
identity of an entity and bind a public key to it, but it doesn’t
offer the transparency of transactions or their immutability.

**Option C is incorrect.** Symmetric encryption is used to encrypt
and decrypt data using a single secret key, but it doesn’t provide
transaction transparency or immutability.

**Option D is incorrect.** Secure file transfer protocol (SFTP) is a
method to securely transfer files over a network, but it doesn’t
maintain an open public ledger of transactions.

**Question 63.** A company is implementing a system to ensure
that code released to production is both unaltered and approved
by a specific team member. Which of the following
cryptographic techniques should they implement?
**(A)** Symmetric encryption of the code
**(B)** Hashing the code with SHA-256
**(C)** Encrypting the code with the team member's public key
**(D) Digital signature by the team member**

**Explanation 63. Correct Answer: D. Digital signature by the
team member.** Digital signatures provide both integrity and
non-repudiation. By having the specific team member digitally
sign the code, the company can ensure that the code has not
been altered (integrity) and that it was approved by the
designated individual (non-repudiation).

**Option A is incorrect.** Symmetric encryption provides
confidentiality, but it doesn’t provide the needed integrity and
non-repudiation in this scenario.


**Option B is incorrect.** Hashing the code provides a mechanism
to check for alterations (integrity), but it does not provide non-
repudiation or evidence of the specific team member’s approval.

**Option C is incorrect.** Encrypting with the team member’s
public key doesn’t provide non-repudiation. Moreover, only the
team member with the corresponding private key would be able
to decrypt it, which might not be desirable for production
releases.

**Question 64.** Your company has recently deployed an update to
its CRM application. Post-update, users are experiencing
connectivity issues. As a security administrator, which of the
following steps should you take FIRST to address the
connectivity problem without causing data loss?
**(A)** Restart the application immediately
**(B)** Disconnect all users and then restart the application
**(C) Validate the update's integrity and then restart the
application
(D)** Reinstall the previous version of the CRM application

**Explanation 64. Correct Answer: C. Validate the update’s
integrity and then restart the application.** Before making any
changes, it’s essential to ensure the update’s integrity. This
means confirming that the update was correctly applied and that
there were no issues during its installation. Once the update’s
integrity is confirmed, a restart can help apply any changes that
may not have taken effect immediately.


**Option A is incorrect.** Restarting the application immediately
without validation might cause further complications if the
update was not correctly applied.

**Option B is incorrect.** While disconnecting users might be
necessary at some point, doing so without validating the
update’s integrity can result in further disruptions.

**Option D is incorrect.** Reinstalling the previous version is a
drastic step and might not be necessary if the update’s integrity
can be validated and issues resolved with a restart.

**Question 65.** TechDynamics, a growing tech startup, plans to
scale its operations and serve a global clientele. Given that their
client base operates in multiple time zones, when should
TechDynamics schedule their system maintenance to ensure
minimal disruption?
**(A)** During the busiest hours for their headquarters' local
time
**(B)** Staggered based on the peak hours of their global clients
**(C)** Only when a system breakdown occurs
**(D) Establish a consistent maintenance window during
off-peak hours for the majority of their clientele**

**Explanation 65. Correct Answer: D. Establish a consistent
maintenance window during off-peak hours for the majority
of their clientele.** When serving a global clientele operating in
various time zones, it’s crucial to establish a maintenance
window during hours when the majority of clients are least
active. This minimizes disruptions and ensures smooth
operations for most clients.


**Option A is incorrect.** Focusing only on the headquarters’ local
time disregards the operational hours of global clients. This
approach might cause disruptions for clients in other time
zones.

**Option B is incorrect.** While staggering maintenance based on
peak hours of global clients seems logical, it could lead to a
complex and hard-to-manage maintenance schedule, especially
as the client base grows.

**Option C is incorrect.** Waiting for a system breakdown to
perform maintenance is reactive rather than proactive. This
approach might lead to more extended and unpredictable
downtimes, resulting in greater disruptions and potential
security risks.

**Question 66.** During an IT audit, a company’s encryption
practices come under scrutiny. The IT auditor recommends
increasing the encryption key length for certain applications to
improve security. What is the PRIMARY reason to increase the
encryption key length?
**(A)** To speed up encryption and decryption processes
**(B)** To ensure compatibility with older systems
**(C) To reduce the possibility of a brute force attack
(D)** To reduce the key management overhead

**Explanation 66. Correct Answer: C. To reduce the
possibility of a brute force attack.** Increasing the encryption
key length primarily enhances the security of the encryption by
making it more resistant to brute-force attacks. A brute force
attack involves trying all possible key combinations, and a


longer key length means exponentially more possible
combinations, making the attack vastly more time-consuming
and difficult.

**Option A is incorrect.** Longer key lengths generally slow down
the encryption and decryption processes, as more computational
power is required.

**Option B is incorrect.** Increasing key length might make the
encryption incompatible with older systems that do not support
the newer, longer key lengths.

**Option D is incorrect.** Key management overhead typically
increases with longer key lengths, as more data must be
managed and kept secure.

**Question 67.** Sarah is working on a project where she needs to
validate the integrity and authenticity of assets over time,
without a centralized authority. Which technology would be
most appropriate for this use case?
**(A)** Digital signature
**(B)** Key escrow
**(C) Blockchain
(D)** Key management system

**Explanation 67. Correct Answer: C. Blockchain.** Blockchain
technology allows for the validation of the integrity and
authenticity of assets over time in a decentralized manner. Each
transaction or asset is verified by the network’s participants and
added to the chain, ensuring its authenticity and making it
tamper-evident.


**Option A is incorrect.** While digital signatures can validate the
authenticity and integrity of a message or document, they do not
provide a decentralized ledger of assets over time.

**Option B is incorrect.** Key escrow is a method where
cryptographic keys are held in trust so that a third party can
have access under certain conditions. It doesn’t help in
validating the integrity and authenticity of assets over time
without central authority.

**Option D is incorrect.** A key management system manages the
life cycle of cryptographic keys but does not inherently validate
the integrity and authenticity of assets over time.

**Question 68.** A graphic design company frequently works with
large files such as videos and high-resolution images. These
files are stored on a dedicated storage volume in their server.
While they need to secure this data, they don’t want to encrypt
individual files due to the volume of data and frequent access
needs. Which encryption approach is most appropriate for this
scenario?
**(A)** File-level Encryption
**(B)** Full-disk Encryption
**(C)** Transport-layer Encryption
**(D) Volume-level Encryption**

**Explanation 68. Correct Answer: D. Volume-level
Encryption.** Volume-level Encryption allows the encryption of
a specific volume or logical drive. In this case, it would allow
the company to encrypt the entire volume where these large
files are stored without having to encrypt individual files. This


ensures the entire volume’s contents are encrypted while
providing seamless access when required.

**Option A is incorrect.** File-level Encryption, while effective,
would be tedious for encrypting numerous large files
individually.

**Option B is incorrect.** Full-disk Encryption would encrypt the
entire disk, which may not be necessary if only a specific
volume requires encryption.

**Option C is incorrect.** Transport-layer Encryption pertains to
data in transit and doesn’t address the need of encrypting stored
data at the volume level.

**Question 69.** An e-commerce company stores millions of
customer transaction records in their primary database. They
have decided to enhance their security posture by applying
encryption to protect sensitive data. However, they don’t want
to encrypt the entire server storage, just the data within the
database. Which encryption approach should the company adopt
to meet their objective?
**(A)** Full-disk Encryption
**(B)** File-level Encryption
**(C)** Volume-level Encryption
**(D) Database-level Encryption**

**Explanation 69. Correct Answer: D. Database-level
Encryption.** Database-level Encryption provides encryption


specifically for data within a database. It ensures that the data
remains encrypted even when backed up, replicated, or moved.
This approach is ideal for the e-commerce company as it
focuses on encrypting the sensitive transaction records without
affecting other data on the server.

**Option A is incorrect.** Full-disk Encryption would encrypt the
entire server’s storage, which may not be required by the
company.

**Option B is incorrect.** File-level Encryption would require
encrypting individual files, which may not be efficient for a
database with millions of records.

**Option C is incorrect.** Volume-level Encryption encrypts
specific volumes or logical drives, not just the database data.

**Question 70.** Your organization plans to upgrade its database
system. To maintain security during this process, which of the
following actions should be RESTRICTED until the upgrade is
validated?
**(A)** Monitoring the database for any anomalies
**(B) Allowing end-users to access the upgraded database
(C)** Making regular backups of the database
**(D)** Reviewing the database system logs

**Explanation 70. Correct Answer: B. Allowing end-users to
access the upgraded database.** Until the upgraded system is
validated and any potential issues are addressed, end-user
access should be restricted. This ensures that any vulnerabilities


or problems introduced by the upgrade don’t compromise data
or allow unauthorized activities.

**Option A is incorrect.** Monitoring the database is crucial to
identify any potential security issues and should not be
restricted.

**Option C is incorrect.** Regular backups should continue, as
they are part of a comprehensive disaster recovery and data
protection strategy.

**Option D is incorrect.** Reviewing logs is essential to monitor
the system’s health and security; hence, it should not be
restricted.

**Question 71.** A journalist wants to send a confidential message
to her editor without raising suspicion. Instead of sending a
coded or encrypted text, she embeds the message within a
harmless-looking photograph. What method is she employing to
keep the message concealed?
**(A)** Digital signature
**(B)** Tunneling
**(C) Steganography
(D)** Chaining

**Explanation 71. Correct Answer: C. Steganography.**
Steganography is a technique used to conceal data within
another piece of data. In this scenario, the journalist is
embedding a confidential message within a photograph, making
it look harmless and unsuspicious.


**Option A is incorrect.** A digital signature is used to verify the
authenticity and integrity of a message or document. It doesn’t
hide information within another piece of data.

**Option B is incorrect.** Tunneling is a method used to
encapsulate one protocol within another, typically used in VPNs
to transport data over a public network.

**Option D is incorrect.** Chaining in the context of cryptography
often refers to modes of operation like Cipher Block Chaining
(CBC). It doesn’t involve hiding data within other data.

**Question 72.** A security administrator needs to apply a
configuration change to a critical service, requiring a service
restart. Before initiating the restart, which of the following steps
is MOST important to ensure continuous service availability?
**(A) Implement automatic service restart on failure
(B)** Announce the restart to all company employees
**(C)** Schedule the restart during off-peak hours
**(D)** Take a backup of the current service configuration

**Explanation 72. Correct Answer: A. Implement automatic
service restart on failure.** Having an automatic service restart
on failure ensures that if any issues arise after applying the
configuration change, the service will attempt to restart itself,
ensuring minimal interruption to its availability.

**Option B is incorrect.** While notifying company employees is
good practice, it doesn’t directly ensure continuous service
availability.


**Option C is incorrect.** Scheduling during off-peak hours
minimizes impact but doesn’t ensure the service will be
available if issues arise post-restart.

**Option D is incorrect.** While taking a backup of the
configuration is crucial for rollback purposes, it doesn’t ensure
the service will remain available immediately post-restart.

**Question 73.** A security analyst at DataCorp is tasked with
preventing unauthorized external applications from connecting
to their server. Which approach should the analyst primarily rely
on to achieve this?
**(A) Implement an allow list for approved applications
(B)** Monitor server CPU usage
**(C)** Regularly patch server software
**(D)** Encrypt data at rest on the server

**Explanation 73. Correct Answer: A. Implement an allow list
for approved applications.** By implementing an allow list, the
analyst can specify which applications are authorized to connect
to the server. Any application not on the list will be prevented
from establishing a connection, effectively stopping
unauthorized external applications.

**Option B is incorrect.** While monitoring server CPU usage can
provide insights into the server’s performance and potential
anomalies, it doesn’t prevent unauthorized applications from
connecting.

**Option C is incorrect.** Regularly patching server software is a
best practice for security to fix known vulnerabilities. However,


patching doesn’t directly prevent specific external applications
from connecting.

**Option D is incorrect.** Encrypting data at rest helps protect
stored data from unauthorized access but does not regulate
which applications can connect to the server.

**Question 74.** Alice needs to provide proof of the authenticity of
a digital document she’s sending to Bob. Which of the
following cryptographic elements should Alice use to
accomplish this task and ensure Bob knows the document came
from her?
**(A)** Encrypt the document with Bob's private key
**(B)** Encrypt the document with her public key
**(C) Sign the document with her private key
(D)** Sign the document with Bob's public key

**Explanation 74. Correct Answer: C. Sign the document with
her private key.** To prove authenticity, a digital signature is
created using the sender’s private key. When Bob receives the
document, he can verify the signature using Alice’s public key.
This proves that the document was signed by Alice and has not
been tampered with during transit.

**Option A is incorrect.** Bob’s private key is known only to Bob
and should never be used by anyone else, including for
encryption.

**Option B is incorrect.** Encrypting with Alice’s public key
doesn’t prove authenticity. It would also mean only Alice’s
private key could decrypt it, which isn’t the intent.


**Option D is incorrect.** One does not sign documents with the
recipient’s public key. Signatures are created using the sender’s
private key.

**Question 75.** Carla, a security analyst, receives an alert that one
of the company’s server certificates may have been exposed in a
recent data breach. What is the most immediate action Carla
should take to ensure that the exposed certificate cannot be used
maliciously?
**(A)** Request a new certificate from the CA
**(B)** Update the company firewall rules
**(C) Add the certificate to the Certificate revocation list
(CRL)
(D)** Perform a vulnerability assessment on the server

**Explanation 75. Correct Answer: C. Add the certificate to
the Certificate revocation list (CRL).** If a certificate is
believed to be compromised, the most immediate action is to
revoke it. This is done by adding the certificate to the Certificate
revocation list (CRL). Systems and applications that check the
CRL before establishing secure communications will then know
not to trust the compromised certificate.

**Option A is incorrect.** While requesting a new certificate may
be necessary after revoking the compromised one, the
immediate action should be revoking the potentially
compromised certificate.

**Option B is incorrect.** Updating firewall rules, while essential
for many security scenarios, doesn’t directly address the misuse
of a potentially exposed certificate.


**Option D is incorrect.** A vulnerability assessment is a broader
action to identify weaknesses in the system. While valuable, it
doesn’t directly address the issue of the compromised
certificate.

**Question 76.** A database administrator is concerned about
identical hashes being produced for users who select the same
password. To mitigate this risk, what cryptographic technique
should the administrator implement?
**(A)** Digital signature
**(B) Salting
(C)** Key stretching
**(D)** Symmetric encryption

**Explanation 76. Correct Answer: B. Salting.** Salting involves
adding a random value to a password before hashing it. This
ensures that even if two users have the same password, their
hashes will be different because of the unique salts. This makes
it difficult for attackers to use precomputed tables (like rainbow
tables) to match hashes to possible plaintext passwords.

**Option A is incorrect.** Digital signatures are primarily used to
ensure the authenticity and integrity of a message or data, not
for hashing passwords.

**Option C is incorrect.** Key stretching involves repeating the
hashing process multiple times to make brute-force attacks
more challenging, but it doesn’t address the problem of identical
hashes for identical passwords.


**Option D is incorrect.** Symmetric encryption uses the same
key for both encryption and decryption and isn’t related to the
scenario of producing unique hashes for passwords.

**Question 77.** An online retailer is considering various methods
to protect its customers’ credit card information. Instead of
storing the actual credit card numbers in their database, they opt
for a solution that replaces the numbers with unrelated, random
values. What is this method called?
**(A)** Symmetric encryption
**(B)** Digital watermarking
**(C)** Hashing
**(D) Tokenization**

**Explanation 77. Correct Answer: D. Tokenization.**
Tokenization is a method where sensitive data is replaced with
non-sensitive substitutes, referred to as “tokens”. These tokens
act as references to the original data but don’t contain the actual
sensitive data, making it a preferred method for protecting
credit card information in many retail environments.

**Option A is incorrect.** Symmetric encryption is a method of
encrypting data using a single key for both encryption and
decryption. It changes the original data into a ciphered format
but doesn’t replace it with random values as tokenization does.

**Option B is incorrect.** Digital watermarking embeds data into a
digital signal, primarily for asserting rights or ownership and
not for replacing sensitive data with random values.


**Option C is incorrect. Hashing converts input data into a
fixed-length string of characters, which is typically a hash
code. It doesn’t produce a random value that can be used as
a reference back to the original data.**

**Question 78.** During a scheduled maintenance window, a
security administrator plans to apply a critical update to the
company’s firewall. Which of the following actions is MOST
crucial to ensure minimized downtime during this process?
**(A)** Notifying the firewall vendor about the update
**(B)** Disabling all firewall rules temporarily
**(C) Creating a rollback plan in case of update failure
(D)** Scheduling the update during peak business hours

**Explanation 78. Correct Answer: C. Creating a rollback
plan in case of update failure.** In change management
processes, having a rollback plan ensures that if there are issues
with the applied update, the system can be reverted to its
previous state, thereby minimizing downtime.

**Option A is incorrect.** While it might be useful to notify the
firewall vendor, it is not the most crucial step to minimize
downtime.

**Option B is incorrect.** Disabling all firewall rules can introduce
significant security risks and might not be related directly to the
downtime.

**Option D is incorrect.** Scheduling updates during peak
business hours could result in maximum disruption and
downtime.


**Question 79.** A security administrator is considering a
cryptographic solution for protecting data in transit between two
servers located in the same data center. The primary goal is to
ensure speed and efficiency in encryption and decryption
processes. Which type of encryption would best meet this
requirement?
**(A)** Asymmetric encryption using RSA
**(B) Symmetric encryption using AES
(C)** Hybrid encryption using a combination of RSA and
AES
**(D)** Asymmetric encryption using ECC

**Explanation 79. Correct Answer: B. Symmetric encryption
using AES.** Symmetric encryption, such as AES, is typically
faster and requires less computational resources than
asymmetric encryption. This makes it suitable for scenarios
where high-speed encryption and decryption are essential, like
for data in transit between servers in a data center.

**Option A is incorrect.** Asymmetric encryption using RSA is
more computationally intensive than symmetric encryption and
may not be the most efficient for the given scenario.

**Option C is incorrect.** While hybrid encryption can provide a
balance of security and speed, using only symmetric encryption
(AES) is more efficient for the described use case.

**Option D is incorrect.** ECC, like RSA, is an asymmetric
encryption method, which means it will typically be slower than
symmetric methods like AES.


**Question 80.** A software developer wants to store user
passwords in a way that even if the database is compromised,
attackers would not be able to retrieve the original passwords.
What technique should the developer use to achieve this?
**(A)** Symmetric encryption
**(B)** Digital signing
**(C) Hashing
(D)** Steganography

**Explanation 80. Correct Answer: C. Hashing.** Hashing is a
technique that takes an input (or ‘message’) and returns a fixed-
size string of bytes, usually in the form of a digest. The output
should ideally be unique (within reason) for every different
input. It’s often used for storing passwords because even a tiny
change in input will produce a dramatically different output, and
it’s computationally hard to reverse the process.

**Option A is incorrect.** Symmetric encryption uses the same key
for both encryption and decryption. If an attacker gains access
to the encryption key, they can decrypt any encrypted data.

**Option B is incorrect.** Digital signing is used to verify the
integrity and authenticity of a message but does not hide the
original content.

**Option D is incorrect.** Steganography involves hiding
information within other information, which isn’t related to
securely storing passwords.

**Question 81.** A software development company is working on a
mobile banking application. They want to ensure that sensitive


operations like cryptographic processes and biometric data
validation are isolated from the main operating system to
prevent potential tampering. Which tool should they consider
implementing to achieve this objective?
**(A)** Hardware Security Module (HSM)
**(B)** Key Management System (KMS)
**(C) Secure enclave
(D)** Trusted Platform Module (TPM)

**Explanation 81. Correct Answer: C. Secure enclave.** A
secure enclave is a hardware-based isolated environment often
found in modern CPUs, especially in mobile devices. It’s
designed to run specific operations such as cryptographic
processes and biometric data validation, ensuring that this data
is kept isolated from the main operating system, making it
resilient to tampering.

**Option A is incorrect.** While an HSM is a dedicated hardware
device that manages cryptographic keys, it doesn’t provide the
isolated environment within a CPU for operations like the
secure enclave does.

**Option B is incorrect.** KMS is primarily focused on the
management of cryptographic keys, not on creating isolated
environments for operations.

**Option D is incorrect.** TPM provides hardware-level security
for individual devices, but it doesn’t create a segregated
environment within the CPU for operations like the secure
enclave.


**Question 82.** A web server hosting the company’s e-commerce
site is set for an OS upgrade. The upgrade is expected to last 30
minutes. What should be a primary consideration to minimize
customer impact due to potential downtime?
**(A) Implementing a load balancer
(B)** Taking a backup of the e-commerce site
**(C)** Posting a maintenance notice a week in advance
**(D)** Upgrading the server's hardware

**Explanation 82. Correct Answer: A. Implementing a load
balancer.** A load balancer can redirect traffic to other servers
while one is undergoing maintenance, ensuring that customers
can still access the e-commerce site and minimizing the impact
of downtime. Load balancers distribute incoming traffic across
multiple servers, allowing one server to be taken offline without
affecting the availability of the service.

**Option B is incorrect.** While backups are crucial for disaster
recovery, they don’t minimize immediate downtime during
upgrades.

**Option C is incorrect.** While informing customers is a good
practice, it doesn’t prevent downtime. Some customers may still
try to access the site during maintenance.

**Option D is incorrect.** Upgrading the server’s hardware might
improve performance but doesn’t directly minimize the
downtime caused by an OS upgrade.

**Question 83.** A project manager is working on a new product
launch and has documents with sensitive financial projections


on her local computer. She occasionally shares these documents
with select board members via email. While she wants to keep
the financial documents secure, she doesn’t want to encrypt all
the data on her computer. Which encryption approach should
she utilize?
**(A)** Full-disk Encryption
**(B)** Transport-layer Encryption
**(C) File-level Encryption
(D)** Partition Encryption

**Explanation 83. Correct Answer: C. File-level Encryption.**
File-level Encryption allows individual files or folders to be
encrypted. In this scenario, the project manager can encrypt
only the sensitive financial documents, allowing her to securely
share them while keeping the rest of her data unencrypted.

**Option A is incorrect.** Full-disk Encryption would encrypt the
entire drive, which is more than what’s required.

**Option B is incorrect.** Transport-layer Encryption protects data
in transit, but does not specifically address encrypting
individual files for storage and sharing.

**Option D is incorrect.** Partition Encryption encrypts entire
partitions or volumes, which isn’t necessary in this scenario.

**Question 84.** A security analyst is evaluating security
enhancements for a series of laptops that will store highly
confidential data. The analyst wants to ensure that stored data
remains encrypted and the integrity of the boot process is
maintained. Which of the following would BEST meet this


requirement?
**(A)** Installing antivirus software on each laptop
**(B)** Enabling a software-based full-disk encryption
**(C)** Implementing a BIOS password
**(D) Utilizing a Trusted Platform Module (TPM)**

**Explanation 84. Correct Answer: D. Utilizing a Trusted
Platform Module (TPM).** A Trusted Platform Module (TPM)
is a specialized chip on an endpoint device that stores RSA
encryption keys specific to the host system. It provides
hardware-based security to enhance the security of the device
by enabling features like hardware-based encryption and
ensuring the integrity of the boot process, among other things.

**Option A is incorrect.** While antivirus software is vital for
protecting against malware, it does not directly address
hardware-based encryption or boot process integrity.

**Option B is incorrect.** Software-based full-disk encryption can
ensure the confidentiality of data, but it does not offer
hardware-level protection or boot process integrity like a TPM.

**Option C is incorrect.** A BIOS password provides a layer of
security, but it does not offer encryption for stored data or
ensure boot process integrity.

**Question 85.** A large e-commerce company is deploying a new
online payment system. The Chief Information Security Officer
(CISO) is concerned about the security of cryptographic keys
and wants to ensure they are protected from potential theft or
compromise. Which tool should the CISO implement to provide


the HIGHEST level of security for these keys?
**(A)** Password vault
**(B)** Software-based key storage
**(C) Hardware Security Module (HSM)
(D)** Cloud-based encryption service

**Explanation 85. Correct Answer: C. Hardware Security
Module (HSM).** A Hardware Security Module (HSM) is a
specialized device specifically designed to manage, protect, and
securely store cryptographic keys. It is built to be tamper-
resistant and provides a high level of security, making it suitable
for environments where the protection of cryptographic keys is
of paramount importance, such as in an e-commerce payment
system.

**Option A is incorrect.** A password vault is designed primarily
for storing and managing passwords, not cryptographic keys
used in payment systems.

**Option B is incorrect.** Software-based key storage solutions do
not provide the same level of physical security that an HSM
offers.

**Option D is incorrect.** While cloud-based encryption services
can provide encryption capabilities, they might not offer the
same level of physical protection and control as an on-premises
HSM.

**Question 86.** Sarah, a security analyst, is concerned about
potential man-in-the-middle attacks on the company’s internal
portal. To mitigate this risk, she recommends obtaining a digital


certificate from a trusted entity. Which of the following is
responsible for issuing such certificates?
**(A)** Key distribution center
**(B) Certificate authority (CA)
(C)** Tokenization system
**(D)** Security incident event manager

**Explanation 86. Correct Answer: B. Certificate authority
(CA).** Certificate authorities (CAs) are trusted entities
responsible for issuing, validating, and revoking digital
certificates. These certificates are used to authenticate entities
on the internet, such as websites, ensuring secure and
authenticated communications, thus mitigating the risk of man-
in-the-middle attacks.

**Option A is incorrect.** A Key distribution center (KDC) is a
part of the Kerberos authentication protocol and is responsible
for distributing session tickets and temporary session keys, not
for issuing digital certificates.

**Option C is incorrect.** Tokenization systems replace sensitive
data with non-sensitive substitutes, known as tokens. They don’t
issue digital certificates.

**Option D is incorrect.** A Security incident event manager
(SIEM) aggregates and analyzes log data from various sources,
providing real-time analysis of security alerts, but does not issue
certificates.

**Question 87.** A financial institution is looking to adopt an
encryption algorithm for its transactions that is considered to be


very secure due to its longer key length, compared to older
standards. Which encryption algorithm best fits this description?
**(A)** DES
**(B)** Blowfish
**(C)** RSA
**(D) AES-256**

**Explanation 87. Correct Answer: D. AES-256.** AES-256, part
of the Advanced Encryption Standard (AES) family, utilizes a
256-bit key length. This extended key length provides a higher
degree of security and is considered resistant to all known
practical attacks when used properly.

**Option A is incorrect.** DES (Data Encryption Standard) has a
key length of only 56 bits and is considered to be insecure
against sufficiently equipped attackers due to its shorter key
length.

**Option B is incorrect.** While Blowfish is a symmetric
encryption algorithm, it’s older and not as universally
recommended for secure transactions as AES.

**Option C is incorrect.** RSA is an asymmetric encryption
algorithm, not typically used directly for encrypting bulk
transaction data.

**Question 88.** Alice receives an email from Bob with an attached
document. She wants to verify both the authenticity of the
sender and the integrity of the attached document. Which of the
following should Bob have used before sending the email?
**(A)** Encrypt the document with his private key


**(B)** Hash the document
**(C)** Encrypt the document with Alice's public key
**(D) Sign the document with his private key**

**Explanation 88. Correct Answer: D. Sign the document with
his private key.** Digital signatures are created by taking a hash
of a message (or document) and then encrypting that hash with
the sender’s private key. When Alice receives the email, she can
decrypt the signature using Bob’s public key to retrieve the
original hash and then compare it with her computed hash of the
document. If they match, it confirms both the sender’s identity
(authenticity) and that the document has not been altered
(integrity).

**Option A is incorrect.** Encrypting the entire document with his
private key isn’t practical for verifying authenticity and
integrity. Instead, the hash of the document is encrypted to
create a signature.

**Option B is incorrect.** Simply hashing the document will
provide a way to check the document’s integrity but does not
verify the authenticity of the sender.

**Option C is incorrect.** Encrypting the document with Alice’s
public key would make it confidential for Alice, but this doesn’t
help in verifying authenticity or integrity.

**Question 89.** During a critical financial quarter, GlobalFin Corp
experienced unexpected outages during peak business hours due
to system maintenance, impacting its operations significantly.
To prevent such occurrences in the future, what should


GlobalFin Corp implement regarding their maintenance
activities?
**(A)** Conduct maintenance activities randomly to avoid
predictability
**(B)** Implement maintenance activities during peak business
hours
**(C) Establish designated maintenance windows
(D)** Reduce the frequency of maintenance activities

**Explanation 89. Correct Answer: C. Establish designated
maintenance windows.** Maintenance windows are specific time
frames designated for system maintenance, ensuring that
disruptions due to updates, patches, or other maintenance
activities don’t occur during critical business hours. By setting
these windows, usually during off-peak times, businesses can
minimize operational disruptions.

**Option A is incorrect.** Conducting maintenance activities
randomly can lead to unpredictable outages, which can be
disruptive to business operations and degrade trust among
stakeholders.

**Option B is incorrect.** Implementing maintenance activities
during peak business hours is precisely what led to the
disruption in the scenario. This approach would likely cause
more operational problems, especially for businesses with
critical operations during these hours.

**Option D is incorrect.** Reducing the frequency of maintenance
activities might decrease disruptions, but it could also lead to


unpatched vulnerabilities, outdated software, or other security
and operational issues.

**Question 90.** A financial institution wants to securely transfer
transaction data between its main office and a branch office.
The data should be encrypted while in transit to prevent any
interception and unauthorized access. Which encryption
solution is most suitable for securing the data during transport?
**(A)** Database-level Encryption
**(B)** Full-disk Encryption
**(C) Transport-layer Encryption
(D)** File-level Encryption

**Explanation 90. Correct Answer: C. Transport-layer
Encryption.** Transport-layer Encryption is specifically designed
to protect data while it is in transit over a network. It ensures
that the data remains confidential and is not tampered with
during transmission. For the financial institution, this approach
would be most effective in securing the transaction data
between offices.

**Option A is incorrect.** Database-level Encryption is used to
secure data stored within a database, not for data in transit.

O **ption B is incorrect.** Full-disk Encryption secures the entire
storage of a device and is not specific to data being transferred
over a network.

**Option D is incorrect.** File-level Encryption encrypts
individual files but may not ensure the confidentiality of the
data while it’s being transmitted over a network.


**Question 91.** After a recent software update, a company’s
intranet portal has been inaccessible to a few employees. The IT
team suspects it could be due to network filtering rules. What
should the IT team review to confirm their suspicions?
**(A)** The content filtering policies
**(B)** The malware detection logs
**(C) The allow list/deny list configurations
(D)** The network bandwidth utilization graphs

**Explanation 91. Correct Answer: C. The allow list/deny list
configurations.** Network accessibility issues, especially after
software or configuration changes, can often arise due to
misconfigured allow lists or deny lists. Reviewing these
configurations can help determine if specific IP addresses or
domains have been incorrectly blocked or not allowed, causing
the inaccessibility issues.

**Option A is incorrect.** Content filtering policies primarily focus
on blocking specific types of content (like social media or adult
sites) rather than causing inaccessibility to specific users or
departments.

**Option B is incorrect.** Malware detection logs track potential
security threats and not network access configurations. They
wouldn’t typically cause a selective inaccessibility issue unless
a specific user’s machine is quarantined.

**Option D is incorrect.** While network bandwidth utilization
graphs might show reduced traffic, they won’t provide details
on specific allow/deny list configurations that might be causing
the inaccessibility.


**Question 92.** A user wants to send a confidential email to their
colleague and ensure that only the intended recipient can read it.
The user also wants to provide assurance to the recipient that
the email was indeed sent by them. Which encryption method
should the user employ to accomplish this?
**(A)** Use symmetric encryption with a shared key
**(B)** Use asymmetric encryption and encrypt the email with
the recipient's public key
**(C)** Use asymmetric encryption, encrypt the email with the
user's private key
**(D) Use asymmetric encryption, first sign the email with
the user's private key, then encrypt it with the recipient's
public key**

**Explanation 92. Correct Answer: D. Use asymmetric
encryption, first sign the email with the user’s private key,
then encrypt it with the recipient’s public key.**

**Option D offers both confidentiality and non-repudiation.**
The email is encrypted with the recipient’s public key, ensuring
only the recipient can decrypt it using their private key. Signing
the email with the sender’s private key allows the recipient to
verify the sender using the sender’s public key.

**Option A is incorrect.** While symmetric encryption provides
confidentiality, it doesn’t offer non-repudiation or sender
verification.

**Option B is incorrect.** Encrypting with the recipient’s public
key provides confidentiality but lacks sender verification.


**Option C is incorrect.** Encrypting an email with the user’s
private key would offer sender verification but won’t provide
confidentiality.

**Question 93.** A user, Amy, wants to securely send a confidential
document to her colleague, Bob. Amy decides to encrypt the
document to ensure its confidentiality. Which of the following
should Amy use to encrypt the document, ensuring only Bob
can decrypt it?
**(A)** Amy's private key
**(B)** Amy's public key
**(C)** Bob's private key
**(D) Bob's public key**

**Explanation 93. Correct Answer: D. Bob’s public key.** In
asymmetric encryption, if a message is encrypted with an
individual’s public key, only the corresponding private key can
decrypt it. Therefore, to ensure Bob is the only person who can
decrypt the document, Amy should encrypt it using Bob’s
public key.

**Option A is incorrect.** Encrypting with Amy’s private key
would allow anyone with Amy’s public key to decrypt it, and it
would also serve as a digital signature rather than ensuring
confidentiality.

**Option B is incorrect.** Using Amy’s public key would not
make sense because then only Amy’s private key could decrypt
it.


**Option C is incorrect.** The private key should never be shared
or used for encryption. Its main use is for decryption and
signing.

**Question 94.** A cybersecurity analyst is investigating a
suspicious image file received via email. Upon closer
examination, the analyst suspects that the image might be
carrying hidden data because the file size is unusually large.
Which technique might the sender have used to embed secret
information within the image?
**(A)** Symmetric encryption
**(B)** Digital watermarking
**(C) Steganography
(D)** Hashing

**Explanation 94. Correct Answer: C. Steganography.**
Steganography is the practice of hiding information within
another form of data. In this case, the analyst suspects that an
image file is carrying hidden data due to its unusually large size,
which is a common indicator of steganographic practices.

**Option A is incorrect.** Symmetric encryption is used for
encrypting data using a single key for both encryption and
decryption. It doesn’t hide data within other data.

**Option B is incorrect.** Digital watermarking embeds
information into a digital signal, but it’s generally used to assert
rights or ownership, not to hide data in the manner of
steganography.


**Option D is incorrect.** Hashing is the process of converting an
input into a fixed-length string of bytes, generally used to verify
data integrity.

**Question 95.** A company is preparing to roll out a new
infrastructure deployment for its internal network. They have a
server that will store both highly confidential customer
information and non-sensitive marketing material. The IT
department wants to ensure that only the confidential data is
encrypted, while the marketing data remains easily accessible.
Which level of encryption would be most suitable for this
scenario?
**(A)** File-level Encryption
**(B)** Full-disk Encryption
**(C) Partition Encryption
(D)** Transport-layer Encryption

**Explanation 95. Correct Answer: C. Partition Encryption.**
Partition Encryption allows specific partitions or volumes of a
storage drive to be encrypted. By encrypting only the partition
that contains confidential data, the company can ensure the
security of sensitive information while leaving other partitions,
such as the one with marketing material, unencrypted for easy
access.

**Option A is incorrect.** File-level Encryption would require
each confidential file to be encrypted individually, which could
be cumbersome.

**Option B is incorrect.** Full-disk Encryption would encrypt the
entire disk, including the non-sensitive marketing material.


**Option D is incorrect.** Transport-layer Encryption protects data
in transit, not data at rest on storage drives.

**Question 96.** Sarah, a cybersecurity analyst, receives a report
that a company laptop was stolen from an employee’s car. The
laptop contained sensitive financial data. Sarah checked the
company’s security configurations and found that the laptop
was equipped with full-disk encryption. How does this impact
the potential data breach situation?
**(A)** The data remains easily accessible, as only the boot
sector was encrypted
**(B) The data is protected, as the entire hard drive's
contents are encrypted
(C)** The data is partially encrypted, with only the user
directories protected
**(D)** The data is vulnerable since full-disk encryption only
applies when the laptop is connected to the company network

**Explanation 96. Correct Answer: B. The data is protected,
as the entire hard drive’s contents are encrypted.** Full-disk
Encryption (FDE) encrypts the entirety of a hard drive, ensuring
that all its contents, including system and user files, are
unreadable without the appropriate decryption key or
credentials. As such, even if the laptop is stolen, the data
remains secured unless the attacker has the decryption key.

**Option A is incorrect.** Full-disk Encryption does not encrypt
only the boot sector; it encrypts the entire disk.

**Option C is incorrect.** Full-disk Encryption doesn’t only
encrypt user directories; it encrypts the whole disk.


**Option D is incorrect.** Full-disk Encryption protects the data at
all times, irrespective of the laptop’s connection to a network.

**Question 97.** A university’s IT department provides access to its
student records for training purposes to new hires. To protect
student identities, they replace the real names and social
security numbers with fictitious ones while maintaining the
database’s original format. Which technique is the IT
department utilizing?
**(A)** Digital signing
**(B) Data masking
(C)** Steganography
**(D)** Data deduplication

**Explanation 97. Correct Answer: B. Data masking.** Data
masking protects the data subject’s data privacy by obscuring
specific data within a database, making the data unreadable and
unusable, especially in non-production environments. It is
commonly used for testing and development purposes.

**Option A is incorrect.** Digital signing involves using a digital
signature to prove the authenticity and integrity of data.

**Option C is incorrect.** Steganography involves hiding
information within other information, such as embedding text
within images, making it undetectable.

**Option D is incorrect.** Data deduplication is the process of
eliminating duplicate copies of repeating data to save storage
space.


**Question 98.** A company is looking for a cryptographic solution
that provides an immutable and transparent record of all
transactions in a distributed ledger system. Which of the
following would BEST meet this requirement?
**(A)** Symmetric key algorithm
**(B)** Public key infrastructure
**(C) Blockchain
(D)** Digital watermark

**Explanation 98. Correct Answer: C. Blockchain.** Blockchain
is a decentralized and distributed ledger technology that
provides an immutable record of transactions. Each block
contains a list of transactions and is linked to the previous
block, creating a chain. The transparency and immutability of
blockchain make it especially suitable for applications where an
irrefutable record is essential.

**Option A is incorrect.** Symmetric key algorithms are
encryption methods where the same key is used for both
encryption and decryption but don’t inherently provide an
immutable record of transactions.

**Option B is incorrect.** Public key infrastructure (PKI) is used
for digital certificates and keys distribution but doesn’t offer an
immutable record of transactions.

**Option D is incorrect.** Digital watermarking embeds
information into a digital signal, but it doesn’t provide an
immutable record of transactions.


**Question 99.** An IT manager is considering solutions to protect
data stored on the laptops provided to remote employees. The
primary concern is to ensure that the entire content of the
laptop’s storage drive is unreadable if a laptop is lost or stolen.
Which encryption level would best address this concern?
**(A)** File-level Encryption
**(B)** Transport-layer Encryption
**(C) Full-disk Encryption
(D)** Database-level Encryption

**Explanation 99. Correct Answer: C. Full-disk Encryption.**
Full-disk Encryption (FDE) encrypts the entire storage drive,
making all data on the drive unreadable without the correct
decryption key or credentials. This is especially useful for
portable devices like laptops, which are more vulnerable to
physical theft.

**Option A is incorrect.** File-level Encryption encrypts
individual files rather than the entire disk, so some data or
system files might remain unencrypted.

**Option B is incorrect.** Transport-layer Encryption protects data
in transit, not data at rest on storage drives.

**Option D is incorrect.** Database-level Encryption encrypts data
within a database and does not apply to other files or data
outside of that database.

**Question 100.** The finance department at a large firm still relies
on a legacy application for their quarterly reporting. This
application is known to have some security flaws, but due to its


critical nature, it cannot be easily replaced. How can the firm
BEST mitigate the risks associated with this application?
**(A)** Train the finance team about the latest cybersecurity
threats
**(B)** Run the legacy application on the latest hardware to
improve performance
**(C) Place the legacy application behind a web
application firewall (WAF)
(D)** Frequently change the passwords of users who have
access to the application

**Explanation 100. Correct Answer: C. Place the legacy
application behind a web application firewall (WAF).** By
placing the application behind a WAF, the firm can filter,
monitor, and block malicious HTTP traffic targeting the
application’s vulnerabilities, thereby offering a layer of
protection against potential security flaws in the legacy
application.

**Option A is incorrect.** While training is essential, it doesn’t
directly address the vulnerabilities in the legacy application.

**Option B is incorrect.** Using the latest hardware might improve
application performance but doesn’t mitigate the security risks
associated with its vulnerabilities.

**Option D is incorrect.** While frequent password changes can
enhance security, they don’t address the inherent vulnerabilities
in the legacy application.


**Question 101.** A multinational corporation is concerned about
the possibility of losing access to encrypted data due to the loss
or compromise of private keys. They’ve approached a third-
party organization for a solution. Which of the following is a
system that allows the third party to securely hold a copy of the
corporation’s cryptographic keys to ensure data recoverability?
**(A)** Public Key Repository
**(B)** Key Generation Center
**(C) Key Escrow
(D)** Key Renewal Service

**Explanation 101. Correct Answer: C. Key Escrow.** Key
escrow is a system in which cryptographic keys are securely
stored with a third party, so they can be retrieved in cases where
the original keys are lost or compromised. This ensures data
recoverability while maintaining security.

**Option A is incorrect.** A Public Key Repository is where public
keys are stored for retrieval, not for backup or recovery
purposes.

**Option B is incorrect.** Key Generation Center is responsible
for creating cryptographic keys, not storing them for recovery
purposes.

**Option D is incorrect.** Key Renewal Service deals with
replacing and updating cryptographic keys as they expire or
need refreshing, not storing them for recovery.

**Question 102.** A financial institution plans to provide access to
its database for third-party developers to create new


applications. However, they want to ensure that the developers
do not see the actual data but instead work with a disguised
version that retains the data’s original structure. What technique
is the financial institution considering?
**(A)** Tokenization
**(B) Data masking
(C)** Encryption
**(D)** Digital watermarking

**Explanation 102. Correct Answer: B. Data masking.** Data
masking is a technique that obscures specific data within a
database, making the data unreadable and unusable. The method
is often employed in non-production environments to protect
the data subject’s data privacy and data security.

**Option A is incorrect.** Tokenization replaces sensitive data with
random tokens, which act as references to the original data.

**Option C is incorrect.** Encryption converts readable data into
an unreadable format to protect its confidentiality. It requires a
key to return the data to its original form.

**Option D is incorrect.** Digital watermarking embeds data into a
digital signal, primarily for asserting rights or ownership.

**Question 103.** NexTech, a cloud-based software company,
recently faced a security breach due to inconsistent practices
among its system administrators. To avoid such inconsistencies
in the future, what should NexTech emphasize in its operations?
**(A)** Rely on system administrators to develop their personal
methods


**(B)** Mandate frequent system reboots
**(C) Implement Standard Operating Procedures (SOPs)
for all technical operations
(D)** Conduct random security audits without notifying
administrators

**Explanation 103. Correct Answer: C. Implement Standard
Operating Procedures (SOPs) for all technical operations.**
Standard Operating Procedures (SOPs) provide a consistent and
documented process that employees can follow. By
implementing SOPs, businesses ensure that tasks are performed
uniformly, reducing the risk of errors and inconsistencies that
might lead to security vulnerabilities.

**Option A is incorrect.** Relying on individual system
administrators to develop their personal methods can lead to
inconsistent practices and operational inefficiencies, increasing
the risk of security breaches.

**Option B is incorrect.** Frequent system reboots, while they
might be part of some SOPs, don’t address the root issue of
inconsistency among system administrators. It’s the consistent
method of operations that prevents errors, not just frequent
restarts.

**Option D is incorrect.** While security audits are essential,
conducting them without notice doesn’t address the core issue
of inconsistency in system administration practices.

**Question 104.** After a series of system enhancements, a
financial organization decided to use a manual method of


documenting changes in separate files rather than implementing
a version control system. During an audit, the cybersecurity
team struggled to determine which version of a critical system
file was the most recent and accurate. What is the PRIMARY
risk of not implementing version control for such
documentation?
**(A)** Increased storage requirements for multiple files
**(B)** Difficulty in collaborating between team members
**(C) Lack of traceability and difficulty in reverting to a
known stable state
(D)** Greater need for training staff on manual documentation

**Explanation 104. Correct Answer: C. Lack of traceability
and difficulty in reverting to a known stable state.** Version
control provides a clear history of changes, ensuring easy
reversion to a known stable state, and identifying the latest
version of a document or system file. In the absence of version
control, identifying the most recent and accurate version can be
challenging.

**Option A is incorrect.** While storage might be a concern, the
primary risk is the inability to trace changes and revert to a
stable state.

**Option B is incorrect.** Collaboration might be hindered, but the
direct risk is associated with traceability and stability.

**Option D is incorrect.** Training staff is always essential, but
the immediate concern is the ability to trace and manage
changes.


**Question 105.** During a security audit, it was found that an
application was using plain hashes for storing passwords. The
security team recommended a method that involves using the
original password along with a salt and then rehashing it
multiple times. What is this method known as?
**(A)** Key clustering
**(B)** Rainbow table prevention
**(C)** Key rotation
**(D) Key stretching**

**Explanation 105. Correct Answer: D. Key stretching.** Key
stretching refers to the process of taking a password and,
usually in combination with a salt, hashing it multiple times.
This repeated hashing process makes brute-force attacks more
time-consuming and difficult because the attacker has to not
only guess the password but also apply the hashing function the
same number of times the original process used.

**Option A is incorrect.** Key clustering pertains to different keys
producing the same ciphertext from the same plaintext, which is
not relevant to the described scenario.

**Option B is incorrect.** While using salts can prevent the
effective use of rainbow tables, the act of rehashing passwords
multiple times is not specifically called “rainbow table
prevention.”

**Option C is incorrect.** Key rotation involves periodically
changing cryptographic keys. It does not relate to hashing
passwords multiple times.


**Question 106.** During a routine update, a web server
application requires a restart. What should the administrator do
FIRST to ensure client connections aren’t abruptly terminated
during the restart?
**(A) Redirect incoming traffic to a backup server
(B)** Increase the server's memory
**(C)** Manually terminate all active client sessions
**(D)** Check for available patches for the application

**Explanation 106. Correct Answer: A. Redirect incoming
traffic to a backup server.** Redirecting incoming traffic
ensures that clients are not abruptly disconnected and instead
can continue their activities on a backup server while the
primary server undergoes a restart.

**Option B is incorrect.** Increasing server memory might
improve performance but doesn’t address the immediate issue
of the service restart.

**Option C is incorrect.** Manually terminating client sessions
would cause abrupt disconnections, which is what the
administrator is trying to avoid.

**Option D is incorrect.** While checking for patches is important,
it doesn’t address the issue of ensuring client connections aren’t
terminated during a service restart.

**Question 107.** Carlos is responsible for managing IT services
for a university. The university has numerous departments, each
with its subdomain, like arts.university.com,
science.university.com, and sports.university.com. Carlos wants


a solution that ensures HTTPS security while being cost-
effective. However, he’s wary of potential risks. What might be
a drawback of using a Wildcard Certificate for the university’s
subdomains?
**(A)** It can secure only one subdomain
**(B) If compromised, all subdomains are at risk
(C)** It only validates the domain ownership, not the
organization's identity
**(D)** It's the most expensive certificate available

**Explanation 107. Correct Answer: B. If compromised, all
subdomains are at risk.** The primary concern with a Wildcard
Certificate is that if its private key is compromised, it
jeopardizes the security of all associated subdomains. This
poses a significant risk since the exposure of a single key could
lead to potential attacks or misuse across all subdomains.

**Option A is incorrect.** A Wildcard Certificate can secure
multiple subdomains under a single domain.

**Option C is incorrect.** While true for Domain Validated (DV)
certificates, Wildcard Certificates can also be available with
Organization Validation (OV), ensuring organizational identity.

**Option D is incorrect.** Wildcard Certificates aren’t necessarily
the most expensive. The cost varies based on the level of
validation and the issuing authority.

**Question 108.** Your organization is preparing to upgrade a
database server that supports an e-commerce application. A
review of the change management documentation has revealed


that multiple applications rely on this particular database server
for various functionalities. Which of the following steps should
be taken FIRST to ensure a smooth upgrade process without
disruptions?
**(A)** Upgrade the database server immediately to benefit
from new features
**(B)** Perform a backup of the database server
**(C) Identify and test all applications that have
dependencies on the database server
(D)** Inform users about potential downtime during the
upgrade

**Explanation 108. Correct Answer: C. Identify and test all
applications that have dependencies on the database server.**
Before making changes, especially to systems with multiple
dependencies, it’s crucial to understand the full scope of
potential impacts. By identifying and testing all dependent
applications, you ensure that the upgrade won’t inadvertently
disrupt other services or functionalities.

**Option A is incorrect.** Upgrading immediately without
considering dependencies can lead to unforeseen disruptions
and complications.

**Option B is incorrect.** While backing up the server is a good
practice, understanding dependencies should come first to plan
the upgrade effectively.

**Option D is incorrect.** Informing users is essential, but
understanding the upgrade’s potential impact should come first
to provide accurate information.


**Question 109.** After a recent data breach, a multinational
corporation is evaluating its cryptographic practices. The Chief
Security Officer (CSO) determines that the manual management
of cryptographic keys has become too complex due to the scale
of the operations. Which tool would BEST address the CSO’s
concern while ensuring robust security practices?
**(A)** Password Management System
**(B)** Secure File Transfer Protocol (SFTP)
**(C)** Trusted Platform Module (TPM)
**(D) Key Management System (KMS)**

**Explanation 109. Correct Answer: D. Key Management
System (KMS).** A Key Management System (KMS) is
specifically designed to handle the generation, distribution,
rotation, and retirement of cryptographic keys in a centralized
and automated manner. For large organizations, using a KMS
streamlines and secures the complex task of key management.

**Option A is incorrect.** While a Password Management System
helps in handling and storing passwords, it does not provide
comprehensive features needed for cryptographic key
management.

**Option B is incorrect.** SFTP is a protocol for securely
transferring files over a network, not for managing
cryptographic keys.

**Option C is incorrect.** While TPM provides hardware-level
security for individual devices, it is not meant for enterprise-
wide key management.


**Question 110.** During a quarterly review, the IT team at a
logistics company decided to change the configuration of their
load balancers to better distribute traffic among their servers.
After the change, a series of technical issues emerged, affecting
customer-facing applications. When troubleshooting the issue, it
was discovered that the network diagrams had not been updated
to reflect the new changes. What is the MAJOR consequence of
not having updated diagrams in such a scenario?
**(A)** The servers might need a hardware upgrade
**(B)** The company might need to revert to the old load
balancer configuration
**(C) It increases the time and complexity of
troubleshooting
(D)** Customers might prefer other logistics companies

**Explanation 110. Correct Answer: C. It increases the time
and complexity of troubleshooting.** Accurate and up-to-date
documentation, including network diagrams, is crucial for
effective troubleshooting. Without it, IT teams can spend
unnecessary time trying to understand the current state of the
system, delaying the resolution of the issue.

**Option A is incorrect.** While server upgrades might be
necessary in some cases, it’s not a direct consequence of
outdated diagrams.

**Option B is incorrect.** Reverting to an old configuration might
be a potential solution, but the primary issue is the increased
troubleshooting complexity due to outdated documentation.


**Option D is incorrect.** While the potential loss of customers
can be an indirect consequence of prolonged technical issues,
the immediate concern of outdated diagrams is the impact on
troubleshooting.