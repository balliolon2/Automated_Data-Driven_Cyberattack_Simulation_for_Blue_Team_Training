**Question 221.** The networking team at SecureNet Corp. is
setting up a new branch office. They want to ensure secure
connectivity between the branch office and the main office.
Which of the following would BEST provide this?
**(A)** Establishing a clear line of sight for wireless antennas
between the two locations
**(B) Implementing a Virtual Private Network (VPN)
between the two sites
(C)** Increasing bandwidth on the public internet connection
at both locations
**(D)** Using multi-factor authentication for all user accounts
in both locations

**Explanation 221. Correct Answer: B. Implementing a
Virtual Private Network (VPN) between the two sites.** VPNs
provide secure, encrypted connections over potentially insecure
networks, such as the internet, making them ideal for securely
connecting remote locations.

**Option A is incorrect.** While a clear line of sight might
facilitate wireless communication, it doesn’t inherently provide
secure connectivity.

**Option C is incorrect.** Increasing bandwidth enhances data
transfer rates but doesn’t inherently secure the connection
between sites.


**Option D is incorrect.** Multi-factor authentication is a robust
security measure for user access but doesn’t secure network
connectivity between sites.

**Question 222.** A rapidly growing e-commerce company has
recently experienced an increase in cross-site scripting (XSS)
and SQL injection attacks. The company wants to deploy a
solution that specifically protects against these threats at the
application layer. Which type of firewall would be most
appropriate?
**(A)** Stateful Packet Inspection Firewall
**(B)** Proxy Firewall
**(C)** Network Layer Firewall
**(D) Web Application Firewall (WAF)**

**Explanation 222. Correct Answer: D. Web Application
Firewall (WAF).** A Web Application Firewall (WAF) is
specifically designed to protect web applications by monitoring,
filtering, and blocking potentially harmful HTTP traffic. This
includes protection against threats like XSS and SQL injection
attacks that are common against web applications.

**Option A is incorrect.** While a Stateful Packet Inspection
Firewall can monitor the state of active connections and make
decisions based on the context of the traffic, it doesn’t
specifically address application layer vulnerabilities like XSS
and SQL injection.

**Option B is incorrect.** Although a Proxy Firewall can provide
some application layer filtering, its primary purpose is not


tailored to protect against specific web application
vulnerabilities.

**Option C is incorrect.** A Network Layer Firewall primarily
operates at the network layer and won’t provide specialized
protection against application layer threats.

**Question 223.** SecureNet, a cybersecurity firm, is implementing
an Intrusion Detection System (IDS) for its enterprise client.
Where should the IDS be placed for optimal detection of
malicious activities?
**(A)** Before the perimeter firewall to capture all inbound
traffic
**(B) Between the perimeter firewall and the internal
network to monitor the filtered traffic
(C)** Inside the DMZ to monitor only external service
requests
**(D)** Adjacent to each workstation for personalized security

**Explanation 223. Correct Answer: B. Between the perimeter
firewall and the internal network to monitor the filtered
traffic.** By placing the IDS between the firewall and the internal
network, it can monitor traffic that has passed through the
firewall’s filters. This position provides a balance between
capturing potential threats and reducing false positives from
benign external traffic.

**Option A is incorrect.** While placing the IDS before the
firewall will capture all inbound traffic, it may result in a high
number of false positives since a lot of benign traffic might
trigger alarms.


**Option C is incorrect.** Placing the IDS only in the DMZ would
limit its visibility to only external service requests and not the
overall internal traffic which might have potential threats.

**Option D is incorrect.** While personal IDS on each workstation
provides granular monitoring, it does not provide a holistic view
of network activities and can be resource-intensive.

**Question 224.** An online gaming platform experiences latency
issues during multiplayer sessions, affecting the gameplay
experience of its users. The company wants to ensure real-time
responsiveness for its players worldwide. Which of the
following solutions would BEST mitigate these latency issues?
**(A) Implementing a Content Delivery Network (CDN)
(B)** Introducing stricter user authentication methods
**(C)** Deploying a centralized database server
**(D)** Reducing the game's graphical fidelity

**Explanation 224. Correct Answer: A. Implementing a
Content Delivery Network (CDN).** A CDN involves
distributing the hosting of content to multiple locations around
the world. For online gaming, this means players can connect to
a server closer to their geographic location, thereby reducing
latency and improving responsiveness.

**Option B is incorrect.** While strong authentication methods are
crucial for security, they don’t address latency or improve the
real-time responsiveness of the gaming platform.


**Option C is incorrect.** Centralizing the database server might
actually increase latency for users who are far from the central
location.

**Option D is incorrect.** Reducing graphical fidelity can improve
game performance on the client side, but it doesn’t address
network latency or the real-time responsiveness of multiplayer
sessions.

**Question 225.** The software development team at ABC Corp.
has created a unique application that utilizes innovative
algorithms. The company wants to ensure that competitors
cannot legally copy or replicate their application’s functionality.
Which of the following would be the MOST suitable method to
protect this intellectual property?
**(A)** Copyright the user interface design
**(B) Apply for a patent for the innovative algorithms
(C)** Store the application code in an encrypted vault
**(D)** Ensure all users sign an acceptable use policy (AUP)

**Explanation 225. Correct Answer: B. Apply for a patent for
the innovative algorithms.** Patents provide legal protection for
inventions, allowing the inventor or patent holder the exclusive
right to the patented process, design, or invention for a certain
period. In this case, patenting the algorithms would prevent
competitors from legally replicating the application’s unique
functionality.

**Option A is incorrect.** Copyrighting the user interface design
would only protect the visual appearance and elements of the


application, not its unique functionality or underlying
algorithms.

**Option C is incorrect.** While storing the application code in an
encrypted vault protects it from unauthorized access, it does not
provide legal protection against replication by competitors if
they develop similar functionality independently.

**Option D is incorrect.** An acceptable use policy (AUP) defines
how users are allowed to use the software or system, but it
doesn’t offer protection against competitors copying the
application’s unique functionality.

**Question 226.** A company is evaluating its data storage options.
They need a solution that provides them with the highest level
of control over their hardware, software, and network
configurations, allowing for customized security controls and
measures. Which deployment model would best suit their
needs?
**(A)** Cloud-based Infrastructure
**(B)** Hybrid Infrastructure
**(C) On-premises Infrastructure
(D)** Community Cloud

**Explanation 226. Correct Answer: C. On-premises
Infrastructure.** On-premises infrastructure allows
organizations to have complete control over their hardware,
software, and network configurations. This gives them the
freedom to customize security controls and measures tailored to
their specific needs.


**Option A is incorrect.** Cloud-based infrastructures are hosted
by third-party providers, which means organizations might not
have the same level of control or customization over their
resources.

**Option B is incorrect.** Hybrid infrastructure combines on-
premises and cloud resources. While it offers more control than
purely cloud-based solutions, it still doesn’t provide the total
control that on-premises infrastructure does.

**Option D is incorrect.** A community cloud is shared by several
organizations with similar requirements. While it offers some
level of control, it doesn’t provide the complete customization
and control that an on-premises infrastructure does.

**Question 227.** An organization is evaluating different security
solutions for their new branch office. They want to ensure that
the chosen solution can be rapidly deployed with minimal
configuration. Which of the following options BEST fulfills this
requirement?
**(A)** A customized Intrusion Prevention System (IPS)
tailored to the organization's unique needs
**(B) A zero-touch provisioning firewall
(C)** An open-source firewall with extensive manual settings
**(D)** A security information and event management (SIEM)
solution requiring manual log source integration

**Explanation 227. Correct Answer: B. A zero-touch
provisioning firewall.** Zero-touch provisioning allows devices
to be provisioned and configured automatically, ensuring rapid
deployment with minimal manual intervention.


**Option A is incorrect.** While a customized IPS might be
tailored for the organization’s needs, it often requires a lot of
configuration and isn’t the fastest to deploy.

**Option C is incorrect.** Open-source firewalls that require
extensive manual settings are not considered rapid to deploy as
they necessitate more time for configuration.

**Option D is incorrect.** A SIEM solution that requires manual
log source integration would not be the quickest to deploy due
to the manual steps involved.

**Question 228.** In a microservices architecture, each service
should be designed with a specific principle to ensure it
performs a specific task and interacts with other services
through well-defined interfaces. What principle is this referring
to?
**(A)** Principle of Least Privilege
**(B) Single Responsibility Principle
(C)** Open-Closed Principle
**(D)** Zero Trust Model

**Explanation 228. Correct Answer: B. Single Responsibility
Principle.** The Single Responsibility Principle dictates that a
class or module (or in the context of microservices, a service)
should have only one reason to change, meaning it should only
have one job or responsibility. In microservices, this ensures
that each service does one thing and does it well, and interacts
with others through clear interfaces.


**Option A is incorrect.** The Principle of Least Privilege is about
ensuring that users or services have only the permissions they
need to perform their tasks and nothing more. It’s not
specifically about the design of the service itself.

**Option C is incorrect.** The Open-Closed Principle is a design
principle stating that software entities should be open for
extension but closed for modification. While relevant for
software design, it’s not the primary principle guiding the
design of individual microservices.

**Option D is incorrect.** The Zero Trust Model is a security
concept where no one, whether inside or outside the
organization, is trusted by default. It’s not a principle specific to
microservice design.

**Question 229.** An organization that processes classified
information is implementing a network infrastructure to ensure
the highest level of data security. The CISO recommends using
a network configuration that ensures the system remains
completely disconnected from unsecured networks and any
external connections. Which of the following describes this type
of configuration?
**(A)** DMZ (Demilitarized Zone)
**(B)** VPN (Virtual Private Network)
**(C)** VLAN (Virtual Local Area Network)
**(D) Air-gapped network**

**Explanation 229. Correct Answer: D. Air-gapped network.**
An air-gapped network is a network that is physically isolated
from unsecured networks. Computers on an air-gapped network


cannot communicate with non-air-gapped computers and aren’t
connected to external networks or the internet, ensuring the
highest level of data security.

**Option A is incorrect.** A DMZ is a buffer zone between a
private network and external networks (typically the internet).
It’s designed to provide an additional layer of security but does
not guarantee complete isolation.

**Option B is incorrect.** VPNs are used to establish secure
connections over an unsecured network, such as the internet.
They encrypt the data being transferred but don’t physically
isolate a network.

**Option C is incorrect. A** VLAN is a logically segmented
network within a physical network. It allows for the separation
of broadcast domains in a Layer 2 network, but it doesn’t
provide complete physical isolation from other networks.

**Question 230.** A large financial institution is planning to
upgrade its IT infrastructure to allow for a more efficient use of
hardware resources, faster deployment of applications, and
reduced server provisioning times. While evaluating different
technologies, which of the following would directly address
these needs?
**(A)** Network Segmentation
**(B)** Intrusion Detection System
**(C) Virtualization
(D)** Multi-Factor Authentication


**Explanation 230. Correct Answer: C. Virtualization.**
Virtualization allows multiple virtual machines to run on a
single physical host, maximizing the use of hardware resources.
It enables rapid provisioning and deployment of new servers
and applications, addressing the requirements of the financial
institution.

**Option A is incorrect.** Network segmentation divides a
network into smaller subnetworks, primarily for security and
performance reasons. While it’s a best practice, it doesn’t
address the efficient use of hardware or rapid deployment of
applications.

**Option B is incorrect.** An Intrusion Detection System (IDS)
monitors and detects malicious activities in a network. While
crucial for security, it doesn’t cater to hardware resource
efficiency or faster server provisioning.

**Option D is incorrect.** Multi-Factor Authentication is a security
process in which a user provides multiple methods of
identification to authenticate their identity. It doesn’t impact
hardware resource efficiency or application deployment times.

**Question 231.** A financial institution wants to enhance the
security of its wired network. The goal is to ensure that only
authorized devices can connect to the network, and the
authentication process should be based on credentials or digital
certificates. Which of the following protocols would best serve
this purpose?
**(A)** SNMPv3
**(B)** SSL/TLS


**(C) 802.1X EAP
(D)** DHCP

**Explanation 231. Correct Answer: C. 802.1X EAP.** 802.1X
Extensible Authentication Protocol (EAP) provides port-based
network access control and is used on wired networks to ensure
that only authenticated devices can connect. It supports multiple
authentication methods including credentials, digital
certificates, and smart cards.

**Option A is incorrect.** SNMPv3 (Simple Network Management
Protocol version 3) is used for network management and
monitoring, not for port-based network access control.

**Option B is incorrect.** SSL/TLS (Secure Sockets Layer/
Transport Layer Security) is a protocol used for securing data
transmission on the internet, not for port-based network access
control.

