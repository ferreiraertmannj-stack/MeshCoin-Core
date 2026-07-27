# MESHCOIN CORE

# ENGINEERING DECISIONS

Version 1.0

Architectural Decision Record (ADR)

-------------------------------------------------------------------------------

# PURPOSE

This document records the engineering philosophy behind MeshCoin Core.

It exists to explain WHY architectural decisions were made.

Every AI must understand these decisions before proposing changes.

If code appears unusual, assume there is an engineering reason.

Do not replace intentional design with conventional patterns.

-------------------------------------------------------------------------------

# ENGINEERING PHILOSOPHY

MeshCoin was not created to imitate existing blockchain projects.

MeshCoin was designed from first principles.

Every subsystem exists to solve a specific engineering problem.

Architecture was chosen before implementation.

Implementation serves architecture.

Never reverse this order.

-------------------------------------------------------------------------------

# WHY MESHCOIN EXISTS

The project was created because modern digital infrastructure depends too much on centralized services.

Current internet architecture assumes:

Permanent connectivity.

Cloud availability.

Centralized authentication.

Centralized DNS.

Centralized payment systems.

Centralized messaging.

MeshCoin rejects those assumptions.

-------------------------------------------------------------------------------

# OFFLINE FIRST

Decision:

The network must continue functioning without internet.

Reason:

Internet is not guaranteed.

Many regions have unstable connectivity.

Natural disasters disable infrastructure.

Governments may censor communications.

Remote communities may lack connectivity.

Therefore:

Internet is considered an optimization.

Not a dependency.

-------------------------------------------------------------------------------

# MOBILE FIRST

Decision:

Smartphones are first-class network nodes.

Reason:

Billions of smartphones already exist.

Most eventually become electronic waste.

Every smartphone already contains:

CPU

GPU

RAM

Storage

Battery

WiFi

Bluetooth

Sensors

Instead of manufacturing new hardware,

reuse existing hardware.

-------------------------------------------------------------------------------

# LOW RESOURCE DESIGN

Decision:

Optimize for low-power devices.

Reason:

The network should be globally accessible.

High-end servers should never become mandatory.

The weakest supported device defines the engineering target.

-------------------------------------------------------------------------------

# DISTRIBUTED BEFORE CENTRALIZED

Decision:

Every feature must operate in distributed mode first.

Reason:

Centralized systems are easier.

Distributed systems are resilient.

Resilience is more important than simplicity.

-------------------------------------------------------------------------------

# EVENTUAL CONSISTENCY

Decision:

Temporary inconsistency is acceptable.

Reason:

Distributed systems naturally diverge.

Synchronization restores consistency later.

Do not attempt to eliminate temporary divergence.

-------------------------------------------------------------------------------

# FORKS

Decision:

Forks are normal.

Reason:

Multiple miners may discover blocks simultaneously.

Forks demonstrate decentralization.

Engineering effort should improve resolution,

not eliminate forks.

-------------------------------------------------------------------------------

# BLOCKCHAIN

Decision:

Blockchain is infrastructure.

Reason:

Blockchain provides:

Consensus.

Ordering.

Trust.

History.

Identity.

It is not the product itself.

-------------------------------------------------------------------------------

# WALLET

Decision:

Wallet only manages identity.

Reason:

Wallet must never own consensus.

Wallet must never own balances.

Wallet displays blockchain state.

Wallet does not define blockchain state.

-------------------------------------------------------------------------------

# BALANCE

Decision:

Balances always originate from blockchain.

Reason:

Multiple copies of balances create inconsistency.

Single source of truth prevents divergence.

-------------------------------------------------------------------------------

# CONSENSUS

Decision:

Consensus owns chain selection.

Reason:

No other subsystem should determine network truth.

Never duplicate consensus logic.

-------------------------------------------------------------------------------

# MESH NETWORK

Decision:

Mesh routing is independent from blockchain.

Reason:

Communication and consensus are different concerns.

Routing packets is not validating transactions.

-------------------------------------------------------------------------------

# STORAGE

Decision:

Storage never performs business logic.

Reason:

Databases persist data.

They do not decide correctness.

-------------------------------------------------------------------------------

# UI

Decision:

Flutter is presentation only.

Reason:

Business rules inside UI become impossible to maintain.

UI reflects state.

UI never owns state.

-------------------------------------------------------------------------------

# MODULARITY

Decision:

One responsibility per module.

Reason:

Lower coupling.

Higher maintainability.

Safer evolution.

-------------------------------------------------------------------------------

# DETERMINISM

Decision:

Protocols must behave deterministically.

Reason:

Every node must reach identical conclusions.

Random behavior creates consensus failures.

-------------------------------------------------------------------------------

# SERIALIZATION

Decision:

Serialization is immutable.

Reason:

Changing packet formats breaks compatibility.

Backward compatibility is mandatory.

-------------------------------------------------------------------------------

# NETWORK RECOVERY

Decision:

Recovery is mandatory.

Reason:

Nodes will disconnect.

Recovery is expected.

Failure is normal.

Recovery is architecture.

-------------------------------------------------------------------------------

# SECURITY

Decision:

Trust nothing.

Reason:

Every packet may be malicious.

Every signature must be verified.

Every block must be validated.

Every transaction must be authenticated.

-------------------------------------------------------------------------------

# PERFORMANCE

Decision:

Correctness before optimization.

Reason:

Incorrect fast software is useless.

Correct slow software can later be optimized.

-------------------------------------------------------------------------------

# TESTING

Decision:

Every subsystem must be independently testable.

Reason:

Distributed debugging is difficult.

Isolation reduces complexity.

-------------------------------------------------------------------------------

# BACKWARD COMPATIBILITY

Decision:

Compatibility is mandatory.

Reason:

Nodes may run different software versions.

Abrupt protocol changes fragment the network.

-------------------------------------------------------------------------------

# WHY WE AVOID AUTOMATIC REFACTORING

Decision:

Architecture is manually controlled.

Reason:

Automatic refactoring tools understand syntax.

They rarely understand distributed protocols.

Architecture must remain intentional.

-------------------------------------------------------------------------------

# WHY AI MUST NOT GUESS

Decision:

Artificial Intelligence must stop when uncertain.

Reason:

Guessing destroys distributed systems.

Asking preserves them.

-------------------------------------------------------------------------------

# WHY PLACEHOLDERS MAY EXIST

Decision:

Some incomplete implementations are intentional.

Reason:

The architecture is designed ahead of implementation.

Not every subsystem is implemented immediately.

Placeholders document future direction.

Do not remove them automatically.

-------------------------------------------------------------------------------

# WHY MOCKS ARE FORBIDDEN

Decision:

Mocks are prohibited in production code.

Reason:

Mocks hide real problems.

Distributed systems must be validated with real behavior whenever possible.

-------------------------------------------------------------------------------

# ARCHITECT'S PRINCIPLE

The architecture defines the code.

The code never defines the architecture.

-------------------------------------------------------------------------------

# AI PRINCIPLE

If implementation conflicts with architecture,

implementation must change.

Architecture must not.

-------------------------------------------------------------------------------

# FINAL ENGINEERING PRINCIPLE

The goal of MeshCoin is not merely to create software.

The goal is to create resilient digital infrastructure capable of surviving unreliable networks, heterogeneous hardware, and decentralized operation while preserving determinism, recoverability, modularity, and long-term maintainability.

Every engineering decision should reinforce that objective.

END OF DOCUMENT