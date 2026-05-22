**Question 311.** A financial institution is shutting down one of its
data centers. Given the highly sensitive nature of the data
stored, the company wants to ensure that there is no possibility
of data retrieval from the storage devices. Which of the
following methods would be the MOST effective in
guaranteeing the destruction of data?
**(A)** Overwriting with zeros
**(B)** Standard Disk Format


**(C) Physical Destruction
(D)** Running a Disk Cleanup utility

**Explanation 311. Correct Answer: C. Physical Destruction.**
Physical destruction, such as shredding or incinerating the
storage devices, ensures complete destruction of the media,
making data retrieval impossible, regardless of the tools or
techniques employed.

**Option A is incorrect.** Overwriting with zeros is a method of
data sanitization, but advanced techniques might still retrieve
some portions of the data.

**Option B is incorrect.** A standard disk format often does not
remove data entirely, and traces can remain that could be
recovered with specialized software.

**Option D is incorrect.** Disk Cleanup utilities are designed to
free up space by removing temporary and unnecessary files but
not to securely erase sensitive data.

**Question 312.** A financial application allows users to transfer
money to other accounts by entering the account number and
the amount to transfer. During a security audit, it was observed
that malicious users could enter SQL code into the account
number field to manipulate the application’s database. Which
security technique should the development team implement to
address this vulnerability?
**(A)** Code obfuscation
**(B) Input validation**


**(C)** Encryption at rest
**(D)** Session timeout

**Explanation 312. Correct Answer: B. Input validation.** Input
validation ensures that only valid and expected data is accepted
by an application. By validating and sanitizing user input, the
application can prevent malicious data, such as SQL code, from
being processed and executed, thus mitigating the risk of SQL
injection attacks.

**Option A is incorrect.** Code obfuscation makes it more difficult
for attackers to understand the application’s code but does not
prevent them from injecting malicious input.

**Option C is incorrect.** Encryption at rest protects stored data,
but it doesn’t prevent SQL injection attacks which exploit the
application’s input handling.

**Option D is incorrect.** Session timeout limits the duration of a
user’s session, but it does not validate or sanitize user input.

**Question 313.** After a series of cyberattacks, BetaTech, a
financial institution, decided to standardize the configurations
across its entire server fleet. They’ve established a secure
baseline configuration for their servers. What should be the
NEXT step in ensuring the servers conform to this new
baseline?
**(A)** Frequently conduct vulnerability scanning on all servers
**(B)** Introduce biometric authentication for server access
**(C) Deploy the secure baseline across all servers
(D)** Monitor network traffic to detect anomalies


**Explanation 313. Correct Answer: C. Deploy the secure
baseline across all servers.** After establishing a secure
baseline, the immediate next step is to deploy or apply this
baseline to the servers. This ensures that all servers start from a
standardized and secure configuration.

**Option A is incorrect.** While vulnerability scanning is crucial,
it is more of a continuous process to identify potential
vulnerabilities and doesn’t directly apply the secure baseline to
the servers.

**Option B is incorrect.** Biometric authentication, though a
valuable security measure, does not address the deployment of
the established secure baseline on servers.

**Option D is incorrect.** Monitoring network traffic is essential
for detecting potential threats, but it does not directly ensure the
deployment of the secure baseline on servers.

**Question 314.** An energy company is looking to enhance the
security of its ICS/SCADA systems. They have realized that
default configurations might have vulnerabilities. Which of the
following is the BEST initial step to take in securing their ICS/
SCADA systems?
**(A)** Connect the ICS/SCADA systems to the internet for
remote monitoring
**(B)** Use commercial off-the-shelf software to add a layer of
security
**(C) Implement a secure baseline configuration tailored
to the ICS/SCADA environment**


**(D)** Increase the number of users with administrative
privileges to ensure rapid response to issues

**Explanation 314. Correct Answer: C. Implement a secure
baseline configuration tailored to the ICS/SCADA
environment.** Industrial Control Systems (ICS) and
Supervisory Control and Data Acquisition (SCADA) systems
are critical in managing industrial processes. Implementing a
secure baseline configuration tailored to these systems ensures
that unnecessary services and potential vulnerabilities are
minimized.

**Option A is incorrect.** Connecting ICS/SCADA systems
directly to the internet exposes them to a wide range of threats
and potential cyber-attacks.

**Option B is incorrect.** While commercial off-the-shelf software
can be beneficial, it’s not the best initial step. Secure baselines
tailored to the ICS/SCADA environment are more effective.

**Option D is incorrect.** Increasing the number of users with
administrative privileges contradicts the principle of least
privilege and can introduce more security vulnerabilities.

**Question 315.** A multinational company is planning to issue
company-owned mobile devices to its executives. Given the
sensitivity of the data the executives handle, what hardening
measure would be MOST effective to ensure the security of
these mobile devices?
**(A)** Regularly updating the company's social media profiles
to mention the security measures taken


**(B) Implementing biometric authentication in addition
to strong passcodes
(C)** Turning off Bluetooth and Wi-Fi when not in use
**(D)** Setting the devices to display brighter screen colors

**Explanation 315. Correct Answer: B. Implementing
biometric authentication in addition to strong passcodes.**
Biometric authentication (like fingerprint or facial recognition)
combined with strong passcodes provides a robust multi-factor
authentication mechanism, making unauthorized access to the
device significantly more challenging.

**Option A is incorrect.** Publicly announcing security measures
on social media does not harden the device. Moreover,
divulging too much about security practices could expose the
organization to targeted attacks.

**Option C is incorrect.** While turning off Bluetooth and Wi-Fi
can reduce certain attack vectors, it’s not as comprehensive a
security measure as implementing strong multi-factor
authentication.

**Option D is incorrect.** Screen colors have no direct correlation
to the security hardening of a device.

**Question 316.** A large e-commerce company wants to ensure
that their newly developed application is free from any code
vulnerabilities before it is deployed to the production
environment. They want to catch any software flaws, especially
those that might lead to potential security risks. Which of the
following methodologies should they employ?


**(A)** Runtime application self-protection (RASP)
**(B)** Penetration testing on the live application
**(C) Static code analysis
(D)** User acceptance testing (UAT)

**Explanation 316. Correct Answer: C. Static code analysis.**
Static code analysis involves analyzing the application’s code
without executing it, focusing on detecting vulnerabilities,
coding flaws, or other potential security threats, making it an
appropriate methodology for the scenario presented.

**Option A is incorrect.** Runtime application self-protection
(RASP) is about real-time application monitoring and
protection, which would not be the primary choice for pre-
deployment vulnerability checks.

**Option B is incorrect.** Penetration testing on the live
application involves testing the application once it is live or
running, which might not be suitable when looking to identify
vulnerabilities in the pre-deployment phase.

**Option D is incorrect.** User acceptance testing (UAT) focuses
on determining if the software meets the business needs and
functions correctly for the end-users. It does not emphasize
finding security vulnerabilities.

**Question 317.** Jenny, the new CIO of a multinational firm,
wants to ensure that every software and hardware asset in the
organization has a clearly defined owner responsible for its
security and maintenance. Which of the following is the MOST
effective way to achieve this?


**(A)** Deploy an automated asset discovery tool and assign
assets to departments based on their location
**(B)** Mandate that every department head is the default
owner of all assets within their department
**(C)** Conduct regular audits and require individual users to
claim ownership of their assets
**(D) Introduce an Asset Management System where
assets are logged with defined ownership as they are
procured or assigned**

**Explanation 317. Correct Answer: D. Introduce an Asset
Management System where assets are logged with defined
ownership as they are procured or assigned.** An Asset
Management System provides a structured way to log, track,
and manage assets throughout their lifecycle, including defining
and recording asset ownership. This ensures clarity and
accountability regarding asset responsibility.

**Option A is incorrect.** While automated discovery tools are
helpful, assigning assets based solely on location does not
necessarily reflect the actual usage or responsibility for the
asset.

**Option B is incorrect.** While department heads should be
aware of assets within their departments, they may not always
be the best individuals to take ownership, especially for specific
or specialized assets.

**Option C is incorrect.** Regular audits are crucial, but relying
solely on users to claim ownership might not capture all assets,
and it’s a reactive rather than proactive approach.


**Question 318.** An audit report indicates that several network
switches in a data center lack security configurations, making
them potential targets for attackers. Which of the following
hardening techniques would BEST reduce the risk associated
with these switches?
**(A)** Configuring port mirroring to monitor network traffic
**(B) Disabling unused switch ports
(C)** Implementing load balancing across the switches
**(D)** Increasing the MAC address table size for performance

**Explanation 318. Correct Answer: B. Disabling unused
switch ports.** By disabling unused switch ports, you minimize
potential points of entry for unauthorized devices or attackers.
This is a basic step in network switch hardening.

**Option A is incorrect.** Port mirroring is used for monitoring
and isn’t directly a hardening technique. While monitoring is
crucial for security, it doesn’t necessarily prevent unauthorized
access.

**Option C is incorrect.** Load balancing is primarily about
distributing network traffic efficiently across multiple paths or
resources, and it doesn’t inherently harden the switches.

**Option D is incorrect.** Increasing the MAC address table size
might improve switch performance in certain scenarios, but it
does not harden the switch against potential security threats.

**Question 319.** DeltaSoft has released a new web application.
The security team is tasked with observing the application’s
behavior and responses when it is running to identify potential


vulnerabilities. Which method is most appropriate for this
purpose?
**(A)** Static Analysis
**(B)** Fuzz Testing
**(C)** Whitebox Testing
**(D) Dynamic Analysis**

**Explanation 319. Correct Answer: D. Dynamic Analysis.**
Dynamic analysis focuses on examining an application during
its execution, or runtime. By observing the application’s
behavior and responses, security teams can identify
vulnerabilities that might not be evident in the static code.

**Option A is incorrect.** Static analysis evaluates an application’s
code without executing the application. It does not focus on
observing the application’s behavior during runtime.

**Option B is incorrect.** Fuzz testing involves submitting random
inputs to an application to observe its behavior and detect
vulnerabilities, but it’s just one subset of dynamic analysis.

**Option C is incorrect.** Whitebox testing involves
understanding the internal structures and workings of an
application while testing, which can be part of static or dynamic
testing, but on its own, it does not specifically focus on runtime
behavior observation.

**Question 320.** A security analyst at ZetaTech is looking to
gather information about emerging threats and vulnerabilities
relevant to their industry. Which of the following would be the
MOST suitable method to obtain real-time, continuously


updated data on potential security issues?
**(A)** Relying solely on automated internal vulnerability
scanners
**(B)** Periodic manual penetration testing
**(C) Subscribing to an OSINT threat feed
(D)** Regularly checking the company's firewall logs

**Explanation 320. Correct Answer: C. Subscribing to an
OSINT threat feed.** An OSINT (Open-source intelligence)
threat feed provides real-time information and is continuously
updated with data on potential security issues sourced from
publicly available information.

**Option A is incorrect.** While automated internal vulnerability
scanners are valuable for identifying vulnerabilities within the
organization’s infrastructure, they don’t provide continuous
real-time data about external emerging threats.

**Option B is incorrect.** Manual penetration testing is periodic
and doesn’t offer continuous updates on emerging threats from
the broader industry.

**Option C is incorrect.** Regularly checking the company’s
firewall logs will give insight into potential malicious activities
targeting the company, but it won’t offer broad, updated data on
industry-wide threats.

**Question 321.** A medium-sized enterprise is preparing to
upgrade its office workstations. The IT department is
considering purchasing devices from a lesser-known, but
cheaper, vendor. Which of the following should be the


PRIMARY consideration before finalizing the acquisition?
**(A)** Whether the vendor offers a longer warranty period
**(B)** The aesthetics and design of the workstations
**(C) The vendor's adherence to industry security
standards and practices
(D)** The amount of training required for IT staff to support
the new devices

**Explanation 321. Correct Answer: C. The vendor’s
adherence to industry security standards and practices.**
Ensuring that the vendor adheres to industry security standards
and practices is critical for maintaining a secure environment.
Devices that don’t meet security standards could introduce
vulnerabilities into the organization.

**Option A is incorrect.** While a longer warranty period might be
attractive, it should not be the primary concern over security
considerations.

**Option B is incorrect.** Aesthetics and design, while potentially
important for branding or user satisfaction, are not as critical as
security considerations.

**Option D is incorrect.** Training the IT staff is important, but it
shouldn’t take precedence over ensuring the devices themselves
are secure.

**Question 322.** ClearView Industries wants to give their
employees the flexibility to choose their own devices for work
while retaining control over the device configurations and
applications. Which deployment model would be the MOST


appropriate for ClearView’s objectives?
**(A)** Bring Your Own Device (BYOD)
**(B) Choose Your Own Device (CYOD)
(C)** Corporate-owned, Personally Enabled (COPE)
**(D)** Fixed Device Deployment (FDD)

**Explanation 322. Correct Answer: B. Choose Your Own
Device (CYOD).** In the CYOD model, employees are allowed
to choose a device from a list of approved devices. The
organization retains control over the device configurations and
applications, ensuring a blend of user preference and corporate
security.

**Option A is incorrect.** BYOD allows employees to bring their
personal devices to work, which means the organization has less
control over the configurations and applications.

**Option C is incorrect.** While COPE provides devices owned
by the corporation and allows personal use, it does not
necessarily offer employees the choice of which device they get.

**Option D is incorrect.** Fixed Device Deployment (FDD) is not
a standard deployment model in the context of mobile solutions,
and it does not reflect the flexibility of device choice.

**Question 323.** A recently hired security analyst at CyberTech
Inc. wants to get a better understanding of the organization’s
network infrastructure. Which of the following activities would
provide a LIST of servers, workstations, printers, switches, and
routers currently active in the network?
**(A)** Vulnerability Scanning


**(B)** Intrusion Detection
**(C) Network Enumeration
(D)** Penetration Testing

**Explanation 323. Correct Answer: C. Network
Enumeration.** Network enumeration is the process of
identifying devices on a network. Through enumeration, an
analyst can gather information about network devices, their
types, and other attributes, thereby getting a clearer picture of
the network’s infrastructure.

**Option A is incorrect.** Vulnerability scanning aims to identify
vulnerabilities in network devices and systems, not necessarily
to list all active devices.

**Option B is incorrect.** Intrusion detection focuses on
monitoring network traffic for malicious activities and potential
security breaches.

**Option D is incorrect.** Penetration testing aims to exploit
vulnerabilities in a network to determine its security posture,
not to list all active devices.

**Question 324.** A network engineer is preparing a new batch of
routers for deployment in a large organization. Which of the
following steps should the engineer prioritize to ensure that the
routers are securely configured from the start?
**(A)** Configure the routers to use DHCP to dynamically
assign IP addresses to connected devices
**(B) Change the default administrative credentials on the
routers**


**(C)** Update the routers' firmware to the latest, most feature-
rich version, regardless of its security posture
**(D)** Customize the routers' LED colors for easy
identification in the server room

**Explanation 324. Correct Answer: B. Change the default
administrative credentials on the routers.** Routers often come
with default credentials that are widely known and can be easily
exploited by attackers. Changing these credentials is essential
for establishing a secure baseline.

**Option A is incorrect.** While DHCP is commonly used in many
networks, enabling it is not necessarily a secure baseline action.
Additionally, DHCP on routers can be exploited if not securely
configured.

**Option C is incorrect.** Blindly updating to the latest firmware
without considering its security posture can introduce
vulnerabilities. It’s important to ensure that the firmware is both
updated and secure.

**Option D is incorrect.** Customizing LED colors relates to
physical identification and does not inherently improve the
security configuration of a router.

**Question 325.** A university is looking to revamp its wireless
network to provide secure access for students and faculty. The
IT department wants an authentication method that leverages a
centralized server to validate user credentials and can integrate
with their existing directory service. Which authentication
protocol should they consider?


**(A)** Pre-shared Key (PSK)
**(B)** Lightweight Extensible Authentication Protocol (LEAP)
**(C) Extensible Authentication Protocol-Transport Layer
Security (EAP-TLS)
(D)** Shared Secret Challenge

**Explanation 325. Correct Answer: C. Extensible
Authentication Protocol-Transport Layer Security (EAP-
TLS).** EAP-TLS provides strong security by using digital
certificates to authenticate both the client and the server. It
operates in conjunction with a RADIUS server, allowing for
centralized management of user credentials and seamless
integration with existing directory services.

**Option A is incorrect.** Pre-shared Key (PSK) does not involve
a centralized server for authentication. Instead, a single shared
key is used for all devices, which may not be scalable or secure
for a large university setting.

**Option B is incorrect.** While LEAP provides centralized
authentication via a RADIUS server, it has known
vulnerabilities and may not offer the level of security the
university is looking for.

**Option D is incorrect.** Shared Secret Challenge is not a
standard wireless authentication protocol. In practice,
challenges based on shared secrets (like PSK) can be vulnerable
if the secret becomes known.

**Question 326.** TechHive Corp. is planning to sell some of its
old servers. Before selling, they want to ensure that no


retrievable personal or business data remains on the hard drives.
Which of the following methods should TechHive use to ensure
the drives are clean and the data cannot be recovered?
**(A)** Simple Format
**(B) Magnetic Wiping
(C)** Physical Destruction
**(D)** Standard Defragmentation

**Explanation 326. Correct Answer: B. Magnetic Wiping.**
Magnetic wiping, also known as degaussing, uses magnetic
fields to permanently erase data from storage devices. It makes
the data irretrievable, ensuring that personal and business data
cannot be recovered even with advanced tools.

**Option A is incorrect.** A simple format often leaves traces of
data, which can potentially be recovered using specialized
software.

**Option C is incorrect.** While physical destruction ensures data
cannot be retrieved, it would make the servers unusable for
resale purposes.

**Option D is incorrect.** Defragmentation is a process to
optimize and organize the data on a drive, but it does not erase
data.

**Question 327.** The network administrator of a rapidly growing
tech firm is concerned about the potential vulnerabilities of the
company’s switches. Which of the following measures is MOST
effective in hardening these network switches against possible
attacks?


**(A)** Assigning static IP addresses to all connected devices
**(B) Implementing strong password policies for switch
access
(C)** Upgrading the switches to support 10Gbps for future
expansion
**(D)** Customizing the switch LED colors for easy
identification

**Explanation 327. Correct Answer: B. Implementing strong
password policies for switch access.** Hardening switches
involves reducing their vulnerability to unauthorized access and
potential misuse. Implementing strong password policies
ensures that only authorized personnel can access and configure
the switches.

**Option A is incorrect.** While assigning static IP addresses
might help in network management, it does not inherently
harden a switch against potential attacks.

**Option C is incorrect.** Upgrading the speed capabilities of
switches, such as supporting 10Gbps, is a performance
enhancement and not directly related to hardening or security.

**Option D is incorrect.** Customizing LED colors is related to
physical identification and does not improve the security
posture of the switch.

**Question 328.** A robotics company is developing an
autonomous vehicle that relies on a Real-Time Operating
System (RTOS) to manage its operations. The development
team wants to ensure that the vehicle’s RTOS has a solid


security posture. What should the team prioritize when
establishing a secure baseline for this RTOS?
**(A)** Installing a robust antivirus software
**(B)** Enabling all features for maximum functionality
**(C)** Regularly backing up the RTOS data to the cloud
**(D) Minimizing the number of services and open ports**

**Explanation 328. Correct Answer: D. Minimizing the
number of services and open ports.** Reducing the number of
services and open ports reduces the attack surface on the RTOS,
making it more challenging for attackers to find vulnerabilities.

**Option A is incorrect.** While antivirus software is essential for
many systems, an RTOS, especially one for an autonomous
vehicle, would prioritize reducing the attack surface and
ensuring real-time performance rather than relying on
traditional antivirus solutions.

**Option B is incorrect.** Enabling all features may introduce
unnecessary vulnerabilities. It’s best to enable only required
functionalities to maintain security.

**Option C is incorrect.** While backups are crucial for data
integrity and recovery, the primary concern for an RTOS,
especially in autonomous vehicles, would be real-time
performance and reducing potential vulnerabilities.

**Question 329.** After deploying wireless access points in a large
manufacturing facility, employees report inconsistent wireless
connectivity in some areas. What tool would be most effective
for the IT team to use to visualize areas of weak wireless signal


strength?
**(A)** Network bandwidth monitor
**(B)** Protocol analyzer
**(C) Heat map software
(D)** Intrusion detection system

**Explanation 329. Correct Answer: C. Heat map software.**
Heat map software allows IT professionals to visually see areas
of strong and weak wireless signal strength, making it easier to
adjust placements or add additional access points as needed.

**Option A is incorrect.** A network bandwidth monitor is used to
measure the amount of data being sent over a network, not to
visualize wireless signal strength.

**Option B is incorrect.** A protocol analyzer is used to capture
and analyze network traffic, not to visually display wireless
coverage.

**Option D is incorrect.** An intrusion detection system (IDS) is
designed to detect unauthorized access or malicious activities
on a network. It does not show areas of weak wireless signal
strength.

**Question 330.** Sarah, an end-user, downloads a software update
from a website. Before installing, she wants to make sure the
software hasn’t been modified maliciously and that it originates
from a trusted source. What should Sarah check to validate this?
**(A)** The SSL certificate of the website
**(B) The application's code signing certificate**