**Option D is incorrect.** DHCP (Dynamic Host Configuration
Protocol) assigns IP addresses to devices on a network but
doesn’t deal with port-based authentication.

**Question 232.** TechBlitz Inc. recently underwent an IT audit,
and one of the suggestions was to reduce the attack surface.
Which of the following measures would be MOST effective in
accomplishing this?
**(A)** Increasing the password length requirement for all users
**(B)** Implementing regular vulnerability assessments
**(C) Deactivating unused services and ports on servers**


**(D)** Implementing a strict BYOD (Bring Your Own Device)
policy

**Explanation 232. Correct Answer: C. Deactivating unused
services and ports on servers.** Reducing the number of active
services and ports directly diminishes the number of potential
entry points for attackers, thus reducing the attack surface.

**Option A is incorrect.** While increasing password length
improves security against brute-force attacks, it doesn’t directly
affect the attack surface.

**Option B is incorrect.** Regular vulnerability assessments
identify potential security gaps, but simply identifying doesn’t
reduce the attack surface unless actions are taken based on
findings.

**Option D is incorrect.** A strict BYOD policy can enhance
security, but it’s focused more on the types of devices and how
they connect rather than reducing the number of potential attack
points.

**Question 233.** SafeMed, a medical facility, uses a life-saving
medical device with embedded software. Recently, a security
vulnerability was found in the software, but due to the device’s
FDA regulatory status and the software’s design, it cannot be
patched immediately. How should SafeMed address the security
concerns related to this device?
**(A)** Disconnect the device from all networks and only use it
in standalone mode
**(B)** Inform patients about the vulnerability and let them


decide whether to use the device
**(C) Implement network segmentation and strictly
control access to the device
(D)** Return the device to the manufacturer for a full refund

**Explanation 233. Correct Answer: C. Implement network
segmentation and strictly control access to the device.** By
implementing network segmentation, SafeMed can isolate the
vulnerable device from other parts of the network, reducing the
risk of potential exploitation. Strictly controlling access ensures
only authorized personnel can use or interact with the device.

**Option A is incorrect.** Completely disconnecting the device
might limit its functionality, especially if it needs network
access for updates, data transfer, or other essential operations.

**Option B is incorrect.** While transparency is important, simply
informing patients without taking protective measures may not
be enough to ensure safety and could lead to panic or
misinformation.

**Option D is incorrect.** Returning the device might not be
feasible, especially if there are no immediate replacements
available, and it is vital for patient care.

**Question 234.** A smart city project is deploying various IoT
sensors across the city to gather data on traffic patterns,
weather, pollution levels, and more. Which of the following is
the MOST critical security consideration when deploying these
sensors?
**(A)** Ensuring high data transfer speeds to cater to the


volume of data from the IoT sensors
**(B) Limiting the IoT devices to communicate only with
specific, pre-defined servers
(C)** Installing physical locks on IoT devices to prevent theft
**(D)** Allowing IoT devices to connect to any available
network for data redundancy

**Explanation 234. Correct Answer: B. Limiting the IoT
devices to communicate only with specific, pre-defined
servers.** By restricting IoT devices to communicate only with
specific, trusted servers, unauthorized access and data
tampering risks can be minimized. This measure ensures that
data is only sent to and received from legitimate sources.

**Option A is incorrect.** While data transfer speeds are important
for performance and real-time analytics, from a security
standpoint, the integrity and confidentiality of the data is more
crucial.

**Option C is incorrect.** While physical security is an essential
aspect, especially in a public setting, the most significant risks
with IoT often pertain to data security and unauthorized access.

**Option D is incorrect.** Allowing IoT devices to connect to any
available network can introduce significant security
vulnerabilities, such as man-in-the-middle attacks,
eavesdropping, or unauthorized data tampering.

**Question 235.** A multinational corporation is looking to replace
its current firewalls at all its global branches. The IT director
wants a solution that can perform stateful inspection of packets,


application-level filtering, and integrate threat intelligence feeds
for updated threat awareness. Which of the following would be
the most suitable solution?
**(A)** Stateful Packet Inspection Firewall
**(B)** Proxy Server
**(C)** Web Application Firewall (WAF)
**(D) Next-Generation Firewall (NGFW)**

**Explanation 235. Correct Answer: D. Next-Generation
Firewall (NGFW).** Next-Generation Firewalls (NGFWs) are
designed to offer traditional firewall capabilities such as stateful
inspection but also come with advanced features like
application-level filtering and the ability to integrate with threat
intelligence feeds. This makes NGFWs suitable for modern,
complex environments that demand multi-layered security
features.

**Option A is incorrect.** While a Stateful Packet Inspection
Firewall can monitor the state of active connections and analyze
the packets, it doesn’t offer advanced features like application-
level filtering or the integration of threat intelligence feeds that
a NGFW does.

**Option B is incorrect.** A Proxy Server controls and filters
requests from clients to servers, typically for web access. It
doesn’t possess the comprehensive security functionalities of a
NGFW.

**Option C is incorrect.** A Web Application Firewall (WAF)
specifically protects web applications from targeted attacks.


While valuable, it doesn’t offer the breadth of features and
integrations that a NGFW does.

**Question 236.** A multinational organization with multiple
branch offices is looking to simplify their WAN connectivity
and reduce costs while ensuring that their inter-office data
transfers remain secure. Which technology would best fit their
needs?
**(A)** VLAN
**(B)** MPLS
**(C) SD-WAN
(D)** DMZ

**Explanation 236. Correct Answer: C. SD-WAN.** Software-
defined wide area network (SD-WAN) offers enterprises the
ability to leverage any combination of transport services,
including MPLS, LTE, and broadband internet services, to
securely connect users to applications. SD-WAN can reduce
costs by enabling the use of lower-cost internet connections and
can also simplify WAN management.

**Option A is incorrect.** VLAN (Virtual Local Area Network) is
used to segment a local network into different broadcast
domains, but it does not facilitate WAN connectivity or reduce
WAN costs.

**Option B is incorrect.** MPLS (Multi-Protocol Label Switching)
is a type of WAN technology that can be expensive. While it
provides reliable and fast connections, it doesn’t offer the cost-
saving benefits or flexibility of SD-WAN.


**Option D is incorrect.** DMZ (Demilitarized Zone) is a buffer
zone between an organization’s internal network and the
external, untrusted networks. It is not related to WAN
connectivity or cost reduction.

**Question 237.** A financial company wants to improve its web
browsing security by intercepting and inspecting web traffic to
prevent users from accessing malicious sites or downloading
malware. They are looking for a solution that can act as an
intermediary for requests from clients seeking resources from
other servers. What should the company implement?
**(A)** Network IDS
**(B)** VPN Concentrator
**(C) Proxy server
(D)** Jump server

**Explanation 237. Correct Answer: C. Proxy server.** A Proxy
server serves as an intermediary between the user’s computer
and the internet. It intercepts web requests and can provide
functions such as filtering web content, caching web requests,
and inspecting traffic for security reasons.

**Option A is incorrect.** A Network Intrusion Detection System
(IDS) monitors and detects malicious activities on a network but
does not function as an intermediary for web requests.

**Option B is incorrect.** A VPN Concentrator provides secure
remote access to an organization’s network but doesn’t act as an
intermediary for web browsing requests.


**Option D is incorrect.** A Jump server is used for secure and
auditable access to internal servers but isn’t designed for
intercepting web traffic.

**Question 238.** In an IaaS (Infrastructure as a Service) model,
which of the following tasks is typically the responsibility of the
cloud customer in a standard Cloud Responsibility Matrix?
**(A)** Physical security of data centers
**(B)** Patching of host operating systems
**(C)** Network infrastructure maintenance
**(D) Patching of guest operating systems**

**Explanation 238. Correct Answer: D. Patching of guest
operating systems.** In an IaaS model, the customer is typically
responsible for managing their own virtual machines, which
includes patching the guest operating systems. The cloud
provider is generally responsible for the physical infrastructure,
network, and the host operating system.

**Option A is incorrect. P** hysical security of data centers is the
responsibility of the cloud provider. The customer does not have
control over the physical infrastructure in a cloud environment.

**Option B is incorrect.** Patching of host operating systems is the
cloud provider’s responsibility. They manage the underlying
infrastructure, including the host systems.

**Option C is incorrect.** Maintenance of network infrastructure
is typically handled by the cloud provider, ensuring connectivity
and uptime for the resources they offer.


**Question 239.** In a cloud environment, which of the following
matrices defines the shared responsibilities between the cloud
provider and the customer for specific cloud service models?
**(A)** Shared Accountability Matrix
**(B)** Cloud Resource Allocation Table
**(C)** Cloud Security Posture Matrix
**(D) Cloud Responsibility Matrix**

**Explanation 239. Correct Answer: D. Cloud Responsibility
Matrix.** The Cloud Responsibility Matrix defines the shared
responsibilities between the cloud provider and the customer in
a cloud environment, delineating what security measures the
provider will handle and which ones are the responsibility of the
customer.

**Option A is incorrect.** While the term “Shared Accountability
Matrix” may sound relevant, there’s no standard matrix in cloud
computing with this name. Shared responsibility is the concept,
but the actual term used in the context of cloud security is the
“Cloud Responsibility Matrix.”

**Option B is incorrect.** A “Cloud Resource Allocation Table” is
not a standard term associated with delineating responsibilities
in the cloud. It sounds more related to how resources, such as
compute or storage, might be allocated.

**Option C is incorrect.** “Cloud Security Posture Matrix” is not a
recognized term for defining shared responsibilities in cloud
environments. Instead, it sounds more like a potential term for
an organization’s stance or positioning regarding cloud security.


**Question 240.** A financial organization collects and stores
personally identifiable information (PII) of its customers. The
company operates within jurisdictions that have strict
regulations concerning the storage, transmission, and processing
of such data. Which of the following strategies would be MOST
appropriate for the company to ensure compliance with these
regulations?
**(A)** Use open source encryption algorithms without
validation
**(B)** Only store customer data in physical, on-site servers
**(C) Implement data classification and labeling
procedures
(D)** Limit the number of administrators with access to the
data

**Explanation 240. Correct Answer: C. Implement data
classification and labeling procedures.** By implementing data
classification and labeling procedures, the organization can
clearly identify which data is regulated, ensuring appropriate
controls, access rights, and protective measures are in place.
This aids in achieving compliance by setting clear guidelines on
data handling based on the classification.

**Option A is incorrect.** While encryption is essential for
protecting sensitive data, using open source algorithms without
validation might not ensure that the data is protected to the
standards required by regulations.

**Option B is incorrect.** Storing customer data only on physical,
on-site servers doesn’t necessarily guarantee compliance.


Regulations often require specific controls regardless of where
the data is stored.

**Option D is incorrect.** While limiting administrative access can
reduce the risk of insider threats, it doesn’t directly ensure that
the regulated data is handled in compliance with legal
requirements.

**Question 241.** During an annual review, a company discovered
that one of its critical systems had several unscheduled
downtimes over the year. The CTO has recommended a move
towards high availability architecture to address this. What is
the PRIMARY concern when implementing high availability?
**(A) Ensuring that there are no single points of failure
(B)** Ensuring that the system is patched regularly
**(C)** Implementing multi-factor authentication
**(D)** Storing backups in multiple geographical locations

**Explanation 241. Correct Answer: A. Ensuring that there
are no single points of failure.** In a high availability (HA)
system, eliminating single points of failure is of utmost
importance. By ensuring redundancy at every potential failure
point, the system can remain operational even if a component
fails.

**Option B is incorrect.** While patching is important for security,
it is not the primary concern when implementing high
availability.


**Option C is incorrect.** Multi-factor authentication is essential
for secure access but does not directly relate to high availability
architecture.

**Option D is incorrect.** Having backups in various geographical
locations is more relevant to disaster recovery than to high
availability.

**Question 242.** After a recent service outage, a hospital’s IT
team is reviewing the availability of its patient record system.
They want to ensure the system remains operational, even in the
event of hardware failures. Which of the following
considerations is MOST relevant to this requirement?
**(A) Implementing database mirroring
(B)** Regularly updating the system's antivirus definitions
**(C)** Using strong encryption for data at rest
**(D)** Conducting penetration testing on the system

**Explanation 242. Correct Answer: A. Implementing
database mirroring.** Database mirroring is a solution for
increasing the availability of a SQL Server database. Mirroring
is implemented on a per-database basis and works only with
databases that use the full recovery model. This ensures that in
the event of a hardware failure, there’s a mirrored copy of the
data available.

**Option B is incorrect.** While updating antivirus definitions is
crucial for preventing malware infections, it is not directly


related to ensuring high availability in the face of hardware
failures.

