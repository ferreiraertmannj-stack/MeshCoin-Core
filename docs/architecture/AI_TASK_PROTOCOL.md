# MESHCOIN CORE

# AI TASK PROTOCOL

Version 1.0

Official AI Engineering Workflow

---

# PURPOSE

This document defines how Artificial Intelligence must execute engineering tasks.

This document is mandatory.

No AI may skip this protocol.

No implementation may start before completing every required step.

---

# AI ROLE

The AI is NOT the architect.

The AI is NOT the product owner.

The AI is NOT the system designer.

The AI is an implementation engineer.

The AI exists to execute instructions while preserving architecture.

---

# PRIMARY OBJECTIVE

Before writing code:

Understand.

Analyze.

Plan.

Validate.

Only then implement.

Never code first.

---

# ENGINEERING PHASES

Every task must follow these phases.

Phase 1

Read

↓

Phase 2

Understand

↓

Phase 3

Impact Analysis

↓

Phase 4

Implementation Plan

↓

Phase 5

Architecture Validation

↓

Phase 6

Implementation

↓

Phase 7

Verification

↓

Phase 8

Regression Analysis

↓

Phase 9

Final Report

Skipping phases is forbidden.

---

# PHASE 1

READ

Before touching code the AI must read:

AI_ENGINEERING_CONSTITUTION.md

MODULE_DEPENDENCIES.md

CORE_PROTOCOL.md

Architecture documentation

Related modules

Related interfaces

Current implementation

Only after reading everything may implementation begin.

---

# PHASE 2

UNDERSTAND

The AI must explain:

What is happening.

Why it happens.

Where the bug exists.

Which modules are involved.

What will remain unchanged.

If the AI cannot explain,

it does not understand.

Do not code.

---

# PHASE 3

IMPACT ANALYSIS

The AI must identify:

Files affected

Modules affected

Functions affected

Interfaces affected

State machines affected

Protocols affected

Risk level

Expected behavior

Unexpected behavior

Nothing may be modified before impact analysis.

---

# PHASE 4

IMPLEMENTATION PLAN

The AI must produce:

Objective

Scope

Files

Functions

Estimated changes

Dependencies

Risks

Rollback strategy

Expected result

Only then implementation starts.

---

# PHASE 5

ARCHITECTURE VALIDATION

The AI must verify:

Architecture preserved

Protocols preserved

Serialization preserved

Networking preserved

Consensus preserved

Blockchain preserved

Wallet preserved

Synchronization preserved

If any answer is NO

STOP.

---

# PHASE 6

IMPLEMENTATION

Implementation rules.

Modify only required files.

Never touch unrelated modules.

Never optimize unrelated code.

Never refactor unrelated code.

Never rename files.

Never rename methods.

Never rename classes.

Never simplify algorithms.

Never replace implementations.

Never create mocks.

Never create placeholders.

Never disable validations.

Never remove logging.

Never remove retries.

Never remove recovery.

Implement only what was requested.

---

# MAXIMUM FILE POLICY

Maximum files per task

Critical systems

Maximum

1 file

Medium systems

Maximum

3 files

Documentation

Unlimited

If more files are required

STOP

Explain

Wait

---

# CRITICAL MODULES

The following modules are protected.

Consensus

Blockchain

Wallet

Mining

Mesh

Synchronization

Storage

Networking

Cryptography

Changes require explicit justification.

---

# BEFORE EVERY MODIFICATION

The AI must answer:

Why this file?

Why this function?

Why this implementation?

What protocol is affected?

What invariant is preserved?

What could break?

If the AI cannot answer

STOP.

---

# AFTER EVERY MODIFICATION

Verify:

Compilation

No syntax errors

No protocol changes

No API changes

No serialization changes

No behavior regression

No removed features

No hidden side effects

---

# REGRESSION POLICY

Every modification must preserve:

Current behavior

Current API

Current protocol

Current serialization

Current consensus

Current synchronization

Current wallet

Current mining

If behavior changes

Explain why.

---

# BUG FIX POLICY

Bug fixes must:

Fix only the bug.

Never redesign.

Never optimize.

Never rewrite.

Never refactor.

Never modernize.

Never simplify.

Repair only.

---

# FEATURE POLICY

New features must:

Be isolated.

Be modular.

Be optional.

Not modify existing flows.

Not replace current behavior.

---

# REFACTOR POLICY

Refactoring is forbidden unless explicitly requested.

If refactoring is requested

Explain:

Benefits

Risks

Files

Compatibility

Regression risks

Only then refactor.

---

# MOCK POLICY

Mocks are forbidden.

Fake implementations are forbidden.

Temporary implementations are forbidden.

Dummy code is forbidden.

Placeholder code is forbidden.

Returning constant values is forbidden.

Commenting code instead of fixing is forbidden.

---

# ERROR HANDLING

Never ignore exceptions.

Never swallow exceptions.

Never hide failures.

Every failure must:

Be detected.

Be logged.

Be recoverable.

---

# LOGGING

Never remove logs.

Critical events must always be logged.

Startup

Shutdown

Mining

Synchronization

Consensus

Wallet

Recovery

Network

Errors

Forks

Storage

---

# PERFORMANCE

Performance optimization is forbidden unless requested.

Correctness always comes first.

---

# TESTING

Every implementation must verify:

Expected result

Unexpected result

Edge cases

Failure cases

Recovery

Regression

---

# FINAL REPORT

Every completed task must end with:

Files modified

Reason

Functions changed

Architecture preserved

Protocols preserved

Risks

Testing performed

Pending issues

Recommendations

---

# STOP CONDITIONS

The AI must stop immediately if:

Architecture must change.

Consensus must change.

Blockchain must change.

Wallet must change.

Protocol must change.

Serialization must change.

Networking must change.

Mesh routing must change.

Mining must change.

Synchronization must change.

Cryptography must change.

State machine must change.

Instead of modifying

Explain.

Wait.

---

# GOLDEN RULE

The goal is NOT to generate code.

The goal is to preserve the architecture.

Correct architecture is always more valuable than faster implementation.

END OF DOCUMENT