**(C)** The application's open-source repositories
**(D)** The software's user reviews

**Explanation 330. Correct Answer: B. The application’s code
signing certificate.** By checking the application’s code signing
certificate, Sarah can validate that the software was indeed
issued by a trusted entity and hasn’t been altered since it was
signed. A valid code signing certificate gives users confidence
in the authenticity and integrity of the software.

**Option A is incorrect.** While the SSL certificate of a website
ensures secure data transmission between the server and the
browser, it doesn’t guarantee the integrity or authenticity of the
software being downloaded.

**Option C is incorrect.** Open-source repositories may contain
the source code of a software, but checking these repositories
doesn’t necessarily validate the integrity of the compiled
software version Sarah has downloaded.

**Option D is incorrect.** User reviews can provide insights into
the software’s functionality and user experience, but they cannot
be relied upon to confirm software authenticity or integrity.

**Question 331.** A local coffee shop offers free Wi-Fi to its
customers. Recently, there have been reports of man-in-the-
middle attacks on the network. The owner decides to upgrade
the wireless security and wants to implement a cryptographic
protocol to secure data transmissions. Which protocol would
provide a balance between security and performance for the
public Wi-Fi users?


**(A) Advanced Encryption Standard (AES)
(B)** Wired Equivalent Privacy (WEP)
**(C)** RC4 Stream Cipher
**(D)** Open Wireless

**Explanation 331. Correct Answer: A. Advanced Encryption
Standard (AES).** AES offers a good balance between security
and performance. It is a modern encryption standard that
provides strong security without causing significant
performance overhead, making it suitable for a public Wi-Fi
setting where both security and user experience are important.

**Option B is incorrect.** WEP is an outdated encryption protocol
with known vulnerabilities. It is insecure and can be easily
cracked, making it unsuitable for securing a public Wi-Fi
network.

**Option C is incorrect.** RC4 is a stream cipher that was used in
older wireless protocols like WEP. It has been found to be
vulnerable to various attacks and is no longer considered secure
for wireless networks.

**Option D is incorrect.** An open wireless network does not
implement any encryption, leaving data transmissions
vulnerable to eavesdropping and other attacks. It would not
provide the desired security for the coffee shop’s Wi-Fi users.

**Question 332.** A software development company has decided to
host their applications in a multi-cloud environment. Before
deploying, they are looking to enhance the security of their
cloud-based resources. Which of the following is the BEST


practice for hardening their cloud infrastructure?
**(A)** Ensure that all cloud storage buckets or containers are
publicly accessible for easier data sharing
**(B) Apply consistent security configurations and policies
across all cloud providers
(C)** Use the same SSH key pairs across all cloud instances
for uniformity
**(D)** Limit the use of Identity and Access Management
(IAM) roles to senior staff only

**Explanation 332. Correct Answer: B. Apply consistent
security configurations and policies across all cloud
providers.** Maintaining consistent security configurations and
policies across all cloud providers ensures that there are no
weak links in the multi-cloud setup and reduces the complexity
of managing multiple sets of policies.

**Option A is incorrect.** Making all cloud storage buckets or
containers publicly accessible can expose sensitive data and is a
common misconfiguration that leads to data breaches.

**Option C is incorrect.** Using the same SSH key pairs across all
instances can be risky. If an attacker obtains the SSH key, they
can gain unauthorized access to all instances that use that key.

**Option D is incorrect.** While it’s important to limit access
based on the principle of least privilege, IAM roles should be
appropriately assigned to staff based on their responsibilities
and not just their seniority. This ensures that users have the
necessary permissions to do their jobs without unnecessary
access to sensitive resources.


**Question 333.** After running a vulnerability scan on the
company’s infrastructure, a security analyst notices a reported
vulnerability on a server. However, after manual verification,
the analyst determines that the vulnerability doesn’t actually
exist on the server. What is this situation best described as?
**(A)** A false negative
**(B)** A true positive
**(C) A false positive
(D)** A confirmation bias

**Explanation 333. Correct Answer: C. A false positive.** A false
positive in vulnerability management occurs when a system
incorrectly flags a threat or vulnerability that isn’t truly present.
This can lead to wasted resources on investigating and
attempting to remediate a non-existent issue.

**Option A is incorrect.** A false negative would mean that the
vulnerability exists, but the system failed to detect it, which is
the opposite of the scenario described.

**Option B is incorrect.** A true positive means that the
vulnerability was correctly identified and truly exists, which
isn’t the case here.

**Option D is incorrect.** Confirmation bias is a cognitive bias
where one favors information that confirms their existing
beliefs. It’s not relevant to the scenario of incorrectly detected
vulnerabilities.

**Question 334.** An e-commerce platform recently suffered a data
breach where attackers exploited cookies to impersonate user


sessions. A security analyst is tasked with recommending
measures to secure user cookies. Which of the following
measures will ensure that cookies are transmitted securely
between the user’s browser and the server?
**(A)** Storing cookies in the database
**(B) Implementing the "Secure" attribute for cookies
(C)** Increasing the cookie expiration time
**(D)** Base64 encoding the cookie content

**Explanation 334. Correct Answer: B. Implementing the
“Secure” attribute for cookies.** The “Secure” attribute ensures
that a cookie is only sent over secure, encrypted HTTPS
connections. By implementing this attribute, the cookie won’t
be transmitted over unencrypted HTTP connections, reducing
the risk of interception by malicious actors.

**Option A is incorrect.** Storing cookies in the database doesn’t
necessarily secure the transmission of cookies between the
client and server. It’s more about storage security rather than
transmission security.

**Option C is incorrect.** Increasing the cookie expiration time
can actually increase the window of opportunity for an attacker
to exploit a cookie, making it a less secure practice.

**Option D is incorrect.** Base64 encoding is not encryption; it’s
just a way to represent binary data in an ASCII string format.
Encoding can be easily reversed, offering little to no security.

**Question 335.** A security analyst is reviewing a vulnerability
report and sees a reference to CVE-2023-12345 with a CVSS


score of 9.5. Which of the following conclusions can the analyst
draw based on this information?
**(A) The vulnerability was first identified in the year 2023
(B)** The vulnerability is of low severity
**(C)** The vulnerability affects only software produced in
2023
**(D)** CVE-2023-12345 is the software vendor's internal code
for the vulnerability

**Explanation 335. Correct Answer: A. The vulnerability was
first identified in the year 2023.** The CVE identifier’s format
begins with the year the vulnerability was made public, so
CVE-2023-12345 indicates a vulnerability identified in 2023.

**Option B is incorrect.** CVSS scores range from 0 to 10, with
higher scores indicating higher severity. A score of 9.5 is
considered critical severity, not low.

**Option C is incorrect.** The year in a CVE identifier refers to
when the vulnerability was published, not the year the software
was produced.

**Option D is incorrect.** The CVE system is a standard method
for identifying vulnerabilities and doesn’t represent a software
vendor’s internal coding.

**Question 336.** AlphaTech is seeking a comprehensive source of
intelligence about the latest cyber threats targeting its specific
industry. While OSINT provides valuable data, the company is
considering investing in a more specialized solution. Which of
the following would best address AlphaTech’s needs?


**(A)** Implementing internal honeypots to trap attackers
**(B) Subscribing to a third-party threat intelligence feed
(C)** Regularly attending cyber security conferences
**(D)** Using open-source vulnerability scanners

**Explanation 336. Correct Answer: B. Subscribing to a third-
party threat intelligence feed.** A third-party threat intelligence
feed, especially one tailored for a specific industry, provides
specialized, often real-time information about cyber threats,
offering insights beyond what’s available in the public domain.

**Option A is incorrect.** While honeypots can help understand
the tactics of attackers targeting the organization, they don’t
provide comprehensive intelligence about industry-wide threats.

**Option C is incorrect.** Attending cyber security conferences
can offer insights and updates, but it’s not a continuous or real-
time source of threat intelligence.

**Option D is incorrect.** Open-source vulnerability scanners help
identify vulnerabilities within an organization’s infrastructure
but don’t offer specialized intelligence about industry-specific
cyber threats.

**Question 337.** ABC Corp recently adopted a Bring Your Own
Device (BYOD) policy. The IT department is concerned about
the potential risks associated with personal devices accessing
the corporate network. Which of the following solutions would
be MOST effective for enforcing security policies on these
personal mobile devices?
**(A)** Installing antivirus software on each device


**(B)** Establishing a separate guest Wi-Fi network for mobile
devices
**(C) Using Mobile Device Management (MDM) to enforce
security policies
(D)** Mandating that employees use strong passwords on
their personal devices

**Explanation 337. Correct Answer: C. Using Mobile Device
Management (MDM) to enforce security policies.** MDM
solutions provide centralized control to enforce security
policies, manage device features, and ensure that personal
devices meet the organization’s security standards before
accessing corporate resources.

**Option A is incorrect.** While installing antivirus software is a
good security measure, it doesn’t provide the comprehensive
policy enforcement capabilities that MDM does.

**Option B is incorrect.** A separate guest Wi-Fi may restrict
access to the internal network, but it doesn’t manage or enforce
security policies on the devices themselves.

**Option D is incorrect.** While using strong passwords is
essential, it’s just one aspect of device security. MDM offers
broader policy enforcement capabilities.

**Question 338.** David, an IT administrator, noticed an unusually
high data usage on several company-owned mobile devices
even when they are connected to the corporate Wi-Fi. He
suspects these devices might be using cellular data in the
background. Which of the following solutions should David


implement to ensure that company devices use only the
corporate Wi-Fi for data transactions when they’re in the office?
**(A)** Enable Airplane mode on all devices
**(B)** Set up a Wi-Fi whitelist
**(C) Implement a mobile device management (MDM)
policy to prioritize Wi-Fi
(D)** Disable cellular antennas in the office area

**Explanation 338. Correct Answer: C. Implement a mobile
device management (MDM) policy to prioritize Wi-Fi.** An
MDM solution allows administrators to enforce policies on
mobile devices. By implementing a policy that prioritizes Wi-Fi
connections, David can ensure that company devices will use
the corporate Wi-Fi network when available, thus reducing
cellular data usage.

**Option A is incorrect.** While enabling Airplane mode would
cut off cellular data, it would also disable all other forms of
communication including Wi-Fi, making the device unusable
for its intended purposes.

**Option B is incorrect.** Setting up a Wi-Fi whitelist does not
prevent a device from using cellular data. It only restricts which
Wi-Fi networks a device can connect to.

**Option D is incorrect.** Disabling cellular antennas is a drastic
measure that could affect other devices and services in the
vicinity. It’s also impractical and potentially illegal depending
on local regulations.


**Question 339.** A security team recently upgraded their intrusion
detection system (IDS). Since the upgrade, the system hasn’t
flagged any intrusions, even though intrusion attempts are a
regular occurrence. What is this situation best characterized as?
**(A)** A true negative
**(B) A false negative
(C)** A true positive
**(D)** A confirmation feedback

**Explanation 339. Correct Answer: B. A false negative.** A
false negative occurs when a system fails to detect a threat or
vulnerability that is actually present. In this scenario, the IDS
isn’t detecting real intrusion attempts, which could lead to
undetected breaches.

**Option A is incorrect.** A true negative would mean that the
system correctly identified that there were no intrusions when
there truly weren’t any. This isn’t the scenario described as
intrusion attempts are expected.

**Option C is incorrect.** A true positive means that the threat or
vulnerability was correctly identified and truly exists. Since no
intrusions are being detected, this isn’t a true positive.

**Option D is incorrect.** “Confirmation feedback” isn’t a
standard term related to vulnerability detection and doesn’t
apply to the described scenario.

**Question 340.** XYZ Corporation is planning to deploy a new
wireless infrastructure in their newly acquired office building.
The IT manager wants to ensure optimal wireless coverage


throughout the premises. Which of the following should the IT
team prioritize before installing the wireless access points?
**(A)** Purchase the most expensive wireless access points to
ensure maximum range
**(B) Conduct a site survey to determine the best locations
for access points
(C)** Deploy all access points near windows to enhance
signal strength
**(D)** Ensure all users have 5GHz capable devices

**Explanation 340. Correct Answer: B. Conduct a site survey
to determine the best locations for access points.** A site
survey will help identify the optimal placements for access
points to achieve consistent and robust wireless coverage across
the entire premises.

**Option A is incorrect.** While quality equipment is essential, the
costliest access points might not always guarantee the best
results. Placement and environment play crucial roles in
wireless performance.

**Option C is incorrect.** Deploying access points near windows
can cause signal leakage, potentially making the signal available
outside the intended area and presenting a security risk.

**Option D is incorrect.** While ensuring users have 5GHz
capable devices is a good practice, it doesn’t relate directly to
the optimal installation of wireless access points.

**Question 341.** XYZ Company uses MDM to manage company-
owned and employee-owned mobile devices. An employee


reported losing their personal phone over the weekend. What
MDM feature should the IT department use to ensure that
sensitive company data on the lost phone isn’t accessed?
**(A)** Monitor the device's location
**(B)** Force update the device's apps
**(C) Remote wipe the device
(D)** Change the user's email password

**Explanation 341. Correct Answer: C. Remote wipe the
device.** Using the remote wipe feature, the IT department can
erase all data, including company data, from the lost device,
ensuring that sensitive information doesn’t fall into the wrong
hands.

**Option A is incorrect.** Monitoring the device’s location might
help in finding it but doesn’t prevent unauthorized access to the
data on the device.

**Option B is incorrect.** Force updating the device’s apps might
address vulnerabilities in the apps but doesn’t directly protect
the company data on the device.

**Option D is incorrect.** Changing the user’s email password can
prevent unauthorized access to the user’s email, but it doesn’t
secure other sensitive company data that might be on the device.

**Question 342.** After a series of cyber incidents, AlphaTech
Corp. wants to take proactive measures to identify
vulnerabilities in their network. They aim to obtain a
comprehensive report of potential weaknesses without
exploiting them. Which of the following would best meet this


objective?
**(A)** Penetration test
**(B) Vulnerability scan
(C)** Red team assessment
**(D)** Port security

**Explanation 342. Correct Answer: B. Vulnerability scan.** A
vulnerability scan is designed to inspect systems, networks, and
applications to identify potential weaknesses or vulnerabilities.
Unlike some of the other options, it doesn’t attempt to exploit
these vulnerabilities; it merely identifies and reports them.

**Option A is incorrect.** A penetration test goes a step further
than a vulnerability scan. While it identifies vulnerabilities, it
also attempts to exploit them to understand the potential impact
of a breach.

**Option C is incorrect.** A red team assessment is a goal-based
assessment where a simulated adversary (the red team) tries to
achieve specific objectives. It often involves exploitation, which
goes beyond the mere identification of vulnerabilities.

**Option D is incorrect.** Port security is a feature at the data link
layer to control MAC address-based access on a port-by-port
basis. It doesn’t offer a comprehensive vulnerability report.

**Question 343.** GammaTech is in the final stages of deploying a
new application. Before the deployment, the security team
wants to examine the application’s code without executing it to
identify any potential vulnerabilities. Which vulnerability
identification method should the team employ?


**(A)** Penetration Testing
**(B)** Dynamic Analysis
**(C) Static Analysis
(D)** Fuzz Testing

**Explanation 343. Correct Answer: C. Static Analysis.** Static
analysis involves examining an application’s code, bytecode, or
binary code without executing it to identify vulnerabilities. This
method allows security teams to find potential security issues in
the codebase before the application runs.

**Option A is incorrect.** Penetration testing involves actively
trying to exploit vulnerabilities in a system or application. It
does not focus on code examination without execution.

**Option B is incorrect.** Dynamic analysis examines an
application during its runtime, observing its behavior to identify
vulnerabilities.

**Option D is incorrect.** Fuzz testing, or fuzzing, involves
providing a program with a series of random inputs to see if any
of them cause crashes or other unexpected behavior, helping
identify vulnerabilities.

**Question 344.** During a routine vulnerability assessment,
TechInc discovers a weakness in their system that, if exploited,
would allow an attacker to modify existing user accounts,
including privileges. Which classification best describes this
vulnerability?
**(A) Elevation of privilege vulnerability
(B)** Disclosure vulnerability


**(C)** Replay vulnerability
**(D)** Remote code execution vulnerability

**Explanation 344. Correct Answer: A. Elevation of privilege
vulnerability.** Elevation of privilege vulnerabilities allow
attackers to increase their privileges within a system, often
giving them more access than intended. The scenario describes
a situation where user accounts and their privileges can be
modified, aligning with this classification.

**Option B is incorrect.** Disclosure vulnerabilities involve
unauthorized access to information but not the modification of
user privileges or data.

**Option C is incorrect.** Replay vulnerabilities occur when
attackers capture and later retransmit valid data transmissions to
fool a system, which is not described in the given scenario.

**Option D is incorrect.** Remote code execution vulnerabilities
allow attackers to execute arbitrary commands or code on a
target system. While these can be severe, the given scenario
specifically describes privilege changes, not arbitrary code
execution.

**Question 345.** A healthcare organization uses embedded
systems in various medical devices. They are aware of the
potential threats these systems can pose if not properly secured.
Which of the following is NOT a recommended practice when
hardening embedded systems in this context?
**(A)** Regularly patching and updating the firmware of the
devices


**(B) Allowing unrestricted access to the devices for ease of
use by the medical staff
(C)** Disabling unnecessary services and features not
required for the device's primary function
**(D)** Changing default credentials and using strong, unique
passwords for device access

**Explanation 345. Correct Answer: B. Allowing unrestricted
access to the devices for ease of use by the medical staff.**
Unrestricted access can lead to unintentional or deliberate
misconfigurations or misuse of the device, compromising its
security.

**Option A is incorrect.** Regularly patching and updating
firmware is a critical aspect of maintaining the security of
embedded systems, especially with the evolving nature of
threats.

**Option C is incorrect.** Disabling unnecessary services and
features reduces the attack surface and is a fundamental
principle of system hardening.

**Option D is incorrect.** Changing default credentials and
ensuring strong, unique passwords are in place is essential to
prevent unauthorized access.

**Question 346.** TechSoft Corp. is implementing a new asset-
tracking system to monitor its vast array of computing
resources. Which of the following should be the PRIMARY
reason for maintaining an up-to-date hardware and software
inventory?


**(A)** To ensure software licenses are renewed on time
**(B) To identify and respond to unauthorized devices or
software promptly
(C)** To aid in the procurement of new hardware and software
**(D)** To provide employees with an understanding of
available resources

**Explanation 346. Correct Answer: B. To identify and
respond to unauthorized devices or software promptly.
While all the options have valid reasons for maintaining an
i** nventory, from a security standpoint, identifying unauthorized
devices or software promptly is crucial. Unauthorized devices
or software can pose a significant security risk, including
potential data breaches or malware infections.

**Option A is incorrect.** While renewing software licenses is
important, it’s not the primary security reason for maintaining
an inventory.

**Option C is incorrect.** Procurement decisions benefit from
inventory data but do not directly impact immediate security
concerns in the way that identifying unauthorized assets does.

**Option D is incorrect.** While providing employees with
resource understanding is beneficial, it isn’t as crucial for
security as promptly identifying unauthorized assets.

**Question 347.** An enterprise is deploying IoT-based security
cameras across multiple office locations. As the lead security
professional, what recommendation would you prioritize to
establish a secure baseline for these devices?


**(A)** Setting the devices to public mode so all employees can
access the feed for transparency
**(B) Regularly updating the device firmware to patch
known vulnerabilities
(C)** Enabling Universal Plug and Play (UPnP) to ensure
easy connectivity for all devices on the network
**(D)** Using the same password for all cameras for ease of
management

**Explanation 347. Correct Answer: B. Regularly updating
the device firmware to patch known vulnerabilities.** Regular
firmware updates ensure that the IoT devices are protected
against identified vulnerabilities, helping maintain their security
posture.

**Option A is incorrect.** Setting devices to a public mode can
lead to potential breaches of privacy and unauthorized access.

**Option C is incorrect.** UPnP can introduce vulnerabilities by
automatically opening ports and allowing devices to set their
configurations, potentially leading to security risks.

**Option D is incorrect.** Using the same password for all devices
creates a single point of failure. If one device’s password is
compromised, all devices become vulnerable.

**Question 348.** A medium-sized enterprise is concerned about
the security of its office workstations after a series of malware
infections. As a security analyst, which of the following
recommendations would BEST improve the security baseline of
the workstations?


**(A)** Install multiple antivirus solutions to ensure maximum
detection
**(B)** Set up screensavers with cyber hygiene tips to educate
users
**(C) Disable unnecessary services and ports on the
workstations
(D)** Frequently change the desktop wallpaper to prevent
monotony

**Explanation 348. Correct Answer: C. Disable unnecessary
services and ports on the workstations.** By disabling
unnecessary services and ports, you reduce the number of
potential attack vectors and vulnerabilities, thus improving the
security baseline of the workstations.

**Option A is incorrect.** Installing multiple antivirus solutions
can lead to conflicts and may degrade system performance. It’s
better to have one robust, updated antivirus solution.

**Option B is incorrect.** While cyber hygiene tips can be
informative, they don’t directly contribute to the technical
security baseline of a workstation.

**Option D is incorrect.** Changing the desktop wallpaper does
nothing for security. Its primary purpose is aesthetics.