**Option C is incorrect.** Encryption is vital for data security but
doesn’t directly address the high availability needs of a system.

**Option D is incorrect.** Penetration testing identifies
vulnerabilities in the system but doesn’t provide a solution for
high availability in case of hardware failures.

**Question 243.** An enterprise wants to configure its firewall so
that if a malfunction occurs, the firewall should automatically
allow traffic to ensure business continuity. Which failure mode
should be implemented?
**(A)** Fail-safe
**(B)** Fail-over
**(C)** Fail-closed
**(D) Fail-open**

**Explanation 243. Correct Answer: D. Fail-open.** In a fail-
open mode, when the system or device fails, it defaults to an
“open” state, allowing traffic to pass through. This mode
prioritizes availability over strict security.

**Option A is incorrect.** “Fail-safe” is a term that typically refers
to a system’s ability to default to a safe condition in case of a
failure. However, in the context of firewalls and access control
systems, “fail-open” or “fail-closed” are the preferred terms.

**Option B is incorrect.** “Fail-over” refers to switching to a
backup system or component in the event of a failure. It doesn’t


describe the behavior of allowing or denying traffic during a
failure.

**Option C is incorrect.** Fail-closed means that when a
malfunction occurs, the system would block all traffic by
default, which is the opposite of what the enterprise wants.

**Question 244.** An organization with a single physical network
infrastructure wants to separate the traffic of its finance
department from that of the HR department. They do not want
to set up entirely new physical networks but want to ensure that
data packets from one department do not mix with the other’s.
What should the organization implement?
**(A)** Air-gapped network
**(B)** DMZ (Demilitarized Zone)
**(C) VLAN (Virtual Local Area Network)
(D)** VPN (Virtual Private Network)

**Explanation 244. Correct Answer: C. VLAN (Virtual Local
Area Network).** A VLAN allows a network administrator to
create a logical network within a physical network. This logical
segmentation can separate the traffic of different departments,
ensuring that data packets from one VLAN do not mix with
another, effectively isolating them.

**Option A is incorrect.** An air-gapped network is a physically
isolated network and would require setting up a new physical
network.


**Option B is incorrect.** A DMZ is a buffer zone between a
private network and external networks. It doesn’t help in
segmenting traffic within the internal network.

**Option D is incorrect.** A VPN is used to create secure
connections over an unsecured network. It does not logically
segment traffic within a single physical network.

**Question 245.** A large organization is considering deploying a
solution that will allow employees to securely access company
resources remotely using their personal devices. The
organization wants a solution that can provide strong
authentication and ensure that the data remains confidential
during transit. Which technology should be adopted?
**(A)** Kerberos
**(B)** Remote Desktop Services (RDS)
**(C) Remote Access VPN
(D)** SNMP

**Explanation 245. Correct Answer: C. Remote Access VPN.**
Remote Access VPN provides a secure connection from a
remote device to an organization’s internal network. It uses
encryption to ensure data confidentiality and can also be set up
with strong authentication methods.

**Option A is incorrect.** Kerberos is an authentication protocol
that uses tickets to allow nodes to prove their identity in a
network. While it provides strong authentication, it doesn’t
inherently provide a solution for remote access with encrypted
communication.


**Option B is incorrect.** Remote Desktop Services (RDS) allows
users to access a remote desktop or application. While it can be
secured, it doesn’t always ensure encryption during transit,
especially if not paired with another technology like VPN.

**Option D is incorrect.** SNMP (Simple Network Management
Protocol) is used for managing and monitoring network devices.
It isn’t related to providing secure remote access for users.

**Question 246.** An e-commerce company wants to ensure that
their customers’ credit card data remains confidential while in
transit over the internet. They are seeking a protocol that can
help in securing their website’s communication. Which protocol
would best fit this requirement?
**(A)** IPSec
**(B)** SSH
**(C) TLS
(D)** ICMP

**Explanation 246. Correct Answer: C. TLS.** Transport Layer
Security (TLS) is a cryptographic protocol that ensures data
confidentiality and integrity over the internet. Websites use TLS
to secure all communications between their servers and web
browsers, making it suitable for e-commerce platforms to secure
sensitive customer data.

**Option A is incorrect.** IPSec is a suite of protocols that secures
IP communications by encrypting and authenticating all IP
packets. It is more commonly used for VPNs rather than
securing web communications.


**Option B is incorrect.** SSH (Secure Shell) is primarily used for
secure remote access to systems, not for securing web-based
communications.

**Option D is incorrect.** ICMP (Internet Control Message
Protocol) is used by network devices to send error messages and
operational information, not for securing data in transit.

**Question 247.** A rapidly growing e-commerce platform has
been facing intermittent downtimes, especially during sale
events. To ensure high availability and even distribution of
traffic among servers, the company is considering deploying a
specific type of network appliance. Which of the following
should the company deploy?
**(A)** Intrusion Detection System (IDS)
**(B)** VPN concentrator
**(C) Load balancer
(D)** Proxy server

**Explanation 247. Correct Answer: C. Load balancer.** A load
balancer is designed to distribute incoming network traffic
across multiple servers to ensure that no single server is
overwhelmed with too much traffic. This helps in ensuring high
availability and optimizing resource utilization.

**Option A is incorrect.** An Intrusion Detection System (IDS) is
primarily used for detecting malicious activities in the network,
not for distributing traffic among servers.


**Option B is incorrect.** A VPN concentrator is used to create
and manage VPN connections, not for balancing load among
servers.

**Option D is incorrect.** While a proxy server can distribute
client requests to different servers, its primary role is to act as
an intermediary, not to ensure high availability or evenly
distribute traffic.

**Question 248.** An organization is transitioning its IT
infrastructure to be cloud-centric and aims to adopt a zero-trust
network approach. They are looking for a solution that
integrates cloud security, zero-trust access, and WAN
capabilities, ensuring employees have consistent secure access
regardless of their location. Which of the following best
addresses their needs?
**(A)** Remote Desktop Services (RDS)
**(B) Secure access service edge (SASE)
(C)** Content Delivery Network (CDN)
**(D)** Virtual Local Area Network (VLAN)

**Explanation 248. Correct Answer: B. Secure access service
edge (SASE).** SASE integrates multiple network and security
functions traditionally delivered in separate point solutions. This
includes capabilities such as Secure Web Gateway (SWG),
Cloud Access Security Broker (CASB), Firewall-as-a-Service
(FWaaS), and Zero Trust Network Access (ZTNA), all delivered
from a globally distributed set of cloud-native platforms.

**Option A is incorrect.** Remote Desktop Services (RDS)
enables users to connect to a graphical interface of a remote


computer, but it doesn’t inherently combine cloud security,
zero-trust access, and WAN capabilities like SASE.

**Option C is incorrect.** A Content Delivery Network (CDN) is
designed to deliver web content and web applications to users
based on their geographical location. It doesn’t provide
integrated cloud security and WAN capabilities.

**Option D is incorrect.** A Virtual Local Area Network (VLAN)
segments a physical network into multiple isolated networks. It
doesn’t offer the combined benefits of cloud security, zero-trust
access, and WAN capabilities like SASE.

**Question 249.** A medium-sized e-commerce company recently
experienced a data breach due to an external attack. Post-
incident analysis revealed that while there were indications of
the attack in their network traffic, no alarms were raised at the
time of the attack. The company now wants to implement a
solution to actively monitor and take action against malicious
network traffic. Which of the following should they deploy?
**(A)** Intrusion Detection System (IDS)
**(B)** Network Access Control (NAC)
**(C)** Proxy server
**(D) Intrusion Prevention System (IPS)**

**Explanation 249. Correct Answer: D. Intrusion Prevention
System (IPS).** An Intrusion Prevention System (IPS) not only
detects malicious network activities but also takes active steps
to prevent or block them, which is ideal for a company looking
to bolster its defenses against real-time threats.


**Option A is incorrect.** An Intrusion Detection System (IDS)
will detect and alert on malicious activity but does not take
active steps to prevent the threat.

**Option B is incorrect.** Network Access Control (NAC) is
primarily used for controlling network access based on policies,
not for active monitoring and prevention of malicious network
traffic.

**Option C is incorrect.** A Proxy server acts as an intermediary
for requests from clients seeking resources but is not
specifically designed to detect or prevent intrusions.

**Question 250.** GreenTech, a data center company, is planning to
expand its operations in a region known for frequent power
outages. To maintain security posture and ensure continuity of
operations, which of the following should be their PRIMARY
consideration regarding power?
**(A)** Using power-efficient servers to reduce electricity costs
**(B)** Setting up solar panels to promote green energy
**(C) Investing in redundant power supplies and
uninterruptible power systems (UPS)
(D)** Running operations only during peak daylight hours to
ensure natural lighting

**Explanation 250. Correct Answer: C. Investing in
redundant power supplies and uninterruptible power
systems (UPS).** Given the frequent power outages in the region,
having redundant power supplies and UPS ensures that
operations continue smoothly without abrupt interruptions,


which can lead to system crashes, data corruption, or security
breaches.

**Option A is incorrect.** While power-efficient servers can
reduce costs, they do not address the primary concern of power
outages and maintaining operations.

**Option B is incorrect.** Solar panels promote green energy but
might not provide consistent and immediate power backup
during outages, especially during nighttime or cloudy days.

**Option D is incorrect.** Limiting operations to daylight hours is
not a practical solution for a data center, as it restricts
operational capacity and doesn’t directly address the issue of
power outages.

**Question 251.** A startup company anticipates rapid growth in its
user base over the next year. They are considering an
architectural model for their application that can handle the
projected growth without performance issues. Which of the
following would be the BEST design consideration for this
situation?
**(A)** Implementing strict password policies
**(B)** Using a monolithic application design
**(C)** Integrating a DDoS protection mechanism
**(D) Adopting a microservices architecture**

**Explanation 251. Correct Answer: D. Adopting a
microservices architecture.** Microservices architecture breaks
down an application into small, independent services that run as


separate processes. This allows for better scalability as each
service can be scaled individually based on the demand.

**Option A is incorrect.** While strict password policies are good
for security, they don’t directly address the scalability concerns
of the application architecture.

**Option B is incorrect.** Monolithic designs are often harder to
scale as changes or scaling in one area can affect the entire
application.

**Option C is incorrect.** While DDoS protection is essential for
defending against certain types of cyberattacks, it doesn’t
address the scalability of the application’s core architecture.

**Question 252.** A healthcare provider is updating its network
infrastructure. Due to the sensitive nature of the medical data
they handle, they want to ensure that any anomalies or
malicious activities in the network are immediately detected and
alerted. Which system should they primarily consider?
**(A)** Intrusion Prevention System (IPS)
**(B) Intrusion Detection System (IDS)
(C)** DHCP server
**(D)** VPN concentrator

**Explanation 252. Correct Answer: B. Intrusion Detection
System (IDS).** An Intrusion Detection System (IDS) actively
monitors network traffic for any signs of malicious activities or
policy violations and alerts the administrators. It is suitable for
organizations that want to ensure anomalies are promptly
detected.


**Option A is incorrect.** An Intrusion Prevention System (IPS)
does actively monitor and detect malicious traffic but also takes
measures to prevent it. While an IPS could be beneficial for the
healthcare provider, the primary requirement mentioned was for
detection and alerting, which is the primary function of an IDS.

**Option C is incorrect.** A DHCP (Dynamic Host Configuration
Protocol) server assigns IP addresses to devices in a network. It
does not monitor or alert on malicious network activities.

**Option D is incorrect.** A VPN concentrator is used to create
and manage VPN connections, providing secure access to a
network, but it doesn’t primarily detect intrusions.

**Question 253.** The IT security team at a large corporation is
evaluating monitoring tools for network traffic. They need a
solution that can inspect network packets without introducing
any potential latency or altering the network flow. Which type
of device attribute should they consider?
**(A)** Active IDS
**(B)** Passive firewall
**(C)** Active firewall
**(D) Passive IDS**

**Explanation 253. Correct Answer: D. Passive IDS.** Passive
IDS (Intrusion Detection System) is designed to monitor and
analyze network traffic without influencing the traffic flow or
causing potential latency. It observes traffic in real-time but
doesn’t take active actions on its own.


**Option A is incorrect.** An Active IDS can detect potential
security breaches and take action, which may introduce latency
or alter the traffic flow.

**Option B is incorrect.** While the term “passive firewall” is not
standard, traditional firewalls can actively block or allow traffic,
which can introduce changes to the network flow.

**Option C is incorrect.** An Active firewall actively filters
network traffic based on configured policies, potentially
affecting the network flow.

