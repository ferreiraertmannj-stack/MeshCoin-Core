# MESHCOIN CORE

# PROJECT MEMORY

Version 1.0

Permanent Engineering Knowledge Base

-------------------------------------------------------------------------------

# PURPOSE

This document stores permanent engineering knowledge about MeshCoin Core.

Source code explains HOW.

Architecture explains WHAT.

This document explains WHY previous engineering decisions were made.

Every AI must read this document before proposing architectural changes.

-------------------------------------------------------------------------------

# ENGINEERING MEMORY

Software evolves.

People forget.

AI has no long-term memory.

This file preserves engineering knowledge that must never be lost.

If implementation appears unusual,

assume there is historical context.

Read this document first.

-------------------------------------------------------------------------------

# PROJECT ORIGIN

MeshCoin was never intended to become another cryptocurrency.

The project started from the idea of creating resilient communication and
distributed infrastructure capable of operating with or without internet.

The blockchain was introduced as a trust mechanism.

Not as the final product.

-------------------------------------------------------------------------------

# PROJECT EVOLUTION

Initial goal

↓

Wallet

↓

Blockchain

↓

Mining

↓

Peer Communication

↓

Mesh Networking

↓

Offline Synchronization

↓

Distributed Infrastructure

↓

Nebula Cloud Integration

↓

Distributed Computing

MeshCoin evolved beyond blockchain.

Never reduce it back to "just a cryptocurrency".

-------------------------------------------------------------------------------

# IMPORTANT ARCHITECTURAL DECISION

Decision:

Blockchain became only one subsystem.

Reason:

The long-term vision became decentralized infrastructure.

Never redesign the project around blockchain only.

-------------------------------------------------------------------------------

# IMPORTANT LESSON

Distributed systems behave differently from centralized software.

Temporary inconsistency is normal.

Do not attempt to force immediate consistency everywhere.

-------------------------------------------------------------------------------

# KNOWN BEHAVIOR

Forks happen.

Forks are expected.

Forks are not necessarily bugs.

Improve fork resolution.

Do not eliminate forks by simplifying consensus.

-------------------------------------------------------------------------------

# KNOWN ENGINEERING ISSUE

Synchronization is one of the hardest problems in the project.

Several synchronization problems may appear to be simple bugs.

They are usually protocol problems.

Always investigate protocol before changing implementation.

-------------------------------------------------------------------------------

# LESSON ABOUT WALLET

Wallet must never become the source of truth.

Wallet only reflects blockchain state.

Any implementation that stores authoritative balances inside Wallet is wrong.

-------------------------------------------------------------------------------

# LESSON ABOUT UI

UI often appears to be the easiest place to "fix" problems.

It is usually the wrong place.

UI reflects system state.

It never owns system state.

-------------------------------------------------------------------------------

# LESSON ABOUT NETWORKING

Networking failures are expected.

Reconnect.

Recover.

Retry.

Never assume stable internet.

-------------------------------------------------------------------------------

# LESSON ABOUT MESH

Mesh routing is fundamentally different from blockchain.

Do not merge those responsibilities.

-------------------------------------------------------------------------------

# LESSON ABOUT MINING

Mining only proposes work.

Consensus decides.

Mining must never decide network truth.

-------------------------------------------------------------------------------

# LESSON ABOUT CONSENSUS

Consensus is intentionally isolated.

Do not spread consensus logic across the project.

One owner.

One responsibility.

-------------------------------------------------------------------------------

# LESSON ABOUT STORAGE

Storage should never become intelligent.

Persistence stores data.

Protocols decide correctness.

-------------------------------------------------------------------------------

# LESSON ABOUT RECOVERY

Recovery code often looks unnecessary.

It is not.

Recovery exists because failures are expected.

Never remove recovery logic because it appears unused.

-------------------------------------------------------------------------------

# LESSON ABOUT PLACEHOLDERS

Some placeholders exist intentionally.

They represent planned architecture.

Do not replace them automatically.

Ask first.

-------------------------------------------------------------------------------

# LESSON ABOUT MOCKS

Previous AI assistants attempted to replace unfinished implementations with mocks.

This damaged the architecture.

Project policy:

Mocks are forbidden unless explicitly requested.

-------------------------------------------------------------------------------

# LESSON ABOUT REFACTORING

Large automatic refactoring previously introduced regressions.

Project policy:

Small changes.

Incremental improvements.

Architecture preservation.

-------------------------------------------------------------------------------

# LESSON ABOUT DISTRIBUTED SYSTEMS

Distributed systems cannot be debugged like CRUD applications.

Symptoms often appear far away from the actual cause.

Always investigate upstream.

-------------------------------------------------------------------------------

# LESSON ABOUT PERFORMANCE

Performance problems should never be solved by removing correctness.

Correctness comes first.

Optimization comes later.

-------------------------------------------------------------------------------

# LESSON ABOUT AI

AI tends to:

Rename modules.

Simplify code.

Merge responsibilities.

Replace implementations.

Remove recovery paths.

Create mocks.

This behavior is prohibited.

-------------------------------------------------------------------------------

# AI MEMORY

When uncertain:

Read architecture.

Read protocols.

Read invariants.

Read engineering decisions.

Read this document.

Only then modify code.

-------------------------------------------------------------------------------

# FUTURE ROADMAP MEMORY

The architecture is expected to evolve toward:

Better synchronization

More robust mesh routing

Improved fork resolution

Better node recovery

Nebula Cloud integration

Distributed computing

Resource marketplace

Edge AI

Offline-first infrastructure

These directions are intentional.

Do not remove supporting architecture.

-------------------------------------------------------------------------------

# KNOWN NON-GOALS

MeshCoin is NOT trying to become:

Another Bitcoin clone

Another Ethereum clone

Another DeFi platform

A centralized cloud

A server-only application

A desktop-only application

A wallet-only application

-------------------------------------------------------------------------------

# ENGINEERING PRINCIPLE

Whenever implementation conflicts with accumulated engineering knowledge,

the implementation should be questioned first.

Historical decisions usually exist for a reason.

-------------------------------------------------------------------------------

# BEFORE CHANGING ANY CODE

Ask yourself:

Why was this written?

What historical problem does it solve?

Has this approach already been tried?

Was another solution rejected?

Could this change reintroduce an old bug?

If you cannot answer,

stop.

-------------------------------------------------------------------------------

# AFTER EVERY TASK

Record:

Problem solved

Files modified

Reason

Protocol affected

Architecture affected

Lessons learned

Future recommendations

This document should evolve with the project.

-------------------------------------------------------------------------------

# FINAL MESSAGE

Knowledge is part of the architecture.

History is part of the architecture.

Experience is part of the architecture.

Never lose them.

END OF DOCUMENT