**Question 349.** AlphaCorp is migrating to cloud infrastructure
and wants to ensure all virtual machines (VMs) are securely
configured from the onset. Before deploying multiple VM
instances, what should AlphaCorp do to ensure each VM starts
from a secure configuration?


**(A)** Use the default VM templates provided by the cloud
provider
**(B) Establish a secure baseline for VM configurations
and use it for deployment
(C)** Regularly backup all VMs
**(D)** Use multi-factor authentication for cloud access

**Explanation 349. Correct Answer: B. Establish a secure
baseline for VM configurations and use it for deployment.**
Establishing a secure baseline for VM configurations ensures
that each VM is deployed with a set of standard security
settings, reducing vulnerabilities from default configurations or
potential misconfigurations.

**Option A is incorrect.** Relying solely on default VM templates
provided by cloud providers may not meet the specific security
requirements of an organization. Customizing and creating a
secure baseline is more effective.

**Option C is incorrect.** While regular backups are essential for
data recovery, they do not directly ensure that VMs start from a
secure configuration.

**Option D is incorrect.** Multi-factor authentication is a crucial
security measure for accessing cloud resources, but it doesn’t
ensure that VMs are deployed with secure configurations.

**Question 350.** A multinational company is deploying a new set
of servers in its data centers across various countries. Which of
the following steps should be taken FIRST to ensure the servers
are secured against potential threats?


**(A)** Set up a monitoring system to alert the IT team of any
irregular activities
**(B)** Deploy all the software applications the company might
need in the future
**(C)** Use the server's default configuration to ensure
manufacturer's best practices are maintained
**(D) Disable any unused services and ports on the server**

**Explanation 350. Correct Answer: D. Disable any unused
services and ports on the server.** When setting up a new
server, it’s crucial to minimize its attack surface. Disabling
unused services and ports ensures that only necessary services
run on the server, reducing potential vulnerabilities.

**Option A is incorrect.** While monitoring is vital for security,
before setting up a monitoring system, it’s more crucial to
harden the server to minimize vulnerabilities. Monitoring
should complement hardening measures.

**Option B is incorrect.** Deploying all the software applications
the company might need in the future can introduce unnecessary
vulnerabilities and overhead. It’s better to install only what’s
needed and keep the server lean.

**Option C is incorrect.** Relying on a server’s default
configuration can be risky, as it might not be tailored for an
organization’s specific needs and could have open ports or
services that are unnecessary.

**Question 351.** While analyzing a vulnerability in a company’s
web application, the security team refers to a specific CVE to


understand the vulnerability’s details. They further assess its
CVSS score to decide on the remediation urgency. Which of the
following best describes the purpose of the CVE and CVSS in
this context?
**(A)** CVE provides a severity score, while CVSS gives a
unique identifier for the vulnerability
**(B)** CVE and CVSS both offer a scoring mechanism to rank
vulnerabilities
**(C) CVE provides a unique identifier, while CVSS offers
a standardized severity score
(D)** CVE and CVSS are regulatory requirements for all
software applications

**Explanation 351. Correct Answer: C. CVE provides a
unique identifier, while CVSS offers a standardized severity
score.** CVE (Common Vulnerabilities and Exposures) offers a
standardized identifier for vulnerabilities, ensuring clear
communication and reference. CVSS (Common Vulnerability
Scoring System) provides a standardized method for rating the
severity of vulnerabilities, enabling organizations to prioritize
remediation.

**Option A is incorrect.** It’s the opposite. CVE gives a unique
identifier, while CVSS provides a severity score.

**Option B is incorrect.** Only CVSS offers a scoring mechanism
to rank vulnerabilities. CVE offers unique identifiers.

**Option D is incorrect.** While using CVE and CVSS can be
considered best practices, they aren’t regulatory requirements
for all software applications.


**Question 352.** The network administrator at a university wants
to ensure that when students log onto the campus wireless
network, their credentials are verified by the university’s central
authentication server. Additionally, the administrator wants to
make sure that the data between the wireless access point and
the central server is encrypted. Which solution should the
administrator implement?
**(A)** WPA3 with SAE
**(B)** WPA2-Personal with AES
**(C) WPA2-Enterprise with RADIUS
(D)** Open wireless with VPN

**Explanation 352. Correct Answer: C. WPA2-Enterprise
with RADIUS.** WPA2-Enterprise provides a more robust
authentication method suitable for larger organizations like
universities. By using RADIUS, the university can centralize
the authentication process, ensuring students’ credentials are
verified against the central server. RADIUS also encrypts the
data between the wireless access point and the central server,
providing an additional layer of security.

**Option A is incorrect.** WPA3 with SAE is primarily used for
secure handshakes between devices and does not directly relate
to centralized authentication against a server like RADIUS.

**Option B is incorrect.** WPA2-Personal with AES is designed
for personal or small office use where a pre-shared key is used.
It doesn’t support centralized authentication like RADIUS.

**Option D is incorrect.** While a VPN can encrypt data between
a device and a network, it doesn’t centralize wireless


authentication in the way that RADIUS does with WPA2-
Enterprise.

**Question 353.** As part of the company’s vulnerability
management initiative, the security team has decided to conduct
a series of penetration tests. Which of the following is the
PRIMARY reason for incorporating penetration testing as a
threat identification method?
**(A)** To ensure compliance with regulatory requirements
**(B)** To validate the efficiency of security awareness training
**(C) To actively exploit vulnerabilities and assess
potential impact
(D)** To identify misconfigurations in the SIEM system

**Explanation 353. Correct Answer: C. To actively exploit
vulnerabilities and assess potential impact.** Penetration
testing is a method used to actively exploit vulnerabilities in an
environment. Its primary purpose is to determine the potential
impact and risk of those vulnerabilities in a real-world scenario,
thereby allowing the organization to understand and prioritize
remediation efforts.

**Option A is incorrect.** While some regulations might require
penetration testing, the primary goal of the test is not just for
compliance but to understand vulnerabilities’ potential impact.

**Option B is incorrect.** Although penetration testing can
sometimes be used to gauge the effectiveness of security
training (for instance, in social engineering tests), it’s not the
primary reason for conducting these tests.


**Option D is incorrect.** SIEM systems are used for logging and
event management. While a pen test might uncover
misconfigurations in various systems, its primary purpose is not
to focus on the SIEM.

**Question 354.** An organization has recently received a new
software patch for its critical infrastructure. Before deploying it
to production, the security team wants to understand its
behavior and ensure it doesn’t contain any malicious code.
Which of the following methods would be MOST effective for
safely executing and observing the patch’s behavior?
**(A)** Deploying the patch during a maintenance window
**(B) Running the patch within a sandbox environment
(C)** Conducting a code review of the patch
**(D)** Installing the patch on a virtual machine

**Explanation 354. Correct Answer: B. Running the patch
within a sandbox environment.** Running the patch within a
sandbox environment allows the security team to execute and
observe the software’s behavior in an isolated environment,
ensuring it doesn’t interfere with or harm the actual production
environment.

**Option A is incorrect.** Deploying the patch during a
maintenance window reduces operational disruptions but
doesn’t allow for safe observation of the patch’s behavior.

**Option C is incorrect.** While a code review can identify
potential security concerns in the patch, it may not reveal the
actual behavior when executed.


**Option D is incorrect.** Installing the patch on a virtual machine
is a form of isolation, but a sandbox provides a more controlled
and restrictive environment specifically designed for observing
software behavior.

**Question 355.** GreenTech Inc. is selling a set of old servers to
another company. Before the transaction, they want to ensure
the data on these servers is irretrievable and they can prove that
due diligence was performed. What should GreenTech seek to
assure the buyer of proper data destruction?
**(A)** A receipt of sale for the servers
**(B)** A detailed log of the server's usage
**(C) A certificate of data sanitization
(D)** A user manual of the servers

**Explanation 355. Correct Answer: C. A certificate of data
sanitization.** A certificate of data sanitization from a reputable
entity provides assurance that all data on the servers was
securely and completely wiped. Such certificates confirm that
specific techniques were used to ensure data is irretrievable,
giving confidence to the buyer.

**Option A is incorrect.** A receipt of sale only proves the
transaction took place but does not address data sanitization or
destruction.

**Option B is incorrect.** A log of the server’s usage might detail
its operational history but does not provide assurance of data
destruction.


**Option D is incorrect.** A user manual provides instructions on
how to use the server but does not address the state or security
of the data it once held.

**Question 356.** Jane, the CISO at a financial institution, is
overseeing the decommissioning of several old servers. She is
aware that while some data must be destroyed, other data must
be retained due to industry regulations. Which principle should
Jane primarily focus on to ensure compliance?
**(A)** Minimum necessary principle
**(B)** Principle of least privilege
**(C) Data retention policy
(D)** Mandatory vacation policy

**Explanation 356. Correct Answer: C. Data retention policy.**
A data retention policy specifically defines how long data
should be retained and the manner of its storage based on
business needs, regulatory requirements, and other factors. In
the context of decommissioning servers, adhering to this policy
will ensure that Jane complies with regulations related to data
retention.

**Option A is incorrect.** The minimum necessary principle
pertains to healthcare and refers to only using, disclosing, or
requesting the amount of information necessary to accomplish a
specific task.

**Option B is incorrect.** The principle of least privilege pertains
to access controls and means giving users the minimum levels
of access necessary to perform their jobs.


**Option D is incorrect.** A mandatory vacation policy pertains to
ensuring employees take time off so that potential fraudulent
activities might be detected in their absence. It doesn’t directly
address data retention.

**Question 357.** CyberFirm, a leading software development
company, recently updated their server OS due to new features
and patches. Given that they have already established and
deployed a secure baseline in the past, what should CyberFirm
do NEXT to ensure continued security?
**(A)** Conduct a complete system reboot for all servers
**(B)** Re-deploy the same baseline without any modifications
**(C) Update the secure baseline to include new
configurations and then deploy it
(D)** Implement a new firewall rule for the servers

**Explanation 357. Correct Answer: C. Update the secure
baseline to include new configurations and then deploy it.**
When there are significant updates or changes to systems, it’s
crucial to review and update the secure baseline accordingly,
ensuring it remains relevant and effective for the current
environment. After updating the baseline, it should then be re-
deployed.

**Option A is incorrect.** A complete system reboot does not
address the maintenance of the secure baseline after an OS
update.

**Option B is incorrect.** Simply re-deploying the same baseline
without adjustments might miss out on specific configurations
or considerations needed due to the OS update.


**Option D is incorrect.** While firewall rules are essential for
security, they don’t directly address the maintenance and
updating of a secure baseline.

**Question 358.** MatrixCorp recently adopted a mobile strategy
where employees are provided with company-owned devices.
These devices are also allowed for personal use, but the
organization retains the ability to manage and monitor them.
Which deployment model is MatrixCorp using?
**(A)** Bring Your Own Device (BYOD)
**(B)** Choose Your Own Device (CYOD)
**(C) Corporate-owned, Personally Enabled (COPE)
(D)** Public Device Deployment (PDD)

**Explanation 358. Correct Answer: C. Corporate-owned,
Personally Enabled (COPE).** The COPE model involves
organizations providing employees with company-owned
devices that they can also use for personal tasks. However, the
company retains full control, management, and monitoring
capabilities over these devices.

**Option A is incorrect.** In the BYOD model, employees use
their personal devices for work, not company-provided ones.

**Option B is incorrect.** CYOD allows employees to select a
device from a list of company-approved devices, but the
emphasis is on choice rather than personal use of corporate-
owned devices.


**Option D is incorrect.** Public Device Deployment (PDD) is not
a recognized deployment model in the context of mobile
solutions.

**Question 359.** A company has recently upgraded its wireless
infrastructure and wants to ensure that the data transmitted over
its wireless network is protected using the most recent and
secure encryption standards. Which of the following should the
company configure on its wireless access points?
**(A)** WEP
**(B)** WPA
**(C)** WPA2
**(D) WPA3**

**Explanation 359. Correct Answer: D. WPA3.** WPA3 is the
latest iteration of Wi-Fi Protected Access, designed to improve
upon the security features of WPA2. It offers enhanced
protections against brute-force attacks, superior encryption
methods, and a more secure handshake process.

**Option A is incorrect.** WEP (Wired Equivalent Privacy) is an
outdated and insecure protocol that can be easily cracked. It
should not be used for securing wireless networks.

**Option B is incorrect.** While WPA (Wi-Fi Protected Access)
was a significant improvement over WEP, it has been surpassed
by both WPA2 and WPA3 in terms of security features.

**Option C is incorrect.** WPA2 has been a strong standard for
many years but has vulnerabilities that WPA3 addresses,
making WPA3 the more secure option.


**Question 360.** A security analyst is exploring ways to
proactively identify vulnerabilities within the organization’s
infrastructure. Which of the following provides the BEST
method for the analyst to receive real-time threat intelligence
from the dark web?
**(A)** Utilizing a vulnerability scanner on the organization's
internal network
**(B) Subscribing to a dark web threat intelligence feed
(C)** Conducting regular penetration tests on external-facing
systems
**(D)** Reviewing daily reports from the organization's SIEM
system

**Explanation 360. Correct Answer: B. Subscribing to a dark
web threat intelligence feed.** Threat intelligence feeds,
especially those focusing on the dark web, provide real-time
information about emerging threats and vulnerabilities. This
proactive approach allows analysts to be aware of potential
threats even before they may affect the organization.

**Option A is incorrect.** While vulnerability scanners are
effective at identifying known vulnerabilities within a network,
they don’t provide real-time threat intelligence from external
sources like the dark web.

**Option C is incorrect.** Penetration tests are indeed useful for
identifying vulnerabilities in systems and applications.
However, they don’t specifically provide threat intelligence
from the dark web.


**Option D is incorrect.** SIEM systems aggregate and analyze
log data from various sources to identify and respond to security
incidents. While they can alert to current threats within the
environment, they don’t offer real-time dark web threat
intelligence.

**Question 361.** XYZ Corp is designing their new web
application infrastructure. They want to ensure that all web
traffic to and from their application is encrypted. In addition to
selecting HTTPS as the protocol, which default port should they
configure for this encrypted traffic?
**(A)** 21
**(B)** 80
**(C) 443
(D)** 25

**Explanation 361. Correct Answer: C. 443.** Port 443 is the
default port used for HTTPS (HTTP over SSL/TLS) traffic. It
ensures that web traffic is encrypted and provides a secure
communication channel between the client and the server.

**Option A is incorrect.** Port 21 is the default port for FTP (File
Transfer Protocol), which is used for transferring files, not
encrypted web traffic.

**Option B is incorrect.** Port 80 is the default port for HTTP,
which is unencrypted.

**Option D is incorrect.** Port 25 is the default port for SMTP
(Simple Mail Transfer Protocol), which is used for email
transmission, not encrypted web traffic.


**Question 362.** TechWorld Corp is concerned about
cybercriminals sending emails that appear to come from its
domain to deceive its clients. The company wants to implement
a solution that would allow receiving email servers to validate
that an email claiming to come from TechWorld Corp’s domain
indeed originates from an approved server. Which of the
following should the company implement?
**(A)** SMTP authentication
**(B) DKIM
(C)** POP3 over SSL
**(D)** S/MIME

**Explanation 362. Correct Answer: B. DKIM.** DomainKeys
Identified Mail (DKIM) is a method used to authenticate the
origin and integrity of an email. With DKIM, a digital signature
is added to the headers of outgoing emails. Receiving email
servers then use this signature to validate the email by checking
it against the sending domain’s public key, which is published in
the DNS records.

**Option A is incorrect.** SMTP authentication ensures that users
provide valid credentials before they can send emails through a
server, but it doesn’t validate the domain of the sender to the
receiver.

**Option C is incorrect.** POP3 over SSL (Secure Sockets Layer)
is about securely retrieving email messages from a server. It
doesn’t provide a way to authenticate the sender’s domain.

**Option D is incorrect.** S/MIME (Secure/Multipurpose Internet
Mail Extensions) is used to encrypt and sign email messages.


While it ensures the integrity and confidentiality of the email
content, it does not validate the sender’s domain to the recipient.

**Question 363.** After a security incident, a forensic investigation
revealed that a compromised internal workstation was
communicating with a known malicious IP address. To prevent
further communication, the security team decided to take
immediate action. Which of the following is the BEST
immediate action to ensure the workstation cannot communicate
with that IP?
**(A) Implement a block rule on the web filter for the IP
address
(B)** Disable the network port of the compromised
workstation
**(C)** Use a honeypot to divert the traffic from the malicious
**(D)** Update the firewall's firmware

**Explanation 363. Correct Answer: A. Implement a block
rule on the web filter for the IP address.** By implementing a
block rule specifically for the known malicious IP address on
the web filter, any attempt to communicate with that IP would
be stopped, effectively preventing further malicious
communications.

**Option B is incorrect.** Disabling the network port of the
compromised workstation would disconnect the machine from
the network, but it wouldn’t specifically prevent communication
with the malicious IP if the port were to be re-enabled.


**Option C is incorrect.** While honeypots can be used to divert
or study attacker behavior, they are not the most direct way to
block communication with a known malicious IP address.

**Option D is incorrect.** Updating the firewall’s firmware is a
general maintenance task and does not specifically address
blocking communications with the known malicious IP.

**Question 364.** Amy, a network administrator, is researching
tools to assist with automating the evaluation of her
organization’s systems against a specific security baseline. She
comes across SCAP and wants to implement it. Which of the
following BEST describes the primary function of the Security
Content Automation Protocol (SCAP)?
**(A)** To facilitate the real-time transfer of threat intelligence
feeds
**(B)** To provide an interface for user authentication against
Active Directory
**(C) To allow for automated vulnerability management
and policy compliance evaluation
(D)** To offer encrypted communication channels for remote
system management

**Explanation 364. Correct Answer: C. To allow for
automated vulnerability management and policy compliance
evaluation.** SCAP (Security Content Automation Protocol) is a
suite of open standards that enhances the ability of
organizations to automate vulnerability management,
measurement, and policy compliance evaluation on systems.


**Option A is incorrect.** While threat intelligence is crucial for
security, SCAP is not designed primarily for real-time threat
intelligence feed transfers.

**Option B is incorrect.** SCAP does not directly deal with user
authentication against any directory services.

**Option D is incorrect.** SCAP’s primary function isn’t to offer
encrypted communication channels for remote system
management. Other protocols and tools serve this purpose.

**Question 365.** A pharmaceutical company is working on a new
drug formula that promises to revolutionize the treatment of a
particular disease. The R&D team has detailed documentation
on the components, procedures, and results of the drug trials.
How should this documentation be classified to ensure that only
the right people within the company have access?
**(A)** Implement a Domain Name System (DNS) firewall
**(B) Employ URL scanning to identify and block
malicious URLs
(C)** Rely on manual reporting of suspicious URLs by
employees
**(D)** Use a Virtual Private Network (VPN) to redirect all
employee web traffic

**Explanation 365. Correct Answer: B. Employ URL scanning
to identify and block malicious URLs.** URL scanning can
identify potentially malicious sites by examining the complete
URL and comparing it with known malicious URLs or patterns.
This method can block access to URLs that are designed to look
similar to legitimate ones.


**Option A is incorrect.** While a DNS firewall can help block
access to malicious domains, URL scanning provides more
granular control by examining the full URL, not just the
domain.

**Option C is incorrect.** Relying solely on manual reporting by
employees may not be as efficient or effective as an automated
URL scanning solution. It would also place undue responsibility
on employees to recognize and report every malicious URL.

**Option D is incorrect.** Using a VPN simply changes the route
of the traffic and provides encrypted communication. It does not
inherently offer URL scanning or filtering capabilities.

**Question 366.** A financial institution wants to ensure that any
unauthorized access to customer data triggers an immediate
alert to the security team. Which of the following approaches
would be the MOST effective in achieving this requirement?
**(A)** Configure alerts for any modification to database
records
**(B)** Set up alerts for successful logins during off-business
hours
**(C) Establish alerting thresholds based on anomalous
user behavior
(D)** Send daily reports of all access attempts to the security
team for review

**Explanation 366. Correct Answer: C. Establish alerting
thresholds based on anomalous user behavior.** By setting
alerting thresholds based on anomalous behavior, the system
can detect unauthorized access based on patterns that deviate


from typical user behavior, providing timely alerts for potential
breaches.

**Option A is incorrect.** Configuring alerts for any modification
to database records may generate a high number of false
positives, especially in a dynamic environment like a financial
institution where legitimate changes occur regularly.

**Option B is incorrect.** While setting up alerts for off-hours
access can catch some unauthorized attempts, it might miss
breaches occurring during business hours.

**Option D is incorrect.** Daily reports may delay the detection
and response to unauthorized access, as it doesn’t offer real-
time alerting.

**Question 367.** A marketing team is collaborating on a new
campaign and requires access to a shared folder. However, they
shouldn’t be able to modify files created by others. How should
permissions be set on this shared folder?
**(A)** Assign the marketing team full control
**(B)** Allow the marketing team read-only access
**(C)** Assign the marketing team write-only access
**(D) Assign the marketing team modify permission but
deny the delete permission**

**Explanation 367. Correct Answer: D. Assign the marketing
team modify permission but deny the delete permission.** By
providing modify permissions but denying delete permissions,
team members can create and edit their own files, but they
won’t be able to delete or modify files created by others.