**Question 254.** A developer at your company is excited about
the scalability benefits of serverless architecture and has
deployed a new service using it. However, you notice an
increased bill due to the service even when it’s not in use.
Which of the following could be a contributing factor?
**(A) The serverless functions are continuously triggered
by unintended events
(B)** The server hardware is outdated
**(C)** The load balancer is misconfigured
**(D)** The organization lacks a Content Delivery Network
(CDN)

**Explanation 254. Correct Answer: A. The serverless
functions are continuously triggered by unintended events.**
Serverless architectures charge based on the number of function
invocations and the execution time. If there are unintended
events, such as rogue requests or misconfigured triggers,
continuously invoking the serverless functions, it could lead to
unexpected costs.


**Option B is incorrect.** In a serverless architecture, the
responsibility of server hardware management lies with the
service provider. The customer is abstracted from the hardware
details.

**Option C is incorrect.** A misconfigured load balancer might
affect availability or performance but is not directly related to
unexpected cost hikes in a serverless deployment.

**Option D is incorrect.** While CDNs are valuable for optimizing
content delivery, they are not directly related to the cost
implications of unintended function invocations in a serverless
architecture.

**Question 255.** After a recent security breach, CyberCorp is
reviewing its software vendors for their responsiveness to
vulnerabilities. Which of the following metrics would BEST
assist CyberCorp in determining the timeliness and efficiency of
security patches from a vendor?
**(A)** The frequency of software updates released by the
vendor
**(B)** The vendor's quarterly financial reports
**(C) Time between vulnerability disclosure and patch
release by the vendor
(D)** The number of features added by the vendor in the last
software update

**Explanation 255. Correct Answer: C. Time between
vulnerability disclosure and patch release by the vendor.** The
time taken by a vendor to release a security patch after a
vulnerability is disclosed is a direct measure of their


responsiveness to security threats. A shorter duration indicates a
higher prioritization of security concerns.

**Option A is incorrect.** Frequency of software updates does not
necessarily correlate with timely security patches. A vendor
might release frequent updates but still be slow in addressing
security issues.

**Option B is incorrect.** A vendor’s quarterly financial reports
might indicate the financial health of the company but does not
directly reflect their responsiveness to security vulnerabilities.

**Option D is incorrect.** The number of features added in a
software update does not indicate the timeliness or effectiveness
of security patches.

**Question 256.** XYZ Corp. has recently developed a new
manufacturing process that reduces production costs by 50%.
This process is not yet patented and is considered a trade secret.
The company wants to ensure that employees do not disclose
this process to competitors. Which of the following would be
the MOST effective strategy to achieve this?
**(A)** Providing employees with a bonus for keeping the
process confidential
**(B)** Conducting random checks of employee
communications
**(C) Implementing a mandatory non-disclosure
agreement (NDA) for all employees
(D)** Hosting quarterly seminars to educate employees about
the value of the trade secret


**Explanation 256. Correct Answer: C. Implementing a
mandatory non-disclosure agreement (NDA) for all
employees.** An NDA is a legally binding contract that prohibits
employees from disclosing confidential information, like trade
secrets. If an employee breaches the NDA, the company has
legal grounds to seek damages or other remedies.

**Option A is incorrect.** While bonuses might provide an
incentive for employees, it doesn’t provide a binding legal
framework to protect the trade secret.

**Option B is incorrect.** Random checks can deter some
employees but might not be effective across the board and could
also lead to trust issues within the organization.

**Option D is incorrect.** While education is essential and can
instill a sense of responsibility, it doesn’t provide a legal means
to protect the trade secret as an NDA does.

**Question 257.** A large e-commerce platform is facing
challenges during peak sale periods, where the influx of users
causes slowdowns and occasional outages. Which of the
following solutions would BEST improve scalability during
these high-demand times?
**(A)** Implement a centralized logging system
**(B) Employ auto-scaling cloud solutions
(C)** Increase the frequency of data backups
**(D)** Mandate regular security training for employees

**Explanation 257. Correct Answer: B. Employ auto-scaling
cloud solutions.**


Auto-scaling in cloud solutions automatically adjusts the
number of computational resources based on the actual demand.
During peak periods, resources can be automatically increased
to handle the demand, ensuring the system remains scalable and
responsive.

**Option A is incorrect.** A centralized logging system is
beneficial for monitoring and troubleshooting, but it does not
directly address scalability during peak times.

**Option C is incorrect.** Increasing the frequency of data
backups is a good practice for data integrity and recovery but
does not handle scalability concerns directly.

**Option D is incorrect.** Regular security training is crucial for a
company’s cybersecurity posture, but it does not address the
architectural scalability of the platform.

**Question 258.** An e-commerce company is preparing for an
upcoming Black Friday sale, expecting a surge in web traffic.
To ensure their systems remain responsive during the sale,
which of the following would be the MOST effective strategy to
implement?
**(A)** Increasing password complexity for all users
**(B)** Limiting the number of products on sale
**(C) Implementing a content delivery network (CDN)
(D)** Conducting a yearly security audit

**Explanation 258. Correct Answer: C. Implementing a
content delivery network (CDN).** CDNs distribute the traffic
load among multiple servers, often geographically dispersed.


This not only speeds up content delivery to users but also helps
in handling traffic surges, ensuring the system remains
responsive during high-demand periods such as Black Friday
sales.

**Option A is incorrect.** Increasing password complexity can
enhance security but does not directly impact system
responsiveness during high traffic periods.

**Option B is incorrect.** Limiting the number of products on sale
may reduce server load, but it might also decrease potential
revenue and is not a direct method for improving system
responsiveness.

**Option D is incorrect.** While yearly security audits are
essential for identifying vulnerabilities, they don’t directly
address system responsiveness during high traffic periods.

**Question 259.** A multinational corporation has data centers
located in different countries. Due to regulatory constraints,
remote access to these data centers is highly restricted. The
company’s IT administrators need a centralized way to access
all data centers securely without directly accessing them from
their workstations. Which solution should the company
consider?
**(A)** Setting up a DMZ
**(B) Implementing a Jump server
(C)** Deploying an Active Directory
**(D)** Using a local Proxy


**Explanation 259. Correct Answer: B. Implementing a Jump
server.** A Jump server, also known as a bastion host, acts as an
intermediary server allowing users to connect to it before
accessing another server or network. It provides a controlled
means of access between two networks, such as an internal
network and an external network.

**Option A is incorrect.** A DMZ (Demilitarized Zone) is a
physical or logical subnetwork that contains and exposes an
organization’s external-facing services to an untrusted network,
usually the internet. It doesn’t provide a consolidated access
point like a Jump server.

**Option C is incorrect.** Active Directory is a directory service
for Windows domain networks. It’s not designed to provide
controlled remote access to multiple data centers.

**Option D is incorrect.** A local Proxy might be used to control
internet access or cache web content but doesn’t act as a
centralized access point for data centers like a Jump server.

**Question 260.** Acme Corp is restructuring its internal network
to improve its security posture. They aim to separate areas with
different levels of trust. Which of the following approaches
would BEST achieve this objective?
**(A)** Implementing VLANs based on organizational
departments
**(B)** Setting up a perimeter firewall to segment external and
internal traffic
**(C) Designing network zones based on data sensitivity**


**and access requirements
(D)** Using a single, flat network for simplicity

**Explanation 260. Correct Answer: C. Designing network
zones based on data sensitivity and access requirements.**
Zoning based on data sensitivity and access requirements
ensures that segments of the network with varying trust levels
are isolated. This reduces the risk of a breach in one zone
affecting another.

**Option A is incorrect.** While VLANs can segment traffic,
doing so based purely on organizational departments may not
align with varying trust and data sensitivity requirements.

**Option B is incorrect.** While perimeter firewalls are crucial for
security, they primarily separate internal from external traffic
and don’t segment areas of varying trust within the internal
network.

**Option D is incorrect.** A single, flat network doesn’t provide
segmentation, making it vulnerable to lateral movement if a
threat actor gains access.

**Question 261.** Your organization is implementing Infrastructure
as Code (IaC) to deploy and manage its cloud infrastructure. As
part of a security review, what is a primary concern regarding
the use of IaC scripts?
**(A)** Lack of graphical interface for infrastructure
visualization
**(B) Hardcoding sensitive data within the scripts**


**(C)** Inability to scale the infrastructure dynamically
**(D)** Incompatibility with non-cloud environments

**Explanation 261. Correct Answer: B. Hardcoding sensitive
data within the scripts.** Infrastructure as Code (IaC) scripts are
meant to automate infrastructure deployment. If sensitive data
such as passwords or API keys are hardcoded into these scripts,
they could be exposed, leading to potential security breaches.
It’s essential to use secure methods, like secrets management or
encrypted variables, to handle sensitive data.

**Option A is incorrect.** While IaC focuses on code-driven
infrastructure management, many tools provide ways to
visualize the infrastructure or are compatible with tools that do.

**Option C is incorrect.** One of the benefits of IaC is the ability
to scale infrastructure dynamically. It allows for automated
scaling based on the script or configuration files.

**Option D is incorrect.** IaC can be used in various
environments, not just cloud. It depends on the tool and the
platform it supports.

**Question 262.** A cloud-based SaaS company wants to ensure its
infrastructure can handle a potential influx of a large number of
users in the future. Which of the following approaches would
BEST meet this scalability consideration?
**(A) Implement a horizontal scaling strategy
(B)** Introduce multi-factor authentication
**(C)** Deploy deep packet inspection tools
**(D)** Implement a centralized logging system


**Explanation 262. Correct Answer: A. Implement a
horizontal scaling strategy.** Horizontal scaling involves adding
more machines or nodes to a system to handle increased load,
which can be especially effective in a cloud environment. It
allows the infrastructure to accommodate more users by
distributing the load among multiple servers.

**Option B is incorrect.** Multi-factor authentication is essential
for security, but it does not address the scalability of the
infrastructure to handle a large number of users.

**Option C is incorrect.** Deep packet inspection tools are used
for analyzing network traffic for security and management
purposes, but they don’t directly enhance the scalability of an
infrastructure.

**Option D is incorrect.** Centralized logging systems help
consolidate and analyze logs from various sources, but they do
not address the scalability needs of the infrastructure itself.

**Question 263.** A software development company is looking to
migrate its legacy applications to a more modern infrastructure.
They want to ensure the applications can be deployed
consistently across multiple environments without the
challenges of varying dependencies and configurations. Which
approach would best achieve this goal?
**(A)** Virtual Machine Deployment
**(B)** Bare-Metal Deployment
**(C) Containerization
(D)** Serverless Computing


**Explanation 263. Correct Answer: C. Containerization**
Containerization encapsulates an application and its
dependencies into a ‘container.’ This allows the application to
be consistently deployed across different environments
regardless of variations in configurations or dependencies.

**Option A is incorrect.** While virtual machines provide an
isolated environment for applications, they include the overhead
of an entire guest OS and don’t directly address the challenge of
application dependencies varying across environments.

**Option B is incorrect.** Bare-metal deployment involves
installing applications directly on hardware without
virtualization or containers. This method does not inherently
ensure consistent deployment across multiple environments.

**Option D is incorrect.** Serverless computing abstracts the
infrastructure away from the developer, focusing on individual
functions. While it provides scalability and simplicity, it doesn’t
directly address consistent deployment across environments due
to dependencies and configurations.

**Question 264.** A large enterprise is considering a transition to a
more flexible and programmable network architecture. They
want to centralize the control plane, allowing for automated,
programmable network configurations and rapid provisioning.
Which network architecture model should they consider?
**(A)** VLAN (Virtual Local Area Network)
**(B)** MPLS (Multiprotocol Label Switching)
**(C)** VPN (Virtual Private Network)
**(D) SDN (Software-Defined Networking)**


**Explanation 264. Correct Answer: D. SDN (Software-
Defined Networking).** Software-Defined Networking (SDN)
centralizes the control plane, decoupling it from the data plane.
This allows for a more flexible and programmable network,
enabling automated configurations, and rapid provisioning.

**Option A is incorrect.** A VLAN is used to create logical
networks within a physical network but doesn’t provide
centralized control or programmability across the entire
network.

**Option B is incorrect.** MPLS is a type of data-carrying
technique for high-performance telecommunications networks.
It doesn’t centralize network control or enhance
programmability as SDN does.

**Option C is incorrect.** A VPN is designed to provide a secure
tunnel between networks over potentially insecure mediums like
the internet. It does not offer the centralized, programmable
configuration benefits of SDN.

**Question 265.** A global corporation has undergone several IT
incidents in the past year, including outages due to natural
disasters and cyber attacks. The CEO wants to ensure the
organization’s IT infrastructure can withstand or rapidly recover
from disruptive events. Which of the following best
encapsulates this requirement?
**(A)** Adopting a Zero Trust Architecture
**(B)** Implementing a strict password policy
**(C) Establishing a Business Continuity Plan (BCP) with**


**emphasis on resilience
(D)** Regularly updating firewall rules

**Explanation 265. Correct Answer: C. Establishing a
Business Continuity Plan (BCP) with emphasis on resilience.**

Resilience in this context refers to the ability of the IT
infrastructure to rapidly recover and continue functioning even
after disruptive events. A Business Continuity Plan with a focus
on resilience would help the organization prepare for, respond
to, and recover from both natural and man-made incidents.

**Option A is incorrect.** Zero Trust Architecture emphasizes not
trusting any user or system, both inside and outside the
perimeter, but it doesn’t address resilience against disruptive
events directly.

**Option B is incorrect.** While password policies are essential
for security, they are not centered around ensuring resilience
against disruptive events.

**Option D is incorrect.** Updating firewall rules is crucial for
keeping out unwanted traffic and potential threats but doesn’t
directly address the resilience of an IT infrastructure against
major disruptive events.

**Question 266.** A medical company has recently deployed a
device to monitor patient heart rates in real time. This device
uses a real-time operating system (RTOS) to guarantee
immediate response times. The security team is concerned about
potential risks. Which of the following would be a KEY


recommendation to enhance the security of such devices?
**(A)** Ensure real-time data analysis capabilities
**(B)** Integrate the device with the corporate cloud for
backups
**(C) Implement strict network segmentation for the
device
(D)** Increase the storage capacity of the device

**Explanation 266. Correct Answer: C. Implement strict
network segmentation for the device.** To protect RTOS
devices, which often prioritize performance over security, it’s
crucial to minimize their exposure to potential threats. By
segmenting the network, you can isolate the device from other
systems and reduce the risk of a security incident.

**Option A is incorrect.** Real-time data analysis is more about
performance and functionality than security.

**Option B is incorrect.** Integrating the device with the corporate
cloud could introduce more security concerns, especially if the
cloud environment is not secured properly.

**Option D is incorrect.** Storage capacity is a matter of device
functionality, not a direct security enhancement for an RTOS
device.

**Question 267.** A small business wants to deploy a single
network security device that can handle multiple security
functions such as firewall protection, intrusion detection, anti-
malware, and content filtering. Which of the following would
be the most suitable solution?


**(A)** Network Intrusion Detection System (NIDS)
**(B)** Web Application Firewall (WAF)
**(C) Unified Threat Management (UTM)
(D)** Proxy Server

**Explanation 267. Correct Answer: C. Unified Threat
Management (UTM).** Unified Threat Management (UTM)
devices are designed to combine multiple security features into
a single appliance. This makes them ideal for smaller
organizations that require a comprehensive range of security
functions but might not have the resources to deploy and
manage multiple standalone devices.

**Option A is incorrect.** While NIDS can identify and notify of
potential malicious activities, it does not offer the
comprehensive multi-feature capabilities found in a UTM.

**Option B is incorrect.** A Web Application Firewall (WAF)
specifically protects web applications from certain types of
attacks like XSS and SQL injection. It does not encompass the
broader range of security functions that a UTM does.

**Option D is incorrect.** A Proxy Server primarily controls
internet access and might provide some caching and content
filtering capabilities, but it doesn’t offer the wide range of
security functions that a UTM does.

**Question 268.** An e-commerce company is experiencing attacks
that specifically target the shopping cart feature of its web
application. They want to implement a firewall that can
understand web application-specific commands and provide


protection. Which type of firewall should they consider?
**(A)** Layer 4 Firewall
**(B)** Layer 2 Firewall
**(C) Layer 7 Firewall
(D)** Packet Filtering Firewall

**Explanation 268. Correct Answer: C. Layer 7 Firewall.**
Layer 7 firewalls, often known as Application Layer firewalls,
can understand and filter traffic based on application-specific
data, commands, or functions. They operate at the highest layer
in the OSI model and can make decisions based on the actual
content of the traffic.

**Option A is incorrect.** Layer 4 firewalls, also known as
transport layer firewalls, primarily deal with data based on port
numbers and protocol. They wouldn’t be as effective in filtering
application-specific commands as a Layer 7 firewall.

**Option B is incorrect.** Layer 2 firewalls operate at the data link
layer, primarily dealing with MAC addresses. They aren’t
equipped to analyze application-specific content.

**Option D is incorrect.** Packet Filtering Firewalls operate
primarily at the network layer, making decisions based on
source/destination IP addresses, port numbers, and protocol
types, not application-specific content.

**Question 269.** A financial organization’s high-security data
center has an authentication system for its main entry. If the
system encounters an unexpected error, the organization wants
to ensure that no one can gain access to the data center until the


system is fixed. Which configuration should the authentication
system be set to?
**(A)** Fail-open
**(B) Fail-closed
(C)** Fail-secure
**(D)** Fail-passive

**Explanation 269. Correct Answer: B. Fail-closed.** Fail-
closed, also known as fail-secure, is a mode where if a system
fails, it denies all requests by default, effectively “closing”
access. In high-security environments, it’s often preferred to
prevent any unauthorized access during system failures.

**Option A is incorrect.** Fail-open is the opposite of what the
organization wants. In a fail-open scenario, if the system fails, it
would allow all requests by default.

**Option C is incorrect.** While “fail-secure” is another term for
“fail-closed”, the use of both terms in the options can confuse
the test-taker. In this context, option B is the more appropriate
choice.

**Option D is incorrect.** Fail-passive isn’t a standard term in the
context of authentication failure modes. Fail-open and fail-
closed (or fail-secure) are the primary recognized terms.

**Question 270.** A multinational company wants to allow its
remote employees to securely access the corporate intranet over
the Internet. The company’s primary concern is to ensure data
confidentiality and integrity during transit. Which solution
should the company implement?


**(A)** VLAN
**(B) VPN
(C)** NAC
**(D)** DMZ

**Explanation 270. Correct Answer: B. VPN.** A Virtual Private
Network (VPN) allows remote users to establish a secure
connection to an organization’s internal network over the public
internet. It encrypts the data in transit, ensuring both
confidentiality and integrity.

**Option A is incorrect.** A VLAN (Virtual Local Area Network)
is used to segment network traffic within a local network, but it
does not provide encryption or a method for secure remote
access.

**Option C is incorrect.** NAC (Network Access Control) is a
security solution that enforces policy on devices attempting to
access network resources, but it doesn’t inherently provide
encrypted communication over the internet.

**Option D is incorrect.** DMZ (Demilitarized Zone) is a network
segment that acts as a buffer between an organization’s internal
network and untrusted external networks. It doesn’t provide
encrypted communication for remote users.

**Question 271.** After conducting a business impact analysis, a
local library determined that they could afford several days of
downtime without a significant impact on their operations. They
want a disaster recovery solution that offers a balance between
cost and recovery capabilities. Which site consideration is the


most suitable for the library’s needs?
**(A)** Mobile site with a full set of IT equipment
**(B)** Hot site with daily data replication
**(C) Cold site
(D)** Warm site with weekly backups

**Explanation 271. Correct Answer: C. Cold site.** A cold site is
an offsite facility that is pre-configured to accept equipment and
data backups but doesn’t have any of its own. Given the
library’s tolerance for several days of downtime, a cold site
would be a cost-effective solution that aligns with their recovery
capabilities.

**Option A is incorrect.** A mobile site with a full set of IT
equipment would be more suitable for immediate recovery
needs and might be overkill for a library that can tolerate
several days of downtime.

**Option B is incorrect.** A hot site with daily data replication
would be more suited for businesses with a low tolerance for
downtime and would likely be more expensive than the library’s
budget allows.

**Option D is incorrect.** Even though a warm site with weekly
backups might seem like a viable option, a cold site would still
be more cost-effective for the library given their tolerance for
downtime.

**Question 272.** An online payment gateway is evaluating
methods to enhance the security of its payment processing
system. They want a solution that replaces sensitive cardholder


data with a unique identifier that has no intrinsic value by itself.
Which method should they implement?
**(A)** Hashing the card data
**(B)** Encrypting the card data
**(C)** Masking the card data
**(D) Tokenizing the card data**

**Explanation 272. Correct Answer: D. Tokenizing the card
data.** Tokenization involves replacing sensitive data with a non-
sensitive equivalent, referred to as a token, that has no extrinsic
or exploitable meaning or value.

**Option A is incorrect.** Hashing would produce a fixed-size
string, which cannot be reversed to its original value, but it
doesn’t create a “token” that can be swapped back for the
original data.

**Option B is incorrect.** Encryption translates data into another
form, or code, so that only people with access to a secret key or
password can read it. It does not replace the data with a unique
identifier with no intrinsic value.

**Option C is incorrect.** Masking involves hiding parts of data,
such as showing only the last four digits of a credit card
number. It doesn’t replace the data with a unique identifier.

**Question 273.** A software company is developing a new cloud-
based application where clients can store and manage their
customer contact details. The stored data does not include
financial or medical information but does contain phone
numbers and email addresses. How should this information be


classified?
**(A)** Public
**(B)** Restricted
**(C) Sensitive
(D)** Classified

**Explanation 273. Correct Answer: C. Sensitive.** Phone
numbers and email addresses are considered Personally
Identifiable Information (PII) and should be classified as
sensitive since unauthorized access could lead to privacy
violations or misuse.

**Option A is incorrect.** Public classification is for data that is
intended for general access and holds no confidential
information.

**Option B is incorrect.** Restricted typically applies to data that
is more confidential than sensitive data, such as financial or
strategic data.

**Option D is incorrect.** Classified is a more general term and
can encompass various levels of data classification, including
Top Secret, Secret, or Confidential, typically used in
government sectors.

**Question 274.** Cybertech Corp. is reviewing its backup
protocols to enhance security. They decided that before
transferring backups to an offsite location, the data should be
rendered unreadable to unauthorized individuals. Which backup
strategy will ensure this requirement is met?
**(A)** Use deduplication before storing backups


**(B)** Store backups in proprietary formats
**(C) Encrypt backups before transfer
(D)** Compress backups using standard tools

**Explanation 274. Correct Answer: C. Encrypt backups
before transfer.** Encryption transforms data into a format that
can only be read by someone who possesses the decryption key.
By encrypting backups before transferring, Cybertech Corp.
ensures that even if unauthorized individuals access the backup
data, they won’t be able to understand its contents.

**Option A is incorrect.** Deduplication reduces the storage space
required for backups by eliminating redundant data but doesn’t
make data unreadable to unauthorized individuals.

**Option B is incorrect.** Using proprietary formats may obscure
the data for some unauthorized users, but it does not provide the
robust protection encryption does.

**Option D is incorrect.** Compressing backups using standard
tools optimizes storage but does not ensure data confidentiality.

**Question 275.** After a major power outage, TechWave Corp.
wants to ensure their primary data center’s operations can be
swiftly and seamlessly shifted to another facility. The company
wants this process to be automated and immediate to prevent
any service interruption. What type of test should TechWave
Corp. conduct to validate this functionality?
**(A)** Conduct a vulnerability assessment on both data centers
**(B) Execute a failover test**


**(C)** Engage in a tabletop exercise
**(D)** Perform a routine backup test

**Explanation 275. Correct Answer: B. Execute a failover test.**
A failover test involves simulating a failure in the primary
system to ensure that the backup system can seamlessly take
over without any service interruption. It helps in verifying the
resilience and recovery capabilities of an organization’s
infrastructure.

**Option A is incorrect.** A vulnerability assessment identifies
potential weaknesses in systems but does not test the seamless
transition between primary and backup systems.

**Option C is incorrect.** A tabletop exercise is a discussion-based
session and does not involve the live testing of systems or
processes.

**Option D is incorrect.** A routine backup test ensures that
backups can be restored successfully but does not test the
immediate switchover capability between systems.

**Question 276.** A pharmaceutical company is working on a new
drug formula that promises to revolutionize the treatment of a
particular disease. The R&D team has detailed documentation
on the components, procedures, and results of the drug trials.
How should this documentation be classified to ensure that only
the right people within the company have access?
**(A)** Unclassified
**(B)** Public


**(C) Confidential
(D)** Sensitive

**Explanation 276. Correct Answer: C. Confidential.** This drug
formula and its documentation represent valuable intellectual
property for the pharmaceutical company. To protect against
theft, corporate espionage, or accidental disclosure, it should be
classified as confidential to ensure limited, controlled access.

**Option A is incorrect.** Unclassified data doesn’t require any
special protection or confidentiality and is inappropriate for
valuable intellectual property.

**Option B is incorrect.** Public classification would make the
information accessible to everyone, which is not suitable for
proprietary drug formulas.