**Option A is incorrect.** Full control would allow team members
to modify or delete any file, which is not desired.

**Option B is incorrect.** Read-only access would prevent team
members from creating or editing any files.

**Option C is incorrect.** Write-only access would prevent team
members from viewing existing files in the folder.

**Question 368.** Alice, a cybersecurity analyst, is tasked with
identifying potential weaknesses in a newly deployed web
application’s infrastructure before it goes live. She wants a tool
that can proactively discover and report on system
vulnerabilities, missing patches, and misconfigurations. Which
of the following should Alice utilize for this purpose?
**(A)** Intrusion Detection System (IDS)
**(B)** Network sniffer
**(C) Vulnerability scanner
(D)** Security Information and Event Management (SIEM)
system.

**Explanation 368. Correct Answer: C. Vulnerability scanner.**
Vulnerability scanners are designed to discover and report
potential vulnerabilities in systems by probing and analyzing
them. They can detect issues such as missing patches,
misconfigurations, and known software flaws.

**Option A is incorrect.** An IDS detects and alerts on potential
malicious activities based on specific signatures or heuristics
but does not proactively scan for vulnerabilities.


**Option B is incorrect.** A network sniffer captures and analyzes
network traffic but does not perform vulnerability assessments
on infrastructure components.

**Option D is incorrect.** While a SIEM system can aggregate and
analyze log and event data from various sources, it does not
proactively scan for vulnerabilities like a vulnerability scanner.

**Question 369.** Globex Industries is expanding its data centers
across multiple geographic locations. The IT team wants to have
a centralized system to get real-time status, outages, and metrics
of all data center infrastructures. Which of the following
solutions would be the MOST effective for this purpose?
**(A)** Data Loss Prevention (DLP) tools
**(B)** Distributed Denial of Service (DDoS) protection
**(C)** Security Information and Event Management (SIEM)
**(D) Infrastructure Management Platform (IMP)**

**Explanation 369. Correct Answer: D. Infrastructure
Management Platform (IMP).** IMPs provide centralized
control and monitoring for infrastructure across multiple sites.
They deliver insights into infrastructure health, outages, and
metrics, ensuring that the IT team can address issues promptly.

**Option A is incorrect.** DLP tools focus on preventing
unauthorized data transfers and exfiltrations, not on monitoring
infrastructure health and status.

**Option B is incorrect.** DDoS protection tools safeguard against
denial-of-service attacks but aren’t designed to monitor the
overall health or metrics of data center infrastructures.


**Option C is incorrect.** While SIEM solutions centralize log
data and provide insights into security events, they are not
designed primarily to monitor data center infrastructure health
across multiple locations.

**Question 370.** The IT department of XYZ Corp is keen on
preventing users from changing specific system settings, such as
altering the firewall configurations. The majority of their
infrastructure is based on Windows operating systems. Which of
the following would be the most effective way to achieve this?
**(A)** Use SELinux to enforce strict access controls
**(B) Utilize Group Policy to set and enforce policies
related to system settings
(C)** Deploy a third-party software solution to lock system
settings
**(D)** Implement a user training program to guide users on
system settings best practices

**Explanation 370. Correct Answer: B. Utilize Group Policy to
set and enforce policies related to system settings.** Group
Policy provides centralized management and configuration of
operating systems, applications, and users’ settings in an Active
Directory environment. This can be used to enforce specific
policies related to system settings in a Windows environment.

**Option A is incorrect.** SELinux (Security-Enhanced Linux) is
primarily used in Linux environments to enforce mandatory
access controls. It wouldn’t be appropriate for a primarily
Windows-based infrastructure.


**Option C is incorrect.** While third-party solutions may exist,
leveraging built-in features like Group Policy is often more
streamlined, cost-effective, and efficient.

**Option D is incorrect.** While user training is essential, relying
solely on training without implementing technical controls is
not effective in preventing users from altering system settings.

**Question 371.** A company plans to upgrade its email server to
ensure that email transmission between their mail server and
client applications is encrypted. Which of the following
protocols would be the most appropriate for this purpose?
**(A)** HTTP
**(B)** FTP
**(C) IMAP over SSL/TLS
(D)** SNMP

**Explanation 371. Correct Answer: C. IMAP over SSL/TLS.**
IMAP over SSL/TLS (commonly known as IMAPS) is a
protocol used by email clients to retrieve messages from a mail
server over a secure channel. The use of SSL/TLS provides
encryption for the transmitted data, ensuring confidentiality of
the email transmission.

**Option A is incorrect.** HTTP is a protocol for transferring web
content and does not provide encryption by default.
Furthermore, it’s not designed for email transmission.

**Option B is incorrect.** FTP is a protocol for transferring files
and does not provide secure email transmission capabilities.


**Option C is incorrect.** SNMP (Simple Network Management
Protocol) is used for network management and monitoring, not
for email transmission.

**Question 372.** Lisa, a security administrator, is using a popular
benchmark to ensure the web servers in her organization are
configured securely. She wants to make sure that unnecessary
services are disabled, and appropriate permissions are set.
Which of the following organizations is MOST likely the source
of the benchmark she is using?
**(A)** PCI DSS
**(B)** OWASP
**(C) CIS
(D)** GDPR

**Explanation 372. Correct Answer: C. CIS.** The Center for
Internet Security (CIS) is known for its CIS Benchmarks, which
provide prescriptive guidance for configuring systems securely.

**Option A is incorrect.** PCI DSS is a set of security standards
designed to ensure that all companies that accept, process, store,
or transmit credit card information maintain a secure
environment.

**Option B is incorrect.** OWASP (Open Web Application
Security Project) is known for its top ten list of web application
vulnerabilities and not for system configuration benchmarks.

**Option D is incorrect.** GDPR (General Data Protection
Regulation) is a regulation that requires businesses to protect


the personal data and privacy of EU citizens. It doesn’t provide
benchmarks for secure system configurations.

**Question 373.** The IT department at TechCorp Ltd has been
instructed to ensure that critical system files remain unchanged
to avoid potential security breaches. They want to implement a
system that can provide alerts whenever there is an
unauthorized change to these files. Which of the following
would best serve this purpose?
**(A)** Data Loss Prevention (DLP)
**(B)** Intrusion Detection System (IDS)
**(C) File Integrity Monitoring (FIM)
(D)** Remote Monitoring and Management (RMM)

**Explanation 373. Correct Answer: C. File Integrity
Monitoring (FIM).** File Integrity Monitoring (FIM) tools
monitor and detect changes in files that may indicate a breach,
malware infection, or other unauthorized activity. By tracking
and reporting on file change patterns, FIM helps ensure the
integrity and confidentiality of data.

**Option A is incorrect.** Data Loss Prevention (DLP) focuses on
preventing unauthorized data transfers or leaks from the
organization, not on monitoring file changes.

**Option B is incorrect.** An Intrusion Detection System (IDS)
primarily detects unauthorized access or attacks on a network.
While it can identify some unauthorized file changes, it isn’t
specifically designed for file integrity monitoring.


**Option D is incorrect.** Remote Monitoring and Management
(RMM) tools primarily allow IT professionals to monitor and
manage endpoints, networks, and computers remotely. They
don’t specifically focus on file integrity.

**Question 374.** ABC Tech has a mixed environment with both
Linux and Windows servers. They want to ensure that processes
running on their Linux servers only have access to specific
resources and are restricted from performing certain actions.
Which of the following tools would be most appropriate for this
task?
**(A)** Use Group Policy on their Windows servers and apply it
to Linux servers
**(B)** Implement a strict user training regimen to inform users
about security best practices
**(C) Enable Security-Enhanced Linux (SELinux) in
enforcing mode
(D)** Limit user access to Linux servers

**Explanation 374. Correct Answer: C. Enable Security-
Enhanced Linux (SELinux) in enforcing mode.** SELinux
provides mandatory access controls by limiting the actions a
process can perform to only those defined in its policy, ensuring
it only has access to specific resources. By setting SELinux in
enforcing mode, these policies are actively enforced.

**Option A is incorrect.** Group Policy is specific to Windows and
cannot be applied directly to Linux servers.


**Option B is incorrect.** While user training is a good security
practice, it does not provide the technical control required to
restrict processes on Linux servers.

**Option D is incorrect.** Limiting user access to Linux servers
does not directly control the actions of processes running on
those servers.

**Question 375.** Sarah, a security administrator, is implementing
a monitoring solution for her organization’s server
infrastructure. She wants a solution that does not require any
additional software to be installed on the servers themselves.
Which type of monitoring approach should Sarah choose?
**(A)** Agent-based monitoring
**(B)** Intrusion Detection System (IDS)
**(C) Agentless monitoring
(D)** Network-based Application Performance Monitoring
(APM)

**Explanation 375. Correct Answer: C. Agentless monitoring.**
Agentless monitoring solutions do not require any software
agents to be installed on the servers they monitor. They
typically rely on standard protocols and interfaces to collect
performance and other data.

**Option A is incorrect.** Agent-based monitoring requires the
installation of software agents on each server or device that is to
be monitored.


**Option B is incorrect.** While an Intrusion Detection System
(IDS) is a security tool, it is not specifically designed for
agentless infrastructure monitoring.

**Option D is incorrect.** Network-based Application
Performance Monitoring (APM) is focused on monitoring
application performance on the network, not on agentless server
monitoring.

**Question 376.** PharmaCorp, a pharmaceutical company, wants
to ensure that its researchers cannot transfer proprietary
formulas and research data to external storage devices or cloud
storage. The company needs a solution to prevent such transfers
while allowing other types of data to be transferred. What
should they implement?
**(A)** Web Application Firewall (WAF)
**(B)** Data Encryption Tool
**(C) Data Loss Prevention (DLP)
(D)** Virtual Private Network (VPN)

**Explanation 376. Correct Answer: C. Data Loss Prevention
(DLP).** Data Loss Prevention (DLP) is designed to detect
potential data breach attempts and prevent the unauthorized
transfer of sensitive data. By setting up rules and criteria in a
DLP solution, PharmaCorp can restrict the transfer of
proprietary information while allowing other data to be
transferred as needed.

**Option A is incorrect.** A Web Application Firewall (WAF) is
used to protect web applications by filtering and monitoring


HTTP traffic. It does not handle data transfers to external
storage or cloud storage specifically.

**Option B is incorrect.** A Data Encryption Tool encrypts data to
protect it from unauthorized access. While it adds a layer of
security, it doesn’t prevent the transfer of data to external
devices or cloud storage.

**Option D is incorrect.** A Virtual Private Network (VPN)
provides a secure tunnel for data transmission between two
endpoints. It does not prevent or monitor the type of data being
transferred.

**Question 377.** CyberFirm has been facing issues with phishing
campaigns where attackers spoof their domain to send
fraudulent emails. They already implemented DKIM to sign
their emails but want an additional measure to specify which
mail servers are authorized to send emails on behalf of their
domain. Which security measure should CyberFirm adopt?
**(A) SPF
(B)** PGP
**(C)** SSL certificate
**(D)** IMAP

**Explanation 377. Correct Answer: A. SPF.** Sender Policy
Framework (SPF) is designed to prevent email spoofing. It
allows domain owners to specify which mail servers are
authorized to send emails on behalf of their domain. Receiving
servers can then verify the sending server against the SPF
record in the domain’s DNS.


**Option B is incorrect.** PGP (Pretty Good Privacy) is used for
encrypting and decrypting texts, emails, files, directories, and
whole disk partitions, but it doesn’t define authorized mail
servers for a domain.

**Option C is incorrect.** An SSL certificate provides a secure
connection between a user’s web browser and a website,
ensuring data confidentiality. It’s not specifically designed to
validate the sending mail servers of a domain.

**Option D is incorrect.** IMAP (Internet Message Access
Protocol) is a protocol used by email clients to retrieve
messages from a mail server. It doesn’t help in specifying
authorized mail servers for a domain.

**Question 378.** A school wants to prevent its students from
accessing inappropriate websites during class hours. The IT
department decides to implement a solution that blocks requests
to specific domain names associated with inappropriate content.
Which of the following security solutions would best address
this need?
**(A)** Firewall filtering based on IP addresses
**(B)** Intrusion Detection System monitoring
**(C)** Virtual Private Network (VPN) enforcement
**(D) DNS filtering with a blacklist**

**Explanation 378. Correct Answer: D. DNS filtering with a
blacklist.** DNS filtering is a technique that can be used to
prevent users from accessing specific websites by blocking
requests to certain domain names. A blacklist can be created to
list the domain names associated with inappropriate content,


and when users try to access these sites, the DNS filter will not
resolve the domain name, thereby preventing access.

**Option A is incorrect.** While firewalls can block traffic based
on IP addresses, this method is not as effective for blocking
specific domain names, especially when a domain can have
multiple IP addresses or if the IPs can change.

**Option B is incorrect.** Intrusion Detection Systems (IDS) are
primarily used for detecting and alerting on malicious activities
but not for blocking access to specific domain names.

**Option C is incorrect.** VPNs are used to create a secure
communication channel and are not inherently designed to
block domain name resolutions.

**Question 379.** A large financial institution recently experienced
a security breach where an attacker was able to bypass its
intrusion detection system (IDS). Upon investigation, the
security team found out that the attacker utilized a zero-day
exploit. In the aftermath, what should the institution do to
enhance the capability of its IDS?
**(A) Switch from a signature-based IDS to a behavior-
based IDS
(B)** Disable the IDS and rely solely on firewall rules
**(C)** Update the IDS with the latest threat intelligence feeds
and signatures
**(D)** Reduce the frequency of IDS signature updates

**Explanation 379. Correct Answer: A. Switch from a
signature-based IDS to a behavior-based IDS.** Behavior-


based IDS (or anomaly-based IDS) monitors network traffic and
compares it against an established baseline, allowing it to
potentially detect zero-day exploits and other novel attacks that
signature-based systems might miss.

**Option B is incorrect.** Relying solely on firewall rules without
IDS would lessen the security depth and expose the institution
to more risks.

**Option C is incorrect.** While updating the IDS with the latest
signatures is important, it wouldn’t necessarily detect zero-day
exploits until a signature for that exploit has been created and
disseminated.

**Option D is incorrect.** Reducing the frequency of IDS
signature updates would make the system even more vulnerable
to recent threats.

**Question 380.** A security analyst at CyberSecure Corp. reviews
a vulnerability report concerning an application that could allow
attackers to upload malicious scripts. Once these scripts are
executed, they can grant attackers complete control over the
application. How should this vulnerability be primarily
classified?
**(A)** Integrity vulnerability
**(B)** Availability vulnerability
**(C) Remote code execution vulnerability
(D)** Disclosure vulnerability

**Explanation 380. Correct Answer: C. Remote code
execution vulnerability.** Remote code execution (RCE)


vulnerabilities allow attackers to run arbitrary commands or
scripts on a targeted system. In the scenario, the application
vulnerability grants attackers the ability to upload and execute
malicious scripts, which aligns with this classification.

**Option A is incorrect.** While the vulnerability can affect the
integrity of the application, the primary concern here is the
ability of attackers to execute code remotely, making it an RCE
vulnerability.

**Option B is incorrect.** Availability vulnerabilities primarily
concern the disruption of services or systems, preventing
authorized users from accessing them. The scenario doesn’t
indicate any disruption of availability.

**Option D is incorrect.** Disclosure vulnerabilities involve
unauthorized viewing or accessing of information. The
described vulnerability goes beyond mere disclosure, allowing
execution of malicious scripts.

**Question 381.** After a recent security incident in the
organization, the IT team noticed that several legitimate
activities were being flagged by the intrusion detection system,
resulting in a high number of false positives. What is the MOST
appropriate action to improve the system’s accuracy and reduce
unnecessary alerts?
**(A)** Disable the intrusion detection system for a week to
observe regular network traffic patterns
**(B)** Set up a stricter firewall rule to block all external traffic
**(C) Implement alert tuning to refine the system's**


**detection criteria
(D)** Encourage employees to reduce their internet usage

**Explanation 381. Correct Answer: C. Implement alert
tuning to refine the system’s detection criteria.** Alert tuning
involves adjusting the detection rules or criteria of a system to
better differentiate between legitimate and malicious activities,
thereby reducing false positives.

**Option A is incorrect.** Disabling the intrusion detection system
can expose the organization to real threats, making it a risky
approach.

**Option B is incorrect.** Setting up stricter firewall rules doesn’t
directly address the issue of false positives from the intrusion
detection system and could block legitimate business
operations.

**Option D is incorrect.** Reducing internet usage does not
necessarily correlate with a decrease in false positives. The
issue lies with the system’s criteria, not the amount of traffic.

**Question 382.** ABC Corp has recently faced a security breach
due to a contractor connecting an infected laptop to the
corporate network. Management wants to implement a solution
that would ensure that any device connecting to the corporate
network meets the company’s security standards, including up-
to-date antivirus definitions. Which solution should ABC Corp
consider?
**(A)** Intrusion Detection System (IDS)
**(B)** Virtual Private Network (VPN)


**(C) Network Access Control (NAC)
(D)** Web Application Firewall (WAF)

**Explanation 382. Correct Answer: C. Network Access
Control (NAC).** Network Access Control (NAC) allows
organizations to set policies for device connectivity to corporate
networks. NAC can assess the security posture of a device
before it connects to ensure it meets predefined criteria, such as
updated antivirus definitions, required patches, etc.

**Option A is incorrect.** An Intrusion Detection System (IDS)
monitors network traffic for suspicious activities and issues
alerts but does not evaluate device security postures before
allowing network access.

**Option B is incorrect.** A Virtual Private Network (VPN) allows
secure remote access to a network but does not inherently
evaluate the security posture of devices.

**Option D is incorrect.** A Web Application Firewall (WAF)
focuses on protecting web applications by monitoring and
filtering HTTP traffic. It is not used for evaluating device
security postures.

**Question 383.** Global Corp received a report that some of its
customers received phishing emails that seemed to originate
from the company’s domain. The IT team checked and
confirmed that SPF and DKIM configurations were correctly
set. What additional email security measure can Global Corp
implement to provide clear policies on how the emails should
be treated if they don’t align with SPF and DKIM?


**(A)** Enabling TLS encryption
**(B) Implementing DMARC policies
(C)** Setting up a new SMTP server
**(D)** Increasing email retention period

**Explanation 383. Correct Answer: B. Implementing
DMARC policies.** By implementing DMARC policies, Global
Corp can define how email receivers should handle emails from
its domain that don’t align with the specified SPF and DKIM
records. DMARC can be set to monitor, quarantine, or reject
emails that fail these checks, providing more robust protection
against email spoofing and phishing.

**Option A is incorrect.** While TLS encryption is essential for
protecting the content of email in transit, it doesn’t address the
issue of spoofing or provide guidelines on how to handle emails
that don’t match SPF and DKIM.

**Option C is incorrect.** Setting up a new SMTP server can help
with sending emails, but it doesn’t inherently protect against
email spoofing or provide guidance for emails that don’t align
with SPF and DKIM.

**Option D is incorrect.** Increasing the email retention period
affects how long emails are stored but doesn’t offer protection
against spoofing or guidance for handling misaligned emails.

**Question 384.** A financial firm has just experienced a cyber
attack, and the IT team identified a piece of malware that
evaded their traditional antivirus solutions. The CISO now
wants to not only detect but also be able to analyze and respond


to such advanced threats in real-time. Which solution should the
firm consider implementing?
**(A)** Vulnerability Scanner
**(B)** Intrusion Prevention System (IPS)
**(C) Endpoint Detection and Response (EDR)
(D)** Patch Management System

**Explanation 384. Correct Answer: C. Endpoint Detection
and Response (EDR).** Endpoint Detection and Response
(EDR) provides real-time monitoring and analysis of endpoint
events, allowing an organization to detect, investigate, and
respond to potential security threats. EDR tools can identify
behaviors that might indicate advanced threats that evade
traditional antivirus solutions.

**Option A is incorrect.** Vulnerability Scanners are used to
identify vulnerabilities in a system or network but do not
provide real-time monitoring and response capabilities for
threats.

**Option B is incorrect.** Intrusion Prevention Systems (IPS)
monitor network traffic to prevent potential threats, but they
might not provide in-depth analysis and response at the
endpoint level like EDR solutions do.

**Option D is incorrect.** Patch Management Systems are used to
manage the distribution and installation of software updates but
do not offer real-time threat detection and response.

**Question 385.** After the recent cyber-attack on Acme Corp, the
IT security team decided to enhance their proactive defense


mechanism. They want to start with identifying unpatched and
vulnerable systems on their network. Which of the following
scanning activities would BEST assist them in this endeavor?
**(A)** Conducting a passive scan during business hours
**(B)** Implementing a full open port scan on all systems
**(C) Running a credentialed vulnerability scan on their
network
(D)** Scanning the external perimeter for domain name
resolutions

**Explanation 385. Correct Answer: C. Running a
credentialed vulnerability scan on their network.** A
credentialed vulnerability scan uses valid user credentials to
access and scan the target system, allowing for a deeper and
more comprehensive check for vulnerabilities, including
unpatched systems.

**Option A is incorrect.** A passive scan is non-intrusive and only
monitors network traffic, limiting its capability to identify
unpatched systems actively.

**Option B is incorrect.** While a full open port scan can identify
open ports, it doesn’t necessarily identify unpatched systems or
specific vulnerabilities.