**Option D is incorrect.** While the drug formula is sensitive
information, the “Confidential” classification offers stricter
controls and is more suitable for such valuable intellectual
property.

**Question 277.** A software development company maintains a
shared code repository. The company wants to ensure that only
developers can make changes to the code, but testers should be
able to view the code without modifying it. Which approach
should be implemented?
**(A)** Assign all employees read-only permissions
**(B)** Provide testers with administrative rights
**(C) Implement role-based access controls (RBAC)
(D)** Use data encryption on the repository


**Explanation 277. Correct Answer: C. Implement role-based
access controls (RBAC).** RBAC assigns permissions based on
roles within an organization. Developers can be given a role that
allows them to modify code, while testers are given a role that
only allows them to read the code.

**Option A is incorrect.** Assigning read-only permissions to all
employees will prevent developers from making necessary
changes to the code.

**Option B is incorrect.** Providing testers with administrative
rights would give them more access than necessary, potentially
allowing them to modify the code.

**Option D is incorrect.** While data encryption ensures the
confidentiality of the data, it does not address the need for
specific permission levels for different user roles.

**Question 278.** A software development company has just
expanded its team and wants to ensure that in the event of a
disaster, they can resume operations within a day. However,
they have a limited budget for disaster recovery. Which of the
following disaster recovery site types would be the most
suitable for the company’s needs?
**(A)** Hot site with hourly data replication
**(B)** Cold site with monthly data backups
**(C) Warm site with daily backups
(D)** Offsite tape backups

**Explanation 278. Correct Answer: C. Warm site with daily
backups.** For the software development company’s need to


resume operations within a day, a warm site with daily backups
would be ideal. The warm site has equipment and some level of
data ready, which would facilitate a faster recovery compared to
a cold site or just offsite backups.

**Option A is incorrect.** A hot site with hourly data replication
might provide the quickest recovery, but it’s also the most
expensive option which may not align with the company’s
budget constraints.

**Option B is incorrect.** A cold site with monthly data backups
would be cost-effective but might not allow for operations to
resume within a day due to the time needed to set up equipment
and restore the latest data.

**Option D is incorrect.** Offsite tape backups alone would not
provide the infrastructure or equipment needed for quick
recovery. It would also take considerable time to restore
operations just from tape backups.

**Question 279.** A financial institution wants to ensure that if a
cyber attacker gains unauthorized access to one section of their
network, the attacker cannot easily move laterally to more
sensitive parts of the network. Which method can help achieve
this objective?
**(A)** Deploying a honeypot in every segment
**(B) Implementing network segmentation
(C)** Applying encryption on all data traffic
**(D)** Enabling two-factor authentication for all users


**Explanation 279. Correct Answer: B. Implementing
network segmentation.** Network segmentation divides a
network into multiple segments or subnets. Each segment
operates independently and can have its own security and access
controls. This means that if an attacker compromises one
segment, they won’t necessarily have access to other segments.

**Option A is incorrect.** A honeypot is a decoy system to attract
potential attackers, but it doesn’t prevent lateral movement
within the network once an attacker has access.

**Option C is incorrect.** While encrypting data traffic ensures
confidentiality, it doesn’t prevent lateral movement within the
network.

**Option D is incorrect.** Two-factor authentication strengthens
access controls but doesn’t inherently stop lateral movement
within a network if an attacker gains access.

**Question 280.** DataFin, a financial analytics firm, experienced a
minor fire incident in one of its server rooms. Fortunately, they
had backups stored in another wing of the building, allowing for
quick data recovery. However, management realizes that in a
major disaster, both primary and backup data might be
destroyed. To address this, which backup strategy should
DataFin consider?
**(A)** Mirror Backup
**(B)** Local Storage Backup
**(C)** Incremental Backup
**(D) Offsite Backup**


**Explanation 280. Correct Answer: D. Offsite Backup.** Offsite
Backup involves storing backup data in a different geographical
location from the primary data. This ensures that even if a
disaster impacts the primary location, the backup remains safe
and can be used for recovery.

**Option A is incorrect.** Mirror Backup refers to a backup that is
an exact copy of the source data. It doesn’t specify the
geographic location of the backup.

**Option B is incorrect.** Local Storage Backup typically means
backups are stored within the same physical premises or close
vicinity of the primary data, which wouldn’t address DataFin’s
concern of potential data loss in a major disaster.

**Option C is incorrect.** Incremental Backup involves backing
up only the data that has changed since the last backup,
regardless of where it’s stored. It doesn’t address the concern of
geographic separation.

**Question 281.** A healthcare organization with patients
worldwide is planning to set up a backup site for its medical
data repository. They have been advised to consider geographic
dispersion as part of their disaster recovery plan. Which of the
following reasons is the LEAST valid for geographic dispersion
in this scenario?
**(A)** Mitigate risks of regional natural disasters
**(B)** Offer redundancy in case of local power outages
**(C)** Benefit from varying peak load times in different
regions
**(D) Ensure faster access speeds for global patients**


**Explanation 281. Correct Answer: D. Ensure faster access
speeds for global patients.** Geographic dispersion’s primary
goal in a disaster recovery context is resilience and redundancy,
not necessarily optimizing access speeds. While having servers
in different locations can enhance speed for local users, in the
context of a backup site for disaster recovery, the primary focus
is on availability and redundancy, not speed.

**Option A is incorrect.** Geographic dispersion can help mitigate
risks associated with regional natural disasters by ensuring that
backup sites aren’t impacted by the same disaster that affects
the primary site.

**Option B is incorrect.** Regional power outages can be
mitigated by having backup sites in different locations.

**Option C is incorrect.** Geographic dispersion can allow
organizations to manage and distribute load more effectively by
leveraging off-peak times in different regions.

**Question 282.** A multinational corporation is expanding its
operations in various countries. The company has decided to
restrict access to its internal network based on geolocation to
ensure that only employees from specific countries can access
certain data. Which of the following would be the MOST
appropriate solution to implement this requirement?
**(A)** Deploy a VPN with multi-factor authentication
**(B)** Use MAC address filtering on all company devices
**(C) Implement a geolocation-based access control system
(D)** Set up region-specific SSIDs for the company's Wi-Fi
network


**Explanation 282. Correct Answer: C. Implement a
geolocation-based access control system.** Geolocation-based
access control systems determine users’ physical locations and
can grant or deny access based on predefined geographic
boundaries. This solution is most fitting for restricting access
based on country-specific requirements.

**Option A is incorrect.** While a VPN with multi-factor
authentication improves security, it doesn’t inherently restrict
access based on geolocation.

**Option B is incorrect.** MAC address filtering restricts access
based on device hardware addresses and not on geolocation.

**Option D is incorrect.** Region-specific SSIDs might limit Wi-
Fi access in certain areas, but they don’t enforce geolocation-
based restrictions on a broader scale.

**Question 283.** DigitalFront, an e-commerce company, is
expecting a surge in traffic during their upcoming annual sale
event. They want to ensure that their website and applications
can handle the anticipated increase in user activity without any
performance degradation. Which of the following steps is
MOST relevant to achieving this goal?
**(A)** Increasing the frequency of security audits
**(B) Implementing capacity planning specifically focused
on technology
(C)** Adopting multi-factor authentication for all users
**(D)** Investing in advanced threat intelligence solutions


**Explanation 283. Correct Answer: B. Implementing
capacity planning specifically focused on technology.**
Capacity planning with a focus on technology ensures that all
technological resources, such as servers, bandwidth, storage,
and software, are adequately prepared to handle the expected
increase in load or traffic, ensuring smooth performance during
peak periods.

**Option A is incorrect.** Security audits are important for
identifying vulnerabilities, but they don’t address the direct
need for scaling technology resources to manage increased
traffic.

**Option C is incorrect.** Multi-factor authentication improves
security for users but does not aid in handling a surge in website
traffic or user activity.

**Option D is incorrect.** Advanced threat intelligence solutions
provide insights into potential threats, but they don’t address
technology scaling to accommodate increased traffic.

**Question 284.** DeltaTech, a financial institution, operates its
primary site on a UNIX-based platform. For disaster recovery
purposes, they are considering setting up a backup site on a
different platform. Which of the following is NOT a primary
benefit of introducing platform diversity in this case?
**(A) It reduces the organization's learning curve by using
familiar technologies
(B)** It provides resilience against attacks targeting UNIX-
based systems
**(C)** It ensures that platform-specific outages don't affect


both primary and backup sites
**(D)** It diversifies the attack surface, reducing the impact of
specific platform vulnerabilities

**Explanation 284. Correct Answer: A. It reduces the
organization’s learning curve by using familiar technologies.**
Introducing a new platform usually increases the learning curve
as staff need to be trained on the new system. The primary
benefits of platform diversity are resilience, redundancy, and
diversifying the attack surface.

**Option B is incorrect.** This is one of the benefits of platform
diversity. If a vulnerability or threat targets UNIX-based
systems, having a backup on a different platform ensures that
the backup isn’t compromised by the same threat.

**Option C is incorrect.** This is a benefit of platform diversity.
Platform-specific outages or issues would only affect systems
on that platform, so having a backup on a different platform
adds resilience.

**Option D is incorrect.** This is one of the benefits of platform
diversity. By using different platforms, the organization reduces
the risk associated with vulnerabilities specific to a single
platform.

**Question 285.** After a recent system upgrade, CloudTech Corp.
decided to validate the efficiency and reliability of its new data
processing system. To do this, they run the new system
alongside the old one and compare the outcomes. This way, they
aim to ensure that the new system is both robust and capable of


handling the current workload. What kind of testing is
CloudTech Corp. utilizing?
**(A)** Load Testing
**(B)** Failover Testing
**(C) Parallel Processing Testing
(D)** Simulation Testing

**Explanation 285. Correct Answer: C. Parallel Processing
Testing.** Parallel Processing Testing involves running two
systems simultaneously (typically the new one and the old one)
to compare the outcomes. This kind of testing ensures that a
new system is as effective and reliable as the previous one.

**Option A is incorrect.** Load Testing checks the system’s ability
to handle the expected volume of transactions and to see if it
can maintain acceptable response times.

**Option B is incorrect.** Failover Testing is done to ensure that a
system can switch to a backup or secondary system in the case
of a failure, rather than comparing two systems for consistency.

**Option D is incorrect.** Simulation Testing involves creating a
model of the system under test and stimulating it with virtual
users or devices to understand its behavior under various
conditions.

**Question 286.** TechSolutions Inc., a rapidly growing startup, is
expanding its workforce to meet its customer demands. As part
of this expansion, they need to ensure their IT infrastructure can
accommodate the influx of new employees without
compromising performance or security. Which of the following


should be TechSolutions’ primary focus during this expansion
phase?
**(A)** Adopting a Zero Trust Network Architecture
**(B)** Increasing the frequency of vulnerability assessments
**(C) Implementing capacity planning
(D)** Deploying additional firewalls and intrusion detection
systems

**Explanation 286. Correct Answer: C. Implementing
capacity planning.** Capacity planning ensures that an
organization’s infrastructure, technology, and people are scaled
appropriately to meet growth and performance needs. In the
scenario, with an influx of new employees, ensuring that the IT
infrastructure can handle the increased demand is crucial.

**Option A is incorrect.** While adopting a Zero Trust Network
Architecture might enhance security, it doesn’t address the need
to scale resources to accommodate new employees.

**Option B is incorrect.** Vulnerability assessments are crucial for
security, but they don’t directly address the infrastructure’s
ability to handle growth.

**Option D is incorrect.** While firewalls and IDS can improve
security, they don’t directly address the organization’s capacity
to handle more employees

**Question 287.** StreamNet, a popular online streaming service, is
planning to launch in three new countries. They anticipate a
substantial increase in users and concurrent streams. To ensure
that the service remains uninterrupted and provides a seamless


experience to new users, which action related to capacity
planning should StreamNet prioritize?
**(A)** Investing in content encryption and DRM
**(B)** Increasing marketing and promotional activities in the
new countries
**(C)** Implementing stronger user authentication methods
**(D) Expanding and optimizing their infrastructure to
handle the projected growth**

**Explanation 287. Correct Answer: D. Expanding and
optimizing their infrastructure to handle the projected
growth.** Infrastructure capacity planning will allow StreamNet
to predict the necessary resources required and make the needed
upgrades or adjustments to their servers, networks, storage, and
other infrastructure components. This ensures a smooth user
experience even with the influx of new users.

**Option A is incorrect.** While content encryption and DRM
(Digital Rights Management) are essential for protecting
content, they don’t address the infrastructure capacity needed to
handle more users.