**Option D is incorrect.** Scanning the external perimeter for
domain name resolutions can help in gathering information
about domain names but won’t directly assist in identifying
unpatched systems.


**Question 386.** A software developer in a company notices that a
legitimate software tool they use is repeatedly flagged and
quarantined by the company’s security solution. Which of the
following is the BEST action the cybersecurity team can take to
address this without compromising security?
**(A)** Turn off the antivirus solution
**(B) Whitelist the software tool in the antivirus settings
(C)** Decrease the security level of the antivirus
**(D)** Install a different antivirus solution

**Explanation 386. Correct Answer: B. Whitelist the software
tool in the antivirus settings.** Whitelisting allows the
cybersecurity team to specify software or applications that are
considered safe and should not be flagged or quarantined by the
antivirus solution.

**Option A is incorrect.** Turning off the antivirus solution would
leave the system vulnerable to malware and other malicious
threats.

**Option C is incorrect.** Decreasing the security level of the
antivirus might reduce its effectiveness in detecting and
blocking genuine threats.

**Option D is incorrect.** Simply installing a different antivirus
solution does not guarantee that the tool won’t be flagged again,
and frequent switches can also be costly and time-consuming.

**Question 387.** AlphaTech, a growing SaaS company, has
multiple applications deployed across different cloud providers.
The security team struggles to manage and analyze logs from


these disparate sources. Which solution would BEST help
AlphaTech centralize their logs for a more streamlined analysis?
**(A)** Network Intrusion Detection System (NIDS)
**(B) Log Aggregation Tool
(C)** Data Loss Prevention (DLP) software
**(D)** Vulnerability Scanner

**Explanation 387. Correct Answer: B. Log Aggregation Tool.**
Log aggregation tools are specifically designed to gather,
centralize, and manage logs from various sources, making it
easier to analyze and correlate events.

**Option A is incorrect.** While a NIDS can help detect malicious
activity on a network, it doesn’t centralize logs from different
application sources.

**Option C is incorrect.** DLP software focuses on preventing
unauthorized data transfers and does not serve the purpose of
centralizing logs.

**Option D is incorrect.** Vulnerability Scanners are designed to
identify vulnerabilities in a system but don’t aggregate logs
from various sources.

**Question 388.** BetaTech, a tech manufacturing firm, wants to
ensure that a potential compromise of its IoT devices will not
endanger its primary manufacturing control systems. Which of
the following approaches would be most effective in achieving
this?
**(A)** Using a single robust firewall for the entire network
**(B)** Periodic password changes for IoT devices


**(C) Segmenting the IoT devices from the manufacturing
control systems
(D)** Enabling automatic updates for all IoT devices

**Explanation 388. Correct Answer: C. Segmenting the IoT
devices from the manufacturing control systems.** By
segmenting the IoT devices from the primary manufacturing
control systems, BetaTech ensures that a compromise of the IoT
devices doesn’t immediately put the control systems at risk.
Segmentation acts as a barrier to restrict the potential spread of
malicious activity.

**Option A is incorrect.** While a robust firewall is crucial for
network security, it does not replace the need for segmentation,
especially with varied devices and risk profiles.

**Option B is incorrect.** Although periodic password changes
can enhance the security of IoT devices, it does not prevent a
compromised IoT device from affecting other parts of the
network.

**Option D is incorrect.** Automatic updates can fix known
vulnerabilities in IoT devices, but they don’t provide the
isolation that segmentation offers to prevent a compromise from
affecting other network segments.

**Question 389.** A global manufacturing company wants to
ensure its employees worldwide do not access websites
promoting hate speech, gambling, or explicit content during
working hours. To meet this requirement, which web filtering
technique would be the most efficient?


**(A)** Deploy a centralized proxy with location-based filtering
**(B)** Use a blacklist of specific URLs known to contain such
content
**(C) Implement content categorization and block
undesired categories
(D)** Monitor internet usage logs and reprimand violators

**Explanation 389. Correct Answer: C. Implement content
categorization and block undesired categories.** Content
categorization is a technique where websites are categorized
based on their content type, such as “gaming,” “social media,”
“news,” etc. By using this technique, organizations can block
entire categories of content, such as “hate speech” or
“gambling,” ensuring broad coverage without the need to
identify every problematic URL individually.

**Option A is incorrect.** While centralized proxies can help
manage internet access, location-based filtering focuses more
on geographic locations rather than content categories.

**Option B is incorrect.** Relying solely on a blacklist can be
inefficient, as it requires constant updates to catch every
possible harmful URL, and it might not cover newly created
sites quickly enough.

**Option D is incorrect.** Simply monitoring internet usage logs
and reprimanding violators is reactive and does not proactively
prevent access to undesired content.

**Question 390.** TechCo, a medium-sized enterprise, is planning
to implement a solution to monitor, control, and restrict web


access for its employees to improve productivity and enhance
security. They also want to cache frequently accessed web
content to reduce bandwidth consumption. Which solution
would BEST fit TechCo’s requirements?
**(A)** Deploy a decentralized proxy on each departmental
network
**(B)** Set up a DNS-based filtering service
**(C) Use a centralized proxy with caching capabilities
(D)** Recommend browser extensions for web filtering to all
employees

**Explanation 390. Correct Answer: C. Use a centralized
proxy with caching capabilities.** A centralized proxy allows
organizations to filter, monitor, and control web access from a
central point. Additionally, proxies with caching capabilities can
store frequently accessed web content, reducing the need for
repeated downloads and thus saving bandwidth.

**Option A is incorrect.** A decentralized approach would make it
challenging to consistently apply and enforce policies across the
enterprise. It may also lead to inefficiencies in bandwidth usage
as content caching would not be centralized.

**Option B is incorrect.** While a DNS-based filtering service can
help in blocking access to certain malicious or inappropriate
sites, it does not offer the centralized control or caching
capabilities that a proxy does.

**Option D is incorrect.** Relying on browser extensions is a
decentralized method that depends on user compliance and may
not provide centralized control or caching capabilities.


**Question 391.** An organization is planning to deploy a new web
application that will be accessible from both the internal
network and the internet. The application will communicate
exclusively over HTTPS. The security administrator is asked to
configure the firewall to allow the necessary traffic. Which of
the following should the administrator configure?
**(A)** Allow port 21 and block all others
**(B) Allow port 443 and block all others
(C)** Allow port 80 and block all others
**(D)** Allow port 23 and block all others

**Explanation 391. Correct Answer: B. Allow port 443 and
block all others.** HTTPS primarily uses port 443 for secure
communication. Thus, allowing port 443 and blocking all others
would ensure the secure operation of the web application.

**Option A is incorrect.** Port 21 is used for FTP (File Transfer
Protocol), which is not relevant to HTTPS communication.

**Option C is incorrect.** While port 80 is used for HTTP, it does
not provide the encryption that HTTPS does on port 443.

**Option D is incorrect.** Port 23 is used for Telnet, which is
unrelated to secure web communication.

**Question 392.** The company’s security administrator observes
that there are multiple unauthorized access attempts originating
from IP addresses in a specific range. The administrator wants
to prevent these IP addresses from accessing the corporate
network temporarily. Which of the following firewall
configurations would BEST address this requirement?


**(A) Configure an implicit deny rule for the specific IP
range
(B)** Set up a honeypot for the specific IP range
**(C)** Allow the IP range but set a bandwidth limit
**(D)** Add the IP range to a whitelist

**Explanation 392. Correct Answer: A. Configure an implicit
deny rule for the specific IP range.** By configuring an implicit
deny rule for that specific IP range, the firewall will block any
traffic from those addresses, preventing them from accessing
the network.

**Option B is incorrect.** While a honeypot can be used to
monitor and analyze attacker behavior, it does not block the
access of the specified IP range to the corporate network.

**Option C is incorrect.** Allowing the IP range and setting a
bandwidth limit would not prevent access; it would only restrict
the amount of data they could send/receive.

**Option D is incorrect.** Adding the IP range to a whitelist would
grant them access, which is opposite to the required action.

**Question 393.** The IT department of Globex Corp is concerned
about the increasing number of malicious websites being
accessed from company laptops while employees are working
remotely. They want to ensure that the web filter policies set in
the corporate network are enforced even when devices are
offsite. What would be the BEST solution to address this
concern?
**(A)** Implement a cloud-based web filtering solution


**(B)** Use a VPN to force all remote traffic through the
corporate network
**(C) Deploy an agent-based web filter on all company
laptops
(D)** Periodically send reminders to employees about
acceptable web usage

**Explanation 393. Correct Answer: C. Deploy an agent-based
web filter on all company laptops.** Agent-based web filters
can enforce web filtering policies on a device regardless of its
location. This ensures that the policies apply consistently
whether the device is on or off the corporate network.

**Option A is incorrect.** While cloud-based solutions can offer
offsite filtering, they might not be as consistent as an agent-
based solution that directly enforces corporate policies on the
device itself.

**Option B is incorrect.** Using a VPN would force all traffic
through the corporate network, which could cause latency and
might not be feasible for all remote work scenarios.

**Option D is incorrect.** Sending reminders is a passive approach
and may not effectively prevent access to malicious websites.

**Question 394.** Lisa, a cybersecurity analyst, is setting up a
centralized system to correlate logs from multiple sources,
detect malicious activities in real-time, and produce
comprehensive security reports. Which tool should Lisa
consider for this purpose?
**(A)** Network Intrusion Detection System (NIDS)


**(B)** Web Application Firewall (WAF)
**(C)** Vulnerability Scanner
**(D) Security Information and Event Management
(SIEM)**

**Explanation 394. Correct Answer: D. Security Information
and Event Management (SIEM).** SIEM tools are designed to
aggregate, correlate, and analyze logs and events from various
sources in an organization. They help in detecting and
responding to security incidents in real-time and generating
detailed security reports.

**Option A is incorrect.** While NIDS monitors and analyzes
network traffic for signs of malicious activities, it doesn’t
provide centralized logging and reporting functionalities like a
SIEM.

**Option B is incorrect.** A Web Application Firewall (WAF)
protects web applications from web-based attacks. It does not
offer the centralized log correlation and analysis features of a
SIEM.

**Option C is incorrect.** A Vulnerability Scanner identifies
vulnerabilities in systems and applications but doesn’t aggregate
and analyze logs from various sources.

**Question 395.** Lucy, the IT security manager of a financial
company, receives an automated alert that an employee
attempted to email a document containing social security
numbers to an external email address. Which of the following
tools most likely generated this alert?


**(A)** Network Intrusion Detection System (NIDS)
**(B) Data Loss Prevention (DLP) solution
(C)** Vulnerability Scanner
**(D)** Packet Analyzer

**Explanation 395. Correct Answer: B. Data Loss Prevention
(DLP) solution.** DLP solutions are specifically designed to
monitor and control data transfers across an organization’s
network. In this case, the DLP detected sensitive data—social
security numbers—being sent outside of the organization and
alerted Lucy.

**Option A is incorrect.** NIDS detects and alerts on malicious
activity on a network, but it doesn’t typically scan for specific
data types being transferred.

**Option C is incorrect.** Vulnerability scanners identify and
report on vulnerabilities in a system but don’t monitor data
transfers.

**Option D is incorrect.** While packet analyzers can capture and
analyze network traffic, they don’t inherently generate alerts
based on the specific content of data being transferred.

**Question 396.** An online banking platform wants to improve its
customer verification process when users open a new account.
Which of the following identity proofing methods would be the
MOST secure for this purpose?
**(A)** Asking users to select a security question and answer
from a list
**(B) Requiring users to upload a photo of a government-**


**issued ID and a selfie
(C)** Sending a verification code to the user's email address
**(D)** Prompting users to provide their favorite color

**Explanation 396. Correct Answer: B. Requiring users to
upload a photo of a government-issued ID and a selfie.** This
method provides a high level of assurance by comparing a
user’s live image (selfie) with a government-issued ID. The
combination ensures the person is who they claim to be.

**Option A is incorrect.** Security questions, especially from a
pre-defined list, can be easily guessed or obtained by attackers.

**Option C is incorrect.** Verification codes via email can
enhance identity proofing but aren’t as robust as checking
against a government ID and selfie.

**Option D is incorrect.** Personal preferences, like favorite
colors, are weak indicators of identity and can be easily
guessed.

**Question 397.** A company has recently noticed an increased
number of employees accessing social media sites during work
hours, leading to decreased productivity. To counter this, the
security administrator decides to limit access to these websites
during peak working hours. Which firewall rule modification
should the administrator make?
**(A)** Implement an Intrusion Prevention System (IPS) rule to
block social media content
**(B) Change the firewall rule to deny access to known
social media IP addresses between 9 AM and 5 PM**


**(C)** Use the firewall's URL filtering capability to blacklist
social media URLs
**(D)** Increase the firewall's bandwidth to accommodate the
excess traffic

**Explanation 397. Correct Answer: B. Change the firewall
rule to deny access to known social media IP addresses
between 9 AM and 5 PM.** Implementing a time-based rule that
denies access to specific IP addresses (or ranges) associated
with social media can be an effective way to restrict access
during specified hours.

**Option A is incorrect.** An IPS is designed to detect and prevent
malicious activities based on signatures. Blocking social media
content is not typically its primary function.

**Option C is incorrect.** URL filtering would block access to the
URLs entirely. The requirement is to block them only during
specific hours.

**Option D is incorrect.** Increasing the firewall’s bandwidth
doesn’t address the problem of employees accessing social
media during work hours.

**Question 398.** A company wants to host a public-facing website
but ensure that even if the website gets compromised, attackers
cannot gain access to sensitive internal data. Which of the
following is the BEST configuration to achieve this?
**(A)** Place the web server on the internal network and strictly
monitor the traffic
**(B) Place the web server in the DMZ with a firewall in**


**front of it and another firewall between the DMZ and the
internal network
(C)** Directly connect the web server to the internet without a
firewall and move sensitive data off the server
**(D)** Place the web server in the DMZ and connect it directly
to the internal network without a firewall

**Explanation 398. Correct Answer: B. Place the web server
in the DMZ with a firewall in front of it and another firewall
between the DMZ and the internal network.** By placing the
web server in a DMZ, and having two firewalls (one facing the
internet and another facing the internal network), the company
can ensure that even if the public-facing web server is
compromised, the attacker would still need to bypass another
firewall to reach the internal network.

**Option A is incorrect.** Placing a public-facing web server on
the internal network, even with monitoring, exposes the
network to unnecessary risks.

**Option C is incorrect.** Connecting a server directly to the
internet without any form of firewall is highly risky, even if
sensitive data is moved.

**Option D is incorrect.** Without a firewall between the DMZ
and the internal network, it becomes much easier for a
compromised server in th **e** DMZ to impact or access the internal
network.

**Question 399.** A Security Analyst at BetaTech is reviewing the
monitoring tools deployed across the organization. She wants to


ensure that every tool can detect unauthorized changes made to
system files and configurations. Which of the following tools is
BEST suited for this purpose?
**(A)** Network protocol analyzer
**(B) File integrity monitoring (FIM) system
(C)** Bandwidth monitoring tool
**(D)** Passive vulnerability scanner

**Explanation 399. Correct Answer: B. File integrity
monitoring (FIM) system.** File integrity monitoring systems
are designed to detect and alert when unauthorized changes are
made to system files and configurations, ensuring the integrity
of these critical components.

**Option A is incorrect.** While network protocol analyzers can
capture and analyze network traffic, they are not specifically
designed to monitor changes to system files and configurations.

**Option C is incorrect.** Bandwidth monitoring tools primarily
track network usage and bandwidth consumption, not
modifications to system files.

**Option D is incorrect.** Passive vulnerability scanners monitor
network traffic to detect vulnerabilities but do not actively track
changes to system files.

**Question 400.** A company has noticed an increase in malware
infections over the past month. After investigating, it was
determined that the infections were caused by employees
visiting websites that were newly registered but had malicious
intent. Which of the following would be the BEST approach to


mitigate this threat?
**(A)** Implement a block rule to deny access to all websites
**(B) Use a web filter that incorporates domain reputation
checks and blocks domains registered recently
(C)** Set the web filter to block all websites not categorized
as "Business"
**(D)** Enforce multi-factor authentication for all internet-
based applications

**Explanation 400. Correct Answer: B. Use a web filter that
incorporates domain reputation checks and blocks domains
registered recently.** Reputation-based web filters can evaluate
the trustworthiness of domains. One common heuristic is to be
suspicious of newly registered domains, as cybercriminals often
use these for phishing or malware distribution.

**Option A is incorrect.** Blocking access to all websites is an
extreme measure that would hinder business operations and
employee productivity.

**Option C is incorrect.** Simply blocking websites not
categorized as “Business” does not specifically target the threat
of newly registered malicious domains. Additionally, some
business-relevant websites might not be categorized properly.

**Option D is incorrect.** While multi-factor authentication can
enhance security, it does not address the threat of employees
visiting malicious websites.

**Question 401.** At AlphaTech, the security team is assessing
vulnerabilities in a newly deployed cloud infrastructure. While


analyzing potential risks, they consider factors such as the
physical location of data centers, local laws and regulations, and
natural disaster frequencies. What are these considerations
known as in the context of vulnerability management?
**(A)** Asset valuation factors
**(B)** Risk response variables
**(C)** Threat intelligence variables
**(D) Environmental variables**

**Explanation 401. Correct Answer: D. Environmental
variables.** Environmental variables in vulnerability
management refer to external factors that can influence or affect
the security posture of an organization. These can include
physical location, local laws and regulations, and the potential
for natural disasters.

**Option A is incorrect.** Asset valuation factors primarily deal
with determining the value of an asset to the organization and
don’t typically consider external factors like local laws or
natural disaster frequencies.

**Option B is incorrect.** Risk response variables pertain to the
organization’s strategies and actions to respond to identified
risks rather than the external factors influencing risk.

**Option C is incorrect.** Threat intelligence variables revolve
around information regarding potential threats or threat actors.
They don’t typically encompass physical environment
considerations like those described in the scenario.


**Question 402.** Caroline, a security analyst, receives an alert that
an unfamiliar file has been detected on a mission-critical server.
She suspects it might be malware. What is the BEST immediate
action Caroline should take regarding this potential threat?
**(A)** Delete the file immediately to prevent further damage
**(B) Quarantine the file to prevent it from executing or
spreading
(C)** Make a copy of the file for further analysis
**(D)** Notify all employees about the suspicious file

**Explanation 402. Correct Answer: B. Quarantine the file to
prevent it from executing or spreading.** Quarantining
suspicious files isolates them, preventing potential execution or
spread while allowing further investigation without immediate
deletion.

**Option A is incorrect.** While deleting the file seems proactive,
it removes the chance for further analysis and can affect
forensic investigations.

**Option C is incorrect.** Making a copy is important for analysis,
but the immediate priority should be to prevent potential
execution or spread of the suspicious file. Quarantine first, then
analyze.

**Option D is incorrect.** Notifying all employees about a specific
suspicious file may cause panic or confusion. It’s more
appropriate to manage the incident first and then communicate
relevant information in a structured manner.


**Question 403.** Jennifer, an IT administrator, is asked to onboard
a new remote employee for a sales role. Which of the following
is the BEST approach for provisioning the user account?
**(A)** Assign the new user the same access privileges as the
CEO because they might require all resources
**(B)** Provide the new user with administrative rights to
ensure they can install and configure any needed software
**(C) Use the access privileges from a template of a
salesperson to provide the required resources
(D)** Allow the new user to decide and self-select the
necessary access based on their job role

**Explanation 403. Correct Answer: C. Use the access
privileges from a template of a salesperson to provide the
required resources.** Provisioning users based on role templates
ensures that users have just the access they need, adhering to the
principle of least privilege.

**Option A is incorrect.** This goes against the principle of least
privilege and can introduce significant security risks.

**Option B is incorrect.** Giving administrative rights to
employees without a proper need can expose the organization to
unnecessary risks.

**Option D is incorrect.** Users should not self-select access
rights as they might not be aware of potential security
implications.

**Question 404.** AlphaTech, a leading IT company, recently
identified a critical vulnerability in its primary software product.


They have developed a patch to address the vulnerability.
Before distributing the patch to its customers, which of the
following should AlphaTech ideally perform?
**(A)** Deploy the patch on all company systems
**(B)** Notify the media about the vulnerability
**(C) Test the patch in a controlled environment
(D)** Offer compensation to affected customers

**Explanation 404. Correct Answer: C. Test the patch in a
controlled environment.** Before deploying or distributing a
patch, especially for a critical vulnerability, it’s essential to test
it in a controlled environment. This ensures that the patch
doesn’t introduce new issues and that it effectively addresses
the vulnerability.

**Option A is incorrect.** Deploying the patch immediately on all
company systems without testing it could lead to unforeseen
issues or even exacerbate the problem.

**Option B is incorrect.** Notifying the media about the
vulnerability, especially before it’s been effectively addressed
and without a coordinated disclosure plan, can lead to panic and
potential exploitation by malicious actors.

**Option D is incorrect.** Offering compensation is reactive and
doesn’t directly address the vulnerability. The primary goal after
identifying a vulnerability should be to address and mitigate it.

**Question 405.** After a major security incident, DeltaTech
implemented several security patches to address vulnerabilities
in their infrastructure. To ensure the effectiveness of these