**Option B is incorrect.** Marketing and promotional activities
can attract users but don’t directly impact the infrastructure’s
capacity to support those users.

**Option C is incorrect.** Stronger authentication methods can
improve security but don’t address the capacity or scalability of
the infrastructure.


**Question 288.** GlobalTech is implementing a disaster recovery
plan and wants to ensure continuous availability with no data
loss. They have decided to use replication as a backup strategy.
Which of the following replication techniques should
GlobalTech implement to achieve their objective?
**(A)** Periodic replication scheduled daily
**(B)** Asynchronous replication with hourly synchronization
**(C) Synchronous replication
(D)** Snapshot replication every 30 minutes

**Explanation 288. Correct Answer: C. Synchronous
replication.** Synchronous replication ensures that data is written
to the primary and secondary locations simultaneously, ensuring
no data loss and continuous availability. This method is suitable
for GlobalTech’s requirement of no data loss.

**Option A is incorrect.** Periodic replication scheduled daily
could result in up to a day’s worth of data loss in the event of a
disaster.

**Option B is incorrect.** Asynchronous replication with hourly
synchronization could lead to an hour’s worth of data loss if an
incident occurs just before synchronization.

**Option D is incorrect.** Snapshot replication every 30 minutes
could result in 30 minutes of data loss in the event of an
incident.

**Question 289.** A software company has developed a new
product. They want to release a user manual that details how to
use the software, its features, and basic troubleshooting steps.


What should be the classification of this user manual?
**(A)** Confidential
**(B)** Restricted
**(C) Public
(D)** Internal

**Explanation 289. Correct Answer: C. Public.** A user manual
is intended for all users of the software and should be easily
accessible. It doesn’t contain sensitive or proprietary
information about the software’s underlying code or algorithms.
Hence, it should be classified as “Public.”

**Option A is incorrect.** “Confidential” classification is too
restrictive for a document meant to be distributed with every
software copy and to guide users.

**Option B is incorrect.** “Restricted” would limit the distribution
of the manual unnecessarily, making it harder for users to access
the information they need.

**Option D is incorrect.** “Internal” classification would suggest
the manual is only for internal company use, whereas it’s
intended for all users of the software.

**Question 290.** A global finance firm has recently faced
downtime due to unexpected disasters in its main operational
region. The firm wishes to have a backup site that would allow
them to continue their operations with minimal downtime and
no data loss. Which type of backup site would be the most
appropriate for the firm?
**(A)** Cold site


**(B)** Warm site
**(C) Hot site
(D)** Mobile site

**Explanation 290. Correct Answer: C. Hot site.** A hot site is a
fully operational offsite data center configured and ready to
continue operations immediately after a disaster. It has all the
necessary equipment and up-to-date data backups, making it
ideal for scenarios where minimal downtime and no data loss
are essential.

**Option A is incorrect.** A cold site is a backup facility with
space and infrastructure but without updated equipment or data.
It requires time to set up and is not suitable for immediate
recovery.

**Option B is incorrect.** A warm site is a middle ground between
a cold and hot site. It has some equipment and backups but will
require more time to get fully operational compared to a hot
site.

**Option D is incorrect.** A mobile site is usually a portable setup,
often on a truck, that can be deployed to various locations. It
might not be equipped for large-scale operations like a global
finance firm would require.

**Question 291.** An international bank is setting up a new online
portal for its customers to access their financial statements.
Which measure should the bank implement to ensure that
financial data in transit between the customer’s browser and the
bank’s servers is kept confidential?


**(A)** Use file-level encryption for all financial statements
**(B)** Implement a Web Application Firewall (WAF)
**(C) Use Secure Sockets Layer (SSL) or Transport Layer
Security (TLS) for the portal
(D)** Store all financial data in an encrypted database

**Explanation 291. Correct Answer: C. Use Secure Sockets
Layer (SSL) or Transport Layer Security (TLS) for the
portal.** SSL/TLS is designed to secure data in transit. By
encrypting the data between the customer’s browser and the
bank’s servers, it ensures the data’s confidentiality while in
transit.

**Option A is incorrect.** File-level encryption is great for data at
rest but doesn’t secure data in transit between a client and
server.

**Option B is incorrect.** While a WAF can provide protection
against web application attacks, it does not inherently encrypt
data in transit.

**Option D is incorrect.** Storing financial data in an encrypted
database safeguards the data at rest but doesn’t ensure the
confidentiality of data in transit to the client.

**Question 292.** A global e-commerce website wants to allow its
customer service representatives to assist clients with order
issues without exposing the full credit card details of the clients.
Which method should the IT department employ to achieve
this?
**(A)** Replace all digits of the credit card number with random


characters
**(B) Display only the last four digits of the credit card
number while masking the rest
(C)** Encrypt the credit card number with a symmetric key
**(D)** Use a hash function to represent the credit card number

**Explanation 292. Correct Answer: B. Display only the last
four digits of the credit card number while masking the rest.**
Displaying only the last four digits and masking the rest ensures
that customer service representatives have enough information
to assist customers without being exposed to the entire credit
card number.

**Option A is incorrect.** Replacing all digits with random
characters would render the information useless for customer
service representatives as they would not have any reference
point.

**Option C is incorrect.** Encrypting the card number can protect
the data, but the customer service representatives would need a
way to decrypt it to access even partial information, defeating
the purpose of not exposing them to the full details.

**Option D is incorrect.** Hashing the number would not allow
any portion of the original data, like the last four digits, to be
viewed, making it ineffective for this use case.

**Question 293.** XYZ Corp recently faced a ransomware attack
that encrypted critical data files. The company’s IT team was
unable to decrypt the files but had a recent backup available.
However, when they tried to restore the data, they found out the


backup was corrupted. Which of the following best practices
would have helped XYZ Corp in ensuring the integrity of their
backups?
**(A) Regularly testing backup restoration processes
(B)** Storing backups in the same directory as original files
**(C)** Increasing the frequency of backups to every hour
**(D)** Encrypting backups with a strong encryption algorithm

**Explanation 293. Correct Answer: A. Regularly testing
backup restoration processes.** By regularly testing the
restoration process from backups, organizations can identify and
correct any issues with the backup files or processes before they
are needed in a critical recovery situation.

**Option B is incorrect.** Storing backups in the same directory as
the original files increases the risk of both the original and
backup files being compromised or corrupted simultaneously.

**Option C is incorrect.** While increasing the frequency of
backups can reduce data loss, it doesn’t address the integrity or
validity of the backups themselves.

**Option D is incorrect.** While encrypting backups provides
security against unauthorized access, it doesn’t ensure the
integrity or usability of the backup.

**Question 294.** A financial institution is updating its
infrastructure to ensure that customer financial data is kept
secure from both internal and external threats. Which of the
following would be the MOST effective measure to protect
customer financial data from being accessed by unauthorized


internal users?
**(A)** Deploying perimeter firewalls around the institution's
network
**(B)** Implementing two-factor authentication (2FA) for all
customer accounts
**(C) Enforcing strict access controls based on the
principle of least privilege
(D)** Conducting yearly cybersecurity awareness training for
all employees

**Explanation 294. Correct Answer: C. Enforcing strict access
controls based on the principle of least privilege.** The
principle of least privilege (PoLP) ensures that users are given
only the permissions they need to perform their job functions,
nothing more. By enforcing strict access controls using this
principle, the financial institution can minimize the risk of
unauthorized internal users accessing sensitive financial data.

**Option A is incorrect.** Perimeter firewalls are more focused on
preventing external threats and don’t specifically address
unauthorized internal access.

**Option B is incorrect.** Implementing 2FA is primarily for
authenticating users and doesn’t prevent unauthorized internal
users from accessing data if they are already authenticated.

**Option D is incorrect.** While cybersecurity awareness training
is crucial, it’s not the most effective specific measure to prevent
unauthorized internal access to financial data.


**Question 295.** A company’s proprietary algorithm is being
targeted by competitors aiming to replicate its functionality. To
safeguard its intellectual property without changing the
algorithm’s behavior, the company wants a method that
disguises the original code structure. What should they
implement?
**(A)** Data masking on the algorithm’s output
**(B) Obfuscation on the algorithm's code
(C)** Encryption of the algorithm's storage location
**(D)** Implementing a hashing mechanism within the
algorithm

**Explanation 295. Correct Answer: B. Obfuscation on the
algorithm’s code.** Obfuscating the algorithm’s code will make
it more challenging to understand, reverse engineer, or replicate,
without altering its behavior or results.

**Option A is incorrect.** Data masking on the algorithm’s output
would only hide or change certain parts of the output. It doesn’t
protect or disguise the algorithm’s actual code or logic.

**Option C is incorrect.** Encrypting the algorithm’s storage
location protects the stored data from unauthorized access, but
once accessed (e.g., for legitimate use), the original code
structure would still be visible.

**Option D is incorrect.** Implementing a hashing mechanism
within the algorithm might change its behavior and doesn’t
obscure the algorithm’s logic or structure.


**Question 296.** A multinational e-commerce company is
expanding its infrastructure to handle increasing traffic. The
primary goal is to distribute the incoming web traffic across
multiple servers to ensure that no single server is overwhelmed.
Which method should the company use?
**(A)** Deploy a web application firewall
**(B)** Implement server clustering
**(C)** Use hardware-based firewalls
**(D) Set up a load balancer**

**Explanation 296. Correct Answer: D. Set up a load balancer.**
Load balancers distribute incoming traffic across multiple
servers to ensure that no single server is overwhelmed, which
enhances the availability and fault tolerance of applications.

**Option A is incorrect.** While a web application firewall can
protect against web-based threats, it doesn’t distribute incoming
traffic across servers.

**Option B is incorrect.** Server clustering primarily focuses on
providing redundancy and failover capabilities rather than
distributing incoming traffic.

**Option C is incorrect.** Hardware-based firewalls are primarily
used to filter traffic and protect networks from external threats,
not to distribute incoming web traffic.

**Question 297.** A law firm is transitioning to a digital storage
system and wants to ensure that client records and case files are
protected from unauthorized access. Which of the following
would be the BEST strategy to ensure the confidentiality of


legal information stored digitally?
**(A)** Conducting regular penetration testing on the digital
storage system
**(B) Encrypting the client records and case files
(C)** Applying watermarks to digital documents
**(D)** Limiting physical access to the server room

**Explanation 297. Correct Answer: B. Encrypting the client
records and case files.** Encryption is the process of converting
information into an unreadable format unless one has the
appropriate decryption key. By encrypting legal documents,
unauthorized individuals, even if they gain access to the storage
system, won’t be able to comprehend the content of the
documents.

**Option A is incorrect.** While regular penetration testing is
essential to identify vulnerabilities in a system, it does not
directly protect the confidentiality of stored legal information.

**Option C is incorrect.** Watermarking documents can deter
copying or unauthorized distribution, but it doesn’t prevent
unauthorized access or reading of the documents.

**Option D is incorrect.** Limiting physical access to the server
room can prevent unauthorized physical access, but it does not
safeguard against digital breaches or protect the content of the
files.

**Question 298.** A healthcare provider stores vast amounts of
patient data on its servers. While they have strong perimeter
defenses, they want an additional layer of security to ensure


patient data remains confidential even in the event of
unauthorized access. Which of the following would be the
MOST effective solution for this requirement?
**(A)** Use hash algorithms on all patient data
**(B)** Implement data deduplication techniques
**(C) Encrypt the stored patient data
(D)** Use a web application firewall (WAF)

**Explanation 298. Correct Answer: C. Encrypt the stored
patient data.** Encrypting patient data ensures that even if
malicious actors gain access to the data, they won’t be able to
understand or utilize it without the proper decryption key.

**Option A is incorrect.** Hashing is primarily used for verifying
data integrity or storing passwords securely. Hashing would not
allow the healthcare provider to retrieve and use the patient data
when needed.

**Option B is incorrect.** Data deduplication is about reducing
storage requirements by removing duplicate data; it doesn’t
provide confidentiality.

**Option D is incorrect.** While a WAF provides protection
against web-based threats, it doesn’t ensure the confidentiality
of stored data in the event of unauthorized access.

**Question 299.** After a major outage, CloudTech Services is
reviewing their disaster recovery strategy. The company found
out that after restoring from backup, several applications did not
function properly due to configuration discrepancies. What
would be the best approach to ensure a successful recovery in


the future?
**(A)** Prioritize applications for backup based on their
importance
**(B)** Implement differential backups in addition to full
backups
**(C) Regularly conduct a full system recovery in a test
environment
(D)** Use a third-party backup solution instead of an in-house
solution