patches, what should be DeltaTech’s primary next step?
**(A)** Deploy additional firewalls at the network perimeter
**(B)** Provide cybersecurity training to all employees
**(C) Rescan the systems to check if vulnerabilities are
effectively addressed
(D)** Change all user passwords across the organization

**Explanation 405. Correct Answer: C. Rescan the systems to
check if vulnerabilities are effectively addressed.** After
implementing security patches, it’s essential to rescan the
systems to ensure that the identified vulnerabilities have been
effectively addressed and the patches have been implemented
correctly.

**Option A is incorrect.** While firewalls are crucial for security,
deploying them is not directly related to validating the
effectiveness of newly implemented patches.

**Option B is incorrect.** Cybersecurity training for employees is
vital, but it doesn’t directly validate the success of the applied
patches.

**Option D is incorrect.** Changing user passwords can be a
necessary step after a breach, but it doesn’t validate if the
patches have successfully addressed vulnerabilities.

**Question 406.** An e-commerce company is rolling out a new
web application to facilitate online payments. The IT
department wants to be immediately notified of any application
errors or unauthorized modifications to the application’s
codebase. Which of the following tools should they implement?


**(A)** Web Application Firewall (WAF)
**(B) Application Performance Monitoring (APM)
(C)** Domain Name System (DNS) monitoring tool
**(D)** Network flow analyzer

**Explanation 406. Correct Answer: B. Application
Performance Monitoring (APM).** APM tools are designed to
monitor the performance of applications and can detect
application errors, anomalies, and unauthorized code changes,
thereby ensuring application stability and security.

**Option A is incorrect.** While a WAF protects web applications
from various cyber threats by filtering and monitoring HTTP
traffic, it doesn’t typically monitor for application errors or
unauthorized code changes.

**Option C is incorrect.** DNS monitoring tools focus on ensuring
the availability and integrity of DNS services and do not
monitor application performance or codebase changes.

**Option D is incorrect.** Network flow analyzers examine data
flows on the network but do not specifically monitor application
performance or codebase modifications.

**Question 407.** Paul, a network administrator, has configured
various networking devices in his organization to send alerts in
the event of specific failures. After a switch experienced a
power supply failure, Paul received an immediate notification.
Which of the following did Paul most likely utilize to receive
this notification?
**(A)** Syslog server


**(B) Simple Network Management Protocol (SNMP)
traps
(C)** Packet sniffer
**(D)** Firewall logs

**Explanation 407. Correct Answer: B. Simple Network
Management Protocol (SNMP) traps.** SNMP traps are
unsolicited alert messages sent by a device to notify an SNMP
management station of specific events. In this case, the switch
sent an SNMP trap to Paul when it detected the power supply
failure.

**Option A is incorrect.** While a Syslog server can be used to
collect logs from various devices, it doesn’t proactively send
alerts based on specific events like an SNMP trap does.

**Option C is incorrect.** A packet sniffer captures and analyzes
network traffic but doesn’t actively alert administrators to
specific device events.

**Option D is incorrect.** Firewall logs are specific to firewalls
and track traffic that passes through the firewall, but they don’t
typically send unsolicited alerts about network device health.

**Question 408.** DeltaCorp, a retail company, has assessed that a
security breach might result in a loss of $1 million in sales. The
company has determined that they can tolerate a loss of up to
$500,000, but anything beyond that would severely impact
operations. To cover the potential financial loss beyond their
tolerance level, they decide to purchase cybersecurity insurance.
Which of the following terms best describes the $500,000


figure?
**(A)** Risk appetite
**(B) Risk threshold
(C)** Risk capacity
**(D)** Risk assessment

**Explanation 408. Correct Answer: B. Risk threshold.** The
risk threshold is the specific level of risk an organization is
willing to accept. In this case, DeltaCorp is willing to accept
potential losses up to $500,000, marking that as their threshold.

**Option A is incorrect.** Risk appetite is a broader term that
reflects the general level of risk an organization is willing to
accept in pursuit of its objectives. It’s more about strategic
intent than specific figures.

**Option C is incorrect.** Risk capacity refers to the total amount
of risk an organization can absorb without significantly
impacting its strategic objectives or viability. It’s a broader
measure than the specific tolerance threshold.

**Option D is incorrect.** Risk assessment is the process of
identifying, analyzing, and evaluating risks. It’s not a specific
figure or threshold.

**Question 409.** Samantha, a security analyst, has been tasked
with creating a monthly report for senior management detailing
the security posture of the company. Which of the following is
the MOST important element to include to ensure the report
effectively communicates the company’s current security status?
**(A)** Detailed technical logs of all security incidents


**(B) Graphical representation of incidents by category
(C)** A complete list of all users and their access levels
**(D)** Copies of recent phishing emails for demonstration

**Explanation 409. Correct Answer: B. Graphical
representation of incidents by category.** A graphical
representation by category allows senior management to quickly
understand the types and frequency of security incidents, which
can help in decision-making and resource allocation.

**Option A is incorrect.** While technical logs are crucial for
incident analysis, they may be too detailed and technical for a
senior management report.

**Option C is incorrect.** While it’s essential to manage user
access levels, a complete list of all users and their access might
be excessive for a monthly senior management report focused
on the company’s security posture.

**Option D is incorrect.** While examples of phishing emails can
be educational, they are not crucial for a monthly report meant
to provide an overview of the company’s security status.

**Question 410.** After a recent security incident, Sarah, a network
security analyst, wants to analyze the flow data of network
traffic to identify patterns and potential threats. She wants to
collect metadata about IP traffic flow and gather details like IP
addresses, ports, and protocols used. Which tool should Sarah
employ to obtain this information?
**(A)** Intrusion Detection System (IDS)
**(B)** Syslog server


**(C) NetFlow collector
(D)** Simple Network Management Protocol (SNMP) traps

**Explanation 410. Correct Answer: C. NetFlow collector.**
NetFlow is a network protocol developed by Cisco for
collecting IP traffic information and monitoring network traffic.
A NetFlow collector can provide insights into traffic flow
patterns and volume, making it suitable for Sarah’s
requirements.

**Option A is incorrect.** While an IDS can provide alerts on
malicious activities based on specific signatures or heuristics, it
does not provide detailed flow data analysis like NetFlow.

**Option B is incorrect.** A Syslog server is mainly used for
collecting and storing log data from various devices. It does not
focus on detailed network traffic flow like NetFlow.

**Option D is incorrect.** SNMP traps are for sending unsolicited
alert messages from a device to a management station regarding
specific events. They don’t provide traffic flow analysis.

**Question 411.** A security analyst has been tasked with
investigating a possible data breach. While reviewing the
network logs, the analyst noticed an unusual increase in
outbound traffic to an unfamiliar IP address during non-business
hours. The traffic appears to be encrypted and is associated with
a known server containing sensitive data. Which of the
following is the MOST likely explanation for this behavior?
**(A)** The server is downloading patches
**(B)** An employee is accessing the server remotely


**(C)** A backup of the server is being performed
**(D) Data exfiltration is occurring**

**Explanation 411. Correct Answer: D. Data exfiltration is
occurring.** Given that the traffic is encrypted, associated with a
sensitive server, and is being sent to an unfamiliar IP during
non-business hours, the most likely scenario is that
unauthorized data is being taken out of the network, which is
known as data exfiltration.

**Option A is incorrect.** While servers do download patches,
these are usually inbound traffic from a known update source,
not outbound to unfamiliar IPs.

**Option B is incorrect.** While employees might access servers
remotely, the traffic being encrypted and sent during non-
business hours to an unfamiliar IP makes this less likely.

**Option C is incorrect.** Backups generally don’t result in
encrypted outbound traffic to unfamiliar IP addresses, especially
during non-business hours.

**Question 412.** ExamsDigest Enterprises wants to streamline
their permission assignments. They decide that rather than
assigning permissions to each user individually, they will group
users based on departmental roles and then assign permissions
to these groups. For example, all members of the “Marketing”
role would have access to the marketing database. Which access
control method is ExamsDigest Enterprises employing?
**(A)** Rule-based access control
**(B)** Mandatory Access Control (MAC)


**(C)** Discretionary Access Control (DAC)
**(D) Role-Based Access Control (RBAC)**

**Explanation 412. Correct Answer: D. Role-Based Access
Control (RBAC).** RBAC involves grouping users based on
roles (in this case, departments like “Marketing”) and then
assigning permissions to these roles. Individuals are then placed
into these roles, which determines their access. The scenario
described by ExamsDigest Enterprises is a clear example of
implementing RBAC.

**Option A is incorrect.** Rule-based access control often deals
with predefined rules for access, commonly used in firewalls or
routers, and is not about grouping users based on roles.

**Option B is incorrect.** MAC is more about classifying
information and having users with the appropriate clearance
levels. It doesn’t deal with departmental roles like the scenario
mentioned.

**Option C is incorrect.** DAC allows resource owners to grant or
deny permissions. It does not inherently involve assigning
permissions based on roles or job functions.

**Question 413.** BetaTech is implementing a new authentication
mechanism for its data center technicians. Instead of using key
cards, technicians will now have to look into a device that maps
a specific pattern to authenticate their identity. Which of the
following is BetaTech likely implementing?
**(A)** Password system
**(B) Retina scanning**


**(C)** Hardware token
**(D)** Knowledge-based questions

**Explanation 413. Correct Answer: B. Retina scanning.**
Retina scanning is a biometric method that analyzes the unique
patterns of a person’s retina to authenticate their identity. It’s
categorized under the “something you are” factor as it relies on
a unique physical characteristic.

**Option A is incorrect.** A password system pertains to the
“something you know” factor since users need to remember
their passwords to authenticate.

**Option C is incorrect.** A hardware token represents the
“something you have” factor, as it’s a device the user needs to
possess.

**Option D is incorrect.** Knowledge-based questions, such as
“What’s the name of your first pet?”, fall under the “something
you know” factor.

**Question 414.** A global financial company experiences sporadic
cyber attacks on its infrastructure. The company notices that
attacks that occur during non-business hours often result in
more significant damage due to delayed responses. Which of the
following measures would BEST decrease the reaction time to
these off-hour attacks?
**(A)** Train the security staff to handle larger volumes of
incidents during business hours


**(B) Implement an automated intrusion detection and
response system
(C)** Increase the number of security staff during non-
business hours
**(D)** Send email notifications to security personnel when
attacks are detected

**Explanation 414. Correct Answer: B. Implement an
automated intrusion detection and response system.** By
implementing an automated intrusion detection and response
system, the company can ensure that attacks are detected and
responded to in real-time, regardless of when they occur. This
drastically reduces the reaction time compared to manual
interventions.

**Option A is incorrect.** Training staff to handle more incidents
during business hours doesn’t address the issue of delayed
responses during non-business hours.

**Option C is incorrect.** Increasing staff during non-business
hours may help, but it might not be as efficient or cost-effective
as automation, and there’s still potential for human delay.

**Option D is incorrect.** Sending email notifications might still
lead to delays, especially if staff is not checking emails
promptly during off-hours.

**Question 415.** A digital forensics investigator has just
concluded an investigation regarding a potential insider threat.
Before presenting the findings to the organization’s board,
which of the following should the investigator ensure about the


forensic report?
**(A)** The report includes technical jargon to showcase the
depth of the investigation
**(B)** The report emphasizes the investigator's credentials and
experience
**(C) The report provides a clear, concise summary of
findings without unnecessary technical details
(D)** The report contains detailed logs of every action taken
by the investigator

**Explanation 415. Correct Answer: C. The report provides a
clear, concise summary of findings without unnecessary
technical details.** A forensic report’s primary purpose is to
convey the results of an investigation in a clear and
understandable manner, especially to non-technical
stakeholders. By avoiding unnecessary technical jargon and
providing a concise summary, it ensures that the report’s
findings are accessible to all intended readers.

**Option A is incorrect.** While the depth of an investigation is
important, inundating a report with technical jargon can make it
difficult for non-technical individuals, such as board members,
to understand the findings.

**Option B is incorrect.** While the investigator’s credentials and
experience might be important, they should not be the emphasis
of the report. The focus should remain on the investigation’s
findings and their implications.

**Option D is incorrect.** While maintaining a detailed log of
every action taken is crucial for the investigator’s records and


ensuring the integrity of the investigation, including every detail
in the report can be overwhelming and detract from its main
findings.

**Question 416.** MegaCorp is transitioning to a cloud-based
infrastructure and wants to allow its employees to access
multiple cloud services without re-entering their credentials
every time. They currently have an on-premises LDAP directory
in place. Which approach should MegaCorp take to provide a
seamless authentication experience?
**(A)** MegaCorp should abandon their LDAP directory and
create individual accounts for each cloud service
**(B) Integrate their LDAP with a Single Sign-On (SSO)
solution that supports cloud services
(C)** Store passwords in a plaintext file for users to access
and login to cloud services manually
**(D)** Force users to change passwords every day to enhance
security across all cloud platforms

**Explanation 416. Correct Answer: B. Integrate their LDAP
with a Single Sign-On (SSO) solution that supports cloud
services.** By integrating the on-premises LDAP with an SSO
solution, MegaCorp can leverage its existing user directory to
authenticate and provide access to multiple cloud services
without requiring users to log in separately for each service.

**Option A is incorrect.** Abandoning the existing LDAP
directory would negate the benefits of central management and
create administrative overhead.


**Option C is incorrect.** Storing passwords in plaintext is a
significant security risk and goes against best practices.

**Option D is incorrect.** Forcing users to change passwords daily
is impractical, could result in weaker passwords, and doesn’t
provide a seamless authentication experience.

**Question 417.** At ExamsDigest, employees can access the
company’s cloud-based storage system. However, access to
certain files within the storage is determined by the employee’s
department, job title, and years of service. For instance, senior
managers in the finance department with more than five years
of service can view the company’s financial forecasts. Which
access control model is ExamsDigest using?
**(A)** Rule-Based Access Control (RAC)
**(B)** Role-Based Access Control (RBAC)
**(C) Attribute-Based Access Control (ABAC)
(D)** Discretionary Access Control (DAC)

**Explanation 417. Correct Answer: C. Attribute-Based
Access Control (ABAC).** Attribute-Based Access Control
(ABAC) determines access based on attributes of the user,
resource, and environment. In the scenario, the employee’s
department, job title, and years of service are the attributes that
determine their access to specific files.

**Option A is incorrect.** RAC works based on predefined rules,
typically without involving multiple user attributes like
department or job title.


**Option B is incorrect.** While RBAC is close in functionality, it
assigns permissions based on roles (like “manager” or “clerk”),
not on a combination of attributes.

**Option D is incorrect.** DAC allows resource owners to specify
who can access their resources based on their discretion. It
doesn’t involve a combination of attributes to determine access.

**Question 418.** You are an IT security professional for a large
corporation. After receiving reports about some users being
unable to access external websites, you decided to review the
firewall logs. Which of the following would be a PRIMARY
indicator in the logs that a rule is blocking outbound traffic?
**(A)** Multiple entries of the same external IP address being
ALLOWED
**(B)** Timestamps showing large gaps between entries
**(C) Entries showing DROP/REJECT action for
outbound traffic to port 80 and 443
(D)** Logs showing inbound traffic from multiple unknown
external IP addresses

**Explanation 418. Correct Answer: C. Entries showing
DROP/REJECT action for outbound traffic to port 80 and**

**443.** Port 80 and 443 are standard ports for HTTP and HTTPS
respectively, which are commonly used for accessing websites.
If users are unable to access external websites, it would make
sense to check for DROP or REJECT actions for these ports in
the firewall logs.


**Option A is incorrect.** If an external IP address is being
ALLOWED multiple times, it wouldn’t be the cause of users
being unable to access websites.

**Option B is incorrect.** Large gaps between timestamps in logs
can indicate various issues, but they aren’t a direct indicator of a
specific rule blocking outbound traffic.

**Option D is incorrect.** While inbound traffic from unknown IP
addresses might be of concern, it doesn’t specifically indicate an
outbound traffic rule blocking users from accessing websites.

**Question 419.** DeltaCorp has a password policy in place which
mandates users to change their passwords every 30 days.
However, some users complain that this results in them
choosing simpler passwords or writing them down to remember
them. How can DeltaCorp maintain security while addressing
these concerns?
**(A) Reduce the password change frequency but
introduce more complexity requirements
(B)** Eliminate password changes and rely solely on two-
factor authentication
**(C)** Ask users to change passwords every week to improve
security
**(D)** Allow users to reuse any of their last three passwords to
ease the transition

**Explanation 419. Correct Answer: A. Reduce the password
change frequency but introduce more complexity
requirements.** By reducing the frequency of password changes,
users are less burdened with the task of remembering new


passwords frequently. Adding complexity requirements can
compensate for the longer duration by ensuring that passwords
are strong.

**Option B is incorrect.** Relying solely on two-factor
authentication (2FA) without any password requirements can
introduce risks if the 2FA method is compromised. Additionally,
2FA can be inconvenient for users if not implemented correctly.

**Option C is incorrect.** Asking users to change passwords even
more frequently would exacerbate the issue and likely lead to
even weaker password practices.

**Option D is incorrect.** Allowing users to reuse recent
passwords undermines the purpose of having them change
passwords in the first place and reduces overall security.

**Question 420.** During a review of IDS logs, a security specialist
notices a series of alerts indicating that a single external IP has
been sending payloads that exploit a known vulnerability.
However, the internal system to which these payloads are sent is
patched and is not vulnerable to the exploit. Which of the
following describes this type of IDS alert?
**(A)** False positive
**(B)** False negative
**(C) True positive
(D)** True negative

**Explanation 420. Correct Answer: C. True positive.** A true
positive means the IDS correctly identified malicious or
anomalous traffic. In this case, even though the internal system


isn’t vulnerable, the IDS correctly flagged the traffic because
it’s genuinely malicious.

**Option A is incorrect.** A false positive would mean the IDS
incorrectly flagged benign traffic as malicious, which is not the
case here.

**Option B is incorrect.** A false negative means the IDS failed to
identify malicious or anomalous traffic, but in this scenario, the
IDS did correctly identify the malicious traffic.

**Option D is incorrect.** A true negative means the IDS correctly
identified benign traffic as benign. The traffic in this scenario
was malicious, so this doesn’t apply.

**Question 421.** A popular social media platform allows third-
party applications to access user data and post on behalf of
users. To avoid sharing user passwords with third-party
applications and provide limited, scoped access, which
authentication method should the platform use?
**(A)** Embed user passwords in the application's code
**(B)** Use basic authentication with username and password
for every request
**(C) Implement Single Sign-On (SSO) using OAuth to
provide token-based access
(D)** Rely solely on CAPTCHA for third-party app
authentication

**Explanation 421. Correct Answer: C. Implement Single
Sign-On (SSO) using OAuth to provide token-based access.**
OAuth provides token-based access and lets third-party


applications operate on behalf of users without exposing user
passwords. OAuth tokens can also be scoped to limit the range
of actions a third-party application can perform.

**Option A is incorrect.** Embedding passwords in application
code is insecure and goes against best practices.

**Option B is incorrect.** Basic authentication exposes user
credentials and doesn’t offer the scoped access OAuth provides.

**Option D is incorrect.** CAPTCHA is designed to differentiate
between human and automated access but doesn’t handle
authentication or authorization.

**Question 422.** An IT department in a large corporation spends
several hours each day manually deploying patches and updates
to thousands of workstations. Which of the following solutions
would BEST enhance the efficiency of this process and save
time for the IT team?
**(A)** Disable automatic updates and conduct monthly
patching sessions
**(B) Implement an automated patch management system
(C)** Designate a dedicated team for patching that operates in
shifts
**(D)** Educate users to install updates on their own

**Explanation 422. Correct Answer: B. Implement an
automated patch management system.** By implementing an
automated patch management system, the IT department can
streamline the deployment of patches and updates across all


workstations, ensuring consistency, reducing manual efforts,
and saving valuable time.

**Option A is incorrect.** Disabling automatic updates and
conducting monthly patching sessions does not address the
inefficiency of manual patching and might expose the systems
to vulnerabilities for a longer time.

**Option C is incorrect.** While designating a dedicated team
might distribute the workload, it doesn’t eliminate the
inefficiencies associated with manual patching.

**Option D is incorrect.** Relying on users to install updates
introduces inconsistency, potential delays, and additional risks,
as not all users might have the technical knowledge or
remember to update regularly.

**Question 423.** An international company, GlobalTech, is using
several web applications hosted by different vendors. To ensure
their employees can access these applications without having to
remember multiple sets of credentials, they want to implement a
solution that can securely exchange user authentication
information between the company and the service providers.
What should GlobalTech implement?
**(A)** Integrate each application with an independent LDAP
server
**(B) Implement SSO using Security Assertions Markup
Language (SAML)
(C)** Embed encrypted user credentials within the URL of
each application


**(D)** Rely on public API keys shared between the company
and each vendor

**Explanation 423. Correct Answer: B. Implement SSO using
Security Assertions Markup Language (SAML).** SAML is an
XML-based standard for exchanging authentication and
authorization data between parties. It’s designed to facilitate
single sign-on for web applications. By implementing SSO with
SAML, GlobalTech can allow its employees to authenticate
once and gain access to multiple applications without re-
authenticating.

**Option A is incorrect.** Having independent LDAP servers for
each application defeats the purpose of SSO and complicates
user management.

**Option C is incorrect.** Embedding encrypted user credentials
in URLs is insecure and not a recommended practice.

**Option D is incorrect.** API keys are used for system-to-system
communication and not for user authentication. Moreover,
public API keys shouldn’t be shared recklessly.

**Question 424.** A company wants to implement a solution that
verifies the software integrity of remote servers before allowing
them to connect to the primary network. Which of the following
solutions BEST achieves this objective through attestation?
**(A)** Host-based firewall
**(B)** Whitelisting application
**(C) Remote attestation
(D)** VPN tunneling


**Explanation 424. Correct Answer: C. Remote attestation.**
Remote attestation is a process where a device (like a server)
proves to a remote entity (like a network controller) that it is
running genuine, unmodified software. It allows for the
verification of the software integrity of remote devices before
they connect to a primary network.

**Option A is incorrect.** A host-based firewall is used to control
inbound and outbound network traffic to and from a device
based on a set of configurable rules. It does not verify the
software integrity of the device itself.

**Option B is incorrect.** A whitelisting application only allows
specified software to run on a system. While it can enhance
security by ensuring only approved software runs, it does not
attest to the state or integrity of the device or its software when
connecting to another network.

**Option D is incorrect.** VPN tunneling encrypts the connection
between two points over the internet. While it ensures secure
communication, it doesn’t verify the software integrity of
devices.

**Question 425.** TechCorp is collaborating with SoftTech, a
business partner. To streamline collaboration without managing
multiple accounts, TechCorp wants its employees to use their
existing credentials to access SoftTech’s online project
management system. Which of the following approaches would
BEST enable this functionality?
**(A)** TechCorp should create new accounts for its employees
on SoftTech's system


**(B)** SoftTech should allow anonymous access for
TechCorp's employees
**(C) TechCorp should implement federation between its
identity provider and SoftTech's service provider
(D)** SoftTech should reset all passwords and provide them to
TechCorp's employees

**Explanation 425. Correct Answer: C. TechCorp should
implement federation between its identity provider and
SoftTech’s service provider.** Federation allows two
organizations to trust each other’s identity systems. TechCorp’s
employees can use their existing credentials to access services
on SoftTech’s system without the need to create new accounts.

**Option A is incorrect.** Creating new accounts for every user in
a collaborating organization isn’t scalable and negates the
advantages of federation.

**Option B is incorrect.** Allowing anonymous access would
compromise security and wouldn’t guarantee identity
verification.

**Option D is incorrect.** Resetting all passwords and providing
them anew is not a practical or secure approach to collaboration
between two organizations.

**Question 426.** An organization recently experienced a malware
infection on one of its workstations. A security analyst has been
tasked with reviewing the endpoint logs of the infected system
to gather more information about the incident. Which of the
following entries in the endpoint logs would be MOST


indicative of the initial malware infection point?
**(A)** Logs indicating successful user login and logout events
**(B)** Entries showing periodic system health-check status as
"OK"
**(C) Logs documenting a recently installed and executed
unknown .exe file from a temporary directory
(D)** Entries detailing network connectivity checks to the
domain controller

**Explanation 426. Correct Answer: C. Logs documenting a
recently installed and executed unknown .exe file from a
temporary directory.** Endpoint logs that document the
installation and execution of an unknown .exe file, especially
from a temporary directory, are strong indicators of potentially
malicious activity. Such logs can pinpoint the initial infection
point of malware on a system.

**Option A is incorrect.** User login and logout events are routine
logs and do not provide specific information about malware
infections unless associated with other suspicious activities.

**Option B is incorrect.** System health-check status entries are
meant to provide general information about the system’s health
and do not specify actions or changes made on the system
related to malware.

**Option D is incorrect.** Network connectivity checks to domain
controllers are routine in many network environments and don’t
directly indicate malware activity.


**Question 427.** GammaTech has a new remote access policy for
its employees. Whenever an employee attempts to access the
corporate network from an unfamiliar location, the system
requests additional verification before granting access. Which
factor of authentication is being emphasized in this policy?
**(A)** Knowledge-based questions the employee answers
**(B)** A fingerprint scan from the employee
**(C) The physical coordinates of the employee's access
point
(D)** An SMS code sent to the employee's phone

**Explanation 427. Correct Answer: C. The physical
coordinates of the employee’s access point.** By verifying the
location or coordinates of an access point, GammaTech is
utilizing the “somewhere you are” factor in multifactor
authentication. This emphasizes the geographic location of the
user.

**Option A is incorrect.** Knowledge-based questions fall under
the “something you know” factor since users answer based on
information they recall.

**Option B is incorrect.** A fingerprint scan pertains to the
“something you are” factor as it’s a biometric, a unique physical
characteristic of the individual.

**Option D is incorrect.** An SMS code sent to a phone belongs to
the “something you have” factor, as it’s sent to a device in the
user’s possession.


**Question 428.** AlphaTech’s IT department is rolling out a new
authentication protocol for remote workers. As part of the
multifactor authentication process, employees are required to
provide information that is memorized and cannot be physically
taken from them. Which of the following represents this type of
authentication factor?
**(A)** Fingerprint
**(B)** Smart card
**(C) PIN
(D)** USB security key

**Explanation 428. Correct Answer: C. PIN.** A Personal
Identification Number (PIN) represents the “something you
know” factor in multifactor authentication. This type of
information is memorized by the user and is not a physical item
that can be taken or a biological trait.

**Option A is incorrect.** A fingerprint represents the “something
you are” factor, which pertains to biometrics.

**Option B is incorrect.** A smart card represents the “something
you have” factor, as it’s a physical item that a user possesses.

**Option D is incorrect.** A USB security key also falls under the
“something you have” factor. It’s a physical device rather than
memorized information.

**Question 429.** A company has set up its firewall to allow web
traffic through port 80 and port 443, while denying all other
traffic by default. This setup is an example of which type of
access control?


**(A)** Role-Based Access Control (RBAC)
**(B)** Mandatory Access Control (MAC)
**(C)** Discretionary Access Control (DAC)
**(D) Rule-Based Access Control (RAC)**

**Explanation 429. Correct Answer: D. Rule-Based Access
Control (RAC).** Rule-Based Access Control (RAC) operates
based on predefined rules set by administrators. In the scenario
described, the firewall is using rules to allow traffic on certain
ports (80 and 443) while denying all others, making this a clear
example of RAC.

**Option A is incorrect.** RBAC assigns permissions based on
roles within an organization. Firewall rules are not typically
assigned based on user roles.

**Option B is incorrect.** MAC involves classifying information
and matching user clearance levels to these classifications. It is
unrelated to firewall rule settings.

**Option C is incorrect.** DAC allows resource owners to specify
who can access their resources. Firewall rule settings don’t
operate based on individual discretion.

**Question 430.** The security team at WidgetCorp is trying to
identify potential insider threats. They have set up a SIEM
solution with a custom dashboard showing unusual activities.
Which of the following dashboard views would be MOST
effective for quickly identifying an employee uploading large
amounts of proprietary data to an external cloud storage
service?


**(A)** Display of users who logged in during off-hours
**(B) Graph of highest network bandwidth users
(C)** List of most frequently used applications
**(D)** Visualization of failed login attempts

**Explanation 430. Correct Answer: B. Graph of highest
network bandwidth users.** When an employee uploads large
amounts of data to an external service, it typically results in a
significant spike in network bandwidth. Therefore, a dashboard
view that visually displays the highest network bandwidth users
can quickly alert the security team to potential data exfiltration
activities.

**Option A is incorrect.** While logging in during off-hours can
be suspicious, it doesn’t directly correlate to data upload
activities.

**Option C is incorrect.** A list of the most frequently used
applications might help in determining the common tools used
within the organization, but it doesn’t specifically point to data
upload actions.

**Option D is incorrect.** Visualization of failed login attempts
could show potential brute-force or unauthorized access
attempts, but it doesn’t directly indicate data uploading
activities.

**Question 431.** Sarah is a project manager and is working on a
document that she owns. She wants to grant specific
permissions to certain team members, allowing some to edit and
others only to view the document. Which of the following


access control models would BEST allow Sarah to accomplish
this?
**(A)** Mandatory Access Control (MAC)
**(B)** Role-Based Access Control (RBAC)
**(C) Discretionary Access Control (DAC)
(D)** Attribute-Based Access Control (ABAC)

**Explanation 431. Correct Answer: C. Discretionary Access
Control (DAC).** DAC allows the owner of the resource (in this
case, Sarah) to specify who can access it and what permissions
they have (e.g., read, write, execute). This flexibility is what
Sarah needs to grant specific permissions to individual team
members based on her discretion.

**Option A is incorrect.** Mandatory Access Control (MAC) is
based on classifications and clearance levels. It wouldn’t be
suitable for Sarah’s needs in this scenario.

**Option B is incorrect.** Role-Based Access Control (RBAC)
assigns permissions based on roles in the organization. It
wouldn’t allow Sarah the fine-grained control she needs over
individual permissions.

**Option D is incorrect.** Attribute-Based Access Control
(ABAC) bases access on attributes of the user, environment,
and resource itself. While flexible, it’s not centered around the
owner’s discretion in the same way DAC is.

**Question 432.** CyberSec Corp’s CISO wants to determine if
there have been any anomalies in user behavior over the past
month. Specifically, they’re concerned about unauthorized data


transfers outside of regular business hours. Which of the
following automated reports would be MOST useful in this
investigation?
**(A) After-hours network activity reports
(B)** User password change frequency reports
**(C)** Hardware inventory audit reports
**(D)** Software licensing compliance reports

**Explanation 432. Correct Answer: A. After-hours network
activity reports.** For the specific concern of unauthorized data
transfers outside of regular business hours, the after-hours
network activity reports would be most useful. These reports
would provide details on network activities, including data
transfers, that took place outside typical working hours.

**Option B is incorrect.** While user password change frequency
reports might indicate if users are frequently changing
passwords, which could be a sign of suspicious activity, it
wouldn’t directly address the concern of data transfers outside
of regular hours.

**Option C is incorrect.** Hardware inventory audit reports would
provide information about the hardware assets of the company
but wouldn’t give insights into data transfer activities.

**Option D is incorrect.** Software licensing compliance reports
would detail the compliance status of software licenses, which
isn’t relevant to the concern about unauthorized data transfers
after hours.


**Question 433.** After detecting suspicious activity on a network,
a digital forensic analyst is dispatched to acquire data from a
potential compromised system. The analyst decides to capture
an image of the affected system’s memory. This technique of
capturing volatile data is particularly beneficial because:
**(A)** It helps identify deleted files
**(B) It can capture data in real-time operations
(C)** It provides information on patch levels
**(D)** It offers insights into firewall configurations

**Explanation 433. Correct Answer: B. It can capture data in
real-time operations.** Memory acquisition allows forensic
analysts to capture data in its current state, including data about
running processes, open network connections, and contents of
the system’s RAM. This can provide insights into malware or
unauthorized activities that occurred in real-time.

**Option A is incorrect.** While memory acquisition can
sometimes provide information about recently accessed files,
identifying deleted files is typically done through disk imaging
and not memory acquisition.

**Option C is incorrect.** Information about patch levels is usually
gleaned from system configurations and logs, not directly from
memory acquisition.

**Option D is incorrect.** Firewall configurations are typically
found in system configurations and logs, not directly from a
memory acquisition.


**Question 434.** During a suspected security incident involving
unauthorized access to sensitive data, Jake, an IT administrator,
immediately disconnected the affected server from the network.
Later, a digital forensic expert criticized Jake’s action. Which of
the following is the MOST likely reason for the criticism?
**(A)** Jake should have left the server connected to capture
more evidence from the attacker
**(B)** Jake should have immediately informed the company's
legal department
**(C) Jake should have taken an image of the server's
memory before disconnecting it
(D)** Jake should have updated the server's software to
prevent further unauthorized access

**Explanation 434. Correct Answer: C. Jake should have
taken an image of the server’s memory before disconnecting
it.** Preserving the current state of a system, especially its volatile
memory, is crucial during a digital forensic investigation.
Volatile memory can contain critical evidence about an incident,
but this evidence is lost once the system is powered off or
restarted.

**Option A is incorrect.** While sometimes monitoring an attacker
can be valuable, it’s often more critical to prioritize the safety of
data and systems over collecting additional evidence.

**Option B is incorrect.** Although informing the legal
department is an important step in many incident response
processes, preserving the integrity and state of potential
evidence comes first.


**Option D is incorrect.** While updating the server’s software
might be a future step to prevent incidents, the immediate
priority during a suspected security incident is to preserve
evidence.

**Question 435.** A large enterprise is deploying a new automation
system that will allow various teams, including development,
operations, and QA, to provision and configure their own
environments. The security team is concerned about potential
misconfigurations or excessive permissions being granted.
Which solution can be used within the automation to ensure
security standards are met without limiting the agility of the
teams?
**(A)** Implementing a zero-trust model for all teams
**(B)** Manually reviewing all requests before provisioning
**(C) Setting up guard rails within the automation scripts
to define boundaries and prevent misconfigurations
(D)** Disabling the automation system for all teams except
the security team

**Explanation 435. Correct Answer: C. Setting up guard rails
within the automation scripts to define boundaries and
prevent misconfigurations.** Guard rails in automation scripts
act as safeguards, ensuring that certain actions, configurations,
or provisions stay within defined security and operational
boundaries without impeding the benefits of automation.

**Option A is incorrect.** While a zero-trust model is beneficial
for security, it does not directly address the challenge of
misconfigurations in automation processes.


**Option B is incorrect.** Manually reviewing all requests negates
the efficiency benefits of automation and slows down the
provisioning process.

**Option D is incorrect.** Disabling the automation system for all
teams except the security team defeats the purpose of having an
automation system for diverse teams and reduces agility.

**Question 436.** After a security breach, Jake, a digital forensics
investigator, arrives at the scene to collect a hard drive for
examination. He labels the hard drive, records its serial number,
photographs the scene, and ensures the hard drive is transported
securely to the forensics lab. These steps are crucial to:
**(A)** Preserve the data's integrity on the hard drive
**(B) Maintain the chain of custody
(C)** Decrypt the data on the hard drive
**(D)** Implement a legal hold on the data

**Explanation 436. Correct Answer: B. Maintain the chain of
custody.** Maintaining a chain of custody is crucial to ensure that
evidence is authentic and unchanged. This involves
documenting each step of the evidence handling process, from
collection to analysis, ensuring its validity in legal proceedings.

**Option A is incorrect.** While preserving the data’s integrity is
crucial in forensics, the specific steps mentioned are primarily
for maintaining the chain of custody.

**Option C is incorrect.** Decrypting the data focuses on making
encrypted data readable. The steps Jake took are related to
documenting evidence handling, not decryption.


**Option D is incorrect.** Implementing a legal hold ensures that
specific data is preserved for legal reasons. The steps mentioned
by Jake focus on documenting how the evidence was handled
and preserved, not on a directive to retain it for legal purposes.

**Question 437.** After deploying a new version of your
company’s internal application, several users reported issues
with accessing specific features. To investigate the root cause,
you decided to review the application logs. What entry in the
logs would most directly indicate a software bug or error related
to the recent deployment?
**(A)** Entries showing successful user authentication
timestamps
**(B)** Entries detailing the number of transactions completed
by the application
**(C) Entries with "ERROR" or "EXCEPTION" related
to the specific feature being accessed
(D)** Entries showing routine data backup operations

**Explanation 437. Correct Answer: C. Entries with
“ERROR” or “EXCEPTION” related to the specific feature
being accessed.** In the context of application logs, entries
labeled as “ERROR” or “EXCEPTION” generally indicate that
the application encountered a problem. If these entries are
related to the feature users are having trouble with, it points
towards a software bug or issue related to the deployment.

**Option A is incorrect.** Successful user authentication entries
would indicate that users are able to log into the application
successfully, but they don’t provide insights into feature-
specific issues.


**Option B is incorrect.** While the number of transactions could
provide performance metrics or usage patterns, it doesn’t
directly indicate a software bug or deployment-related issue.

**Option D is incorrect.** Routine data backup operations are
unrelated to application feature functionalities and won’t help in
identifying deployment-related errors.

**Question 438.** After a major data breach in XYZ Corporation,
the management decided to understand the primary reason
behind the incident to prevent such occurrences in the future.
Which of the following approaches should the incident response
team prioritize to determine the fundamental cause of the
breach?
**(A)** Perform vulnerability scanning on all servers
**(B)** Review firewall logs for the past week
**(C) Conduct a root cause analysis
(D)** Upgrade all security software

**Explanation 438. Correct Answer: C. Conduct a root cause
analysis.** Root cause analysis (RCA) is a systematic process for
identifying the origin of problems or faults and deciding on the
most suitable approach to take to prevent recurrence. In the
context of a security incident, RCA helps in determining the
primary reason behind the breach.

**Option A is incorrect.** While vulnerability scanning is essential
for understanding potential weaknesses in servers, it does not
directly identify the fundamental cause of a past incident.


**Option B is incorrect.** Firewall logs can provide insights about
traffic patterns and potential attacks but may not directly reveal
the root cause of a breach.

**Option D is incorrect.** Upgrading security software is a
reactive measure and does not ensure understanding or
addressing the core reason for a breach.

**Question 439.** A cloud infrastructure team frequently receives
performance alerts from various resources in the environment.
They want to ensure that relevant teams are immediately
informed and can act upon any resource that crosses a
performance threshold. What is the BEST way to accomplish
this?
**(A)** Conduct a weekly meeting to review all performance
alerts
**(B) Automate ticket creation for any resource that
crosses the performance threshold and assign it to the
relevant team
(C)** Send all performance alerts to the cloud infrastructure
team's email for review
**(D)** Disable performance monitoring to reduce alert fatigue

**Explanation 439. Correct Answer: B. Automate ticket
creation for any resource that crosses the performance
threshold and assign it to the relevant team.** Automating the
ticket creation process ensures that alerts are not overlooked and
that the appropriate teams are informed in real-time, allowing
for prompt resolution.


**Option A is incorrect.** Conducting a weekly review does not
allow for immediate action upon critical performance issues.

**Option C is incorrect.** Sending all alerts to the team’s email
might lead to alert fatigue and the possibility of overlooking
critical alerts among less important ones.

**Option D is incorrect.** Disabling performance monitoring
would prevent the team from receiving important alerts and is
not a practical solution.

**Question 440.** A development team is working on a mission-
critical application for a financial institution. The team wants to
ensure that any code changes do not introduce vulnerabilities or
break existing functionalities. What is the BEST automation
approach to achieve this objective?
**(A)** Manually review the code changes once a month
**(B) Use continuous integration tools to automatically
compile and test code changes against known vulnerabilities
and functional tests
(C)** Rely on users to report any issues after the application is
deployed
**(D)** Implement a firewall to block potential attacks on the
application

**Explanation 440. Correct Answer: B. Use continuous
integration tools to automatically compile and test code
changes against known vulnerabilities and functional tests.**
Continuous integration tools can be set up to automatically test
code changes as they are committed. This ensures that


vulnerabilities and functional issues are detected early in the
development process.

**Option A is incorrect.** Manually reviewing code changes once
a month does not provide immediate feedback to developers and
could delay the identification of vulnerabilities or functional
issues.

**Option C is incorrect.** Relying on users to report issues is
reactive and could expose the financial institution to risks if
vulnerabilities are exploited.

**Option D is incorrect.** While firewalls are important for
security, they don’t address the need to test code changes for
vulnerabilities or functional issues.

**Question 441.** Acme Corp. is in the early stages of a potential
lawsuit, and their legal department has just issued a notice for e-
discovery related to email communications of a former
executive. As an IT security professional, which of the
following should be your FIRST action?
**(A)** Start a full backup of the company's email server
**(B) Identify and isolate the email accounts related to the
former executive
(C)** Immediately delete all emails that are more than two
years old
**(D)** Inform the media about the upcoming lawsuit

**Explanation 441. Correct Answer: B. Identify and isolate
the email accounts related to the former executive.** In the
context of e-discovery, it’s crucial to locate and preserve


electronically stored information (ESI) that could be relevant to
the lawsuit. The first step would be to identify and isolate the
specific email accounts or data sources relevant to the request to
ensure they are not tampered with or deleted.

**Option A is incorrect.** While backups are essential, a blanket
backup of the email server may not specifically cater to the e-
discovery request. Focusing on the particular data in question is
more pertinent.

**Option C is incorrect.** Deleting potential evidence, especially
after a notice for e-discovery, could lead to legal penalties and is
not a recommended action.

**Option D is incorrect.** Informing the media is not a primary
step in the e-discovery process and can have detrimental effects
on the company’s reputation.

**Question 442.** OmegaHealth, a large healthcare provider, is
integrating automation into its operations. When a new
healthcare worker is hired, they require access to multiple
systems. Why would OmegaHealth automate the user
provisioning process across these systems?
**(A)** To enforce a uniform password for all healthcare
workers.
**(B) To save time by ensuring consistent and
simultaneous account creation across all necessary
platforms
(C)** To prevent the new hires from accessing any system
until their probation period ends


**(D)** To reduce the software licenses needed by delaying
account activation