**Explanation 299. Correct Answer: C. Regularly conduct a
full system recovery in a test environment.** By conducting a
full system recovery in a test environment, CloudTech can
simulate a real-world recovery scenario. This helps in
identifying any discrepancies, configuration issues, or
application dependencies that might be missed during regular
backups, ensuring a successful recovery when it’s critically
needed.

**Option A is incorrect.** Prioritizing applications for backup
based on their importance doesn’t address the configuration
discrepancies or interdependencies between applications.

**Option B is incorrect.** Implementing differential backups,
while beneficial in capturing changes, does not address the
configuration discrepancies found during the recovery.

**Option D is incorrect.** Using a third-party solution instead of
an in-house solution does not inherently guarantee a successful
recovery or address the specific issue of configuration
discrepancies.


**Question 300.** A financial institution is implementing a system
where customers can verify the integrity of their monthly
statements without having access to the original data. Which of
the following techniques would be MOST appropriate for this
task?
**(A)** Encrypting the statements using AES
**(B)** Compressing the statements to reduce file size
**(C) Hashing the statements and providing the hash value
to the customers
(D)** Tokenizing sensitive data within the statements

**Explanation 300. Correct Answer: C. Hashing the
statements and providing the hash value to the customers.**
Hashing can be used to ensure data integrity. Customers can
hash the statement they receive and compare it with the
provided hash value to verify that the data hasn’t been altered.

**Option A is incorrect.** While AES is a strong encryption
method, encrypting the statements won’t allow customers to
verify the integrity of their statements. It ensures confidentiality
more than integrity.

**Option B is incorrect.** Compressing the statements merely
reduces the file size for ease of storage or transmission; it
doesn’t help in verifying data integrity.

**Option D is incorrect.** Tokenization replaces sensitive data
with non-sensitive placeholders. It doesn’t offer a way to verify
the integrity of the entire statement.


**Question 301.** An organization has decided to focus on securing
its database servers where customer details and transaction
records are stored. This data is not being actively accessed or
processed. Which type of security measure would be MOST
appropriate for this type of data?
**(A)** VLData Loss Prevention (DLP) for emailAN
**(B)** Web Application Firewall (WAF)
**(C) Full Disk Encryption (FDE)
(D)** Intrusion Detection System (IDS) for network traffic

**Explanation 301. Correct Answer: C. Full Disk Encryption
(FDE).** Data at rest refers to data that is not actively moving
through the network, such as data stored on hard drives. Full
Disk Encryption (FDE) is a security measure specifically
designed to protect data at rest by encrypting the entire hard
drive.

**Option A is incorrect.** DLP for email primarily focuses on
preventing unauthorized data transfers via email and is more
suited for data in transit rather than data at rest.

**Option B is incorrect.** WAF protects web applications from
various online threats and is not specifically tailored to protect
data at rest.

**Option D is incorrect.** An IDS for network traffic mainly
focuses on monitoring and detecting malicious activity in
network traffic, which pertains to data in transit and not data at
rest.


**Question 302.** A multinational company is considering using a
cloud storage provider based in a foreign country to store
customer data. The company’s home country has strict data
protection laws that require customer data to remain within its
borders. Which of the following considerations is MOST
critical for the company when choosing the cloud storage
provider?
**(A)** The speed of data access from the foreign-based cloud
storage
**(B)** The encryption standards used by the foreign cloud
provider
**(C) Whether the foreign cloud provider offers data
storage exclusively within the company's home country
(D)** The reputation and customer reviews of the foreign
cloud provider

**Explanation 302. Correct Answer: C. Whether the foreign
cloud provider offers data storage exclusively within the
company’s home country.** Data sovereignty refers to the
concept that digital data is subject to the laws of the country in
which it is located. If a company’s home country has regulations
that require customer data to stay within its borders, it’s
essential to ensure that the cloud provider offers storage that
complies with this requirement.

**Option A is incorrect.** While speed of data access is important,
it doesn’t address the data sovereignty concern.

**Option B is incorrect.** While encryption is essential for
security, it doesn’t guarantee compliance with data sovereignty
laws.


**Option D is incorrect.** Reputation is important, but it doesn’t
directly address the specific requirement of data sovereignty.

**Question 303.** After an annual review, BestTech Co. realized
that their IT team was unfamiliar with the protocols to follow
during a data breach. To ensure the team understands the steps
and decision points without launching a live drill, what should
the company implement?
**(A)** Upgrade their firewall systems
**(B) Engage in a tabletop exercise
(C)** Conduct a red team exercise
**(D)** Implement multi-factor authentication for all users

**Explanation 303. Correct Answer: B. Engage in a tabletop
exercise.** Tabletop exercises allow teams to discuss and review
specific scenarios, such as a data breach, and ensure everyone
understands their roles and the processes to follow. It’s a cost-
effective way to familiarize the team with response procedures.

**Option A is incorrect.** Upgrading firewall systems may
enhance the company’s security posture but doesn’t familiarize
the IT team with data breach response protocols.

**Option C is incorrect.** A red team exercise is a live simulated
attack to evaluate the organization’s security posture and does
not focus on walking the team through the response steps in a
controlled environment.

**Option D is incorrect.** Multi-factor authentication is a security
measure to validate users’ identities but does not address the
team’s unfamiliarity with data breach response protocols.


**Question 304.** SecureData Inc., a financial firm, recently
experienced a system crash and needed to restore their database.
While they had a full backup from the previous week, they
realized that several days of transactions were missing. To
prevent such data loss in the future, which backup strategy
involving recording transactions can SecureData implement?
**(A)** Implement differential backups
**(B)** Use snapshot backups every hour
**(C) Enable database journaling
(D)** Configure RAID 5 for their storage

**Explanation 304. Correct Answer: C. Enable database
journaling.** Journaling involves keeping a log or ‘journal’ of
every transaction or change that happens. If there’s a crash, the
system can be restored to the last backup and then use the
journal to replay each transaction, thus preventing data loss.

**Option A is incorrect.** Differential backups capture changes
since the last full backup, but they don’t record every
transaction in real-time like journaling.

**Option B is incorrect.** While hourly snapshot backups would
reduce the potential data loss window, they still don’t capture
every transaction in real-time.

**Option D is incorrect.** RAID 5 offers redundancy and can
protect against a disk failure, but it doesn’t capture and record
every database transaction like journaling.

**Question 305.** SecureNet Inc. recently upgraded their security
infrastructure. To validate how the new system would respond


in real-world scenarios without exposing it to actual risks, they
decide to imitate certain cyber threats in a controlled
environment. Which type of test is SecureNet Inc. planning to
conduct?
**(A)** Penetration Testing
**(B) Simulation Testing
(C)** Vulnerability Assessment
**(D)** Failover Testing.

**Explanation 305. Correct Answer: B. Simulation Testing.**
Simulation testing involves creating a model of the system
under test and then stimulating it with virtual users or devices to
understand its behavior under various conditions. In the context
of cybersecurity, this means imitating cyber threats in a
controlled environment to assess how security infrastructure
responds.

**Option A is incorrect.** Penetration Testing involves ethical
hackers attempting to breach an organization’s defenses, which
involves actual risks, rather than just a simulation.

**Option C is incorrect.** A Vulnerability Assessment identifies,
quantifies, and prioritizes vulnerabilities in a system but doesn’t
necessarily imitate threats in a controlled environment.

**Option D is incorrect.** Failover Testing ensures that a system
can switch to a backup or secondary system in case of a failure,
and does not focus on simulating threats.

**Question 306.** ZenTech, a multinational corporation, recently
adopted a multi-cloud strategy, deploying workloads across


multiple cloud service providers. What is a primary security
benefit of this approach?
**(A)** Centralized management of all cloud resources
**(B)** Automatic encryption of data in transit between clouds
**(C) Mitigation against a single point of failure
(D)** Reduction in the cost of cloud storage solutions

**Explanation 306. Correct Answer: C. Mitigation against a
single point of failure.** Using a multi-cloud strategy distributes
workloads across various cloud providers, which reduces the
risk associated with outages or security breaches in any one
provider. This diversification ensures that a failure or
compromise in one cloud environment doesn’t jeopardize the
entire infrastructure.

**Option A is incorrect.** Multi-cloud strategies can sometimes
complicate management due to the need to interface with
different platforms and providers. Centralization isn’t the
primary benefit in this context.

**Option B is incorrect.** While encryption is crucial, adopting a
multi-cloud approach doesn’t automatically encrypt data in
transit between different cloud environments. This would
require specific configurations and solutions.

**Option D is incorrect.** While cost optimization can be a factor
in adopting a multi-cloud strategy, the primary security
advantage is not related to cost reduction.

**Question 307.** GlobalBank has implemented a backup strategy
where only the changes made since the last full backup are


recorded. Considering the need to optimize storage and reduce
backup times, the IT team performs this type of backup every
day. What backup frequency and type is GlobalBank utilizing?
**(A) Incremental Backup daily
(B)** Differential Backup weekly
**(C)** Full Backup bi-weekly
**(D)** Snapshot Backup daily

**Explanation 307. Correct Answer: A. Incremental Backup
daily.** GlobalBank captures only the changes made since the last
full backup, which is characteristic of incremental backups.
Furthermore, they perform this type of backup every day,
indicating a daily frequency.

**Option B is incorrect.** Differential backups capture the changes
made since the last full backup but would cumulatively grow in
size each time until the next full backup, unlike the incremental
which resets with each backup.

**Option C is incorrect.** A full backup would involve backing up
all the data, regardless of changes, and the scenario does not
mention a bi-weekly frequency.

**Option D is incorrect.** Snapshot backups capture the state of a
system at a particular point in time, but the scenario specifically
mentions capturing changes since the last full backup.

**Question 308.** A defense contractor is working on a new missile
system. While the detailed schematics and operational details
are classified, they have an overview document that only a
select group of partners and stakeholders can view. How should


this overview document be classified?
**(A)** Public
**(B) Restricted
(C)** Confidential
**(D)** Top Secret

**Explanation 308. Correct Answer: B. Restricted.**
“Restricted” classification is used when access to the document
or data needs to be limited to a particular group of individuals.
In this scenario, only a select group of partners and stakeholders
are allowed to view the overview document, making
“Restricted” the most appropriate classification.

**Option A is incorrect.** “Public” classification would allow
anyone to access the document, which is not suitable for this
scenario.

**Option C is incorrect.** While “Confidential” classification
would provide a level of protection, it doesn’t specifically
restrict access to a defined group as “Restricted” does.

**Option D is incorrect.** “Top Secret” is a high-level
classification that would be more appropriate for the detailed
schematics and operational details rather than an overview
document.

**Question 309.** A tech startup is developing a mobile application
that offers exclusive content only to users within a specific
country due to licensing agreements. Which method should the
startup use to ensure that only users within that country can
access the content?


**(A)** Integrate a time-based one-time password (TOTP)
system
**(B) Use geolocation-based access controls
(C)** Enable biometric authentication
**(D)** Implement IP whitelisting

**Explanation 309. Correct Answer: B. Use geolocation-based
access controls.** Geolocation-based access controls can identify
a user’s geographic location and allow or deny access based on
predefined geographic boundaries, making it the most suitable
solution for the startup’s requirement.

**Option A is incorrect.** TOTP systems provide a second factor
of authentication based on time, not geographic location.

**Option C is incorrect.** Biometric authentication verifies the
identity of a user based on physical or behavioral characteristics
but does not restrict access based on geography.

**Option D is incorrect.** IP whitelisting allows access only to
specific IP addresses. While it can offer a level of geographic
restriction, it’s not as precise as geolocation-based controls and
may inadvertently block or allow users.

**Question 310.** MetroTech recently experienced an incident
where an employee mistakenly deleted a portion of their
database. The IT team was able to restore the data using a
backup snapshot taken 24 hours prior. However, some data loss
still occurred. Which of the following recommendations would
minimize data loss in a similar situation in the future?
**(A)** Configure backup snapshots to be taken on a weekly


basis
**(B)** Employ a differential backup solution in addition to
snapshots
**(C)** Increase the storage capacity for backups
**(D) Use an hourly snapshot backup schedule**

**Explanation 310. Correct Answer: D. Use an hourly
snapshot backup schedule.** Increasing the frequency of
snapshot backups, such as taking them every hour, will ensure
that the maximum amount of data that can be lost is limited to
that hour’s worth. This minimizes the potential for data loss
compared to a 24-hour window.

**Option A is incorrect.** Taking weekly snapshots would actually
increase the potential data loss window, not decrease it.

**Option B is incorrect.** While differential backups capture data
changes since the last full backup, they don’t necessarily reduce
the time window for potential data loss like increasing snapshot
frequency would.

**Option C is incorrect.** Increasing storage capacity allows for
more backups or longer retention but does not in itself reduce
the time window for potential data loss.