**Explanation 442. Correct Answer: B. To save time by
ensuring consistent and simultaneous account creation
across all necessary platforms.** Automating user provisioning,
especially in an environment where access to multiple systems
is needed, can dramatically save time and reduce human errors.
With automation, accounts can be created consistently and
simultaneously across all required platforms.

**Option A is incorrect.** Enforcing a uniform password for all
users is a poor security practice. Automation should focus on
efficiency and security, not creating potential vulnerabilities.

**Option C is incorrect.** Automating user provisioning doesn’t
inherently prevent new hires from accessing systems. The
automation should be designed to provision based on specific
rules and roles.

**Option D is incorrect.** Automation in user provisioning is
about streamlining the creation and management of accounts,
not about reducing software licenses or delaying activations.

**Question 443.** During a regular review of system logs, Alex, a
security analyst, noticed an unusual pattern of network traffic
originating from a single IP address. Instead of waiting for an
automated system to flag this as suspicious, he decides to
manually dive deeper into the data to identify any potential
threats. What is Alex engaging in?
**(A)** Incident management


**(B)** Threat modeling
**(C) Threat hunting
(D)** Security monitoring

**Explanation 443. Correct Answer: C. Threat hunting.** Threat
hunting is a proactive approach where security professionals or
analysts actively and manually search for signs of malicious
activities within their network or systems, especially those
threats that haven’t been automatically detected by traditional
security tools.

**Option A is incorrect.** Incident management refers to the
process followed when managing and responding to a security
incident. In this scenario, Alex is taking a proactive approach to
find potential threats, not responding to an identified incident.

**Option B is incorrect.** Threat modeling involves identifying
potential threats and designing countermeasures to prevent or
mitigate the impact of those threats. It’s more about planning
than actively searching for threats.

**Option D is incorrect.** Security monitoring is the process of
continuously monitoring and analyzing an organization’s
security events. While Alex is reviewing logs, which is part of
monitoring, the manual and proactive deep dive he is taking is
more in line with threat hunting.

**Question 444** OmegaTech’s security team noticed an increase
in account compromises. An internal investigation revealed that
many employees have been using the same passwords across
different company systems and applications. Which password


best practice can OmegaTech enforce to mitigate this issue?
**(A)** Encouraging users to change their passwords every
month
**(B)** Implementing an account lockout policy after three
failed login attempts
**(C) Prohibiting password reuse for at least the last five
password changes
(D)** Mandating that passwords contain only alphabetical
characters for simplicity

**Explanation 444. Correct Answer: C. Prohibiting password
reuse for at least the last five password changes.** By
prohibiting password reuse for a number of iterations, you
discourage users from cycling between a small set of passwords
and, therefore, increase the overall security of user accounts.

**Option A is incorrect.** While frequent password changes can
enhance security, they don’t directly address the issue of
password reuse across different systems and applications.

**Option B is incorrect.** While an account lockout policy can
deter brute-force attacks, it doesn’t prevent users from reusing
the same passwords across different platforms.

**Option D is incorrect.** Restricting passwords to only
alphabetical characters reduces complexity and weakens the
security of the password.

**Question 445.** AlphaCorp’s IT department is reviewing
password policies and wants to adopt a strategy that enhances
security. Which of the following password strategies would be


the MOST secure?
**(A)** Passwords should be at least 6 characters long, with no
other requirements
**(B)** Passwords should be at least 10 characters long and
include both uppercase and lowercase letters
**(C) Passwords should be at least 8 characters long and
include uppercase letters, lowercase letters, numbers, and
special characters
(D)** Passwords should be at least 4 characters long and
include a mix of uppercase and lowercase letters

**Explanation 445. Correct Answer: C. Passwords should be
at least 8 characters long and include uppercase letters,
lowercase letters, numbers, and special characters.** This
option provides a balanced combination of length and
complexity, making it more resistant to brute-force and
dictionary attacks.

**Option A is incorrect.** A 6-character password without any
complexity requirements is easier to crack with modern
computational capabilities.

**Option B is incorrect.** While having a 10-character length is
beneficial, it lacks the additional complexity of numbers and
special characters which can further bolster security.

**Option D is incorrect.** The length is too short, and even with a
mix of uppercase and lowercase, it does not provide the
recommended security for passwords.


**Question 446.** A security analyst is reviewing the IPS logs and
discovers multiple alerts originating from a single IP address
attempting to access various company servers. The analyst is
trying to determine the type of attack. Which of the following
log entries BEST indicates a port scanning activity?
**(A) Multiple consecutive connection attempts to different
ports on a single server in a short time frame
(B)** Repeated connection attempts to port 80 of a web server
every 3 seconds
**(C)** Numerous failed login attempts to an FTP server from
the same IP address
**(D)** Consistent pings to the network gateway every 5
seconds

**Explanation 446. Correct Answer: A. Multiple consecutive
connection attempts to different ports on a single server in a
short time frame.** Port scanning is an activity where an attacker
probes a server or host to determine what services are running.
Rapid connection attempts to various ports indicate that the
attacker is trying to discover open ports and the services
running on them.

**Option B is incorrect.** While repeated connection attempts to a
single port could be suspicious, it doesn’t indicate port
scanning. This could be indicative of a DoS attack.

**Option C is incorrect.** Multiple failed login attempts indicate a
possible brute force attack on the FTP server, not port scanning.


**Option D is incorrect.** Regular pings to the network gateway
might be indicative of network mapping or checking
connectivity but doesn’t suggest port scanning activity.

**Question 447.** As part of a cloud infrastructure project,
AlphaTech plans to deploy multiple virtualized resources for its
new application. The deployment includes databases, web
servers, and load balancers. What is the PRIMARY benefit of
using automation scripts in the resource provisioning process
for this project?
**(A)** It enables AlphaTech to use a single operating system
for all resources
**(B)** It guarantees 100% uptime for all virtualized resources
**(C) It ensures standardized, repeatable, and rapid
deployments across the infrastructure
(D)** It prevents unauthorized users from accessing the cloud
infrastructure

**Explanation 447. Correct Answer: C. It ensures
standardized, repeatable, and rapid deployments across the
infrastructure.** Automation in resource provisioning provides
consistency in deployments, allowing for standardized
configurations, rapid scaling, and the ability to repeat
deployments without human error.

**Option A is incorrect.** While automation can deploy similar or
consistent environments, it does not inherently enforce a single
operating system.

**Option B is incorrect.** Automation can increase efficiency and
reduce human errors, but it cannot guarantee 100% uptime, as


there are other factors involved, like hardware failures or
network issues.

**Option D is incorrect.** While automation can implement
security configurations, it does not, by itself, prevent
unauthorized access. Security measures need to be designed and
implemented separately.

**Question 448.** DeltaTech, a progressive tech firm, is aiming to
improve its security posture by eliminating the vulnerabilities
associated with password use. They are considering deploying a
passwordless authentication system. Which of the following
represents the PRIMARY advantage of such a system?
**(A)** It allows users to choose any password complexity
**(B) It eliminates the need for remembering passwords
(C)** It guarantees protection against all cyber threats
**(D)** It ensures compatibility with all legacy systems

**Explanation 448. Correct Answer: B. It eliminates the need
for remembering passwords.** Passwordless authentication
methods, such as biometrics, hardware tokens, or mobile app
confirmations, remove the need for users to remember and input
passwords. This can reduce risks associated with weak or reused
passwords and enhance user convenience.

**Option A is incorrect.** Passwordless systems remove the
concept of a traditional password altogether, so there’s no
“password complexity” to choose.

**Option C is incorrect.** While passwordless authentication can
significantly enhance security, especially against threats like


password spraying or credential stuffing, it doesn’t guarantee
protection against all cyber threats.

**Option D is incorrect.** Passwordless authentication methods
might not be compatible with all legacy systems without
modifications or updates.

**Question 449.** OmegaTech recently introduced an additional
layer of security for its remote server access. Along with their
usual passwords, employees now need to use a physical device
they have with them to gain access. Which of the following
represents this “something you have” factor in multifactor
authentication?
**(A)** Password hint
**(B)** Facial recognition
**(C) Hardware token
(D)** Voice recognition

**Explanation 449. Correct Answer: C. Hardware token.**
Hardware tokens are devices that generate security codes for
authentication purposes. Users are required to have this physical
device on-hand, making it an example of the “something you
have” factor in multifactor authentication.

**Option A is incorrect.** A password hint is related to “something
you know” factor. Moreover, it’s not an authentication factor
but rather an aid for recalling a password.

**Option B is incorrect.** Facial recognition pertains to the
“something you are” factor, which refers to biometric methods
of authentication.


**Option D is incorrect.** Voice recognition, similar to facial
recognition, falls under the “something you are” factor, as it
uses biometric identification.

**Question 450.** A large corporation is investigating a potential
insider threat incident. A security analyst is tasked with
examining the OS-specific security logs of a Windows server
where sensitive documents are stored. Which of the following
entries in the logs would MOST likely indicate unauthorized
access attempts?
**(A)** Logs displaying Windows Update successful
installations
**(B) Entries showing a large number of failed login
attempts followed by a successful login from a user outside
of regular business hours
(C)** Logs indicating scheduled disk defragmentation tasks
**(D)** Entries detailing successful printer connections and
print jobs

**Explanation 450. Correct Answer: B. Entries showing a
large number of failed login attempts followed by a
successful login from a user outside of regular business
hours.** A sequence of failed login attempts followed by a
successful login, especially outside of regular business hours,
can be a strong indicator of a brute-force attack or unauthorized
access attempt on a system.

**Option A is incorrect.** Successful installations of Windows
Updates are routine maintenance activities and do not directly
imply unauthorized access.


**Option C is incorrect.** Disk defragmentation tasks are part of
regular system maintenance and are not indicative of
unauthorized access.

**Option D is incorrect.** While monitoring printer connections
and print jobs can be relevant in certain security investigations,
they are not direct indicators of unauthorized server access.

**Question 451.** ThetaTech, a financial institution, wants to
upgrade its authentication system for high-net-worth customers
accessing their accounts online. Besides the traditional
password, they want to include a method that captures unique
physical or behavioral characteristics. Which type of
authentication method should ThetaTech consider?
**(A)** Token-based authentication
**(B)** Geolocation tracking
**(C) Biometrics
(D)** Smart card

**Explanation 451. Correct Answer: C. Biometrics.** Biometrics
authentication involves recognizing an individual based on their
unique physical or behavioral characteristics, such as
fingerprints, retina scans, or voice recognition.

**Option A is incorrect.** Token-based authentication provides
users with a device or software token that generates a time-
sensitive code. It doesn’t capture physical or behavioral
characteristics.


**Option B is incorrect.** Geolocation tracking determines a
user’s location but doesn’t involve capturing unique physical or
behavioral traits.

**Option D is incorrect.** A smart card is a physical card that
contains electronic information. It can be used for
authentication, but it does not inherently capture unique
physical or behavioral characteristics.

**Question 452.** The cybersecurity team of XYZ Corp. plans to
assess their organization’s preparedness for a potential data
breach. They aim to evaluate the effectiveness of their response
strategy without performing any real actions. Which of the
following methods would BEST help them achieve this goal?
**(A)** Live fire exercise
**(B)** System hardening test
**(C)** Red team/blue team exercise
**(D) Tabletop exercise**

**Explanation 452. Correct Answer: D. Tabletop exercise.** A
tabletop exercise is a discussion-based session where team
members meet in an informal, classroom setting to discuss their
roles during an emergency and their responses to a particular
emergency situation. It allows for an assessment of an incident
response strategy without performing any actual tasks.

**Option A is incorrect.** A live fire exercise is a real-world test,
often unannounced, and could impact actual operations.


**Option B is incorrect.** System hardening test focuses on
making a system more secure against attacks and does not
evaluate incident response.

**Option C is incorrect.** Red team/blue team exercises involve
simulating real-world cyber attacks to test an organization’s
defense and response capabilities, which is more hands-on than
what is described.

**Question 453.** In preparation for a potential lawsuit, Meg, a
cybersecurity analyst, has been asked to ensure that specific
digital evidence remains intact and is not altered or deleted.
What measure should Meg implement to ensure this
requirement?
**(A)** Encrypt the evidence
**(B) Initiate a legal hold
(C)** Perform a full disk wipe
**(D)** Conduct a vulnerability assessment

**Explanation 453. Correct Answer: B. Initiate a legal hold.** A
legal hold ensures that specific data that could be relevant to a
legal case is preserved and not altered or deleted until the hold
is lifted. In this situation, Meg would initiate a legal hold to
keep the digital evidence intact for the potential lawsuit.

**Option A is incorrect.** Encrypting the evidence can ensure its
confidentiality, but it does not prevent deletion or guarantee its
preservation for legal reasons.


**Option C is incorrect.** Performing a full disk wipe would
eliminate all data on a disk, which is contrary to the requirement
of preserving specific digital evidence.

**Option D is incorrect.** Conducting a vulnerability assessment
is about identifying weaknesses in a system, not about
preserving digital evidence for legal purposes.

**Question 454.** A financial company is designing a new system
that needs to ensure data is accessed based on classifications
and clearance levels of the users. Which of the following access
control models BEST fits this requirement?
**(A)** Role-Based Access Control (RBAC)
**(B)** Discretionary Access Control (DAC)
**(C) Mandatory Access Control (MAC)
(D)** Attribute-Based Access Control (ABAC)

**Explanation 454. Correct Answer: C. Mandatory Access
Control (MAC).** MAC is based on the classification of
information and the clearance level of users. In a MAC model,
the operating system constrains the ability of a subject or
initiator to access or perform some sort of operation on an
object or target. In this scenario, where data is classified and
users are given clearance levels, MAC is the most suitable
model.

**Option A is incorrect.** Role-Based Access Control (RBAC)
assigns permissions to specific roles in an organization. Users
are then assigned to roles. While useful in many contexts, it
doesn’t focus on classifications and clearance levels.


**Option B is incorrect.** Discretionary Access Control (DAC)
allows the owner of the resource to specify who can access it.
It’s more flexible but less restrictive than MAC.

**Option D is incorrect.** Attribute-Based Access Control
(ABAC) uses policies to determine access, based on attributes
of users, the environment, and the resource itself. While it can
be used in scenarios with classifications and clearances, it’s not
as strictly based on these factors as MAC.

**Question 455.** The incident response team at XYZ Corp
received a report that an attacker successfully exploited a
vulnerable web application in their environment. To identify
which server might have been compromised, the team decided
to cross-reference recent vulnerability scan results. Which of the
following information from the vulnerability scan would be
MOST helpful in pinpointing the potentially compromised
server?
**(A)** The timestamp of when the scan was conducted
**(B)** The software version of the scanning tool
**(C) List of hosts with the specific vulnerability related to
the exploit
(D)** The total number of vulnerabilities identified during the
scan

**Explanation 455. Correct Answer: C. List of hosts with the
specific vulnerability related to the exploit.** In the given
scenario, to determine which server may have been
compromised, the team should focus on those hosts identified in
the vulnerability scan as having the specific vulnerability that
matches the exploit used by the attacker.


**Option A is incorrect.** While the timestamp might indicate
when the scan was done, it wouldn’t provide specifics about
which servers had the vulnerability related to the reported
exploit.

**Option B is incorrect.** Knowing the software version of the
scanning tool wouldn’t assist in identifying the potentially
compromised server.

**Option D is incorrect.** The total number of vulnerabilities
identified doesn’t help pinpoint a specific server; it only
provides a high-level overview of the security posture.

**Question 456.** Epsilon Inc. recently hired Jenny as a junior
network administrator. To ensure security, they give Jenny only
the access permissions necessary to complete her specific tasks,
such as monitoring network traffic, but not modifying firewall
rules. This approach of granting Jenny’s permissions aligns with
which security principle?
**(A)** Mandatory Access Control (MAC)
**(B)** Role-Based Access Control (RBAC)
**(C)** Time-of-Day Restrictions
**(D) Least Privilege**

**Explanation 456. Correct Answer: D. Least Privilege.** The
principle of Least Privilege dictates that users should be granted
the minimum levels of access – or the least amount of privileges

- necessary to complete their job functions. In Jenny’s case,
she’s only granted the permissions necessary for her role, which
aligns with this principle.


**Option A is incorrect.** MAC involves classifying information
and matching user clearance levels to these classifications.
Jenny’s access isn’t based on classifications.

**Option B is incorrect.** RBAC assigns permissions based on
roles within an organization. While Jenny’s permissions may be
aligned with her role, the scenario specifically emphasizes
granting the minimal necessary access, which is a characteristic
of Least Privilege.

**Option C is incorrect.** Time-of-Day Restrictions determine
access based on the current time and have no direct relation to
the principle of Least Privilege.

**Question 457.** A company is attempting to verify the legitimacy
of an email sent from a senior executive to a number of
employees. The email requests the recipients to click on a link
and enter their credentials for a “system upgrade.” The security
team wants to ascertain if the email genuinely came from the
executive. Which of the following metadata from the email
would be MOST beneficial in this investigation?
**(A)** The email's subject line
**(B)** The email's send time and date
**(C) The originating IP address in the email headers
(D)** The size of the email in bytes

**Explanation 457. Correct Answer: C. The originating IP
address in the email headers.** Email headers often contain
metadata about the originating IP address of the email. This can
help investigators determine if the email was sent from an


expected location or if it originated from an unfamiliar or
suspicious IP, indicating a possible phishing attempt.

**Option A is incorrect.** The subject line of an email is not
typically indicative of its authenticity.

**Option B is incorrect.** While knowing the send time and date
may provide some context, it won’t necessarily verify the
legitimacy of the email sender.

**Option D is incorrect.** The size of the email in bytes doesn’t
offer significant value in determining the authenticity of the
email’s origin.

**Question 458.** A cloud-based e-commerce company wants to
ensure that its inventory system automatically updates the stock
levels on its website and third-party sales platforms whenever a
sale occurs. What should the company leverage to achieve this
real-time synchronization?
**(A)** Regularly backup the inventory system and restore it on
the website and sales platforms
**(B)** Rely on customers to report discrepancies in stock levels
**(C) Use Application Programming Interfaces (APIs) to
integrate the inventory system with the website and third-
party platforms
(D)** Conduct daily stock audits and manually update all
platforms

**Explanation 458. Correct Answer: C. Use Application
Programming Interfaces (APIs) to integrate the inventory
system with the website and third-party platforms.** Using


APIs ensures real-time data synchronization between systems.
Whenever a sale occurs, the inventory system can automatically
update stock levels across all platforms.

**Option A is incorrect.** Backing up and restoring the inventory
system is a reactive and inefficient approach. It won’t provide
real-time stock updates.

**Option B is incorrect.** Relying on customers to report stock
discrepancies is not proactive and could lead to negative
customer experiences and potential lost sales.

**Option D is incorrect.** Manual updates are time-consuming and
prone to human error, and they don’t support real-time stock
level synchronization.

**Question 459.** After a series of phishing attacks, the IT
department of BetaTech Corp noticed that several employees
were using easily guessable passwords. The security team
decided to recommend the use of password managers to assist
employees in creating and remembering strong passwords.
Which of the following is a PRIMARY benefit of using
password managers in this context?
**(A)** Password managers automatically update the operating
system
**(B) Password managers can generate and store complex
passwords
(C)** Password managers always prevent phishing attacks
**(D)** Password managers allow the reuse of strong passwords
across multiple platforms


**Explanation 459. Correct Answer: B. Password managers
can generate and store complex passwords.** Password
managers can automatically generate complex passwords that
meet a variety of criteria (e.g., length, use of special characters,
avoidance of easily guessable terms) and store them securely.
This helps users maintain strong, unique passwords for every
service they use without needing to remember each one.

**Option A is incorrect.** Password managers are designed for
generating, retrieving, and storing complex passwords, not for
updating the operating system.

**Option C is incorrect.** While password managers can help
reduce the risk of successful phishing attacks by storing
passwords securely and autofilling them on recognized
websites, they don’t always prevent phishing attacks, especially
if a user is deceived into entering credentials on a fake website.

**Option D is incorrect.** One of the primary benefits of password
managers is to ensure that users have unique passwords for each
application or service. Reusing passwords, even if strong, across
multiple platforms poses a security risk.

**Question 460.** A company’s online retail website faces DDoS
attacks that cause significant downtime. Their current setup
relies on manual verification of traffic spikes before mitigation
efforts are deployed. What change could BEST enhance the
company’s reaction time to such attacks in the future?
**(A)** Educate users to report slow website loading times
**(B)** Manually back up the website data every hour
**(C) Deploy a web application firewall with automated**


**DDoS mitigation features
(D)** Increase the website's bandwidth to handle traffic spikes

**Explanation 460. Correct Answer: C. Deploy a web
application firewall with automated DDoS mitigation
features.** Deploying a web application firewall (WAF) with
automated DDoS mitigation can instantly detect and mitigate
attack traffic, significantly improving reaction time to attacks
compared to waiting for manual verification.

**Option A is incorrect.** Relying on users to report slow loading
times is reactive and does not guarantee a swift response to
DDoS attacks.

**Option B is incorrect.** Backing up website data is important for
recovery, but it doesn’t prevent or mitigate active DDoS attacks
or improve reaction time.

**Option D is incorrect.** Merely increasing bandwidth might not
be sufficient against DDoS attacks. An attacker can still
overwhelm the increased capacity, and it doesn’t address the
need for swift detection and mitigation.